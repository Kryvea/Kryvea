package poc

const (
	PocTypeText    = "text"
	PocTypeRequest = "request/response"
	PocTypeImage   = "image"
)

var (
	PocTypes = map[string]struct{}{
		PocTypeText:    {},
		PocTypeRequest: {},
		PocTypeImage:   {},
	}
)
