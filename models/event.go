package models

type Event struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Location    string `json:"location"`
	Date        string `json:"date"`
	Description string `json:"description"`
	ImageURL    string `json:"imageUrl"`
	TicketURL   string `json:"ticketUrl"`
}