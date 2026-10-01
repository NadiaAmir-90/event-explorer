package services

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"event-explorer/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocationService_Autocomplete_Success(t *testing.T) {

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

	service := NewLocationService(client)

	result, err := service.Autocomplete(
		context.Background(),
		"Toronto",
		"test-session",
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Len(t, result, 1)

	assert.Equal(
		t,
		"place-123",
		result[0].PlaceID,
	)

	assert.Equal(
		t,
		"Toronto, Canada",
		result[0].Text,
	)
}

func TestLocationService_Autocomplete_APIError(t *testing.T) {

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.WriteHeader(http.StatusInternalServerError)

			_, err := w.Write([]byte(`{
				"error": "Google Places API failed"
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

	service := NewLocationService(client)

	result, err := service.Autocomplete(
		context.Background(),
		"Toronto",
		"test-session",
	)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestLocationService_GetPlace_Success(t *testing.T) {

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

	service := NewLocationService(client)

	result, err := service.GetPlace(
		context.Background(),
		"place-123",
		"test-session",
	)

	require.NoError(t, err)

	assert.Equal(
		t,
		"Toronto",
		result.City,
	)

	assert.Equal(
		t,
		"CA",
		result.CountryCode,
	)
}

func TestLocationService_GetPlace_APIError(t *testing.T) {

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.WriteHeader(http.StatusInternalServerError)

			_, err := w.Write([]byte(`{
				"error": "Google Places API failed"
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

	service := NewLocationService(client)

	result, err := service.GetPlace(
		context.Background(),
		"invalid-place",
		"test-session",
	)

	assert.Error(t, err)

	assert.Equal(
		t,
		models.LocationDetails{},
		result,
	)
}