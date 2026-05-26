package errors

const (
	CodeSuccess      = 0
	CodeBadRequest   = 40000
	CodeUnauthorized = 40100
	CodeForbidden    = 40300
	CodeNotFound     = 40400
	CodeInternal     = 50000
)

type AppError struct {
	HTTPStatus int
	Code       int
	Message    string
}

func (e *AppError) Error() string {
	return e.Message
}
