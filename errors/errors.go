package errors

import "fmt"

type FunctionalError struct {
	Message string `json:"message,omitempty"`
	Details any    `json:"details,omitempty"`
}

type TechnicalError struct {
	Message string `json:"message,omitempty"`
	Details any    `json:"details,omitempty"`
}

type ResourceNotFoundError struct {
	Message string `json:"message,omitempty"`
	Details any    `json:"details,omitempty"`
}

type ForbiddenError struct {
	Message string `json:"message,omitempty"`
	Details any    `json:"details,omitempty"`
}

type UnauthorizedError struct {
	Message string `json:"message,omitempty"`
	Details any    `json:"details,omitempty"`
}

// ---------------------------------------------------------------------------------------------------------------------

// Functional error
func Functional(message string, details ...any) error {
	return &FunctionalError{Message: message, Details: getDetails(details...)}
}

func (e *FunctionalError) Error() string {
	return fmt.Sprintf("FunctionalError %s", e.Message)
}

// ---------------------------------------------------------------------------------------------------------------------

// Technical error
func Technical(message string, details ...any) error {
	return &TechnicalError{Message: message, Details: getDetails(details...)}
}

func (e *TechnicalError) Error() string {
	return fmt.Sprintf("TechnicalErrror %s", e.Message)
}

// ---------------------------------------------------------------------------------------------------------------------

// ResourceNotFound error
func ResourceNotFound(message string, details ...any) error {
	return &ResourceNotFoundError{Message: message, Details: getDetails(details...)}
}

func (e *ResourceNotFoundError) Error() string {
	return fmt.Sprintf("ResourceNotFoundError %s", e.Message)
}

// ---------------------------------------------------------------------------------------------------------------------

// Forbidden error
func Forbidden(message string, details ...any) error {
	return &ForbiddenError{Message: message, Details: getDetails(details...)}
}

func (e *ForbiddenError) Error() string {
	return fmt.Sprintf("ForbiddenError %s", e.Message)
}

// ---------------------------------------------------------------------------------------------------------------------

// Unauthorized error
func Unauthorized(message string, details ...any) error {
	return &UnauthorizedError{Message: message, Details: getDetails(details...)}
}

func (e *UnauthorizedError) Error() string {
	return fmt.Sprintf("UnauthorizedError %s", e.Message)
}

// ---------------------------------------------------------------------------------------------------------------------

func getDetails(details ...any) any {
	if len(details) == 0 {
		return nil
	}
	return details[0]
}
