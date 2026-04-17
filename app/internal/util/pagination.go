package util

import "strconv"

type Pagination struct {
	Page  int64
	Limit int64
	Skip  int64
}

func GetPagination(pageStr, limitStr string) *Pagination {
	page := int64(1)
	limit := int64(5)

	// Parse page query param
	if pageStr != "" {
		if val, err := strconv.ParseInt(pageStr, 10, 64); err == nil && val > 0 {
			page = val
		}
	}

	// Parse limit query param
	if limitStr != "" {
		if val, err := strconv.ParseInt(limitStr, 10, 64); err == nil && val > 0 {
			limit = val
		}
	}

	// Calculate skip for Mongo
	skip := (page - 1) * limit

	return &Pagination{
		Page:  page,
		Limit: limit,
		Skip:  skip,
	}
}
