package errors

type CheckError struct {
	msg string
}

func NewCheckError(msg string) *CheckError {
	return &CheckError{msg: msg}
}

func (c CheckError) Error() string {
	return c.msg
}
