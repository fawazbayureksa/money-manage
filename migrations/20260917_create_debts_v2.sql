CREATE TABLE IF NOT EXISTS debts (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT UNSIGNED NOT NULL,

    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    debt_type ENUM(
        'credit_card',
        'personal_loan',
        'mortgage',
        'car_loan',
        'student_loan',
        'other'
    ) NOT NULL,
    creditor_name VARCHAR(100) NULL,

    original_amount BIGINT UNSIGNED NOT NULL,
    current_balance BIGINT UNSIGNED NOT NULL,
    interest_rate DECIMAL(5,2) NOT NULL,
    interest_type ENUM('fixed', 'variable') NOT NULL DEFAULT 'fixed',

    minimum_payment BIGINT UNSIGNED NOT NULL,
    payment_due_day TINYINT UNSIGNED NOT NULL,
    current_month_paid BOOLEAN NOT NULL DEFAULT FALSE,

    start_date DATE NOT NULL,
    expected_payoff_date DATE NULL,
    actual_payoff_date DATE NULL,

    status ENUM(
        'active',
        'paid_off',
        'defaulted',
        'settled'
    ) NOT NULL DEFAULT 'active',

    payment_asset_id BIGINT UNSIGNED NULL,

    include_in_net_worth BOOLEAN NOT NULL DEFAULT TRUE,
    auto_track_interest BOOLEAN NOT NULL DEFAULT TRUE,

    notes TEXT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,

    INDEX idx_debts_user_status (user_id, status),
    INDEX idx_debts_payment_due (payment_due_day, status),
    INDEX idx_debts_payment_asset (payment_asset_id),
    INDEX idx_debts_deleted_at (deleted_at)
);


CREATE TABLE IF NOT EXISTS debt_payments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    debt_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,

    amount BIGINT UNSIGNED NOT NULL,
    payment_type ENUM(
        'regular',
        'extra',
        'interest_only',
        'payoff',
        'adjustment'
    ) NOT NULL,

    principal_amount BIGINT UNSIGNED NOT NULL DEFAULT 0,
    interest_amount BIGINT UNSIGNED NOT NULL DEFAULT 0,
    fees_amount BIGINT UNSIGNED NOT NULL DEFAULT 0,

    balance_before BIGINT UNSIGNED NOT NULL,
    balance_after BIGINT UNSIGNED NOT NULL,

    transaction_id BIGINT UNSIGNED NULL,

    payment_date DATE NOT NULL,
    notes TEXT NULL,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_debt_payments_debt_date (debt_id, payment_date),
    INDEX idx_debt_payments_user_date (user_id, payment_date),
    INDEX idx_debt_payments_transaction (transaction_id),
    INDEX idx_debt_payments_type (payment_type)
);


CREATE TABLE IF NOT EXISTS debt_milestones (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,

    debt_id BIGINT UNSIGNED NOT NULL,
    user_id BIGINT UNSIGNED NOT NULL,

    milestone_type ENUM(
        '25_percent',
        '50_percent',
        '75_percent',
        'paid_off',
        'custom'
    ) NOT NULL,

    description VARCHAR(255) NULL,

    reached_at TIMESTAMP NULL,
    is_celebrated BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_debt_milestones_debt (debt_id),
    INDEX idx_debt_milestones_user (user_id),
    INDEX idx_debt_milestones_debt_type (debt_id, milestone_type),
    INDEX idx_debt_milestones_reached (reached_at)
);
