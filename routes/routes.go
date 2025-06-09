package routes

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/vit0-9/li-enricher-api/services"
)

type AppRoutes struct {
	companyService *services.CompanyService
	authService    *services.AuthService
}

func Setup(app *fiber.App) {
	companyService := services.NewCompanyService()
	authService := services.NewAuthService()
	routes := &AppRoutes{
		companyService: companyService,
		authService:    authService,
	}

	api := app.Group("/api/v1")

	api.Get("/auth/validate", routes.handleValidateAuth)
	api.Get("/company/:slug", routes.handleScrapeCompany)
}

// handleScrapeCompany scrapes data for a LinkedIn company page.
// @Summary      Scrape Company Data
// @Description  Scrapes data for a LinkedIn company page. If a session cookie is provided via the 'X-Linkedin-Session-Cookie' header, it performs a full, authenticated scrape. Otherwise, it performs a public scrape for basic JSON-LD data.
// @Tags         Company
// @Accept       json
// @Produce      json
// @Param        slug                        path      string                          true   "Company Slug (e.g., 'google')"
// @Param        X-Linkedin-Session-Cookie   header    string                          false  "LinkedIn 'li_at' session cookie for authenticated scraping"
// @Success      200                         {object}  object{scrapeType=string,data=object}  "Successfully scraped data. 'scrapeType' will be 'full' or 'public'."
// @Failure      400                         {object}  object{error=string}                   "Bad Request - Invalid input"
// @Failure      500                         {object}  object{error=string,details=string}    "Internal Server Error"
// @Router       /company/{slug} [get]
func (r *AppRoutes) handleScrapeCompany(c *fiber.Ctx) error {
	slug := c.Params("slug")
	sessionCookie := c.Get("X-Linkedin-Session-Cookie")

	if slug == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Company slug cannot be empty"})
	}

	// The handler's only job is to call the service and render the response.
	data, scrapeType, err := r.companyService.EnrichCompanyData(slug, sessionCookie)
	if err != nil {
		log.Printf("Error from service: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to process company data",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"scrapeType": scrapeType,
		"data":       data,
	})
}

// handleValidateAuth checks if a given LinkedIn session cookie is valid.
// @Summary      Validate Session Cookie
// @Description  Checks if a given LinkedIn session cookie ('li_at') is valid and active.
// @Tags         Authentication
// @Produce      json
// @Param        X-Linkedin-Session-Cookie   header    string                                 true   "LinkedIn 'li_at' session cookie"
// @Success      200                         {object}  object{valid=bool}
// @Failure      400                         {object}  object{error=string}
// @Failure      500                         {object}  object{error=string,details=string}
// @Router       /auth/validate [get]
func (r *AppRoutes) handleValidateAuth(c *fiber.Ctx) error {
	sessionCookie := c.Get("X-Linkedin-Session-Cookie")
	if sessionCookie == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Header 'X-Linkedin-Session-Cookie' is required"})
	}

	isValid, err := r.authService.ValidateSession(sessionCookie)
	if err != nil {
		log.Printf("Error during session validation: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"valid": isValid})
}
