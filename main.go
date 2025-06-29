package main

import (
    "my-api/config"
    "my-api/models"
    "my-api/routes"
)

func main() {
    config.ConnectDatabase()
    models.AutoMigrate()

    r := routes.SetupRouter()
    r.Run(":8080")
}
