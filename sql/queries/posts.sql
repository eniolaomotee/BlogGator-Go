-- name: CreatePost :one
INSERT INTO posts (id,created_at, updated_at,title, url, description, published_at, feed_id)
VALUES($1,$2,$3,$4,$5,$6,$7,$8)
RETURNING *;


-- name: GetPostsForUser :many
SELECT posts.*, feeds.name AS feed_name 
FROM posts
INNER JOIN feeds ON posts.feed_id = feeds.id
INNER JOIN feed_follows ON feeds.id = feed_follows.feed_id
WHERE feed_follows.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2;



-- name: GetPostsForUserSorted :many
SELECT 
    p.id, p.created_at, p.updated_at, p.title, p.url,
    p.description, p.published_at, p.feed_id,
    f.name AS feed_name
FROM posts p
JOIN feed_follows ff ON p.feed_id = ff.feed_id
JOIN feeds f ON p.feed_id = f.id
WHERE ff.user_id = $1
  AND ($3 = '' OR f.name ILIKE '%' || $3 || '%')
ORDER BY
    -- Title-based sorting (text)
    CASE 
        WHEN $4 = 'title_asc'  THEN p.title
        WHEN $4 = 'title_desc' THEN p.title
        ELSE NULL
    END
    --
    , CASE 
        WHEN $4 = 'title_asc' THEN 1 
        WHEN $4 = 'title_desc' THEN -1 
        ELSE 0 
    END
    --
    -- Timestamp-based sorting (created/published)
    , CASE 
        WHEN $4 = 'created_at_asc'  THEN p.created_at
        WHEN $4 = 'created_at_desc' THEN p.created_at
        WHEN $4 = 'published_at_asc' THEN p.published_at
        WHEN $4 = 'published_at_desc' THEN p.published_at
        ELSE NULL
    END
    --
    , CASE 
        WHEN $4 LIKE '%_asc' THEN 1 
        WHEN $4 LIKE '%_desc' THEN -1 
        ELSE 0 
    END
LIMIT $2;