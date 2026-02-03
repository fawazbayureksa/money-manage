package models

import (
	"gorm.io/gorm"
	"log"
)

// AddAssetIDToTransactions adds asset_id column to transactions table
// This migration is backward compatible - existing v1 API will continue to work
func AddAssetIDToTransactions(db *gorm.DB) error {
	log.Println("Starting migration: Add AssetID to Transactions")

	// Check if asset_id column already exists in the table
	var hasColumn bool
	if err := db.Raw(`
        SELECT COUNT(*) > 0 
        FROM information_schema.columns 
        WHERE table_schema = DATABASE() 
        AND table_name = 'transactions' 
        AND column_name = 'asset_id'
    `).Scan(&hasColumn).Error; err != nil {
		log.Printf("Error checking for asset_id column: %v", err)
		return err
	}

	if hasColumn {
		log.Println("Column asset_id already exists in transactions table")
	} else {
		// Add asset_id column
		if err := db.Migrator().AddColumn(&Transaction{}, "asset_id"); err != nil {
			log.Printf("Error adding asset_id column: %v", err)
			return err
		}

		// Create index on asset_id
		if err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_transactions_asset_id 
        ON transactions(asset_id)
    `).Error; err != nil {
			log.Printf("Error creating index on asset_id: %v", err)
			return err
		}

		// Migrate existing data - map bank_id to asset_id
		// This assumes each bank has a corresponding asset with the same ID
		log.Println("Migrating existing transactions from bank_id to asset_id...")
		result := db.Exec(`
        UPDATE transactions t
        INNER JOIN assets a ON t.bank_id = a.id
        SET t.asset_id = a.id
        WHERE t.asset_id = 0 OR t.asset_id IS NULL
    `)

		if result.Error != nil {
			log.Printf("Error migrating data: %v", result.Error)
			return result.Error
		}

		log.Printf("Migrated %d transactions to use asset_id", result.RowsAffected)

		// Create default assets for banks without matching assets
		log.Println("Creating default assets for banks...")
		result = db.Exec(`
        INSERT INTO assets (user_id, name, type, balance, currency, bank_name, account_no, created_at, updated_at)
        SELECT DISTINCT t.user_id, b.bank_name, 'bank', 0, 'USD', b.bank_name, '', NOW(), NOW()
        FROM transactions t
        INNER JOIN banks b ON t.bank_id = b.id
        WHERE NOT EXISTS (
            SELECT 1 FROM assets a WHERE a.id = t.bank_id
        )
    `)

		if result.Error != nil {
			log.Printf("Error creating default assets: %v", result.Error)
			return result.Error
		}

		log.Printf("Created %d default assets for banks", result.RowsAffected)

		// Update transactions to use newly created assets
		result = db.Exec(`
        UPDATE transactions t
        INNER JOIN (
            SELECT user_id, bank_name, MIN(id) as min_asset_id
            FROM assets
            WHERE type = 'bank'
            GROUP BY user_id, bank_name
        ) a ON t.user_id = a.user_id AND t.asset_id = 0
        SET t.asset_id = a.min_asset_id
    `)

		if result.Error != nil {
			log.Printf("Error updating transactions with new assets: %v", result.Error)
			return result.Error
		}

		log.Printf("Updated %d transactions with new assets", result.RowsAffected)
	}

	// Make bank_id nullable for V2 API
	log.Println("Making bank_id nullable...")
	if err := db.Exec(`
        ALTER TABLE transactions 
        MODIFY COLUMN bank_id INT UNSIGNED NULL
    `).Error; err != nil {
		log.Printf("Error making bank_id nullable: %v", err)
		// Non-fatal error, continue
	}

	// Drop foreign key constraint on bank_id to allow 0/NULL values
	log.Println("Dropping foreign key constraint on bank_id...")
	if err := db.Exec(`
        ALTER TABLE transactions 
        DROP FOREIGN KEY IF EXISTS fk_transactions_bank
    `).Error; err != nil {
		log.Printf("Error dropping bank_id foreign key: %v", err)
		// Non-fatal error, continue
	}

	// Add foreign key constraint for asset_id
	log.Println("Adding foreign key constraint for asset_id...")
	if err := db.Exec(`
        ALTER TABLE transactions 
        ADD CONSTRAINT fk_transactions_asset 
        FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE RESTRICT
    `).Error; err != nil {
		log.Printf("Error adding foreign key constraint: %v", err)
		// Ignore duplicate key errors
		if err.Error() != "Error 1826: Duplicate foreign key constraint name 'fk_transactions_asset'" &&
			err.Error() != "Error 1005 (HY000): Can't create table `money_manage`.`transactions` (errno: 121 \"Duplicate key on write or update\")" {
			return err
		}
	}

	log.Println("Migration completed successfully!")
	return nil
}

// RollbackAddAssetIDToTransactions rolls back the migration
func RollbackAddAssetIDToTransactions(db *gorm.DB) error {
	log.Println("Rolling back migration: Remove AssetID from Transactions")

	// Drop foreign key constraint
	if err := db.Exec(`
        ALTER TABLE transactions DROP FOREIGN KEY IF EXISTS fk_transactions_asset
    `).Error; err != nil {
		log.Printf("Error dropping foreign key: %v", err)
	}

	// Drop index
	if err := db.Exec(`
        DROP INDEX IF EXISTS idx_transactions_asset_id ON transactions
    `).Error; err != nil {
		log.Printf("Error dropping index: %v", err)
	}

	// Drop column
	if err := db.Migrator().DropColumn(&Transaction{}, "asset_id"); err != nil {
		log.Printf("Error dropping column: %v", err)
	}

	log.Println("Rollback completed successfully!")
	return nil
}
