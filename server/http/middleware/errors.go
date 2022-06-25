package middleware

import (
	"github.com/m3dev/dsps/server/domain"
)

var (
	// ErrAuthRejection : auth rejection
	ErrAuthRejection = domain.NewErrorWithCode("dsps.auth.rejected")
)
