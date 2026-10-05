-- +migrate Up
CREATE TABLE transactions(
    `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `balance_request_id` BIGINT UNSIGNED NULL,
    `message_id` BIGINT UNSIGNED NULL,
    `credit` DECIMAL(20,2) NOT NULL DEFAULT 0,
    `debit` DECIMAL(20,2) NOT NULL DEFAULT 0,
    `type` ENUM('sms_charge', 'refund', 'reversal', 'manual_adjustment') NOT NULL,
    CONSTRAINT `fk_transactions_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
    CONSTRAINT `fk_transactions_balance_request_id`
        FOREIGN KEY (`balance_request_id`) REFERENCES `balance_requests`(`id`),
    CONSTRAINT `fk_transactions_message_id`
        FOREIGN KEY (`message_id`) REFERENCES `messages`(`id`),
    INDEX `idx_transactions_user_id` (`user_id`),
    INDEX `idx_transactions_balance_request_id` (`balance_request_id`),
    INDEX `idx_transactions_message_id` (`message_id`)
);

-- +migrate Down
DROP TABLE transactions;
