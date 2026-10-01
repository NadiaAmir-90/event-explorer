package controllers

import (
	"event-explorer/models"
	"event-explorer/services"
	"net/http"
	"net/http/httptest"
	"testing"

	beegoContext "github.com/beego/beego/v2/server/web/context"
)

// createController creates an EventController with a Beego request context.
func createController(
	method string,
	target string,
) (*EventController, *httptest.ResponseRecorder) {

	req := httptest.NewRequest(method, target, nil)
	recorder := httptest.NewRecorder()

	ctx := beegoContext.NewContext()
	ctx.Reset(recorder, req)

	controller := &EventController{
		EventService: nil,
	}

	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	return controller, recorder
}

// ============================================================
// LIST TESTS
// ============================================================

// Test when city and countryCode are missing.
func TestEventController_List_MissingLocation(t *testing.T) {

	controller, recorder := createController(
		http.MethodGet,
		"/events",
	)

	controller.List()

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			recorder.Code,
		)
	}

	if controller.TplName != "listing.tpl" {
		t.Fatalf(
			"expected template listing.tpl, got %s",
			controller.TplName,
		)
	}

	page, ok := controller.Data["Page"].(ListingPageData)

	if !ok {
		t.Fatal("expected Page to contain ListingPageData")
	}

	if page.Error != "Please select a city first." {
		t.Fatalf(
			"expected location error, got %q",
			page.Error,
		)
	}

	if page.HasSearched {
		t.Fatal("expected HasSearched to be false")
	}
}

// Test when only city is provided.
func TestEventController_List_MissingCountryCode(t *testing.T) {

	controller, recorder := createController(
		http.MethodGet,
		"/events?city=Toronto",
	)

	controller.List()

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			recorder.Code,
		)
	}

	page, ok := controller.Data["Page"].(ListingPageData)

	if !ok {
		t.Fatal("expected Page to contain ListingPageData")
	}

	if page.City != "Toronto" {
		t.Fatalf(
			"expected city Toronto, got %q",
			page.City,
		)
	}

	if page.Error != "Please select a city first." {
		t.Fatalf(
			"expected location error, got %q",
			page.Error,
		)
	}
}

// Test when country code is normalized to uppercase.
func TestEventController_List_NormalizesCountryCode(t *testing.T) {

	controller, recorder := createController(
		http.MethodGet,
		"/events?city=Toronto&countryCode=ca",
	)

	// Service is intentionally nil.
	// This allows us to verify that the controller correctly
	// normalizes the country code before checking the service.
	controller.List()

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			recorder.Code,
		)
	}

	page, ok := controller.Data["Page"].(ListingPageData)

	if !ok {
		t.Fatal("expected Page to contain ListingPageData")
	}

	if page.City != "Toronto" {
		t.Fatalf(
			"expected city Toronto, got %q",
			page.City,
		)
	}

	if page.CountryCode != "CA" {
		t.Fatalf(
			"expected country code CA, got %q",
			page.CountryCode,
		)
	}

	if !page.HasSearched {
		t.Fatal("expected HasSearched to be true")
	}

	if page.Error != "Event service is not configured." {
		t.Fatalf(
			"expected service configuration error, got %q",
			page.Error,
		)
	}
}

// Test when EventService is not configured.
func TestEventController_List_ServiceNotConfigured(t *testing.T) {

	controller, recorder := createController(
		http.MethodGet,
		"/events?city=Toronto&countryCode=CA",
	)

	controller.EventService = nil

	controller.List()

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			recorder.Code,
		)
	}

	if controller.TplName != "listing.tpl" {
		t.Fatalf(
			"expected listing.tpl, got %s",
			controller.TplName,
		)
	}

	page, ok := controller.Data["Page"].(ListingPageData)

	if !ok {
		t.Fatal("expected Page to contain ListingPageData")
	}

	if page.Error != "Event service is not configured." {
		t.Fatalf(
			"expected service configuration error, got %q",
			page.Error,
		)
	}
}

// ============================================================
// DETAILS TESTS
// ============================================================

// Test Details when event ID is missing.
func TestEventController_Details_MissingEventID(t *testing.T) {

	controller, recorder := createController(
		http.MethodGet,
		"/events/",
	)

	// Simulate missing :eventId route parameter.
	controller.Ctx.Input.SetParam(":eventId", "")

	controller.Details()

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status 400, got %d",
			recorder.Code,
		)
	}

	if controller.TplName != "details.tpl" {
		t.Fatalf(
			"expected details.tpl, got %s",
			controller.TplName,
		)
	}

	page, ok := controller.Data["Page"].(DetailsPageData)

	if !ok {
		t.Fatal("expected Page to contain DetailsPageData")
	}

	if page.Error != "Invalid event." {
		t.Fatalf(
			"expected invalid event error, got %q",
			page.Error,
		)
	}
}

// Test Details when EventService is not configured.
func TestEventController_Details_ServiceNotConfigured(t *testing.T) {

	controller, recorder := createController(
		http.MethodGet,
		"/events/123",
	)

	controller.Ctx.Input.SetParam(":eventId", "123")

	controller.EventService = nil

	controller.Details()

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status 500, got %d",
			recorder.Code,
		)
	}

	if controller.TplName != "details.tpl" {
		t.Fatalf(
			"expected details.tpl, got %s",
			controller.TplName,
		)
	}

	page, ok := controller.Data["Page"].(DetailsPageData)

	if !ok {
		t.Fatal("expected Page to contain DetailsPageData")
	}

	if page.Error != "Event service is not configured." {
		t.Fatalf(
			"expected service configuration error, got %q",
			page.Error,
		)
	}
}
func TestEventController_List_Success(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			_, _ = w.Write([]byte(`{
				"_embedded": {
					"events": [
						{
							"id": "event-1",
							"name": "Toronto Music Event",
							"url": "https://www.ticketmaster.com/event-1",
							"classifications": [
								{
									"segment": {
										"name": "Music"
									},
									"genre": {
										"name": "Rock"
									}
								}
							],
							"dates": {
								"start": {
									"localDate": "2026-10-10"
								}
							},
							"_embedded": {
								"venues": [
									{
										"name": "Test Venue",
										"city": {
											"name": "Toronto"
										}
									}
								]
							}
						}
					]
				}
			}`))
		}),
	)
	defer server.Close()

	client := models.NewAPIClient(
		server.Client(),
		models.NewCache(),
		server.URL,
		"test-key",
	)

	service := services.NewEventService(client)

	controller, recorder := createController(
		http.MethodGet,
		"/events?city=Toronto&countryCode=CA",
	)

	controller.EventService = service

	controller.List()

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	if controller.TplName != "listing.tpl" {
		t.Fatalf(
			"expected listing.tpl, got %s",
			controller.TplName,
		)
	}

	page, ok := controller.Data["Page"].(ListingPageData)
	if !ok {
		t.Fatal("expected Page to contain ListingPageData")
	}

	if page.City != "Toronto" {
		t.Fatalf("expected Toronto, got %q", page.City)
	}

	if page.CountryCode != "CA" {
		t.Fatalf("expected CA, got %q", page.CountryCode)
	}

	if !page.HasSearched {
		t.Fatal("expected HasSearched to be true")
	}

	if len(page.Music) == 0 {
		t.Fatal("expected Music events")
	}

	if page.Music[0].ID != "event-1" {
		t.Fatalf(
			"expected event-1, got %q",
			page.Music[0].ID,
		)
	}
}
func TestEventController_List_ServiceError(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`server error`))
		}),
	)
	defer server.Close()

	client := models.NewAPIClient(
		server.Client(),
		models.NewCache(),
		server.URL,
		"test-key",
	)

	service := services.NewEventService(client)

	controller, recorder := createController(
		http.MethodGet,
		"/events?city=Toronto&countryCode=CA",
	)

	controller.EventService = service

	controller.List()

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	page, ok := controller.Data["Page"].(ListingPageData)
	if !ok {
		t.Fatal("expected Page to contain ListingPageData")
	}

	if page.Error != "Unable to load events. Please try again." {
		t.Fatalf(
			"expected load error, got %q",
			page.Error,
		)
	}
}
func TestEventController_Details_Success(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			_, _ = w.Write([]byte(`{
				"id": "event-123",
				"name": "Toronto Concert",
				"description": "A test concert",
				"url": "https://www.ticketmaster.com/event-123",
				"dates": {
					"start": {
						"localDate": "2026-10-20"
					}
				},
				"classifications": [
					{
						"segment": {
							"name": "Music"
						},
						"genre": {
							"name": "Rock"
						}
					}
				],
				"_embedded": {
					"venues": [
						{
							"name": "Test Arena",
							"city": {
								"name": "Toronto"
							},
							"state": {
								"stateCode": "ON"
							}
						}
					]
				}
			}`))
		}),
	)
	defer server.Close()

	client := models.NewAPIClient(
		server.Client(),
		models.NewCache(),
		server.URL,
		"test-key",
	)

	service := services.NewEventService(client)

	controller, recorder := createController(
		http.MethodGet,
		"/events/event-123",
	)

	controller.EventService = service
	controller.Ctx.Input.SetParam(":eventId", "event-123")

	controller.Details()

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	if controller.TplName != "details.tpl" {
		t.Fatalf(
			"expected details.tpl, got %s",
			controller.TplName,
		)
	}

	page, ok := controller.Data["Page"].(DetailsPageData)
	if !ok {
		t.Fatal("expected Page to contain DetailsPageData")
	}

	if page.Event == nil {
		t.Fatal("expected Event to be populated")
	}

	event, ok := page.Event.(models.Event)
	if !ok {
		t.Fatal("expected Page.Event to contain models.Event")
	}

	if event.ID != "event-123" {
		t.Fatalf(
			"expected event-123, got %q",
			event.ID,
		)
	}

	if event.Title != "Toronto Concert" {
		t.Fatalf(
			"expected Toronto Concert, got %q",
			event.Title,
		)
	}
}
func TestEventController_Details_ServiceError(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`event not found`))
		}),
	)
	defer server.Close()

	client := models.NewAPIClient(
		server.Client(),
		models.NewCache(),
		server.URL,
		"test-key",
	)

	service := services.NewEventService(client)

	controller, recorder := createController(
		http.MethodGet,
		"/events/missing-event",
	)

	controller.EventService = service
	controller.Ctx.Input.SetParam(":eventId", "missing-event")

	controller.Details()

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status 404, got %d",
			recorder.Code,
		)
	}

	if controller.TplName != "details.tpl" {
		t.Fatalf(
			"expected details.tpl, got %s",
			controller.TplName,
		)
	}

	page, ok := controller.Data["Page"].(DetailsPageData)
	if !ok {
		t.Fatal("expected Page to contain DetailsPageData")
	}

	if page.Error != "Event not found." {
		t.Fatalf(
			"expected event not found error, got %q",
			page.Error,
		)
	}
}
