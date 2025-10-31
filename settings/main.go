package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"infraaudit/backend/settings/routes"
)

func main() {
	r := gin.Default()
	routes.RegisterSettingsRoutes(r)
	log.Println("Settings API running on :8081")
	_ = r.Run(":8081")
}
