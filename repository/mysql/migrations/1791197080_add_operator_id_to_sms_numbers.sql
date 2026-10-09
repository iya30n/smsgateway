-- +migrate Up
ALTER TABLE sms_numbers
    ADD COLUMN `operator_id` BIGINT UNSIGNED NOT NULL AFTER `number`,
    ADD CONSTRAINT `fk_sms_numbers_operator_id`
        FOREIGN KEY (`operator_id`) REFERENCES `operators`(`id`),
    ADD INDEX `idx_sms_numbers_operator_id` (`operator_id`);

-- +migrate Down
ALTER TABLE sms_numbers
    DROP FOREIGN KEY `fk_sms_numbers_operator_id`,
    DROP INDEX `idx_sms_numbers_operator_id`,
    DROP COLUMN `operator_id`;
