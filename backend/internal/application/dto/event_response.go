package dto

type EventResponse struct {
	ID            int64   `json:"id"`
	ApplicationID int64   `json:"application_id"`
	Type          string  `json:"type"`
	Title         string  `json:"title"`
	EventAt       string  `json:"event_at"`
	Notes         *string `json:"notes"`
	RemindAt      *string `json:"remind_at"`
	StatusFrom    *string `json:"status_from"`
	StatusTo      *string `json:"status_to"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}
