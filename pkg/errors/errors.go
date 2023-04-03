package errors

type PreValidationError struct {
	Message string
}

func (e *PreValidationError) Error() string {
	return e.Message
}
