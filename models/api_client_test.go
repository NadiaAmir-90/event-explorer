package models

import (
	"bytes"
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
