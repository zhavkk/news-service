-- +goose Up
-- +goose StatementBegin
CREATE TABLE news (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    category TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    start_time TIMESTAMPTZ NOT NULL,
    end_time   TIMESTAMPTZ NOT NULL
);

CREATE TABLE content_blocks (
    id BIGSERIAL PRIMARY KEY,
    news_id BIGINT NOT NULL REFERENCES news(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('text','image','link')),
    content TEXT NOT NULL,
    position INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (news_id, position)
);

CREATE INDEX idx_content_news_id ON content_blocks(news_id);
CREATE INDEX idx_news_visibility ON news(start_time, end_time);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS content_blocks;
DROP TABLE IF EXISTS news;
-- +goose StatementEnd
