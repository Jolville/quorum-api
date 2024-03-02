// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	srvcustomer "quorum-api/services/customer"
)

type BaseError interface {
	IsBaseError()
	GetMessage() string
	GetPath() []string
}

type GetLoginLinkError interface {
	IsGetLoginLinkError()
}

type SignUpError interface {
	IsSignUpError()
}

type VerifyCustomerTokenError interface {
	IsVerifyCustomerTokenError()
}

type CustomerNotFoundError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

func (CustomerNotFoundError) IsBaseError()            {}
func (this CustomerNotFoundError) GetMessage() string { return this.Message }
func (this CustomerNotFoundError) GetPath() []string {
	if this.Path == nil {
		return nil
	}
	interfaceSlice := make([]string, 0, len(this.Path))
	for _, concrete := range this.Path {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (CustomerNotFoundError) IsGetLoginLinkError() {}

type EmailTakenError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

func (EmailTakenError) IsBaseError()            {}
func (this EmailTakenError) GetMessage() string { return this.Message }
func (this EmailTakenError) GetPath() []string {
	if this.Path == nil {
		return nil
	}
	interfaceSlice := make([]string, 0, len(this.Path))
	for _, concrete := range this.Path {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (EmailTakenError) IsSignUpError() {}

type GetLoginLinkInput struct {
	Email    string `json:"email"`
	ReturnTo string `json:"returnTo"`
}

type GetLoginLinkPayload struct {
	Errors []GetLoginLinkError `json:"errors"`
}

type InvalidEmailError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

func (InvalidEmailError) IsBaseError()            {}
func (this InvalidEmailError) GetMessage() string { return this.Message }
func (this InvalidEmailError) GetPath() []string {
	if this.Path == nil {
		return nil
	}
	interfaceSlice := make([]string, 0, len(this.Path))
	for _, concrete := range this.Path {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (InvalidEmailError) IsSignUpError() {}

func (InvalidEmailError) IsGetLoginLinkError() {}

type InvalidReturnToError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

func (InvalidReturnToError) IsBaseError()            {}
func (this InvalidReturnToError) GetMessage() string { return this.Message }
func (this InvalidReturnToError) GetPath() []string {
	if this.Path == nil {
		return nil
	}
	interfaceSlice := make([]string, 0, len(this.Path))
	for _, concrete := range this.Path {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (InvalidReturnToError) IsSignUpError() {}

func (InvalidReturnToError) IsGetLoginLinkError() {}

type LinkExpiredError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

func (LinkExpiredError) IsBaseError()            {}
func (this LinkExpiredError) GetMessage() string { return this.Message }
func (this LinkExpiredError) GetPath() []string {
	if this.Path == nil {
		return nil
	}
	interfaceSlice := make([]string, 0, len(this.Path))
	for _, concrete := range this.Path {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (LinkExpiredError) IsVerifyCustomerTokenError() {}

type Mutation struct {
}

type Query struct {
}

type SignUpInput struct {
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Email      string `json:"email"`
	Profession string `json:"profession"`
	ReturnTo   string `json:"returnTo"`
}

type SignUpPayload struct {
	Errors []SignUpError `json:"errors"`
}

type VerifyCustomerTokenInput struct {
	Token string `json:"token"`
}

type VerifyCustomerTokenPayload struct {
	Customer *srvcustomer.Customer      `json:"customer,omitempty"`
	NewToken *string                    `json:"newToken,omitempty"`
	Errors   []VerifyCustomerTokenError `json:"errors"`
}
