package controllers

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/server/web"

	"event-explorer/models"
)

type EventController struct {
	web.Controller
	APIClient *models.APIClient
	Template  *template.Template
}

type EventPageData struct {
	Title       string
	Query       string
	Category    string
	Events      []models.Event
	Error       string
	HasSearched bool
}

func NewEventController(apiClient *models.APIClient) *EventController {
	tmpl := template.Must(
		template.ParseFiles("views/index.html"),
	)

	return &EventController{
		APIClient: apiClient,
		Template:  tmpl,
	}
}

// Get handles GET /.
//
// It displays the main Event Explorer page without performing
// an API search.
func (c *EventController) Get() {
	c.render(EventPageData{
		Title:       "Event Explorer",
		Events:      []models.Event{},
		HasSearched: false,
	})
}

// Search handles GET /search.
//
// Example:
//
//	/search?query=concert&category=Music
func (c *EventController) Search() {
	query := strings.TrimSpace(c.Ctx.Input.Query("query"))
	category := strings.TrimSpace(c.Ctx.Input.Query("category"))

	data := EventPageData{
		Title:       "Search Results",
		Query:       query,
		Category:    category,
		Events:      []models.Event{},
		HasSearched: true,
	}

	if c.APIClient == nil {
		data.Error = "Event API client is not configured."
		c.render(data)
		return
	}

	requestURL := c.APIClient.BuildSearchURL(query, category)

	events, err := c.APIClient.FetchEventsConcurrent(
		[]string{requestURL},
	)

	if err != nil {
		data.Error = "Unable to load events. Please try again."
		c.render(data)
		return
	}

	data.Events = events

	c.render(data)
}

func (c *EventController) render(data EventPageData) {
	if c.Template == nil {
		c.Ctx.ResponseWriter.WriteHeader(http.StatusInternalServerError)
		_, _ = c.Ctx.ResponseWriter.Write([]byte("Template is not configured"))
		return
	}

	if err := c.Template.Execute(
		c.Ctx.ResponseWriter,
		data,
	); err != nil {
		// The response may already have started, so we only log the
		// rendering error here.
		return
	}
}