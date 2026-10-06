-- +migrate Up
CREATE TABLE sms_numbers(
    `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `number` VARCHAR(20) NOT NULL,
    `operator_id` BIGINT UNSIGNED NOT NULL,
    `is_active` TINYINT(1) NOT NULL DEFAULT 1,
    CONSTRAINT `fk_sms_numbers_operator_id`
        FOREIGN KEY (`operator_id`) REFERENCES `operators`(`id`),
    INDEX `idx_sms_numbers_operator_id` (`operator_id`)
);

-- +migrate Down
DROP TABLE sms_numbers;
