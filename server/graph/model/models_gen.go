// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package model

type Error struct {
	Message string `json:"message"`
	Reason  string `json:"reason"`
}

type ForgotPasswordInput struct {
	Email string `json:"email"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message              string  `json:"message"`
	AccessToken          *string `json:"accessToken"`
	AccessTokenExpiresAt *int64  `json:"accessTokenExpiresAt"`
	User                 *User   `json:"user"`
}

type ResendVerifyEmailInput struct {
	Email string `json:"email"`
}

type ResetPassowrdInput struct {
	Token           string `json:"token"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type Response struct {
	Message string `json:"message"`
}

type SignUpInput struct {
	FirstName       *string `json:"firstName"`
	LastName        *string `json:"lastName"`
	Email           string  `json:"email"`
	Password        string  `json:"password"`
	ConfirmPassword string  `json:"confirmPassword"`
	Image           *string `json:"image"`
}

type UpdateProfileInput struct {
	OldPassword        *string `json:"oldPassword"`
	NewPassword        *string `json:"newPassword"`
	ConfirmNewPassword *string `json:"confirmNewPassword"`
	FirstName          *string `json:"firstName"`
	LastName           *string `json:"lastName"`
	Image              *string `json:"image"`
	Email              *string `json:"email"`
}

type User struct {
	ID              string  `json:"id"`
	Email           string  `json:"email"`
	SignupMethod    string  `json:"signupMethod"`
	FirstName       *string `json:"firstName"`
	LastName        *string `json:"lastName"`
	EmailVerifiedAt *int64  `json:"emailVerifiedAt"`
	Image           *string `json:"image"`
	CreatedAt       *int64  `json:"createdAt"`
	UpdatedAt       *int64  `json:"updatedAt"`
}

type VerificationRequest struct {
	ID         string  `json:"id"`
	Identifier *string `json:"identifier"`
	Token      *string `json:"token"`
	Email      *string `json:"email"`
	Expires    *int64  `json:"expires"`
	CreatedAt  *int64  `json:"createdAt"`
	UpdatedAt  *int64  `json:"updatedAt"`
}

type VerifyEmailInput struct {
	Token string `json:"token"`
}
