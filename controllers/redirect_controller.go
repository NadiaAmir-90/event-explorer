package controllers

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/beego/beego/v2/server/web"

	"event-explorer/services"
)

type RedirectController struct {
	web.Controller
	EventService *services.EventService
}

func (c *RedirectController) Ticket() {

	eventID := strings.TrimSpace(
		c.Ctx.Input.Param(":eventId"),
	)

	if eventID == "" {
		c.Ctx.ResponseWriter.WriteHeader(
			http.StatusBadRequest,
		)
		return
	}

	event, err := c.EventService.GetEventDetails(
		context.Background(),
		eventID,
	)

	if err != nil || event.TicketURL == "" {
		c.Ctx.ResponseWriter.WriteHeader(
			http.StatusNotFound,
		)
		return
	}

	if !isApprovedTicketURL(event.TicketURL) {
		c.Ctx.ResponseWriter.WriteHeader(
			http.StatusBadGateway,
		)
		return
	}

	c.Redirect(event.TicketURL, http.StatusFound)
}

func isApprovedTicketURL(rawURL string) bool {

	parsed, err := url.Parse(rawURL)

	if err != nil {
		return false
	}

	if parsed.Scheme != "https" {
		return false
	}

	host := strings.ToLower(parsed.Hostname())

	approved := map[string]bool{
		"ticketmaster.com":     true,
		"www.ticketmaster.com": true,
		"ticketmaster.ca":     true,
		"www.ticketmaster.ca": true,
	}

	return approved[host]
}