package config

import (
    "fmt"
    "log"
    "os"

    "my-api/utils"

    "github.com/joho/godotenv"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {

    err := godotenv.Load()
	if err != nil {
		utils.LogError("Error loading .env file")
		log.Fatal("Error loading .env file")
	}

    // Get environment variables
    dbName := os.Getenv("DB_NAME")
    dbUser := os.Getenv("DB_USER")
    dbPass := os.Getenv("DB_PASS")
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")

    utils.LogInfof("Connecting to database: %s@%s:%s/%s", dbUser, dbHost, dbPort, dbName)
    fmt.Println("Connecting to database:", dbUser, dbHost, dbPort, dbName, dbPass)
    // Construct the DSN using environment variables
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        dbUser, dbPass, dbHost, dbPort, dbName)
        
    database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

    if err != nil {
        utils.LogErrorf("Failed to connect to database: %v", err)
        log.Fatal("Gagal koneksi database:", err)
    }

    DB = database
    utils.LogInfo("Successfully connected to database")
    fmt.Println("Berhasil koneksi ke database")
}
