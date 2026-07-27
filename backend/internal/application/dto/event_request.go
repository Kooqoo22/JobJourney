package dto

type CreateEventRequest struct {
	Type     string  `json:"type" binding:"required,oneof=applied phone_screen interview assessment offer follow_up deadline note status_changed"`
	Title    string  `json:"title" binding:"required,max=200"`
	EventAt  string  `json:"event_at" binding:"required"`
	Notes    *string `json:"notes" binding:"omitempty,max=2000"`
	RemindAt *string `json:"remind_at" binding:"omitempty"`
}

type UpdateEventRequest struct {
	Type     *string `json:"type" binding:"omitempty,oneof=applied phone_screen interview assessment offer follow_up deadline note status_changed"`
	Title    *string `json:"title" binding:"omitempty,max=200"`
	EventAt  *string `json:"event_at" binding:"omitempty"`
	Notes    *string `json:"notes" binding:"omitempty,max=2000"`
	RemindAt *string `json:"remind_at" binding:"omitempty"`
}

type EventPathParam struct {
	ID      int64 `uri:"id" binding:"required,gt=0"`
	EventID int64 `uri:"event_id" binding:"required,gt=0"`
}
