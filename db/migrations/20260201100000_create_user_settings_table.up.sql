CREATE TABLE user_settings (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id INT UNSIGNED NOT NULL UNIQUE,
    pay_cycle_type ENUM('calendar', 'last_weekday', 'custom_day', 'bi_weekly') NOT NULL DEFAULT 'calendar',
    pay_day INT DEFAULT NULL COMMENT 'For custom_day: 1-31, For bi_weekly: day of week (0-6)',
    cycle_start_offset INT NOT NULL DEFAULT 1 COMMENT 'Days after payday to start counting expense period',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NULL ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
