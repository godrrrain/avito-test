-- file: 20-create-database.sql
CREATE DATABASE banners;
GRANT ALL PRIVILEGES ON DATABASE banners TO program;

\c banners;

CREATE TABLE banner (
    id         SERIAL PRIMARY KEY,
    feature_id INT NOT NULL,
    content    JSON NOT NULL,
    is_active  BOOLEAN NOT NULL,
    created_at TIMESTAMP   NOT NULL,
    updated_at TIMESTAMP   NOT NULL
);

CREATE TABLE banner_tag (
    banner_id INT REFERENCES banner(id),
    tag_id    INT NOT NULL
);

GRANT ALL ON ALL TABLES IN SCHEMA public TO program;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO program;


-- INSERT INTO banner (feature_id, content, is_active, created_at, updated_at) VALUES (2, '{ "title": "Bob", "text": "bob", "url": "bob.com"}', TRUE, '2024-04-12T10:03:05', '2024-04-16T12:03:05');
-- INSERT INTO banner_tag VALUES (1, 1);

-- INSERT INTO banner (feature_id, content, is_active, created_at, updated_at) VALUES (3, '{ "title": "Alice", "text": "alice", "url": "alice.com"}', true, '2024-04-10T10:03:05', '2024-04-14T12:03:05');
-- INSERT INTO banner_tag VALUES (2, 1);

-- INSERT INTO banner (feature_id, content, is_active, created_at, updated_at) VALUES (3, '{ "title": "AliceGood", "text": "alicegood", "url": "alicegood.com"}', false, '2024-04-10T10:03:05', '2024-04-14T12:03:05');
-- INSERT INTO banner_tag VALUES (3, 1);

-- SELECT b.id, b.content, b.feature_id, array_agg(bt.tag_id) as tag_ids
-- FROM banner b
-- JOIN banner_tag bt ON b.id = bt.banner_id
-- WHERE b.feature_id = 8
-- GROUP BY b.id, b.feature_id
-- HAVING bool_or(bt.tag_id = 10)
-- LIMIT 1
-- OFFSET 1;

-- SELECT b.id, b.feature_id, b.content, b.is_active, b.created_at, b.updated_at
-- FROM banner b
-- JOIN banner_tag bt ON b.id = bt.banner_id
-- WHERE b.feature_id = 8 AND bt.tag_id = 10;

-- SELECT b.content
-- FROM banner b
-- JOIN banner_tag bt ON b.id = bt.banner_id
-- WHERE b.feature_id = 8 AND bt.tag_id = 10;

-- UPDATE banner SET feature_id = 3, content = '{ "title": "Alice", "text": "alice", "url": "alice.com"}', is_active = true, updated_at = '2024-04-12T23:00:00' WHERE id = 2;

