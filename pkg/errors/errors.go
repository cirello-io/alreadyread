package errors

import "fmt"

func Invalid(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("invalid: %w", err)
}

func Invalidf(err error, message string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf("invalid: "+message, args...)
	return fmt.Errorf(msg+": %w", err)
}

func Internal(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("internal: %w", err)
}

func Internalf(err error, message string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf("internal: "+message, args...)
	return fmt.Errorf(msg+": %w", err)
}

func Errorf(err error, message string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(message, args...)
	return fmt.Errorf(msg+": %w", err)
}

func E(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("error: %w", err)
}
