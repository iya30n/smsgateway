-- +migrate Up
CREATE TABLE operators(
    `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `name` VARCHAR(70) NOT NULL,
    `total_tps` INT UNSIGNED NOT NULL,
    `express_reserved_tps` INT UNSIGNED NOT NULL DEFAULT 0,
    `is_active` TINYINT(1) NOT NULL DEFAULT 1,
    CONSTRAINT `chk_operators_express_reserved_fits_total`
        CHECK (`express_reserved_tps` <= `total_tps`),
    UNIQUE INDEX `uk_operators_name` (`name`)
);

-- +migrate Down
DROP TABLE operators;
