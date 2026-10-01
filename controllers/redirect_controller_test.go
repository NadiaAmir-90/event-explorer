package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"event-explorer/models"
	"event-explorer/services"

	beegoContext "github.com/beego/beego/v2/server/web/context"
)

// ============================================================
// isApprovedTicketURL TESTS
// ============================================================

func TestIsApprovedTicketURL_ValidTicketmasterCom(t *testing.T) {

	if !isApprovedTicketURL("https://www.ticketmaster.com/event/123") {
		t.Fatal("expected Ticketmaster URL to be approved")
	}
}

func TestIsApprovedTicketURL_ValidTicketmasterCA(t *testing.T) {

	if !isApprovedTicketURL("https://www.ticketmaster.ca/event/123") {
		t.Fatal("expected Ticketmaster Canada URL to be approved")
	}
}
func TestIsApprovedTicketURL_ValidTicketmasterCANonWWW(t *testing.T) {

	if !isApprovedTicketURL("https://ticketmaster.ca/event/123") {
		t.Fatal("expected ticketmaster.ca URL to be approved")
	}
}
func TestIsApprovedTicketURL_ValidWithoutWWW(t *testing.T) {

	if !isApprovedTicketURL("https://ticketmaster.com/event/123") {
		t.Fatal("expected ticketmaster.com URL to be approved")
	}
}

func TestIsApprovedTicketURL_InvalidScheme(t *testing.T) {

	if isApprovedTicketURL("http://www.ticketmaster.com/event/123") {
		t.Fatal("expected HTTP URL to be rejected")
	}
}

func TestIsApprovedTicketURL_InvalidDomain(t *testing.T) {

	if isApprovedTicketURL("https://example.com/event/123") {
		t.Fatal("expected non-Ticketmaster URL to be rejected")
	}
}

func TestIsApprovedTicketURL_SubdomainAttack(t *testing.T) {

	if isApprovedTicketURL("https://ticketmaster.com.example.com/event/123") {
		t.Fatal("expected fake Ticketmaster subdomain to be rejected")
	}
}

func TestIsApprovedTicketURL_InvalidURL(t *testing.T) {

	if isApprovedTicketURL("not-a-valid-url") {
		t.Fatal("expected invalid URL to be rejected")
	}
}

func TestIsApprovedTicketURL_EmptyURL(t *testing.T) {

	if isApprovedTicketURL("") {
		t.Fatal("expected empty URL to be rejected")
	}
}

// ============================================================
// Ticket CONTROLLER TESTS
// ============================================================

// Missing event ID should return 400.
func TestRedirectController_Ticket_MissingEventID(t *testing.T) {

	req := httptest.NewRequest(
		http.MethodGet,
		"/events//ticket",
		nil,
	)

	recorder := httptest.NewRecorder()

	ctx := NewControllerContext(recorder, req)

	controller := &RedirectController{}
	controller.Ctx = ctx

	controller.Ctx.Input.SetParam(":eventId", "")

	controller.Ticket()

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status 400, got %d",
			recorder.Code,
		)
	}
}

func TestRedirectController_Ticket_ServiceError(t *testing.T) {
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

	req := httptest.NewRequest(
		http.MethodGet,
		"/events/event-1/ticket",
		nil,
	)

	recorder := httptest.NewRecorder()
	ctx := NewControllerContext(recorder, req)

	controller := &RedirectController{
		EventService: service,
	}
	controller.Ctx = ctx

	controller.Ctx.Input.SetParam(":eventId", "event-1")

	controller.Ticket()

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status 404, got %d",
			recorder.Code,
		)
	}
}
func TestRedirectController_Ticket_EmptyTicketURL(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			_, _ = w.Write([]byte(`{
				"id": "event-1",
				"name": "Test Event",
				"url": "",
				"dates": {
					"start": {
						"localDate": "2026-10-10"
					}
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

	req := httptest.NewRequest(
		http.MethodGet,
		"/events/event-1/ticket",
		nil,
	)

	recorder := httptest.NewRecorder()
	ctx := NewControllerContext(recorder, req)

	controller := &RedirectController{
		EventService: service,
	}
	controller.Ctx = ctx

	controller.Ctx.Input.SetParam(":eventId", "event-1")

	controller.Ticket()

	if recorder.Code != http.StatusNotFound {
		t.Fatalf(
			"expected status 404, got %d",
			recorder.Code,
		)
	}
}
func TestRedirectController_Ticket_UnapprovedURL(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			_, _ = w.Write([]byte(`{
				"id": "event-1",
				"name": "Test Event",
				"url": "https://example.com/event/123",
				"dates": {
					"start": {
						"localDate": "2026-10-10"
					}
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

	req := httptest.NewRequest(
		http.MethodGet,
		"/events/event-1/ticket",
		nil,
	)

	recorder := httptest.NewRecorder()
	ctx := NewControllerContext(recorder, req)

	controller := &RedirectController{
		EventService: service,
	}
	controller.Ctx = ctx

	controller.Ctx.Input.SetParam(":eventId", "event-1")

	controller.Ticket()

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf(
			"expected status 502, got %d",
			recorder.Code,
		)
	}
}
func TestRedirectController_Ticket_Success(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			_, _ = w.Write([]byte(`{
				"id": "event-1",
				"name": "Test Event",
				"url": "https://www.ticketmaster.com/event/123",
				"dates": {
					"start": {
						"localDate": "2026-10-10"
					}
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

	req := httptest.NewRequest(
		http.MethodGet,
		"/events/event-1/ticket",
		nil,
	)

	recorder := httptest.NewRecorder()
	ctx := NewControllerContext(recorder, req)

	controller := &RedirectController{
		EventService: service,
	}
	controller.Ctx = ctx

	controller.Ctx.Input.SetParam(":eventId", "event-1")

	controller.Ticket()

	if recorder.Code != http.StatusFound {
		t.Fatalf(
			"expected status 302, got %d",
			recorder.Code,
		)
	}

	location := recorder.Header().Get("Location")

	if location != "https://www.ticketmaster.com/event/123" {
		t.Fatalf(
			"expected redirect location %q, got %q",
			"https://www.ticketmaster.com/event/123",
			location,
		)
	}
}

// EventService is nil.
// The current controller does not check for nil EventService,
// so this test is intentionally NOT included.
//
// Once a nil-service check is added to the controller,
// it should be tested here.

// ============================================================
// Helper
// ============================================================

func NewControllerContext(
	recorder *httptest.ResponseRecorder,
	req *http.Request,
) *beegoContext.Context {

	ctx := beegoContext.NewContext()

	ctx.Reset(
		recorder,
		req,
	)

	return ctx
}
