package models

type Event struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Location    string `json:"location"`
	Date        string `json:"date"`
	Description string `json:"description"`
}

type ticketmasterResponse struct {
	Embedded struct {
		Events []rawEvent `json:"events"`
	} `json:"_embedded"`
}

type rawEvent struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	PleaseNote string `json:"pleaseNote"`

	Classifications []struct {
		Segment struct {
			Name string `json:"name"`
		} `json:"segment"`

		Genre struct {
			Name string `json:"name"`
		} `json:"genre"`
	} `json:"classifications"`

	Dates struct {
		Start struct {
			LocalDate string `json:"localDate"`
		} `json:"start"`
	} `json:"dates"`

	Embedded struct {
		Venues []struct {
			Name string `json:"name"`

			City struct {
				Name string `json:"name"`
			} `json:"city"`

			State struct {
				StateCode string `json:"stateCode"`
			} `json:"state"`
		} `json:"venues"`
	} `json:"_embedded"`
}

func (r rawEvent) toEvent() Event {
	category := ""

	if len(r.Classifications) > 0 {
		segment := r.Classifications[0].Segment.Name
		genre := r.Classifications[0].Genre.Name

		if segment != "" && genre != "" {
			category = segment + " / " + genre
		} else if segment != "" {
			category = segment
		} else {
			category = genre
		}
	}

	location := ""

	if len(r.Embedded.Venues) > 0 {
		venue := r.Embedded.Venues[0]

		parts := []string{}

		if venue.Name != "" {
			parts = append(parts, venue.Name)
		}

		if venue.City.Name != "" {
			parts = append(parts, venue.City.Name)
		}

		if venue.State.StateCode != "" {
			parts = append(parts, venue.State.StateCode)
		}

		location = joinLocation(parts)
	}

	return Event{
		ID:          r.ID,
		Title:       r.Name,
		Category:    category,
		Location:    location,
		Date:        r.Dates.Start.LocalDate,
		Description: r.PleaseNote,
	}
}

func joinLocation(parts []string) string {
	result := ""

	for i, part := range parts {
		if i > 0 {
			result += ", "
		}

		result += part
	}

	return result
}