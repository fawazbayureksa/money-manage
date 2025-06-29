package config

import (
    "fmt"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "log"
)

var DB *gorm.DB

func ConnectDatabase() {
    dsn := "root:@tcp(127.0.0.1:3306)/go_project?charset=utf8mb4&parseTime=True&loc=Local"
    database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

    if err != nil {
        log.Fatal("Gagal koneksi database:", err)
    }

    DB = database
    fmt.Println("Berhasil koneksi ke database")
}
