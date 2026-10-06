-- +migrate Up
CREATE TABLE wages(
    `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    `user_id` BIGINT UNSIGNED NOT NULL,
    `amount` DECIMAL(20,2) NOT NULL,
    `type` ENUM('normal', 'express') NOT NULL DEFAULT 'normal',
    `created_at` BIGINT NOT NULL,
    CONSTRAINT `fk_wages_user_id`
        FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),
    UNIQUE INDEX `uk_wage_type_user_id` (`type`,`user_id`),
    INDEX `idx_wages_user_id` (`user_id`)
);

-- +migrate Down
DROP TABLE wages;
