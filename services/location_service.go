package services

import (
	"context"

	"event-explorer/models"
)

type LocationService struct {
	Client *models.GooglePlacesClient
}

func NewLocationService(
	client *models.GooglePlacesClient,
) *LocationService {
	return &LocationService{
		Client: client,
	}
}

func (s *LocationService) Autocomplete(
	ctx context.Context,
	input string,
	sessionToken string,
) ([]models.LocationSuggestion, error) {

	return s.Client.Autocomplete(
		ctx,
		input,
		sessionToken,
	)
}

func (s *LocationService) GetPlace(
	ctx context.Context,
	placeID string,
	sessionToken string,
) (models.LocationDetails, error) {

	return s.Client.GetPlace(
		ctx,
		placeID,
		sessionToken,
	)
}