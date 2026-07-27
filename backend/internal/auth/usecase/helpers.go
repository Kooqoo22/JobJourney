package usecase

import (
	"errors"

	"github.com/Kooqoo22/JobJourney/backend/pkg/utils"
)

func wrapInternal(err error) error {
	var appErr *utils.AppError
	if errors.As(err, &appErr) {
		return err
	}
	return utils.ErrInternal(err)
}
