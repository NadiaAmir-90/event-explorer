package models

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

const sampleJSON = `{
	"_embedded": {
		"events": [
			{
				"id": "1kAYvP7_GA2u_jk",
				"name": "Oasis Live '27",
				"pleaseNote": "Face Value Exchange only.",
				"dates": {
					"start": {
						"localDate": "2027-08-24"
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
							"name": "Allegiant Stadium",
							"city": {
								"name": "Las Vegas"
							},
							"state": {
								"stateCode": "NV"
							}
						}
					]
				}
			}
		]
	}
}`

func newResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func TestNewAPIClient(t *testing.T) {
	client := &mockHTTPClient{}
	cache := NewCache()

	apiClient := NewAPIClient(
		client,
		cache,
		"https://example.com/discovery/v2",
		"test-api-key",
	)

	if apiClient == nil {
		t.Fatal("expected API client, got nil")
	}

	if apiClient.client != client {
		t.Fatal("expected provided HTTP client to be stored")
	}

	if apiClient.cache != cache {
		t.Fatal("expected provided cache to be stored")
	}

	if apiClient.baseURL != "https://example.com/discovery/v2" {
		t.Fatalf(
			"unexpected base URL: %s",
			apiClient.baseURL,
		)
	}

	if apiClient.apiKey != "test-api-key" {
		t.Fatalf(
			"unexpected API key: %s",
			apiClient.apiKey,
		)
	}
}

func TestNewAPIClientWithConfig(t *testing.T) {
	client := &mockHTTPClient{}
	cache := NewCache()

	apiClient := NewAPIClientWithConfig(
		client,
		cache,
		"https://app.ticketmaster.com/discovery/v2",
		"test-consumer-key",
	)

	if apiClient == nil {
		t.Fatal("expected API client, got nil")
	}

	if apiClient.baseURL != "https://app.ticketmaster.com/discovery/v2" {
		t.Fatalf(
			"unexpected base URL: %s",
			apiClient.baseURL,
		)
	}

	if apiClient.apiKey != "test-consumer-key" {
		t.Fatalf(
			"unexpected API key: %s",
			apiClient.apiKey,
		)
	}
}

func TestNewAPIClient_NilHTTPClient(t *testing.T) {
	apiClient := NewAPIClient(
		nil,
		NewCache(),
		"https://example.com/discovery/v2",
		"test-key",
	)

	if apiClient == nil {
		t.Fatal("expected API client, got nil")
	}

	if apiClient.client == nil {
		t.Fatal("expected default HTTP client")
	}
}
func TestNewAPIClient_NilCache(t *testing.T) {
	client := &mockHTTPClient{}

	apiClient := NewAPIClient(
		client,
		nil,
		"https://example.com/discovery/v2/",
		"test-key",
	)

	if apiClient == nil {
		t.Fatal("expected API client, got nil")
	}

	if apiClient.client != client {
		t.Fatal("expected provided HTTP client to be stored")
	}

	if apiClient.cache == nil {
		t.Fatal("expected default cache to be created")
	}

	if apiClient.baseURL != "https://example.com/discovery/v2" {
		t.Fatalf(
			"unexpected base URL: %s",
			apiClient.baseURL,
		)
	}

	if apiClient.apiKey != "test-key" {
		t.Fatalf(
			"unexpected API key: %s",
			apiClient.apiKey,
		)
	}
}

func TestBuildEventsURL(t *testing.T) {
	apiClient := NewAPIClient(
		nil,
		NewCache(),
		"https://app.ticketmaster.com/discovery/v2",
		"test-api-key",
	)

	requestURL := apiClient.BuildEventsURL(
		"Las Vegas",
		"US",
	)

	if !strings.Contains(
		requestURL,
		"https://app.ticketmaster.com/discovery/v2/events.json",
	) {
		t.Fatalf(
			"unexpected base URL: %s",
			requestURL,
		)
	}

	if !strings.Contains(
		requestURL,
		"apikey=test-api-key",
	) {
		t.Fatalf(
			"API key missing from URL: %s",
			requestURL,
		)
	}

	if !strings.Contains(
		requestURL,
		"city=Las+Vegas",
	) {
		t.Fatalf(
			"city missing from URL: %s",
			requestURL,
		)
	}

	if !strings.Contains(
		requestURL,
		"countryCode=US",
	) {
		t.Fatalf(
			"countryCode missing from URL: %s",
			requestURL,
		)
	}
}

func TestBuildEventsURL_WithoutFilters(t *testing.T) {
	apiClient := NewAPIClient(
		nil,
		NewCache(),
		"https://app.ticketmaster.com/discovery/v2",
		"test-api-key",
	)

	requestURL := apiClient.BuildEventsURL("", "")

	expected := "https://app.ticketmaster.com/discovery/v2/events.json?apikey=test-api-key"

	if requestURL != expected {
		t.Fatalf(
			"expected:\n%s\ngot:\n%s",
			expected,
			requestURL,
		)
	}
}

func TestFetchEventsConcurrent_Success(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return newResponse(
				http.StatusOK,
				sampleJSON,
			), nil
		},
	}

	cache := NewCache()

	apiClient := NewAPIClient(
		client,
		cache,
		"https://example.com",
		"test-key",
	)

	urls := []string{
		"https://example.com/events/1",
		"https://example.com/events/2",
	}

	events, err := apiClient.FetchEventsConcurrent(urls)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(events) != 2 {
		t.Fatalf(
			"expected 2 events, got %d",
			len(events),
		)
	}

	expected := Event{
		ID:          "1kAYvP7_GA2u_jk",
		Title:       "Oasis Live '27",
		Category:    "Music / Rock",
		Location:    "Allegiant Stadium, Las Vegas, NV",
		Date:        "2027-08-24",
		Description: "Face Value Exchange only.",
	}

	if events[0] != expected {
		t.Errorf(
			"unexpected event\nexpected: %+v\ngot: %+v",
			expected,
			events[0],
		)
	}
}

func TestFetchEventsConcurrent_EmptyURLs(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			t.Fatal("HTTP request should not have been made")
			return nil, nil
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com",
		"test-key",
	)

	events, err := apiClient.FetchEventsConcurrent(nil)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(events) != 0 {
		t.Fatalf(
			"expected zero events, got %d",
			len(events),
		)
	}
}

func TestFetchEventsConcurrent_HTTPError(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network failure")
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com",
		"test-key",
	)

	_, err := apiClient.FetchEventsConcurrent(
		[]string{
			"https://example.com/events",
		},
	)

	if err == nil {
		t.Fatal("expected an error")
	}
}

func TestFetchEventsConcurrent_StatusError(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return newResponse(
				http.StatusInternalServerError,
				`{"error":"server error"}`,
			), nil
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com",
		"test-key",
	)

	_, err := apiClient.FetchEventsConcurrent(
		[]string{
			"https://example.com/events",
		},
	)

	if err == nil {
		t.Fatal("expected HTTP status error")
	}
}

func TestFetchEventsConcurrent_InvalidJSON(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return newResponse(
				http.StatusOK,
				`this is not valid json`,
			), nil
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com",
		"test-key",
	)

	_, err := apiClient.FetchEventsConcurrent(
		[]string{
			"https://example.com/events",
		},
	)

	if err == nil {
		t.Fatal("expected JSON decoding error")
	}
}

func TestFetchEventsConcurrent_CacheHitAndClear(t *testing.T) {
	var requestCount int32

	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&requestCount, 1)

			return newResponse(
				http.StatusOK,
				sampleJSON,
			), nil
		},
	}

	cache := NewCache()

	apiClient := NewAPIClient(
		client,
		cache,
		"https://example.com",
		"test-key",
	)

	urls := []string{
		"https://example.com/events",
	}

	// First request:
	// cache miss -> HTTP request -> cache populated.
	_, err := apiClient.FetchEventsConcurrent(urls)

	if err != nil {
		t.Fatalf(
			"first request failed: %v",
			err,
		)
	}

	if atomic.LoadInt32(&requestCount) != 1 {
		t.Fatalf(
			"expected 1 HTTP request, got %d",
			requestCount,
		)
	}

	// Second request:
	// cache hit -> no HTTP request -> value stays in cache.
	_, err = apiClient.FetchEventsConcurrent(urls)

	if err != nil {
		t.Fatalf(
			"second request failed: %v",
			err,
		)
	}

	if atomic.LoadInt32(&requestCount) != 1 {
		t.Fatalf(
			"expected cache hit to avoid HTTP request; got %d requests",
			requestCount,
		)
	}

	// Explicitly clear the cache.
	cache.Clear()

	// Third request:
	// cache was cleared -> cache miss -> HTTP request -> cache populated.
	_, err = apiClient.FetchEventsConcurrent(urls)

	if err != nil {
		t.Fatalf(
			"third request failed: %v",
			err,
		)
	}

	if atomic.LoadInt32(&requestCount) != 2 {
		t.Fatalf(
			"expected 2 HTTP requests after Clear(), got %d",
			requestCount,
		)
	}

	// Fourth request:
	// cache hit again -> no HTTP request.
	_, err = apiClient.FetchEventsConcurrent(urls)

	if err != nil {
		t.Fatalf(
			"fourth request failed: %v",
			err,
		)
	}

	if atomic.LoadInt32(&requestCount) != 2 {
		t.Fatalf(
			"expected cache hit to avoid another HTTP request; got %d",
			requestCount,
		)
	}
}

func TestFetchEventsConcurrent_MultipleURLsAreConcurrent(t *testing.T) {
	var activeRequests int32
	var maximumConcurrent int32

	var mu sync.Mutex

	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			current := atomic.AddInt32(
				&activeRequests,
				1,
			)

			mu.Lock()

			if current > maximumConcurrent {
				maximumConcurrent = current
			}

			mu.Unlock()

			time.Sleep(50 * time.Millisecond)

			atomic.AddInt32(
				&activeRequests,
				-1,
			)

			return newResponse(
				http.StatusOK,
				sampleJSON,
			), nil
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com",
		"test-key",
	)

	urls := []string{
		"https://example.com/events/1",
		"https://example.com/events/2",
		"https://example.com/events/3",
	}

	events, err := apiClient.FetchEventsConcurrent(urls)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if len(events) != 3 {
		t.Fatalf(
			"expected 3 events, got %d",
			len(events),
		)
	}

	if maximumConcurrent < 2 {
		t.Fatalf(
			"expected concurrent HTTP requests, maximum concurrency was %d",
			maximumConcurrent,
		)
	}
}

func TestFetchEventsConcurrent_PartialFailure(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			if req.URL.String() == "https://example.com/fail" {
				return nil, errors.New("request failed")
			}

			return newResponse(
				http.StatusOK,
				sampleJSON,
			), nil
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com",
		"test-key",
	)

	urls := []string{
		"https://example.com/success",
		"https://example.com/fail",
	}

	events, err := apiClient.FetchEventsConcurrent(urls)

	if err == nil {
		t.Fatal("expected an error")
	}

	// The successful request should still return its event.
	if len(events) != 1 {
		t.Fatalf(
			"expected 1 successful event, got %d",
			len(events),
		)
	}
}
func TestBuildSearchURL(t *testing.T) {
	apiClient := NewAPIClient(
		nil,
		NewCache(),
		"https://app.ticketmaster.com/discovery/v2",
		"test-api-key",
	)

	requestURL := apiClient.BuildSearchURL(
		"Oasis",
		"Music",
	)

	if !strings.Contains(
		requestURL,
		"https://app.ticketmaster.com/discovery/v2/events.json",
	) {
		t.Fatalf("unexpected URL: %s", requestURL)
	}

	if !strings.Contains(
		requestURL,
		"apikey=test-api-key",
	) {
		t.Fatalf("API key missing: %s", requestURL)
	}

	if !strings.Contains(
		requestURL,
		"keyword=Oasis",
	) {
		t.Fatalf("keyword missing: %s", requestURL)
	}

	if !strings.Contains(
		requestURL,
		"classificationName=Music",
	) {
		t.Fatalf("classification missing: %s", requestURL)
	}
}
func TestBuildSearchURL_WithoutFilters(t *testing.T) {
	apiClient := NewAPIClient(
		nil,
		NewCache(),
		"https://app.ticketmaster.com/discovery/v2",
		"test-api-key",
	)

	requestURL := apiClient.BuildSearchURL("", "")

	expected := "https://app.ticketmaster.com/discovery/v2/events.json?apikey=test-api-key"

	if requestURL != expected {
		t.Fatalf(
			"expected:\n%s\ngot:\n%s",
			expected,
			requestURL,
		)
	}
}
func TestFetchEventsByLocation_Success(t *testing.T) {
	var mu sync.Mutex
	var requestedURLs []string

	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			mu.Lock()
			requestedURLs = append(requestedURLs, req.URL.String())
			mu.Unlock()

			return newResponse(
				http.StatusOK,
				sampleJSON,
			), nil
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com/discovery/v2",
		"test-key",
	)

	events, err := apiClient.FetchEventsByLocation(
		context.Background(),
		"Toronto",
		"CA",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events["Music"]) != 1 {
		t.Fatalf(
			"expected 1 Music event, got %d",
			len(events["Music"]),
		)
	}

	if len(events["Sports"]) != 1 {
		t.Fatalf(
			"expected 1 Sports event, got %d",
			len(events["Sports"]),
		)
	}

	if len(requestedURLs) != 2 {
		t.Fatalf(
			"expected 2 HTTP requests, got %d",
			len(requestedURLs),
		)
	}

	for _, requestURL := range requestedURLs {
		if !strings.Contains(requestURL, "city=Toronto") {
			t.Errorf(
				"city missing from request URL: %s",
				requestURL,
			)
		}

		if !strings.Contains(requestURL, "countryCode=CA") {
			t.Errorf(
				"countryCode missing from request URL: %s",
				requestURL,
			)
		}

		if !strings.Contains(
			requestURL,
			"classificationName=",
		) {
			t.Errorf(
				"classificationName missing from request URL: %s",
				requestURL,
			)
		}

		if !strings.Contains(requestURL, "size=6") {
			t.Errorf(
				"size=6 missing from request URL: %s",
				requestURL,
			)
		}
	}
}
func TestFetchEventsByLocation_AllCategoriesFail(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network failure")
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com/discovery/v2",
		"test-key",
	)

	events, err := apiClient.FetchEventsByLocation(
		context.Background(),
		"Toronto",
		"CA",
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(
		err.Error(),
		"all category requests failed",
	) {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if events == nil {
		t.Fatal("expected response map")
	}

	if _, ok := events["Music"]; !ok {
		t.Fatal("expected Music category")
	}

	if _, ok := events["Sports"]; !ok {
		t.Fatal("expected Sports category")
	}
}

func TestFetchEventsByLocation_PartialFailure(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			classification := req.URL.Query().Get("classificationName")

			if classification == "Sports" {
				return nil, errors.New("sports request failed")
			}

			return newResponse(
				http.StatusOK,
				sampleJSON,
			), nil
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com/discovery/v2",
		"test-key",
	)

	events, err := apiClient.FetchEventsByLocation(
		context.Background(),
		"Toronto",
		"CA",
	)

	if err != nil {
		t.Fatalf(
			"expected no overall error when one category succeeds: %v",
			err,
		)
	}

	if len(events["Music"]) != 1 {
		t.Fatalf(
			"expected Music event, got %d",
			len(events["Music"]),
		)
	}

	if len(events["Sports"]) != 0 {
		t.Fatalf(
			"expected no Sports events, got %d",
			len(events["Sports"]),
		)
	}
}
func TestFetch_Success(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return newResponse(
				http.StatusOK,
				sampleJSON,
			), nil
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com",
		"test-key",
	)

	events, err := apiClient.fetch(
		"https://example.com/events",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf(
			"expected 1 event, got %d",
			len(events),
		)
	}
}

func TestFetch_InvalidURL(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			t.Fatal("HTTP client should not be called")
			return nil, nil
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com",
		"test-key",
	)

	_, err := apiClient.fetch("http://[::1")

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "create request") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestGetEventDetails_Success(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			if req.URL.Path != "/events/event-123" {
				t.Fatalf(
					"unexpected path: %s",
					req.URL.Path,
				)
			}

			if req.URL.Query().Get("apikey") != "test-key" {
				t.Fatalf("API key missing")
			}

			body := `{
                "id": "event-123",
                "name": "Oasis Live '27",
                "description": "Live concert",
                "url": "https://www.ticketmaster.com/event-123",
                "dates": {
                    "start": {
                        "localDate": "2027-08-24",
                        "localTime": "20:00:00"
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
                            "name": "Allegiant Stadium",
                            "city": {
                                "name": "Las Vegas"
                            },
                            "state": {
                                "stateCode": "NV"
                            }
                        }
                    ]
                }
            }`

			return newResponse(
				http.StatusOK,
				body,
			), nil
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com",
		"test-key",
	)

	event, err := apiClient.GetEventDetails(
		context.Background(),
		"event-123",
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event.ID != "event-123" {
		t.Fatalf(
			"expected event ID event-123, got %s",
			event.ID,
		)
	}

	if event.Title != "Oasis Live '27" {
		t.Fatalf(
			"unexpected title: %s",
			event.Title,
		)
	}

	if event.Category != "Music / Rock" {
		t.Fatalf(
			"unexpected category: %s",
			event.Category,
		)
	}

	if event.Location != "Allegiant Stadium, Las Vegas, NV" {
		t.Fatalf(
			"unexpected location: %s",
			event.Location,
		)
	}

	if event.Date != "2027-08-24" {
		t.Fatalf(
			"unexpected date: %s",
			event.Date,
		)
	}

	if event.Description != "Live concert" {
		t.Fatalf(
			"unexpected description: %s",
			event.Description,
		)
	}
}
func TestGetEventDetails_EmptyID(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			t.Fatal("HTTP request should not have been made")
			return nil, nil
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com",
		"test-key",
	)

	_, err := apiClient.GetEventDetails(
		context.Background(),
		"   ",
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(
		err.Error(),
		"event ID is required",
	) {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}
func TestGetEventDetails_HTTPError(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return newResponse(
				http.StatusNotFound,
				`{"error":"event not found"}`,
			), nil
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com",
		"test-key",
	)

	_, err := apiClient.GetEventDetails(
		context.Background(),
		"event-123",
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(
		err.Error(),
		"Ticketmaster event returned status 404",
	) {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}
func TestGetEventDetails_NetworkError(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return nil, errors.New("network failure")
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com",
		"test-key",
	)

	_, err := apiClient.GetEventDetails(
		context.Background(),
		"event-123",
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(
		err.Error(),
		"request event details",
	) {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}
func TestGetEventDetails_InvalidJSON(t *testing.T) {
	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return newResponse(
				http.StatusOK,
				`this is invalid json`,
			), nil
		},
	}

	apiClient := NewAPIClient(
		client,
		NewCache(),
		"https://example.com",
		"test-key",
	)

	_, err := apiClient.GetEventDetails(
		context.Background(),
		"event-123",
	)

	if err == nil {
		t.Fatal("expected JSON decoding error")
	}

	if !strings.Contains(
		err.Error(),
		"decode event response",
	) {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}
}
func TestFetchWithCacheContext_NilCachedValue(t *testing.T) {
	cache := NewCache()

	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			t.Fatal("HTTP client should not be called")
			return nil, nil
		},
	}

	apiClient := NewAPIClient(
		client,
		cache,
		"https://example.com",
		"test-key",
	)

	requestURL := "https://example.com/events.json?test=nil"

	cache.mu.Lock()
	cache.items[requestURL] = nil
	cache.mu.Unlock()

	_, err := apiClient.fetchWithCacheContext(
		context.Background(),
		requestURL,
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "no events returned") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestFetchWithCacheContext_InvalidCachedValue(t *testing.T) {
	cache := NewCache()

	client := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			t.Fatal("HTTP client should not be called")
			return nil, nil
		},
	}

	apiClient := NewAPIClient(
		client,
		cache,
		"https://example.com",
		"test-key",
	)

	requestURL := "https://example.com/events.json?test=invalid-type"

	cache.mu.Lock()
	cache.items[requestURL] = "not a []Event"
	cache.mu.Unlock()

	_, err := apiClient.fetchWithCacheContext(
		context.Background(),
		requestURL,
	)

	if err == nil {
		t.Fatal("expected error")
	}

	if !strings.Contains(err.Error(), "invalid cached value") {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestTicketmasterEvent_ToEvent_EmptyFields(t *testing.T) {
	raw := ticketmasterEvent{
		ID:              "event-empty",
		Name:            "Empty Event",
		Description:     "",
		PleaseNote:      "",
		Images:          nil,
		Embedded:        nil,
		Classifications: nil,
	}

	event := raw.toEvent()

	if event.ID != "event-empty" {
		t.Fatalf("expected ID event-empty, got %q", event.ID)
	}

	if event.Title != "Empty Event" {
		t.Fatalf("expected title Empty Event, got %q", event.Title)
	}

	if event.Location != "" {
		t.Fatalf("expected empty location, got %q", event.Location)
	}

	if event.Category != "" {
		t.Fatalf("expected empty category, got %q", event.Category)
	}

	if event.ImageURL != "" {
		t.Fatalf("expected empty image URL, got %q", event.ImageURL)
	}

	if event.Description != "" {
		t.Fatalf("expected empty description, got %q", event.Description)
	}
}


func TestTicketmasterEvent_ToEvent_WithImage(t *testing.T) {
	raw := ticketmasterEvent{
		ID:   "event-123",
		Name: "Test Event",
		Images: []struct {
			URL string `json:"url"`
		}{
			{
				URL: "https://example.com/event.jpg",
			},
		},
	}

	event := raw.toEvent()

	if event.ImageURL != "https://example.com/event.jpg" {
		t.Fatalf(
			"expected image URL %q, got %q",
			"https://example.com/event.jpg",
			event.ImageURL,
		)
	}
}
