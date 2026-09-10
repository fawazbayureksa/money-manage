-- Migration: create tags and transaction_tags tables

-- Create tags table
CREATE TABLE IF NOT EXISTS tags (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(50) NOT NULL,
    color VARCHAR(7) DEFAULT '#6366F1' COMMENT 'Hex color code',
    icon VARCHAR(10) NULL COMMENT 'Emoji',
    usage_count INT DEFAULT 0 COMMENT 'For sorting by popularity',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    UNIQUE KEY unique_user_tag (user_id, name),
    INDEX idx_user_usage (user_id, usage_count DESC),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Create transaction_tags junction table
CREATE TABLE IF NOT EXISTS transaction_tags (
    transaction_id BIGINT UNSIGNED NOT NULL,
    tag_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    PRIMARY KEY (transaction_id, tag_id),
    INDEX idx_transaction_id (transaction_id),
    INDEX idx_tag_id (tag_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
