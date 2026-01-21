package models

import "my-api/config"

// Init hook to ensure Asset table is migrated on startup.
func init() {
    if config.DB != nil {
        config.DB.AutoMigrate(&Asset{})
    }
}
