package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"avitotest/src/storage"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type BannerResponse struct {
	ID         int                    `json:"banner_id"`
	Tag_ids    []int                  `json:"tag_ids"`
	Feature_id int                    `json:"feature_id"`
	Content    map[string]interface{} `json:"content"`
	Is_active  bool                   `json:"is_active"`
	Created_at time.Time              `json:"created_at"`
	Updated_at time.Time              `json:"updated_at"`
}

type CreateBannerRequest struct {
	Tag_ids    []int                  `json:"tag_ids"`
	Feature_id int                    `json:"feature_id"`
	Content    map[string]interface{} `json:"content"`
	Is_active  bool                   `json:"is_active"`
}

type CreateBannerResponse struct {
	Banner_id int `json:"banner_id"`
}

type UpdateBannerRequest struct {
	Tag_ids    []int                  `json:"tag_ids"`
	Feature_id *int                   `json:"feature_id"`
	Content    map[string]interface{} `json:"content"`
	Is_active  *bool                  `json:"is_active"`
}

type Handler struct {
	storage storage.Storage
	cacheB  *cache.Cache
}

func NewHandler(storage storage.Storage, cacheB *cache.Cache) *Handler {
	return &Handler{
		storage: storage,
		cacheB:  cacheB,
	}
}

func (h *Handler) GetBannerForUser(c *gin.Context) {

	tag_id, err := strconv.Atoi(c.Query("tag_id"))
	if err != nil {
		fmt.Printf("error: Invalid tag ID %s\n", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid tag ID",
		})
		return
	}

	feature_id, err := strconv.Atoi(c.Query("feature_id"))
	if err != nil {
		fmt.Printf("error: Invalid feature ID %s\n", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid feature ID",
		})
		return
	}

	useLastRevision := false
	useLastRevisionStr := c.Query("use_last_revision")
	if useLastRevisionStr == "true" {
		useLastRevision = true
	}

	role, exists := c.Get("role")
	if !exists {
		fmt.Println("Role not found in context")
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	var banner storage.BannerDb
	var ok bool

	reqCache := fmt.Sprintf("%d %d", feature_id, tag_id)

	cacheValue, found := h.cacheB.Get(reqCache)
	if found && !useLastRevision {
		banner, ok = cacheValue.(storage.BannerDb)
		if !ok {
			found = false
		}
	}
	if !found || useLastRevision {
		banner, err = h.storage.GetBanner(context.Background(), tag_id, feature_id)
		h.cacheB.Set(reqCache, banner, cache.DefaultExpiration)
	}

	if err != nil {
		if err.Error() == "banner not found" {
			fmt.Println(err.Error())
			c.Status(http.StatusNotFound)
			return
		}
		fmt.Printf("failed to get banner %s\n", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	if !banner.Is_active && role == "user" {
		fmt.Println("banner is disable")
		c.Status(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, banner.Content)
}

func (h *Handler) GetBanners(c *gin.Context) {

	params := c.Request.URL.Query()

	featureIdStr := params.Get("feature_id")
	if featureIdStr == "" {
		featureIdStr = "-1"
	}

	featureId, err := strconv.Atoi(featureIdStr)
	if err != nil {
		fmt.Printf("error: Invalid feature ID query %s\n", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid feature ID query",
		})
		return
	}

	tagIdStr := params.Get("tag_id")
	if tagIdStr == "" {
		tagIdStr = "-1"
	}

	tagId, err := strconv.Atoi(tagIdStr)
	if err != nil {
		fmt.Printf("error: Invalid tag ID query %s\n", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid tag ID query",
		})
		return
	}

	limitStr := params.Get("limit")
	if limitStr == "" {
		limitStr = "-1"
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		fmt.Printf("error: Invalid limit query %s\n", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid limit query",
		})
		return
	}

	offsetStr := params.Get("offset")
	if offsetStr == "" {
		offsetStr = "-1"
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		fmt.Printf("error: Invalid offset query %s\n", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid offset query",
		})
		return
	}

	banners, err := h.storage.GetBanners(context.Background(), featureId, tagId, limit, offset)
	if err != nil {
		fmt.Printf("failed to get banners %s\n", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BannersToResponse(banners))
}

func (h *Handler) CreateBanner(c *gin.Context) {

	var reqBanner CreateBannerRequest

	err := json.NewDecoder(c.Request.Body).Decode(&reqBanner)
	if err != nil {
		fmt.Printf("failed to decode body %s\n", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid body",
		})
		return
	}

	var resBanner CreateBannerResponse

	resBanner.Banner_id, err = h.storage.CreateBanner(context.Background(), reqBanner.Tag_ids, reqBanner.Feature_id, reqBanner.Content, reqBanner.Is_active)
	if err != nil {
		fmt.Printf("failed to create banner %s\n", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, resBanner)
}

func (h *Handler) UpdateBanner(c *gin.Context) {

	bannerIdStr := c.Param("id")

	bannerId, err := strconv.Atoi(bannerIdStr)
	if err != nil {
		fmt.Printf("error: Invalid ID param %s\n", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid feature ID query",
		})
		return
	}

	var reqBanner UpdateBannerRequest

	err = json.NewDecoder(c.Request.Body).Decode(&reqBanner)
	if err != nil {
		fmt.Printf("failed to decode body %s\n", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid body",
		})
		return
	}

	feature_id := -1
	if reqBanner.Feature_id != nil {
		feature_id = *reqBanner.Feature_id
	}

	isActiveExist := false
	isActive := false

	if reqBanner.Is_active != nil {
		isActiveExist = true
		isActive = *reqBanner.Is_active

	}

	err = h.storage.UpdateBanner(context.Background(), bannerId, reqBanner.Tag_ids, feature_id, reqBanner.Content, isActive, isActiveExist)
	if err != nil {
		if err.Error() == "banner not found" {
			fmt.Println(err.Error())
			c.Status(http.StatusNotFound)
			return
		}
		fmt.Printf("failed to update banner %s\n", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) DeleteBanner(c *gin.Context) {

	bannerIdStr := c.Param("id")

	bannerId, err := strconv.Atoi(bannerIdStr)
	if err != nil {
		fmt.Printf("error: Invalid ID param %s\n", err.Error())
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid feature ID query",
		})
		return
	}

	err = h.storage.DeleteBanner(context.Background(), bannerId)
	if err != nil {
		if err.Error() == "banner not found" {
			fmt.Println(err.Error())
			c.Status(http.StatusNotFound)
			return
		}
		fmt.Printf("failed to delete banner %s\n", err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) GetHealth(c *gin.Context) {
	c.Status(http.StatusOK)
}

func BannerToResponse(banner storage.Banner) BannerResponse {
	return BannerResponse{
		ID:         banner.ID,
		Tag_ids:    banner.Tag_ids,
		Feature_id: banner.Feature_id,
		Content:    banner.Content,
		Is_active:  banner.Is_active,
		Created_at: banner.Created_at,
		Updated_at: banner.Updated_at,
	}
}

func BannersToResponse(banners []storage.Banner) []BannerResponse {
	if banners == nil {
		return nil
	}

	res := make([]BannerResponse, len(banners))

	for index, value := range banners {
		res[index] = BannerToResponse(value)
	}

	return res
}
