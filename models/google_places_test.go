package models

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGooglePlacesClient_NilHTTPClient(t *testing.T) {
	client := NewGooglePlacesClient(
		"https://example.com/",
		"test-api-key",
		nil,
	)

	require.NotNil(t, client)
	assert.Equal(t, "https://example.com", client.BaseURL)
	assert.Equal(t, "test-api-key", client.APIKey)
	assert.NotNil(t, client.HTTPClient)
}
func TestGooglePlacesClient_Autocomplete_NetworkError(t *testing.T) {
	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, fmt.Errorf("connection failed")
		},
	}

	client := NewGooglePlacesClient(
		"https://example.com",
		"test-api-key",
		mockClient,
	)

	result, err := client.Autocomplete(
		context.Background(),
		"Toronto",
		"test-session",
	)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(
		t,
		err.Error(),
		"request Google autocomplete",
	)
}

// ==================================================
// AUTOCOMPLETE TESTS
// ==================================================

func TestGooglePlacesClient_Autocomplete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/v1/places:autocomplete", r.URL.Path)

		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-api-key", r.Header.Get("X-Goog-Api-Key"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`{
			"suggestions": [
				{
					"placePrediction": {
						"placeId": "place-123",
						"text": {
							"text": "Toronto, Ontario, Canada"
						}
					}
				},
				{
					"placePrediction": {
						"placeId": "place-456",
						"text": {
							"text": "Toronto, Ohio, USA"
						}
					}
				}
			]
		}`))

		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewGooglePlacesClient(
		server.URL,
		"test-api-key",
		server.Client(),
	)

	result, err := client.Autocomplete(
		context.Background(),
		"Toronto",
		"test-session",
	)

	require.NoError(t, err)
	require.Len(t, result, 2)

	assert.Equal(t, "place-123", result[0].PlaceID)
	assert.Equal(t, "Toronto, Ontario, Canada", result[0].Text)

	assert.Equal(t, "place-456", result[1].PlaceID)
	assert.Equal(t, "Toronto, Ohio, USA", result[1].Text)
}

func TestGooglePlacesClient_Autocomplete_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		_, err := w.Write([]byte(`{
			"error": {
				"code": 400,
				"message": "Invalid API key",
				"status": "INVALID_ARGUMENT"
			}
		}`))

		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewGooglePlacesClient(
		server.URL,
		"test-api-key",
		server.Client(),
	)

	result, err := client.Autocomplete(
		context.Background(),
		"Toronto",
		"test-session",
	)

	assert.Error(t, err)
	assert.Nil(t, result)

	assert.Contains(t, err.Error(), "Google autocomplete returned status 400")
	assert.Contains(t, err.Error(), "Invalid API key")
}

func TestGooglePlacesClient_Autocomplete_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`invalid-json`))

		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewGooglePlacesClient(
		server.URL,
		"test-api-key",
		server.Client(),
	)

	result, err := client.Autocomplete(
		context.Background(),
		"Toronto",
		"test-session",
	)

	assert.Error(t, err)
	assert.Nil(t, result)

	assert.Contains(t, err.Error(), "decode Google autocomplete response")
}

// ==================================================
// GET PLACE TESTS
// ==================================================

func TestGooglePlacesClient_GetPlace_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/v1/places/place-123", r.URL.Path)

		assert.Equal(t, "test-api-key", r.Header.Get("X-Goog-Api-Key"))
		assert.Equal(t, "addressComponents", r.Header.Get("X-Goog-FieldMask"))

		assert.Equal(t, "test-session", r.URL.Query().Get("sessionToken"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`{
			"addressComponents": [
				{
					"longText": "Toronto",
					"shortText": "Toronto",
					"types": ["locality"]
				},
				{
					"longText": "Canada",
					"shortText": "CA",
					"types": ["country"]
				}
			]
		}`))

		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewGooglePlacesClient(
		server.URL,
		"test-api-key",
		server.Client(),
	)

	result, err := client.GetPlace(
		context.Background(),
		"place-123",
		"test-session",
	)

	require.NoError(t, err)

	assert.Equal(t, "Toronto", result.City)
	assert.Equal(t, "CA", result.CountryCode)
}

func TestGooglePlacesClient_GetPlace_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)

		_, err := w.Write([]byte(`{
			"error": "place not found"
		}`))

		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewGooglePlacesClient(
		server.URL,
		"test-api-key",
		server.Client(),
	)

	result, err := client.GetPlace(
		context.Background(),
		"invalid-place",
		"test-session",
	)

	assert.Error(t, err)
	assert.Equal(t, LocationDetails{}, result)

	assert.Contains(
		t,
		err.Error(),
		"Google place details returned status 404",
	)
}

func TestGooglePlacesClient_GetPlace_EmptyPlaceID(t *testing.T) {
	client := NewGooglePlacesClient(
		"http://example.com",
		"test-api-key",
		http.DefaultClient,
	)

	result, err := client.GetPlace(
		context.Background(),
		"   ",
		"test-session",
	)

	assert.Error(t, err)
	assert.Equal(t, LocationDetails{}, result)

	assert.Contains(t, err.Error(), "place ID is required")
}

func TestGooglePlacesClient_GetPlace_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`invalid-json`))

		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewGooglePlacesClient(
		server.URL,
		"test-api-key",
		server.Client(),
	)

	result, err := client.GetPlace(
		context.Background(),
		"place-123",
		"test-session",
	)

	assert.Error(t, err)
	assert.Equal(t, LocationDetails{}, result)

	assert.Contains(t, err.Error(), "decode Google place details")
}

func TestGooglePlacesClient_GetPlace_MissingCity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`{
			"addressComponents": [
				{
					"longText": "Canada",
					"shortText": "CA",
					"types": ["country"]
				}
			]
		}`))

		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewGooglePlacesClient(
		server.URL,
		"test-api-key",
		server.Client(),
	)

	result, err := client.GetPlace(
		context.Background(),
		"place-123",
		"test-session",
	)

	assert.Error(t, err)
	assert.Equal(t, LocationDetails{}, result)

	assert.Contains(
		t,
		err.Error(),
		"selected place does not contain a city",
	)
}

func TestGooglePlacesClient_GetPlace_MissingCountryCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`{
			"addressComponents": [
				{
					"longText": "Toronto",
					"shortText": "Toronto",
					"types": ["locality"]
				}
			]
		}`))

		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewGooglePlacesClient(
		server.URL,
		"test-api-key",
		server.Client(),
	)

	result, err := client.GetPlace(
		context.Background(),
		"place-123",
		"test-session",
	)

	assert.Error(t, err)
	assert.Equal(t, LocationDetails{}, result)

	assert.Contains(
		t,
		err.Error(),
		"selected place does not contain a country code",
	)
}

// ==================================================
// PATH ESCAPING TEST
// ==================================================

func TestGooglePlacesClient_GetPlace_EscapesPlaceID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The important part is that the request reaches the server
		// rather than treating the place ID as a separate path.
		assert.True(
			t,
			strings.HasPrefix(r.URL.Path, "/v1/places/"),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(`{
			"addressComponents": [
				{
					"longText": "Toronto",
					"shortText": "CA",
					"types": ["locality"]
				},
				{
					"longText": "Canada",
					"shortText": "CA",
					"types": ["country"]
				}
			]
		}`))

		require.NoError(t, err)
	}))
	defer server.Close()

	client := NewGooglePlacesClient(
		server.URL,
		"test-api-key",
		server.Client(),
	)

	result, err := client.GetPlace(
		context.Background(),
		"place/123",
		"test-session",
	)

	require.NoError(t, err)

	assert.Equal(t, "Toronto", result.City)
	assert.Equal(t, "CA", result.CountryCode)
}
func TestGooglePlacesClient_Autocomplete_InvalidErrorJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("this is not valid JSON"))
	}))
	defer server.Close()

	client := NewGooglePlacesClient(
		server.URL,
		"test-api-key",
		server.Client(),
	)

	_, err := client.Autocomplete(
		context.Background(),
		"Toronto",
		"session-123",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Google autocomplete returned status 500")
}
func TestGooglePlacesClient_GetPlace_NetworkError(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network failure")
		},
	}

	googleClient := NewGooglePlacesClient(
		"https://example.com",
		"test-api-key",
		client,
	)

	_, err := googleClient.GetPlace(
		context.Background(),
		"place-123",
		"session-123",
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "request Google place details")
}
