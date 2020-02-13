package common

type ComplexError struct {
	errno int
	msg   string
}

func (err *ComplexError) Error() string {
	return err.msg
}
