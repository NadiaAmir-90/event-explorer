package controllers

import (
	"context"
	"strings"

	"event-explorer/services"

	"github.com/beego/beego/v2/server/web"
)

type LocationController struct {
	web.Controller

	LocationService *services.LocationService
}


// ==================================================
// GET /api/locations/autocomplete
// ==================================================

func (c *LocationController) Autocomplete() {

	input := strings.TrimSpace(
		c.Ctx.Input.Query("input"),
	)

	sessionToken := strings.TrimSpace(
		c.Ctx.Input.Query("sessionToken"),
	)


	if input == "" {

		c.Ctx.ResponseWriter.WriteHeader(400)

		c.Data["json"] = map[string]string{
			"error": "input is required",
		}

		c.ServeJSON()

		return
	}


	if sessionToken == "" {

		c.Ctx.ResponseWriter.WriteHeader(400)

		c.Data["json"] = map[string]string{
			"error": "sessionToken is required",
		}

		c.ServeJSON()

		return
	}


	if c.LocationService == nil {

		c.Ctx.ResponseWriter.WriteHeader(500)

		c.Data["json"] = map[string]string{
			"error": "location service unavailable",
		}

		c.ServeJSON()

		return
	}


	suggestions, err :=
		c.LocationService.Autocomplete(
			context.Background(),
			input,
			sessionToken,
		)


	if err != nil {

		c.Ctx.ResponseWriter.WriteHeader(502)

		c.Data["json"] = map[string]string{
			"error": err.Error(),
		}

		c.ServeJSON()

		return
	}


	c.Data["json"] = map[string]interface{}{
		"suggestions": suggestions,
	}

	c.ServeJSON()
}


// ==================================================
// GET /api/locations/:placeId
// ==================================================

func (c *LocationController) Details() {

	placeID := strings.TrimSpace(
		c.Ctx.Input.Param(":placeId"),
	)

	sessionToken := strings.TrimSpace(
		c.Ctx.Input.Query("sessionToken"),
	)


	if placeID == "" {

		c.Ctx.ResponseWriter.WriteHeader(400)

		c.Data["json"] = map[string]string{
			"error": "placeId is required",
		}

		c.ServeJSON()

		return
	}


	if sessionToken == "" {

		c.Ctx.ResponseWriter.WriteHeader(400)

		c.Data["json"] = map[string]string{
			"error": "sessionToken is required",
		}

		c.ServeJSON()

		return
	}


	if c.LocationService == nil {

		c.Ctx.ResponseWriter.WriteHeader(500)

		c.Data["json"] = map[string]string{
			"error": "location service unavailable",
		}

		c.ServeJSON()

		return
	}


	location, err :=
		c.LocationService.GetPlace(
			context.Background(),
			placeID,
			sessionToken,
		)


	if err != nil {

		c.Ctx.ResponseWriter.WriteHeader(502)

		c.Data["json"] = map[string]string{
			"error": err.Error(),
		}

		c.ServeJSON()

		return
	}


	c.Data["json"] = location

	c.ServeJSON()
}