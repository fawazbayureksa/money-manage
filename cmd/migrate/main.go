package main

import (
	"flag"
	"log"
	"my-api/config"
	"my-api/models"
)

func main() {
	action := flag.String("action", "up", "Migration action: up or down")
	flag.Parse()

	// Initialize database connection
	config.ConnectDatabase()
	db := config.DB

	switch *action {
	case "up":
		log.Println("Running migration: Add AssetID to Transactions")
		if err := models.AddAssetIDToTransactions(db); err != nil {
			log.Fatal("Migration failed:", err)
		}
		log.Println("Migration completed successfully!")

	case "down":
		log.Println("Rolling back migration: Remove AssetID from Transactions")
		if err := models.RollbackAddAssetIDToTransactions(db); err != nil {
			log.Fatal("Rollback failed:", err)
		}
		log.Println("Rollback completed successfully!")

	default:
		log.Fatal("Invalid action. Use 'up' or 'down'")
	}
}
