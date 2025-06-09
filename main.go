package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/joho/godotenv"
	"github.com/vit0-9/li-enricher-api/routes"

	_ "github.com/vit0-9/li-enricher-api/docs"
)

// @title           LinkedIn Enricher API
// @version         1.0
// @description     An API to enrich company data using LinkedIn.
// @license.name   Apache 2.0
// @license.url    http://www.apache.org/licenses/LICENSE-2.0.html
// @host            localhost:3000
// @BasePath        /api/v1
func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	app := fiber.New()

	app.Get("/swagger/*", swagger.HandlerDefault)

	routes.Setup(app)

	log.Println("Starting server on http://localhost:" + port)
	log.Println("API documentation available at http://localhost:" + port + "/swagger/index.html")
	log.Fatal(app.Listen(":" + port))
}
