package controllers

import (
	"context"
	"net/http"
	"strings"

	"github.com/beego/beego/v2/server/web"

	"event-explorer/services"
)

type EventController struct {
	web.Controller
	EventService *services.EventService
}

type ListingPageData struct {
	Title       string
	City        string
	CountryCode string

	MusicEvents  interface{} // renamed from Music
    SportsEvents interface{}

	Error       string
	HasSearched bool
}

type DetailsPageData struct {
	Title       string
	Event       interface{}
	City        string
	CountryCode string
	Error       string
}

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
		MusicEvents :      []interface{}{},
		SportsEvents:      []interface{}{},
		HasSearched: city != "" && countryCode != "",
	}

	if city == "" || countryCode == "" {
		data.Error = "Please select a city first."
		c.TplName = "listing.tpl"
		c.Data["Page"] = data
		return
	}

	if c.EventService == nil {
		data.Error = "Event service is not configured."
		c.TplName = "listing.tpl"
		c.Data["Page"] = data
		return
	}

	events, err := c.EventService.ListEvents(
		context.Background(),
		city,
		countryCode,
	)

	if err != nil {
		data.Error = "Unable to load events. Please try again."
	}

	if music, ok := events["Music"]; ok {
		data.MusicEvents = music
	}

	if sports, ok := events["Sports"]; ok {
		data.SportsEvents = sports
	}

	c.TplName = "listing.tpl"
	c.Data["Page"] = data
}

func (c *EventController) Details() {

	eventID := strings.TrimSpace(
		c.Ctx.Input.Param(":eventId"),
	)

	data := DetailsPageData{
		Title: "Event Details",
	}

	if eventID == "" {
		data.Error = "Invalid event."
		c.Ctx.ResponseWriter.WriteHeader(
			http.StatusBadRequest,
		)
		c.TplName = "details.tpl"
		c.Data["Page"] = data
		return
	}

	if c.EventService == nil {
		data.Error = "Event service is not configured."
		c.Ctx.ResponseWriter.WriteHeader(
			http.StatusInternalServerError,
		)
		c.TplName = "details.tpl"
		c.Data["Page"] = data
		return
	}

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

	data.Event = event

	c.TplName = "details.tpl"
	c.Data["Page"] = data
}