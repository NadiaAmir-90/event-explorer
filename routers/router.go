package routers

import (
	"net/http"
	"os"

	"github.com/beego/beego/v2/server/web"

	"event-explorer/controllers"
	"event-explorer/models"
	"event-explorer/services"
)

func init() {

	// ==========================================
	// API KEYS (Env Vars with AppConfig Fallback)
	// ==========================================

	ticketmasterKey := os.Getenv("TICKETMASTER_API_KEY")
	if ticketmasterKey == "" {
		ticketmasterKey, _ = web.AppConfig.String("ticketmaster_api_key")
	}

	googleKey := os.Getenv("GOOGLE_PLACES_API_KEY")
	if googleKey == "" {
		googleKey, _ = web.AppConfig.String("google_places_api_key")
	}

	// ==========================================
	// BASE URLS
	// ==========================================

	ticketmasterBaseURL, err := web.AppConfig.String("ticketmaster_base_url")
	if err != nil {
		panic(err)
	}

	googleBaseURL, err := web.AppConfig.String("google_places_base_url")
	if err != nil {
		panic(err)
	}

	// ==========================================
	// SHARED HTTP CLIENT & CACHE
	// ==========================================

	httpClient := &http.Client{}
	cache := models.NewCache()

	// ==========================================
	// CLIENTS
	// ==========================================

	ticketmasterClient := models.NewAPIClient(
		httpClient,
		cache,
		ticketmasterBaseURL,
		ticketmasterKey,
	)

	googleClient := models.NewGooglePlacesClient(
		googleBaseURL,
		googleKey,
		httpClient,
	)

	// ==========================================
	// SERVICES
	// ==========================================

	eventService := services.NewEventService(ticketmasterClient)
	locationService := services.NewLocationService(googleClient)

	// ==========================================
	// CONTROLLERS
	// ==========================================

	homeController := &controllers.HomeController{}

	eventController := &controllers.EventController{
		EventService: eventService,
	}

	locationController := &controllers.LocationController{
		LocationService: locationService,
	}

	redirectController := &controllers.RedirectController{
		EventService: eventService,
	}

	// ==========================================
	// ROUTES
	// ==========================================

	web.Router("/", homeController, "get:Get")
	web.Router("/events", eventController, "get:List")
	web.Router("/events/:eventId", eventController, "get:Details")

	web.Router("/api/locations/autocomplete", locationController, "get:Autocomplete")
	web.Router("/api/locations/:placeId", locationController, "get:Details")

	web.Router("/redirect/:eventId", redirectController, "get:Ticket")
}