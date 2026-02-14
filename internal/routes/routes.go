package routes

import (
	"github.com/akashtripathi12/TBO_Backend/internal/config"
	"github.com/akashtripathi12/TBO_Backend/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, cfg *config.Config, repo *handlers.Repository) {
	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"env":    cfg.Env,
		})
	})

	// API v1 group
	api := app.Group("/api/v1")

	// Auth routes
	auth := api.Group("/auth")
	auth.Post("/login/agent", repo.LoginAgent)
	auth.Post("/login/guest", repo.LoginGuest)
	auth.Post("/logout", repo.Logout)
	auth.Get("/me", repo.GetCurrentUser)

	// Dashboard metrics
	api.Get("/dashboard/metrics", repo.GetMetrics)

	// Event routes
	events := api.Group("/events")
	events.Get("/", repo.GetEvents)
	events.Post("/", repo.CreateEvent)
	events.Get("/:id", repo.GetEvent)
	events.Put("/:id", repo.UpdateEvent)
	events.Delete("/:id", repo.DeleteEvent)
	events.Get("/:id/venues", repo.GetEventVenues)
	events.Get("/:id/allocations", repo.GetEventAllocations)
	events.Get("/:id/guests", repo.GetGuests) // Event-specific guests

	// Guest routes
	guests := api.Group("/guests")
	guests.Get("/:id", repo.GetGuest)
	guests.Put("/:id", repo.UpdateGuest)
	guests.Delete("/:id", repo.DeleteGuest)
	guests.Post("/", repo.CreateGuest)
	guests.Post("/:id/subguests", repo.AddSubGuest)

	// Allocation routes
	allocations := api.Group("/allocations")
	allocations.Post("/", repo.CreateAllocation)
	allocations.Put("/:id", repo.UpdateAllocation)

	locations := api.Group("/locations")
    locations.Get("/countries", repo.GetCountries)
    locations.Get("/cities", repo.GetCities)
}
