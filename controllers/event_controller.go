package controllers

import (
	"context"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/server/web"

	"event-explorer/models"
	"event-explorer/services"
)

type EventController struct {
	web.Controller
	EventService *services.EventService
}

// Data passed to listing.tpl.
type ListingPageData struct {
	Title       string
	City        string
	CountryCode string

	Music  []models.Event
	Sports []models.Event

	Error       string
	HasSearched bool
}

// Data passed to details.tpl.
type DetailsPageData struct {
	Title       string
	Event       interface{}
	City        string
	CountryCode string
	Error       string
}

// List displays events for the selected city.
// GET /events?city=Toronto&countryCode=CA
func (c *EventController) List() {

	
	city := strings.TrimSpace(
		c.Ctx.Input.Query("city"),
	)

	
	countryCode := strings.ToUpper(
		strings.TrimSpace(
			c.Ctx.Input.Query("countryCode"),
		),
	)

	
	data := ListingPageData{
		Title:       "Events",
		City:        city,
		CountryCode: countryCode,

		Music:  []models.Event{},
		Sports: []models.Event{},

		HasSearched: city != "" && countryCode != "",
	}

	
	if city == "" || countryCode == "" {
		data.Error = "Please select a city first."

		c.TplName = "listing.tpl"
		c.Data["Page"] = data
		return
	}

	// Make sure the service was initialized.
	if c.EventService == nil {
		data.Error = "Event service is not configured."

		c.TplName = "listing.tpl"
		c.Data["Page"] = data
		return
	}

	// Fetch Music and Sports events.
	// EventService -> APIClient
	// APIClient fetches Music and Sports concurrently.
	events, err := c.EventService.ListEvents(
		context.Background(),
		city,
		countryCode,
	)

	// If both category requests failed, display an error.
	if err != nil {
		data.Error = "Unable to load events. Please try again."
	}

	// Copy Music events into the page data.
	if music, ok := events["Music"]; ok {
		data.Music = music
	}

	// Copy Sports events into the page data.
	if sports, ok := events["Sports"]; ok {
		data.Sports = sports
	}

	// Render listing page.
	c.TplName = "listing.tpl"
	c.Data["Page"] = data
}

// Details displays information about one event.
// GET /events/:eventId
func (c *EventController) Details() {

	// Get event ID from route parameter.
	eventID := strings.TrimSpace(
		c.Ctx.Input.Param(":eventId"),
	)

	data := DetailsPageData{
		Title: "Event Details",
	}

	// Validate event ID.
	if eventID == "" {
		data.Error = "Invalid event."

		c.Ctx.ResponseWriter.WriteHeader(
			http.StatusBadRequest,
		)

		c.TplName = "details.tpl"
		c.Data["Page"] = data
		return
	}

	// Make sure the service was initialized.
	if c.EventService == nil {
		data.Error = "Event service is not configured."

		c.Ctx.ResponseWriter.WriteHeader(
			http.StatusInternalServerError,
		)

		c.TplName = "details.tpl"
		c.Data["Page"] = data
		return
	}

	// Fetch event details from Ticketmaster.
	event, err := c.EventService.GetEventDetails(
		context.Background(),
		eventID,
	)

	if err != nil {
		data.Error = "Event not found."

		c.Ctx.ResponseWriter.WriteHeader(
			http.StatusNotFound,
		)

		c.TplName = "details.tpl"
		c.Data["Page"] = data
		return
	}

	// Put event into page data.
	data.Event = event

	// Render details page.
	c.TplName = "details.tpl"
	c.Data["Page"] = data
}