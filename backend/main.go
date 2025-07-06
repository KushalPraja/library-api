package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kushalpraja/library-api/db"
	"github.com/kushalpraja/library-api/routes"
)

func main() {
	db.Connect()
	r := gin.Default()
	routes.SetupRoutes(r)
	r.Run(":8080")
}
