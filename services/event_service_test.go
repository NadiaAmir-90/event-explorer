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

func TestEventService_ListEvents_Success(t *testing.T) {

	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(http.StatusOK)

			_, _ = w.Write([]byte(`{
				"_embedded": {
					"events": [
						{
							"id": "event-1",
							"name": "Test Music Event",
							"classifications": [
								{
									"segment": {
										"name": "Music"
									}
								}
							],
							"_embedded": {
								"venues": [
									{
										"name": "Test Venue",
										"city": {
											"name": "Toronto"
										}
									}
								]
							},
							"dates": {
								"start": {
									"localDate": "2026-10-10"
								}
							},
							"url": "https://www.ticketmaster.com/test-event"
						}
					]
				}
			}`))
		}),
	)

	defer server.Close()

	// Create a cache for the API client.
	cache := models.NewCache()

	// Create API client using the mock HTTP server.
	client := models.NewAPIClient(
		server.Client(),
		cache,
		server.URL,
		"test-key",
	)

	service := NewEventService(client)

	events, err := service.ListEvents(
		context.Background(),
		"Toronto",
		"CA",
	)

	require.NoError(t, err)
	require.NotNil(t, events)

	assert.NotEmpty(t, events)

	// Verify Music events were returned.
	assert.NotEmpty(t, events["Music"])

	assert.Equal(
		t,
		"event-1",
		events["Music"][0].ID,
	)

	assert.Equal(
		t,
		"Test Music Event",
		events["Music"][0].Title,
	)
}
func TestEventService_GetEventDetails_Success(t *testing.T) {
	server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			_, _ = w.Write([]byte(`{
				"id": "event-1",
				"name": "Test Music Event",
				"description": "A test event",
				"url": "https://www.ticketmaster.com/test-event",
				"dates": {
					"start": {
						"localDate": "2026-10-10"
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
							"name": "Test Venue",
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

	service := NewEventService(client)

	event, err := service.GetEventDetails(
		context.Background(),
		"event-1",
	)

	require.NoError(t, err)

	assert.Equal(t, "event-1", event.ID)
	assert.Equal(t, "Test Music Event", event.Title)
	assert.Equal(t, "Music / Rock", event.Category)
	assert.Equal(t, "Test Venue, Toronto, ON", event.Location)
	assert.Equal(t, "2026-10-10", event.Date)
	assert.Equal(t, "A test event", event.Description)
	assert.Equal(
		t,
		"https://www.ticketmaster.com/test-event",
		event.TicketURL,
	)
}
func TestEventService_GetEventDetails_Error(t *testing.T) {
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

	service := NewEventService(client)

	event, err := service.GetEventDetails(
		context.Background(),
		"missing-event",
	)

	require.Error(t, err)
	assert.Empty(t, event.ID)
}
