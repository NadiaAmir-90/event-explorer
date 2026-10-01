package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"event-explorer/models"
	"event-explorer/services"

	beegoContext "github.com/beego/beego/v2/server/web/context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newLocationControllerTest(
	t *testing.T,
	method string,
	target string,
	service *services.LocationService,
) (*LocationController, *httptest.ResponseRecorder) {

	t.Helper()

	req := httptest.NewRequest(
		method,
		target,
		nil,
	)

	recorder := httptest.NewRecorder()

	ctx := beegoContext.NewContext()
	ctx.Reset(recorder, req)

	controller := &LocationController{
		LocationService: service,
	}

	controller.Ctx = ctx
	controller.Data = make(map[interface{}]interface{})

	return controller, recorder
}

// ==================================================
// Autocomplete
// ==================================================

func TestLocationController_Autocomplete_MissingInput(t *testing.T) {

	controller, recorder := newLocationControllerTest(
		t,
		http.MethodGet,
		"/api/locations/autocomplete?sessionToken=test-session",
		nil,
	)

	controller.Autocomplete()

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)

	assert.Contains(
		t,
		recorder.Body.String(),
		"input is required",
	)
}

func TestLocationController_Autocomplete_MissingSessionToken(t *testing.T) {

	controller, recorder := newLocationControllerTest(
		t,
		http.MethodGet,
		"/api/locations/autocomplete?input=Toronto",
		nil,
	)

	controller.Autocomplete()

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)

	assert.Contains(
		t,
		recorder.Body.String(),
		"sessionToken is required",
	)
}

func TestLocationController_Autocomplete_ServiceUnavailable(t *testing.T) {

	controller, recorder := newLocationControllerTest(
		t,
		http.MethodGet,
		"/api/locations/autocomplete?input=Toronto&sessionToken=test-session",
		nil,
	)

	controller.Autocomplete()

	assert.Equal(
		t,
		http.StatusInternalServerError,
		recorder.Code,
	)

	assert.Contains(
		t,
		recorder.Body.String(),
		"location service unavailable",
	)
}

func TestLocationController_Autocomplete_ServiceError(t *testing.T) {

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"Google Places API failed"}`))
		}),
	)

	defer server.Close()

	client := models.NewGooglePlacesClient(
		server.URL,
		"test-api-key",
		server.Client(),
	)

	service := services.NewLocationService(client)

	controller, recorder := newLocationControllerTest(
		t,
		http.MethodGet,
		"/api/locations/autocomplete?input=Toronto&sessionToken=test-session",
		service,
	)

	controller.Autocomplete()

	assert.Equal(
		t,
		http.StatusBadGateway,
		recorder.Code,
	)

	assert.Contains(
		t,
		recorder.Body.String(),
		"Google autocomplete returned status 500",
	)
}

func TestLocationController_Autocomplete_Success(t *testing.T) {

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusOK)

			_, err := w.Write([]byte(`{
				"suggestions": [
					{
						"placePrediction": {
							"placeId": "place-123",
							"text": {
								"text": "Toronto, Canada"
							}
						}
					}
				]
			}`))

			require.NoError(t, err)
		}),
	)

	defer server.Close()

	client := models.NewGooglePlacesClient(
		server.URL,
		"test-api-key",
		server.Client(),
	)

	service := services.NewLocationService(client)

	controller, recorder := newLocationControllerTest(
		t,
		http.MethodGet,
		"/api/locations/autocomplete?input=Toronto&sessionToken=test-session",
		service,
	)

	controller.Autocomplete()

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	assert.Contains(
		t,
		recorder.Body.String(),
		"place-123",
	)

	assert.Contains(
		t,
		recorder.Body.String(),
		"Toronto, Canada",
	)
}

// ==================================================
// Details
// ==================================================

func TestLocationController_Details_MissingPlaceID(t *testing.T) {

	controller, recorder := newLocationControllerTest(
		t,
		http.MethodGet,
		"/api/locations/?sessionToken=test-session",
		nil,
	)

	controller.Details()

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)

	assert.Contains(
		t,
		recorder.Body.String(),
		"placeId is required",
	)
}

func TestLocationController_Details_MissingSessionToken(t *testing.T) {

	controller, recorder := newLocationControllerTest(
		t,
		http.MethodGet,
		"/api/locations/place-123",
		nil,
	)

	// Beego route parameters are normally populated by the router.
	controller.Ctx.Input.SetParam(
		":placeId",
		"place-123",
	)

	controller.Details()

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)

	assert.Contains(
		t,
		recorder.Body.String(),
		"sessionToken is required",
	)
}

func TestLocationController_Details_ServiceUnavailable(t *testing.T) {

	controller, recorder := newLocationControllerTest(
		t,
		http.MethodGet,
		"/api/locations/place-123?sessionToken=test-session",
		nil,
	)

	controller.Ctx.Input.SetParam(
		":placeId",
		"place-123",
	)

	controller.Details()

	assert.Equal(
		t,
		http.StatusInternalServerError,
		recorder.Code,
	)

	assert.Contains(
		t,
		recorder.Body.String(),
		"location service unavailable",
	)
}

func TestLocationController_Details_ServiceError(t *testing.T) {

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"Google Places API failed"}`))
		}),
	)

	defer server.Close()

	client := models.NewGooglePlacesClient(
		server.URL,
		"test-api-key",
		server.Client(),
	)

	service := services.NewLocationService(client)

	controller, recorder := newLocationControllerTest(
		t,
		http.MethodGet,
		"/api/locations/place-123?sessionToken=test-session",
		service,
	)

	controller.Ctx.Input.SetParam(
		":placeId",
		"place-123",
	)

	controller.Details()

	assert.Equal(
		t,
		http.StatusBadGateway,
		recorder.Code,
	)

	assert.Contains(
		t,
		recorder.Body.String(),
		"Google place details returned status 500",
	)
}

func TestLocationController_Details_Success(t *testing.T) {

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusOK)

			_, err := w.Write([]byte(`{
				"addressComponents": [
					{
						"longText": "Toronto",
						"shortText": "Toronto",
						"types": [
							"locality"
						]
					},
					{
						"longText": "Canada",
						"shortText": "CA",
						"types": [
							"country"
						]
					}
				]
			}`))

			require.NoError(t, err)
		}),
	)

	defer server.Close()

	client := models.NewGooglePlacesClient(
		server.URL,
		"test-api-key",
		server.Client(),
	)

	service := services.NewLocationService(client)

	controller, recorder := newLocationControllerTest(
		t,
		http.MethodGet,
		"/api/locations/place-123?sessionToken=test-session",
		service,
	)

	controller.Ctx.Input.SetParam(
		":placeId",
		"place-123",
	)

	controller.Details()

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	assert.Contains(
		t,
		recorder.Body.String(),
		"Toronto",
	)

	assert.Contains(
		t,
		recorder.Body.String(),
		"CA",
	)
}

// Keep context import used explicitly for controller/service compatibility.
var _ = context.Background
