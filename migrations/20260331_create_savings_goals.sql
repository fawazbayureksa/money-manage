-- Migration: create savings_goals and savings_contributions tables

CREATE TABLE savings_goals (
  id INT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id INT UNSIGNED NOT NULL,
  asset_id BIGINT UNSIGNED DEFAULT NULL,
  name VARCHAR(255) NOT NULL,
  description VARCHAR(500) DEFAULT '',
  target_amount INT NOT NULL,
  current_amount INT NOT NULL DEFAULT 0,
  currency VARCHAR(10) NOT NULL DEFAULT 'IDR',
  deadline DATETIME DEFAULT NULL,
  status VARCHAR(20) NOT NULL DEFAULT 'active',
  icon VARCHAR(50) DEFAULT '',
  color VARCHAR(20) DEFAULT '',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_user_id (user_id),
  KEY idx_asset_id (asset_id),
  KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE savings_contributions (
  id INT UNSIGNED NOT NULL AUTO_INCREMENT,
  goal_id INT UNSIGNED NOT NULL,
  user_id INT UNSIGNED NOT NULL,
  amount INT NOT NULL,
  note VARCHAR(500) DEFAULT '',
  date DATETIME NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_goal_id (goal_id),
  KEY idx_user_id (user_id),
  CONSTRAINT fk_contributions_goal FOREIGN KEY (goal_id) REFERENCES savings_goals (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
