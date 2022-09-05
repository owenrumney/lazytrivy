package aws

type ErrNoValidCredentials struct {
}

func NewErrNoValidCredentials() error {
	return ErrNoValidCredentials{}
}

func (e ErrNoValidCredentials) Error() string {
	return "No valid credentials found"
}
