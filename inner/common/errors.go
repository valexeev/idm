package common

type RequestValidationError struct {
	Message string
}

func (err RequestValidationError) Error() string {
	return err.Message
}

type AlreadyExistsError struct {
	Message string
}

func (err AlreadyExistsError) Error() string {
	return err.Message
}

type TransactionError struct {
	Message string
	Err     error
}

func (e TransactionError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e TransactionError) Unwrap() error {
	return e.Err
}

type RepositoryError struct {
	Message string
	Err     error
}

func (e RepositoryError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e RepositoryError) Unwrap() error {
	return e.Err
}

type NotFoundError struct {
	Message string
}

func (err NotFoundError) Error() string {
	return err.Message
}
