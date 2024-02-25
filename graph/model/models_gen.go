// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

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

type VerifyUserTokenError interface {
	IsVerifyUserTokenError()
}

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
	Email string `json:"email"`
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

func (LinkExpiredError) IsVerifyUserTokenError() {}

type Mutation struct {
}

type Query struct {
}

type SignUpInput struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

type SignUpPayload struct {
	Errors []SignUpError `json:"errors"`
}

type User struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

type UserNotFoundError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

func (UserNotFoundError) IsBaseError()            {}
func (this UserNotFoundError) GetMessage() string { return this.Message }
func (this UserNotFoundError) GetPath() []string {
	if this.Path == nil {
		return nil
	}
	interfaceSlice := make([]string, 0, len(this.Path))
	for _, concrete := range this.Path {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (UserNotFoundError) IsGetLoginLinkError() {}

type VerifyUserTokenInput struct {
	Token string `json:"token"`
}

type VerifyUserTokenPayload struct {
	User   *User                  `json:"user,omitempty"`
	Errors []VerifyUserTokenError `json:"errors"`
}
