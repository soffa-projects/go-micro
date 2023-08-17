package errors

import "fmt"

type Managed struct {
	Kind    string `json:"kind"`
	Message string `json:"message,omitempty"`
	Details any    `json:"details,omitempty"`
}

type FunctionalError struct {
	Managed
}

type ConflictError struct {
	Managed
}

type TechnicalError struct {
	Managed
}

type ResourceNotFoundError struct {
	Managed
}

type ForbiddenError struct {
	Managed
}

type UnauthorizedError struct {
	Managed
}

func (e *Managed) Error() string {
	return fmt.Sprintf("Managed %s", e.Message)
}

// ---------------------------------------------------------------------------------------------------------------------

// Functional error
func Functional(message string, details ...any) error {
	return &FunctionalError{Managed{Kind: "error.functional", Message: message, Details: getDetails(details...)}}
}

func (e *FunctionalError) Error() string {
	return fmt.Sprintf("FunctionalError %s", e.Message)
}

// ---------------------------------------------------------------------------------------------------------------------

// Technical error
func Technical(message string, details ...any) error {
	return &TechnicalError{Managed{Kind: "error.technical", Message: message, Details: getDetails(details...)}}
}

func (e *TechnicalError) Error() string {
	return fmt.Sprintf("TechnicalErrror %s", e.Message)
}

// ---------------------------------------------------------------------------------------------------------------------

// ResourceNotFound error
func ResourceNotFound(message string, details ...any) error {
	return &ResourceNotFoundError{Managed{Kind: "error.resource_not_found", Message: message, Details: getDetails(details...)}}
}

func (e *ResourceNotFoundError) Error() string {
	return fmt.Sprintf("ResourceNotFoundError %s", e.Message)
}

// ---------------------------------------------------------------------------------------------------------------------

// Forbidden error
func Forbidden(message string, details ...any) error {
	return &ForbiddenError{Managed{Kind: "error.forbidden", Message: message, Details: getDetails(details...)}}
}

func (e *ForbiddenError) Error() string {
	return fmt.Sprintf("ForbiddenError %s", e.Message)
}

// ---------------------------------------------------------------------------------------------------------------------

// Unauthorized error
func Unauthorized(message string, details ...any) error {
	return &UnauthorizedError{Managed{Kind: "error.unauthorized", Message: message, Details: getDetails(details...)}}
}

func (e *UnauthorizedError) Error() string {
	return fmt.Sprintf("UnauthorizedError %s", e.Message)
}

// ---------------------------------------------------------------------------------------------------------------------

// Conflict error
func Conflict(message string, details ...any) error {
	return &ConflictError{Managed{Kind: "error.conflict", Message: message, Details: getDetails(details...)}}
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("Conflict %s", e.Message)
}

// ---------------------------------------------------------------------------------------------------------------------

func getDetails(details ...any) any {
	if len(details) == 0 {
		return nil
	}
	return details[0]
}

func ThrowAny(err error) {
	if err != nil {
		panic(err)
	}
}
