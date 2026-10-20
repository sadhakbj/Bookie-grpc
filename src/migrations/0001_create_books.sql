-- +goose Up
CREATE TABLE books (
    id          BINARY(16)   NOT NULL,
    title       VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL,
    author      VARCHAR(255) NOT NULL,
    price       BIGINT       NOT NULL,
    created_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE = InnoDB;
