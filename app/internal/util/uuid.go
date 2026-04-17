package util

import (
	"fmt"

	"github.com/google/uuid"
)

func ParseUUID(id string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, err
	}

	if parsed.Variant() != uuid.RFC4122 || parsed == uuid.Nil {
		return uuid.Nil, fmt.Errorf("invalid UUID")
	}

	return parsed, nil
}
