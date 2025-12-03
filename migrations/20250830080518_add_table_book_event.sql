-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS book_event(
    id BIGSERIAL NOT NULL,
    book_id BIGINT NOT NULL REFERENCES book(id) ON DELETE CASCADE,
    type SMALLINT NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL DEFAULT 0,
    payload JSONB,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (status, id)
    ) PARTITION BY LIST (status);
CREATE INDEX idx_book_event_book_id ON book_event(book_id);

CREATE TABLE book_event_new PARTITION OF book_event FOR VALUES IN (1);
CREATE TABLE book_event_locked PARTITION OF book_event FOR VALUES IN (2);
CREATE TABLE book_event_unlocked PARTITION OF book_event FOR VALUES IN (3);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE book_event;
-- +goose StatementEnd

