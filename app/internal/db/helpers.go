package db

import (
	"strings"
	"time"

	"github.com/Kryvea/Kryvea/internal/model"
)

func timePtrIfSet(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func timeOrTimeNever(t time.Time) time.Time {
	if t.IsZero() {
		return model.TimeNever
	}
	return t
}

func timePtrOrTimeNever(t time.Time) *time.Time {
	t = timeOrTimeNever(t)
	return &t
}

func emptyMapIfNil[V any](m map[string]V) map[string]V {
	if m == nil {
		return map[string]V{}
	}
	return m
}

func emptyStringsIfNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func escapeLike(s string) string {
	return strings.NewReplacer(
		"\\", "\\\\",
		"%", "\\%",
		"_", "\\_",
	).Replace(s)
}
