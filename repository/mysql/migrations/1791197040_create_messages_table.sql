-- +migrate Up
CREATE TABLE messages(
    `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `idempotency_key` VARCHAR(191) NOT NULL,
    `operator_id` BIGINT UNSIGNED NOT NULL,
    `origin_number` VARCHAR(20) NOT NULL,
    `destination_number` VARCHAR(20) NOT NULL,
    `content` TEXT NOT NULL,
    `status` ENUM('initiated', 'queued', 'sent', 'failed') NOT NULL DEFAULT 'initiated',
    `failed_reason` TEXT NULL,
    `created_at` BIGINT NOT NULL,
    `updated_at` BIGINT NOT NULL,
    CONSTRAINT `fk_messages_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
    CONSTRAINT `fk_messages_operator_id`
        FOREIGN KEY (`operator_id`) REFERENCES `operators`(`id`),
    UNIQUE INDEX `uk_messages_idempotency_key` (`idempotency_key`),
    INDEX `idx_messages_user_id` (`user_id`),
    INDEX `idx_messages_operator_id` (`operator_id`)
);

-- +migrate Down
DROP TABLE messages;
