package errors

type ContextError struct {
	msg string
}

func NewContextError(msg string) *ContextError {
	return &ContextError{msg: msg}
}

func (c ContextError) Error() string {
	return c.msg
}
