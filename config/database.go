package config

import (
    "fmt"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "log"
)

var DB *gorm.DB

func ConnectDatabase() {
    dsn := "money_manage_dev:qsqcfthn132@tcp(127.0.0.1:3306)/money_manage_dev?charset=utf8mb4&parseTime=True&loc=Local"
    database, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

    if err != nil {
        log.Fatal("Gagal koneksi database:", err)
    }

    DB = database
    fmt.Println("Berhasil koneksi ke database")
}
