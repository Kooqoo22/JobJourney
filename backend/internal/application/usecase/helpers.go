package usecase

import (
	"errors"
	"time"

	"github.com/Kooqoo22/JobJourney/backend/internal/application/dto"
	"github.com/Kooqoo22/JobJourney/backend/internal/application/entity"
	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

func validateCreateRequest(req dto.CreateApplicationRequest) error {
	status := req.Status
	if status == "" {
		status = "applied"
	}

	if req.AppliedDate != nil && status != "wishlist" {
		t, err := time.Parse("2006-01-02", *req.AppliedDate)
		if err == nil && t.After(time.Now().UTC()) {
			return utils.ErrUnprocessable("validation failed", []utils.FieldError{
				{Field: "applied_date", Message: "cannot be in the future"},
			})
		}
	}
	return nil
}

func wrapNotFound(err error) error {
	if errors.Is(err, entity.ErrNotFound) {
		return utils.ErrNotFound("application not found")
	}
	return utils.ErrInternal(err)
}
