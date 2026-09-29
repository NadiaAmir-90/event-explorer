package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/beego/beego/v2/core/logs"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type APIClient struct {
	client  HTTPClient
	cache   *Cache
	baseURL string
	apiKey  string
}

type fetchResult struct {
	events []Event
	err    error
}

func NewAPIClient(
	client HTTPClient,
	cache *Cache,
	baseURL string,
	apiKey string,
) *APIClient {
	if client == nil {
		client = &http.Client{}
	}

	return &APIClient{
		client:  client,
		cache:   cache,
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
	}
}

func NewAPIClientWithConfig(
	client HTTPClient,
	cache *Cache,
	baseURL string,
	apiKey string,
) *APIClient {
	return NewAPIClient(
		client,
		cache,
		baseURL,
		apiKey,
	)
}

// BuildEventsURL creates the Ticketmaster events URL.
func (a *APIClient) BuildEventsURL(
	city string,
	countryCode string,
) string {
	u := a.baseURL + "/events.json"

	params := url.Values{}
	params.Set("apikey", a.apiKey)

	if city != "" {
		params.Set("city", city)
	}

	if countryCode != "" {
		params.Set("countryCode", countryCode)
	}

	return u + "?" + params.Encode()
}

func (a *APIClient) BuildSearchURL(
	query string,
	category string,
) string {
	u := a.baseURL + "/events.json"

	params := url.Values{}
	params.Set("apikey", a.apiKey)

	if query != "" {
		params.Set("keyword", query)
	}

	if category != "" {
		params.Set("classificationName", category)
	}

	return u + "?" + params.Encode()
}
// FetchEventsConcurrent fetches multiple resources concurrently.
func (a *APIClient) FetchEventsConcurrent(
	urls []string,
) ([]Event, error) {

	if len(urls) == 0 {
		return []Event{}, nil
	}

	results := make(chan fetchResult, len(urls))

	var wg sync.WaitGroup

	wg.Add(len(urls))

	for _, requestURL := range urls {
		requestURL := requestURL

		go func() {
			defer wg.Done()

			events, err := a.fetchWithCache(requestURL)

			results <- fetchResult{
				events: events,
				err:    err,
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	allEvents := make([]Event, 0)

	var firstErr error

	for result := range results {
		if result.err != nil {
			if firstErr == nil {
				firstErr = result.err
			}

			continue
		}

		allEvents = append(allEvents, result.events...)
	}

	return allEvents, firstErr
}

// fetchWithCache returns cached events when available.
// On a cache miss, it fetches the events and stores them in the cache.
func (a *APIClient) fetchWithCache(requestURL string) ([]Event, error) {
	var fetchErr error

	value := a.cache.GetOrFetch(requestURL, func() interface{} {
		logs.Info("Cache miss: fetching %s", requestURL)

		events, err := a.fetch(requestURL)

		if err != nil {
			fetchErr = err
			return nil
		}

		return events
	})

	// The HTTP request failed.
	if fetchErr != nil {
		return nil, fetchErr
	}

	// No value was returned.
	if value == nil {
		return nil, fmt.Errorf("no events returned for %s", requestURL)
	}

	events, ok := value.([]Event)
	if !ok {
		return nil, fmt.Errorf("invalid cached value for %s", requestURL)
	}

	logs.Info("Using cached result: %s", requestURL)

	return events, nil
}

// fetch performs the actual HTTP request to Ticketmaster.
func (a *APIClient) fetch(requestURL string) ([]Event, error) {
	req, err := http.NewRequest(
		http.MethodGet,
		requestURL,
		nil,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"create request: %w",
			err,
		)
	}

	response, err := a.client.Do(req)

	if err != nil {
		return nil, fmt.Errorf(
			"request Ticketmaster API: %w",
			err,
		)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)

		return nil, fmt.Errorf(
			"Ticketmaster API returned status %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var payload ticketmasterResponse

	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf(
			"decode Ticketmaster response: %w",
			err,
		)
	}

	events := make([]Event, 0, len(payload.Embedded.Events))

	for _, raw := range payload.Embedded.Events {
		events = append(
			events,
			raw.toEvent(),
		)
	}

	return events, nil
}