package errors

type PreValidationError struct {
	Message string
}

type DriftError struct {
	Message string
}

func (e *PreValidationError) Error() string {
	return e.Message
}

func (e *DriftError) Error() string {
	return e.Message
}
