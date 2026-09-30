package services

import (
	"context"

	"event-explorer/models"
)

type EventService struct {
	APIClient *models.APIClient
}

func NewEventService(
	apiClient *models.APIClient,
) *EventService {

	return &EventService{
		APIClient: apiClient,
	}
}

func (s *EventService) ListEvents(
	ctx context.Context,
	city string,
	countryCode string,
) (map[string][]models.Event, error) {

	return s.APIClient.FetchEventsByLocation(
		ctx,
		city,
		countryCode,
	)
}

func (s *EventService) GetEventDetails(
	ctx context.Context,
	eventID string,
) (models.Event, error) {

	return s.APIClient.GetEventDetails(
		ctx,
		eventID,
	)
}