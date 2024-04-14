package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Banner struct {
	ID         int                    `json:"id"`
	Tag_ids    []int                  `json:"tag_ids"`
	Feature_id int                    `json:"feature_id"`
	Content    map[string]interface{} `json:"content"`
	Is_active  bool                   `json:"is_active"`
	Created_at time.Time              `json:"created_at"`
	Updated_at time.Time              `json:"updated_at"`
}

type BannerDb struct {
	ID         int                    `json:"id"`
	Feature_id int                    `json:"feature_id"`
	Content    map[string]interface{} `json:"content"`
	Is_active  bool                   `json:"is_active"`
	Created_at time.Time              `json:"created_at"`
	Updated_at time.Time              `json:"updated_at"`
}

type Storage interface {
	GetBanner(ctx context.Context, tag_id int, feature_id int) (BannerDb, error)
	GetBanners(ctx context.Context, feature_id int, tag_id int, limit int, offset int) ([]Banner, error)
	UpdateBanner(ctx context.Context, id int, tagIds []int, featureId int, content map[string]interface{}, isActive bool, isActiveExist bool) error
	CreateBanner(ctx context.Context, tagIds []int, featureId int, content map[string]interface{}, isActive bool) (int, error)
	DeleteBanner(ctx context.Context, id int) error
}

type postgres struct {
	db *pgxpool.Pool
}

func NewPgStorage(ctx context.Context, connString string) (*postgres, error) {
	var pgInstance *postgres
	var pgOnce sync.Once
	pgOnce.Do(func() {
		db, err := pgxpool.New(ctx, connString)
		if err != nil {
			fmt.Printf("Unable to create connection pool: %v\n", err)
			return
		}

		pgInstance = &postgres{db}
	})

	return pgInstance, nil
}

func (pg *postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

func (pg *postgres) Close() {
	pg.db.Close()
}

func (pg *postgres) GetBanner(ctx context.Context, tag_id int, feature_id int) (BannerDb, error) {
	query := fmt.Sprintf(`SELECT b.id, b.feature_id, b.content, b.is_active, b.created_at, b.updated_at
	FROM banner b 
	JOIN banner_tag bt ON b.id = bt.banner_id
	WHERE b.feature_id = %d AND bt.tag_id = %d;`, feature_id, tag_id)

	rows, err := pg.db.Query(ctx, query)

	var banner BannerDb

	if err != nil {
		return banner, fmt.Errorf("unable to query: %w", err)
	}
	defer rows.Close()

	banner, err = pgx.CollectOneRow(rows, pgx.RowToStructByName[BannerDb])

	if errors.Is(err, pgx.ErrNoRows) {
		return banner, errors.New("banner not found")
	}

	if err != nil {
		fmt.Printf("CollectRows error: %v", err)
		return banner, err
	}

	return banner, nil
}

func (pg *postgres) GetBanners(ctx context.Context, feature_id int, tag_id int, limit int, offset int) ([]Banner, error) {

	query := `SELECT b.id, array_agg(bt.tag_id) as tag_ids, b.feature_id, b.content, b.is_active, b.created_at, b.updated_at
	FROM banner b
	JOIN banner_tag bt ON b.id = bt.banner_id `

	if feature_id >= 0 {
		query += fmt.Sprintf(`WHERE b.feature_id = %d `, feature_id)
	}
	query += `GROUP BY b.id, b.feature_id `

	if tag_id >= 0 {
		query += fmt.Sprintf(`HAVING bool_or(bt.tag_id = %d) `, tag_id)
	}
	if limit > 0 {
		query += fmt.Sprintf(`LIMIT %d `, limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(`OFFSET %d `, offset)
	}
	query += `;`

	var banners []Banner

	rows, err := pg.db.Query(ctx, query)
	if err != nil {
		return banners, fmt.Errorf("unable to query: %w", err)
	}
	defer rows.Close()

	banners, err = pgx.CollectRows(rows, pgx.RowToStructByName[Banner])
	if err != nil {
		fmt.Printf("CollectRows error: %v", err)
		return banners, err
	}

	return banners, nil
}

func (pg *postgres) CreateBanner(ctx context.Context, tagIds []int, featureId int, content map[string]interface{}, isActive bool) (int, error) {

	createdTime := time.Now().UTC().Format("2006-01-02T15:04:05")

	var id int
	err := pg.db.QueryRow(context.Background(), `INSERT INTO banner (feature_id, content, is_active, created_at, updated_at) 
	VALUES ($1, $2, $3, $4, $5) RETURNING id`, featureId, content, isActive, createdTime, createdTime).Scan(&id)
	if err != nil {
		return id, fmt.Errorf("unable to insert banner: %w", err)
	}

	for _, v := range tagIds {
		query := `INSERT INTO banner_tag (banner_id, tag_id)
		VALUES (@banner_id, @tag_id)`
		args := pgx.NamedArgs{
			"banner_id": id,
			"tag_id":    v,
		}
		_, err := pg.db.Exec(ctx, query, args)
		if err != nil {
			return id, fmt.Errorf("unable to insert banner_tag: %w", err)
		}
	}

	return id, nil
}

func (pg *postgres) UpdateBanner(ctx context.Context, id int, tagIds []int, featureId int, content map[string]interface{}, isActive bool, isActiveExist bool) error {

	updateData := ""

	if featureId >= 0 {
		updateData += fmt.Sprintf(`feature_id = %d, `, featureId)
	}
	if content != nil {
		contentJson, err := json.Marshal(content)
		if err != nil {
			fmt.Println(err.Error())
			return fmt.Errorf("cannot marshal content")
		}
		contentJsonStr := string(contentJson)
		updateData += fmt.Sprintf(`content = '%s', `, contentJsonStr)
	}
	if isActiveExist {
		updateData += fmt.Sprintf(`is_active = %t, `, isActive)
	}

	updatedTime := time.Now().UTC().Format("2006-01-02T15:04:05")
	updateData += fmt.Sprintf(`updated_at = '%s'`, updatedTime)

	query := fmt.Sprintf(`UPDATE banner SET %s WHERE id = %d`, updateData, id)

	res, err := pg.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("unable to update row: %w", err)
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("banner not found")
	}

	if tagIds != nil {
		query := fmt.Sprintf(`DELETE FROM banner_tag WHERE banner_id = %d`, id)
		_, err := pg.db.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("unable to update banner_tag: %w", err)
		}

		for _, tagId := range tagIds {
			query := `INSERT INTO banner_tag (banner_id, tag_id)
			VALUES (@banner_id, @tag_id)`
			args := pgx.NamedArgs{
				"banner_id": id,
				"tag_id":    tagId,
			}
			_, err := pg.db.Exec(ctx, query, args)
			if err != nil {
				return fmt.Errorf("unable to update banner_tag: %w", err)
			}
		}
	}

	return nil
}

func (pg *postgres) DeleteBanner(ctx context.Context, id int) error {

	query := fmt.Sprintf(`DELETE FROM banner_tag WHERE banner_id = %d`, id)
	_, err := pg.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("unable to delete banner_tag: %w", err)
	}

	query = fmt.Sprintf(`DELETE FROM banner WHERE id = %d`, id)
	res, err := pg.db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("unable to delete from banner: %w", err)
	}

	rowsAffected := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("banner not found")
	}

	return nil
}
