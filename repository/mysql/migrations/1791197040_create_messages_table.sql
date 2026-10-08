-- +migrate Up
CREATE TABLE messages(
    `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `idempotency_key` VARCHAR(191) NOT NULL,
    `source_number` VARCHAR(20) NOT NULL,
    `receptor_number` VARCHAR(20) NOT NULL,
    `content` TEXT NOT NULL,
    `type` ENUM('normal', 'express') NOT NULL DEFAULT 'normal',
    `status` ENUM('initiated', 'queued', 'sent', 'failed') NOT NULL DEFAULT 'initiated',
    `failed_reason` TEXT NULL,
    `created_at` BIGINT NOT NULL,
    `updated_at` BIGINT NOT NULL,
    CONSTRAINT `fk_messages_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
    UNIQUE INDEX `uk_messages_idempotency_key` (`idempotency_key`),
    INDEX `idx_messages_user_id` (`user_id`)
);

-- +migrate Down
DROP TABLE messages;
