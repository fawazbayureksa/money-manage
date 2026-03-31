CREATE TABLE debts (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id INT UNSIGNED NOT NULL,
    name VARCHAR(200) NOT NULL,
    debt_type VARCHAR(50) NOT NULL,
    original_amount INT NOT NULL,
    current_balance INT NOT NULL,
    interest_rate DOUBLE NOT NULL DEFAULT 0,
    minimum_payment INT NOT NULL DEFAULT 0,
    due_day INT NOT NULL DEFAULT 1,
    start_date DATETIME NOT NULL,
    lender_name VARCHAR(200) DEFAULT '',
    notes VARCHAR(500) DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NULL ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_user_id (user_id),
    INDEX idx_debt_type (debt_type),
    INDEX idx_is_active (is_active),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE debt_payments (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    debt_id INT UNSIGNED NOT NULL,
    user_id INT UNSIGNED NOT NULL,
    amount INT NOT NULL,
    payment_date DATETIME NOT NULL,
    notes VARCHAR(500) DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_debt_id (debt_id),
    INDEX idx_user_id (user_id),
    INDEX idx_payment_date (payment_date),
    FOREIGN KEY (debt_id) REFERENCES debts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE debt_milestones (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    debt_id INT UNSIGNED NOT NULL,
    user_id INT UNSIGNED NOT NULL,
    milestone_type VARCHAR(50) NOT NULL,
    target_value INT NOT NULL,
    message VARCHAR(500) DEFAULT '',
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_debt_id (debt_id),
    INDEX idx_user_id (user_id),
    FOREIGN KEY (debt_id) REFERENCES debts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
