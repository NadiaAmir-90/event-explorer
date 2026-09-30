package models

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type GooglePlacesClient struct {
	BaseURL    string
	APIKey     string
	HTTPClient HTTPClient
}

type LocationSuggestion struct {
	PlaceID string `json:"placeId"`
	Text    string `json:"text"`
}

type LocationDetails struct {
	City        string `json:"city"`
	CountryCode string `json:"countryCode"`
}

func NewGooglePlacesClient(
	baseURL string,
	apiKey string,
	client HTTPClient,
) *GooglePlacesClient {

	if client == nil {
		client = &http.Client{}
	}

	return &GooglePlacesClient{
		BaseURL:    strings.TrimRight(baseURL, "/"),
		APIKey:     apiKey,
		HTTPClient: client,
	}
}

// ==================================================
// AUTOCOMPLETE
// ==================================================

func (g *GooglePlacesClient) Autocomplete(
	ctx context.Context,
	input string,
	sessionToken string,
) ([]LocationSuggestion, error) {

	requestURL := g.BaseURL + "/v1/places:autocomplete"

	body := map[string]interface{}{
		"input": input,
		"includedPrimaryTypes": []string{
			"(cities)",
		},
		"sessionToken": sessionToken,
	}

	bodyBytes, err := json.Marshal(body)

	if err != nil {
		return nil, fmt.Errorf(
			"encode autocomplete request: %w",
			err,
		)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		requestURL,
		bytes.NewReader(bodyBytes),
	)

	if err != nil {
		return nil, fmt.Errorf(
			"create autocomplete request: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"X-Goog-Api-Key",
		g.APIKey,
	)

	req.Header.Set(
		"X-Goog-FieldMask",
		"suggestions.placePrediction.placeId,suggestions.placePrediction.text.text",
	)

	response, err := g.HTTPClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf(
			"request Google autocomplete: %w",
			err,
		)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {

		var errorBody struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Status  string `json:"status"`
			} `json:"error"`
		}

		if err := json.NewDecoder(response.Body).Decode(&errorBody); err == nil {
			return nil, fmt.Errorf(
				"Google autocomplete returned status %d: %s (%s)",
				response.StatusCode,
				errorBody.Error.Message,
				errorBody.Error.Status,
			)
		}

		return nil, fmt.Errorf(
			"Google autocomplete returned status %d",
			response.StatusCode,
		)
	}

	var payload struct {
		Suggestions []struct {
			PlacePrediction struct {
				PlaceID string `json:"placeId"`

				Text struct {
					Text string `json:"text"`
				} `json:"text"`
			} `json:"placePrediction"`
		} `json:"suggestions"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&payload); err != nil {

		return nil, fmt.Errorf(
			"decode Google autocomplete response: %w",
			err,
		)

	}

	result := make(
		[]LocationSuggestion,
		0,
		len(payload.Suggestions),
	)

	for _, item := range payload.Suggestions {

		result = append(
			result,
			LocationSuggestion{
				PlaceID: item.PlacePrediction.PlaceID,
				Text:    item.PlacePrediction.Text.Text,
			},
		)

	}

	return result, nil
}

// ==================================================
// PLACE DETAILS
// ==================================================

func (g *GooglePlacesClient) GetPlace(
	ctx context.Context,
	placeID string,
	sessionToken string,
) (LocationDetails, error) {

	placeID = strings.TrimSpace(placeID)

	if placeID == "" {
		return LocationDetails{},
			fmt.Errorf("place ID is required")
	}

	requestURL :=
		g.BaseURL +
			"/v1/places/" +
			url.PathEscape(placeID)

	query := url.Values{}

	query.Set(
		"sessionToken",
		sessionToken,
	)

	requestURL += "?" + query.Encode()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		requestURL,
		nil,
	)

	if err != nil {
		return LocationDetails{},
			fmt.Errorf(
				"create place details request: %w",
				err,
			)
	}

	req.Header.Set(
		"X-Goog-Api-Key",
		g.APIKey,
	)

	req.Header.Set(
		"X-Goog-FieldMask",
		"addressComponents",
	)

	response, err := g.HTTPClient.Do(req)

	if err != nil {
		return LocationDetails{},
			fmt.Errorf(
				"request Google place details: %w",
				err,
			)
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {

		return LocationDetails{},
			fmt.Errorf(
				"Google place details returned status %d",
				response.StatusCode,
			)

	}

	var payload struct {
		AddressComponents []struct {
			LongText string `json:"longText"`

			ShortText string `json:"shortText"`

			Types []string `json:"types"`
		} `json:"addressComponents"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&payload); err != nil {

		return LocationDetails{},
			fmt.Errorf(
				"decode Google place details: %w",
				err,
			)
	}

	result := LocationDetails{}

	for _, component := range payload.AddressComponents {

		for _, componentType := range component.Types {

			if componentType == "locality" &&
				result.City == "" {

				result.City = component.LongText

			}

			if componentType == "country" &&
				result.CountryCode == "" {

				result.CountryCode =
					component.ShortText

			}

		}

	}

	if result.City == "" {

		return LocationDetails{},
			fmt.Errorf(
				"selected place does not contain a city",
			)

	}

	if result.CountryCode == "" {

		return LocationDetails{},
			fmt.Errorf(
				"selected place does not contain a country code",
			)

	}

	return result, nil
}
