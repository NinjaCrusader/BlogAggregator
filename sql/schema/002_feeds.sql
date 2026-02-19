-- +goose Up
CREATE TABLE feeds (
    "id" UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    "created_at" TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP NOT NULL,
    "name" TEXT UNIQUE NOT NULL,
    "url" TEXT UNIQUE NOT NULL,
    "user_id" TEXT REFERENCES users(id) ON DELETE CASCADE 
);

-- +goose Down
DROP TABLE feeds;