CREATE TABLE indexer_state (
    id SMALLINT PRIMARY KEY DEFAULT 1,
    last_indexed_block BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT single_row CHECK (id = 1)
);
INSERT INTO indexer_state (id, last_indexed_block)
VALUES (1, 0);