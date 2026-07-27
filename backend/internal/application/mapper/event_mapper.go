package mapper

import (
	"time"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

func ToEventResponse(e entity.ApplicationEvent, tz string) dto.EventResponse {
	return dto.EventResponse{
		ID:            e.ID,
		ApplicationID: e.ApplicationID,
		Type:          e.Type,
		Title:         e.Title,
		EventAt:       utils.ToLocal(e.EventAt, tz).Format(time.RFC3339),
		Notes:         e.Notes,
		RemindAt:      toLocalTimePtr(e.RemindAt, tz),
		StatusFrom:    e.StatusFrom,
		StatusTo:      e.StatusTo,
		CreatedAt:     utils.ToLocal(e.CreatedAt, tz).Format(time.RFC3339),
		UpdatedAt:     utils.ToLocal(e.UpdatedAt, tz).Format(time.RFC3339),
	}
}

func toLocalTimePtr(t *time.Time, tz string) *string {
	if t == nil {
		return nil
	}
	s := utils.ToLocal(*t, tz).Format(time.RFC3339)
	return &s
}
