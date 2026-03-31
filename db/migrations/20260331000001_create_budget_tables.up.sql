-- Create budgets table (if not exists)
CREATE TABLE IF NOT EXISTS budgets (
    id          INT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id     INT UNSIGNED NOT NULL,
    category_id INT UNSIGNED NOT NULL,
    amount      INT          NOT NULL,
    period      VARCHAR(20)  NOT NULL,
    start_date  DATETIME     NOT NULL,
    end_date    DATETIME     NOT NULL,
    is_active   TINYINT(1)   NOT NULL DEFAULT 1,
    alert_at    INT          NOT NULL DEFAULT 80,
    description VARCHAR(500),
    created_at  DATETIME,
    updated_at  DATETIME,
    PRIMARY KEY (id),
    INDEX idx_budgets_user_id (user_id),
    INDEX idx_budgets_category_id (category_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create budget_alerts table (if not exists) with alert_type column
CREATE TABLE IF NOT EXISTS budget_alerts (
    id           INT UNSIGNED NOT NULL AUTO_INCREMENT,
    budget_id    INT UNSIGNED NOT NULL,
    user_id      INT UNSIGNED NOT NULL,
    alert_type   VARCHAR(50)  NOT NULL DEFAULT 'threshold',
    percentage   INT          NOT NULL,
    spent_amount INT          NOT NULL,
    message      VARCHAR(500),
    is_read      TINYINT(1)   NOT NULL DEFAULT 0,
    created_at   DATETIME,
    PRIMARY KEY (id),
    INDEX idx_budget_alerts_budget_id (budget_id),
    INDEX idx_budget_alerts_user_id (user_id),
    INDEX idx_budget_alerts_alert_type (alert_type),
    FOREIGN KEY (budget_id) REFERENCES budgets(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Add alert_type column to existing budget_alerts table when it was created without it.
-- Uses an INFORMATION_SCHEMA check for MySQL 5.7 compatibility.
SET @dbname = DATABASE();
SET @tbl = 'budget_alerts';
SET @col = 'alert_type';
SET @sql = (
    SELECT IF(
        (SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS
         WHERE TABLE_SCHEMA = @dbname AND TABLE_NAME = @tbl AND COLUMN_NAME = @col) > 0,
        'SELECT 1 -- column already exists',
        CONCAT('ALTER TABLE `', @tbl, '` ADD COLUMN `', @col, '` VARCHAR(50) NOT NULL DEFAULT ''threshold'' AFTER user_id')
    )
);
PREPARE _stmt FROM @sql;
EXECUTE _stmt;
DEALLOCATE PREPARE _stmt;
