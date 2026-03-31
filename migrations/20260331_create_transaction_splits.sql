-- Migration: create transaction_splits table
-- Stores individual category-level splits for a parent transaction.
CREATE TABLE transaction_splits (
  id INT UNSIGNED NOT NULL AUTO_INCREMENT,
  transaction_id INT UNSIGNED NOT NULL,
  category_id INT UNSIGNED NULL,
  amount INT NOT NULL,
  description VARCHAR(200) NOT NULL DEFAULT '',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_ts_transaction_id (transaction_id),
  KEY idx_ts_category_id (category_id),
  CONSTRAINT fk_ts_transaction FOREIGN KEY (transaction_id) REFERENCES transactions (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
