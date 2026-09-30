package models

import (
	"context"
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

type categoryResult struct {
	category string
	events   []Event
	err      error
}

// NewAPIClient creates a new Ticketmaster API client.
func NewAPIClient(
	client HTTPClient,
	cache *Cache,
	baseURL string,
	apiKey string,
) *APIClient {

	if client == nil {
		client = &http.Client{}
	}

	if cache == nil {
		cache = NewCache()
	}

	return &APIClient{
		client:  client,
		cache:   cache,
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
	}
}

// NewAPIClientWithConfig is kept for compatibility with main.go/tests.
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

// BuildEventsURL creates a Ticketmaster events URL.
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

// BuildSearchURL is kept for compatibility with existing code/tests.
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

// FetchEventsByLocation fetches Music and Sports concurrently.
func (a *APIClient) FetchEventsByLocation(
	ctx context.Context,
	city string,
	countryCode string,
) (map[string][]Event, error) {

	categories := []string{
		"Music",
		"Sports",
	}

	results := make(chan categoryResult, len(categories))

	for _, category := range categories {
		category := category

		go func() {
			events, err := a.fetchCategory(
				ctx,
				city,
				countryCode,
				category,
			)

			results <- categoryResult{
				category: category,
				events:   events,
				err:      err,
			}
		}()
	}

	response := map[string][]Event{
		"Music":  {},
		"Sports": {},
	}

	var errorsFound []string

	for range categories {
		result := <-results

		if result.err != nil {
			errorsFound = append(
				errorsFound,
				result.category+": "+result.err.Error(),
			)
			continue
		}

		response[result.category] = result.events
	}

	// Return an error only when both category requests failed.
	if len(errorsFound) == len(categories) {
		return response, fmt.Errorf(
			"all category requests failed: %s",
			strings.Join(errorsFound, "; "),
		)
	}

	return response, nil
}

// fetchCategory fetches one Ticketmaster category.
func (a *APIClient) fetchCategory(
	ctx context.Context,
	city string,
	countryCode string,
	category string,
) ([]Event, error) {

	requestURL := a.buildCategoryURL(
		city,
		countryCode,
		category,
	)

	return a.fetchWithCacheContext(ctx, requestURL)
}

// buildCategoryURL creates the URL for a category request.
func (a *APIClient) buildCategoryURL(
	city string,
	countryCode string,
	category string,
) string {

	u := a.baseURL + "/events.json"

	params := url.Values{}
	params.Set("apikey", a.apiKey)
	params.Set("city", city)
	params.Set("countryCode", countryCode)
	params.Set("classificationName", category)
	params.Set("size", "6")

	return u + "?" + params.Encode()
}

// FetchEventsConcurrent fetches multiple resources concurrently.
//
// This method is retained because existing tests/code may use it.
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

		allEvents = append(
			allEvents,
			result.events...,
		)
	}

	return allEvents, firstErr
}

// fetchWithCache uses the shared cache.
//
// Cache behavior:
//
// First request:
// MISS -> fetch -> store -> return
//
// Next requests:
// HIT -> return cached data
//
// Cache stays until Clear/ClearAll is called.
func (a *APIClient) fetchWithCache(
	requestURL string,
) ([]Event, error) {

	return a.fetchWithCacheContext(
		context.Background(),
		requestURL,
	)
}

// fetchWithCacheContext is the context-aware version of fetchWithCache.
func (a *APIClient) fetchWithCacheContext(
	ctx context.Context,
	requestURL string,
) ([]Event, error) {

	var fetchErr error

	value := a.cache.GetOrFetch(
		requestURL,
		func() interface{} {

			logs.Info(
				"Cache miss: fetching %s",
				requestURL,
			)

			events, err := a.fetchWithContext(
				ctx,
				requestURL,
			)

			if err != nil {
				fetchErr = err
				return nil
			}

			return events
		},
	)

	// The HTTP request failed.
	if fetchErr != nil {
		return nil, fetchErr
	}

	// Nothing was returned.
	if value == nil {
		return nil, fmt.Errorf(
			"no events returned for %s",
			requestURL,
		)
	}

	events, ok := value.([]Event)

	if !ok {
		return nil, fmt.Errorf(
			"invalid cached value for %s",
			requestURL,
		)
	}

	logs.Info(
		"Using cached result: %s",
		requestURL,
	)

	return events, nil
}

// fetch performs an HTTP request using the background context.
func (a *APIClient) fetch(
	requestURL string,
) ([]Event, error) {

	return a.fetchWithContext(
		context.Background(),
		requestURL,
	)
}

// fetchWithContext performs the actual Ticketmaster request.
func (a *APIClient) fetchWithContext(
	ctx context.Context,
	requestURL string,
) ([]Event, error) {

	req, err := http.NewRequestWithContext(
		ctx,
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

	if payload.Embedded == nil {
		return []Event{}, nil
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

// GetEventDetails retrieves one event from Ticketmaster.
func (a *APIClient) GetEventDetails(
	ctx context.Context,
	eventID string,
) (Event, error) {

	eventID = strings.TrimSpace(eventID)

	if eventID == "" {
		return Event{}, fmt.Errorf(
			"event ID is required",
		)
	}

	endpoint := a.baseURL +
		"/events/" +
		url.PathEscape(eventID)

	params := url.Values{}
	params.Set("apikey", a.apiKey)

	requestURL := endpoint + "?" + params.Encode()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		requestURL,
		nil,
	)

	if err != nil {
		return Event{}, fmt.Errorf(
			"create event request: %w",
			err,
		)
	}

	response, err := a.client.Do(req)

	if err != nil {
		return Event{}, fmt.Errorf(
			"request event details: %w",
			err,
		)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)

		return Event{}, fmt.Errorf(
			"Ticketmaster event returned status %d: %s",
			response.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var raw ticketmasterEvent

	if err := json.NewDecoder(response.Body).Decode(&raw); err != nil {
		return Event{}, fmt.Errorf(
			"decode event response: %w",
			err,
		)
	}

	return raw.toEvent(), nil
}
type ticketmasterResponse struct {
	Embedded *struct {
		Events []ticketmasterEvent `json:"events"`
	} `json:"_embedded"`
}

type ticketmasterEvent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`

	Images []struct {
		URL string `json:"url"`
	} `json:"images"`

	Dates struct {
		Start struct {
			LocalDate string `json:"localDate"`
			LocalTime string `json:"localTime"`
		} `json:"start"`
	} `json:"dates"`

	Classifications []struct {
		Segment struct {
			Name string `json:"name"`
		} `json:"segment"`
	} `json:"classifications"`

	Embedded *struct {
		Venues []struct {
			Name string `json:"name"`
			City struct {
				Name string `json:"name"`
			} `json:"city"`
		} `json:"venues"`
	} `json:"_embedded"`
}

func (raw ticketmasterEvent) toEvent() Event {
	location := ""

	if raw.Embedded != nil &&
		len(raw.Embedded.Venues) > 0 {

		venue := raw.Embedded.Venues[0]

		location = venue.Name

		if venue.City.Name != "" {
			location += ", " + venue.City.Name
		}
	}

	category := ""

	if len(raw.Classifications) > 0 {
		category = raw.Classifications[0].Segment.Name
	}

	imageURL := ""

	if len(raw.Images) > 0 {
		imageURL = raw.Images[0].URL
	}

	return Event{
		ID:          raw.ID,
		Title:       raw.Name,
		Category:    category,
		Location:    location,
		Date:        raw.Dates.Start.LocalDate,
		Description: raw.Description,
		ImageURL:    imageURL,
		TicketURL:   raw.URL,
	}
}