package util

import "strconv"

type Pagination struct {
	Limit int64
	Skip  int64
}

func GetPagination(pageStr, limitStr string) *Pagination {
	page := int64(1)
	limit := int64(5)

	if pageStr != "" {
		if val, err := strconv.ParseInt(pageStr, 10, 64); err == nil && val > 0 {
			page = val
		}
	}

	if limitStr != "" {
		if val, err := strconv.ParseInt(limitStr, 10, 64); err == nil && val > 0 {
			limit = val
		}
	}

	return &Pagination{
		Limit: limit,
		Skip:  (page - 1) * limit,
	}
}
