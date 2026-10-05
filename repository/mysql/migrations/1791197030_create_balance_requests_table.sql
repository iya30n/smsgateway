-- +migrate Up
CREATE TABLE balance_requests(
    `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `amount` DECIMAL(20,2) NOT NULL,
    `status` ENUM('initiated', 'pending', 'approved', 'rejected') NOT NULL DEFAULT 'initiated',
    `created_at` BIGINT NOT NULL,
    CONSTRAINT `fk_balance_requests_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
    INDEX `idx_balance_requests_user_id` (`user_id`)
);

-- +migrate Down
DROP TABLE balance_requests;
