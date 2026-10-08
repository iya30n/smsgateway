-- +migrate Up
CREATE TABLE sms_numbers(
    `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `number` VARCHAR(20) NOT NULL,
    `is_active` TINYINT(1) NOT NULL DEFAULT 1,
    UNIQUE INDEX `uk_sms_numbers_number` (`number`)
);

-- +migrate Down
DROP TABLE sms_numbers;
