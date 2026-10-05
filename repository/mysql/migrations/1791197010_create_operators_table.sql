-- +migrate Up
CREATE TABLE operators(
    `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(70) NOT NULL,
    `driver` VARCHAR(70) NOT NULL,
    `is_active` TINYINT(1) NOT NULL DEFAULT 1
);

-- +migrate Down
DROP TABLE operators;
