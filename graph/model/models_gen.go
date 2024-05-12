// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

import (
	"fmt"
	"io"
	srvcustomer "quorum-api/services/customer"
	srvpost "quorum-api/services/post"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type BaseError interface {
	IsBaseError()
	GetMessage() string
	GetPath() []string
}

type GenerateSignedPostOptionURLError interface {
	IsGenerateSignedPostOptionURLError()
}

type GetLoginLinkError interface {
	IsGetLoginLinkError()
}

type SignUpError interface {
	IsSignUpError()
}

type SubmitVoteError interface {
	IsSubmitVoteError()
}

type UpsertPostError interface {
	IsUpsertPostError()
}

type VerifyCustomerTokenError interface {
	IsVerifyCustomerTokenError()
}

type ClosesAtNotAfterOpensAtError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

func (ClosesAtNotAfterOpensAtError) IsBaseError()            {}
func (this ClosesAtNotAfterOpensAtError) GetMessage() string { return this.Message }
func (this ClosesAtNotAfterOpensAtError) GetPath() []string {
	if this.Path == nil {
		return nil
	}
	interfaceSlice := make([]string, 0, len(this.Path))
	for _, concrete := range this.Path {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (ClosesAtNotAfterOpensAtError) IsUpsertPostError() {}

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

type ErrPostNotOwned struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

func (ErrPostNotOwned) IsBaseError()            {}
func (this ErrPostNotOwned) GetMessage() string { return this.Message }
func (this ErrPostNotOwned) GetPath() []string {
	if this.Path == nil {
		return nil
	}
	interfaceSlice := make([]string, 0, len(this.Path))
	for _, concrete := range this.Path {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (ErrPostNotOwned) IsUpsertPostError() {}

type GenerateSignedPostOptionUrInput struct {
	// Generates a url to upload the file too based off this filename.
	// The name is ignored, but the extension is not.
	FileName    string `json:"fileName"`
	ContentType string `json:"contentType"`
}

type GenerateSignedPostOptionURLPayload struct {
	BucketName string                             `json:"bucketName"`
	FileKey    string                             `json:"fileKey"`
	URL        string                             `json:"url"`
	Errors     []GenerateSignedPostOptionURLError `json:"errors"`
}

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

type OpensAtAlreadyPassedError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

func (OpensAtAlreadyPassedError) IsBaseError()            {}
func (this OpensAtAlreadyPassedError) GetMessage() string { return this.Message }
func (this OpensAtAlreadyPassedError) GetPath() []string {
	if this.Path == nil {
		return nil
	}
	interfaceSlice := make([]string, 0, len(this.Path))
	for _, concrete := range this.Path {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (OpensAtAlreadyPassedError) IsUpsertPostError() {}

type OptionNotFoundError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

func (OptionNotFoundError) IsBaseError()            {}
func (this OptionNotFoundError) GetMessage() string { return this.Message }
func (this OptionNotFoundError) GetPath() []string {
	if this.Path == nil {
		return nil
	}
	interfaceSlice := make([]string, 0, len(this.Path))
	for _, concrete := range this.Path {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (OptionNotFoundError) IsSubmitVoteError() {}

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

type SubmitVoteInput struct {
	OptionID uuid.UUID `json:"optionId"`
	Reason   *string   `json:"reason,omitempty"`
}

type SubmitVotePayload struct {
	Post   *srvpost.Post     `json:"post,omitempty"`
	Errors []SubmitVoteError `json:"errors"`
}

type TooFewOptionsError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

func (TooFewOptionsError) IsBaseError()            {}
func (this TooFewOptionsError) GetMessage() string { return this.Message }
func (this TooFewOptionsError) GetPath() []string {
	if this.Path == nil {
		return nil
	}
	interfaceSlice := make([]string, 0, len(this.Path))
	for _, concrete := range this.Path {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (TooFewOptionsError) IsUpsertPostError() {}

type TooManyOptionsError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

func (TooManyOptionsError) IsBaseError()            {}
func (this TooManyOptionsError) GetMessage() string { return this.Message }
func (this TooManyOptionsError) GetPath() []string {
	if this.Path == nil {
		return nil
	}
	interfaceSlice := make([]string, 0, len(this.Path))
	for _, concrete := range this.Path {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (TooManyOptionsError) IsUpsertPostError() {}

type UnauthenticatedError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

func (UnauthenticatedError) IsSubmitVoteError() {}

func (UnauthenticatedError) IsUpsertPostError() {}

func (UnauthenticatedError) IsBaseError()            {}
func (this UnauthenticatedError) GetMessage() string { return this.Message }
func (this UnauthenticatedError) GetPath() []string {
	if this.Path == nil {
		return nil
	}
	interfaceSlice := make([]string, 0, len(this.Path))
	for _, concrete := range this.Path {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (UnauthenticatedError) IsGenerateSignedPostOptionURLError() {}

type UnsupportedFileTypeError struct {
	Message string   `json:"message"`
	Path    []string `json:"path,omitempty"`
}

func (UnsupportedFileTypeError) IsUpsertPostError() {}

func (UnsupportedFileTypeError) IsBaseError()            {}
func (this UnsupportedFileTypeError) GetMessage() string { return this.Message }
func (this UnsupportedFileTypeError) GetPath() []string {
	if this.Path == nil {
		return nil
	}
	interfaceSlice := make([]string, 0, len(this.Path))
	for _, concrete := range this.Path {
		interfaceSlice = append(interfaceSlice, concrete)
	}
	return interfaceSlice
}

func (UnsupportedFileTypeError) IsGenerateSignedPostOptionURLError() {}

type UpsertPostInput struct {
	ID          uuid.UUID                `json:"id"`
	DesignPhase *DesignPhase             `json:"designPhase,omitempty"`
	Context     *string                  `json:"context,omitempty"`
	Category    *PostCategory            `json:"category,omitempty"`
	OpensAt     *time.Time               `json:"opensAt,omitempty"`
	ClosesAt    *time.Time               `json:"closesAt,omitempty"`
	Options     []*UpsertPostOptionInput `json:"options"`
}

type UpsertPostOptionInput struct {
	ID         uuid.UUID `json:"id"`
	Position   int       `json:"position"`
	BucketName string    `json:"bucketName"`
	FileKey    string    `json:"fileKey"`
}

type UpsertPostPayload struct {
	Post   *srvpost.Post     `json:"post,omitempty"`
	Errors []UpsertPostError `json:"errors"`
}

type VerifyCustomerTokenInput struct {
	Token string `json:"token"`
}

type VerifyCustomerTokenPayload struct {
	Customer *srvcustomer.Customer      `json:"customer,omitempty"`
	NewToken *string                    `json:"newToken,omitempty"`
	Errors   []VerifyCustomerTokenError `json:"errors"`
}

type DesignPhase string

const (
	DesignPhaseWireframe DesignPhase = "WIREFRAME"
	DesignPhaseLoFi      DesignPhase = "LO_FI"
	DesignPhaseHiFi      DesignPhase = "HI_FI"
)

var AllDesignPhase = []DesignPhase{
	DesignPhaseWireframe,
	DesignPhaseLoFi,
	DesignPhaseHiFi,
}

func (e DesignPhase) IsValid() bool {
	switch e {
	case DesignPhaseWireframe, DesignPhaseLoFi, DesignPhaseHiFi:
		return true
	}
	return false
}

func (e DesignPhase) String() string {
	return string(e)
}

func (e *DesignPhase) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = DesignPhase(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid DesignPhase", str)
	}
	return nil
}

func (e DesignPhase) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type PostCategory string

const (
	PostCategoryAnimation    PostCategory = "ANIMATION"
	PostCategoryBranding     PostCategory = "BRANDING"
	PostCategoryIllustration PostCategory = "ILLUSTRATION"
	PostCategoryPrint        PostCategory = "PRINT"
	PostCategoryProduct      PostCategory = "PRODUCT"
	PostCategoryTypography   PostCategory = "TYPOGRAPHY"
	PostCategoryWeb          PostCategory = "WEB"
)

var AllPostCategory = []PostCategory{
	PostCategoryAnimation,
	PostCategoryBranding,
	PostCategoryIllustration,
	PostCategoryPrint,
	PostCategoryProduct,
	PostCategoryTypography,
	PostCategoryWeb,
}

func (e PostCategory) IsValid() bool {
	switch e {
	case PostCategoryAnimation, PostCategoryBranding, PostCategoryIllustration, PostCategoryPrint, PostCategoryProduct, PostCategoryTypography, PostCategoryWeb:
		return true
	}
	return false
}

func (e PostCategory) String() string {
	return string(e)
}

func (e *PostCategory) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PostCategory(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PostCategory", str)
	}
	return nil
}

func (e PostCategory) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type PostStatus string

const (
	PostStatusDraft  PostStatus = "DRAFT"
	PostStatusLive   PostStatus = "LIVE"
	PostStatusClosed PostStatus = "CLOSED"
)

var AllPostStatus = []PostStatus{
	PostStatusDraft,
	PostStatusLive,
	PostStatusClosed,
}

func (e PostStatus) IsValid() bool {
	switch e {
	case PostStatusDraft, PostStatusLive, PostStatusClosed:
		return true
	}
	return false
}

func (e PostStatus) String() string {
	return string(e)
}

func (e *PostStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PostStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PostStatus", str)
	}
	return nil
}

func (e PostStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
