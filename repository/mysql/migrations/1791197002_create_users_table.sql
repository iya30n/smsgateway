-- +migrate Up
CREATE TABLE users(
    `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(70) NOT NULL,
    `username` VARCHAR(70) NOT NULL,
    `balance` DECIMAL(20,2)
);

-- +migrate Down
DROP TABLE users;