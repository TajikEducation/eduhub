package domain_test

import (
	"errors"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

func TestRegisterInput_Validate_ValidInputPasses(t *testing.T) {
	in := domain.RegisterInput{Email: "user@example.com", Password: "password123", DisplayName: "User"}
	if err := in.Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
}

func TestRegisterInput_Validate_EmptyEmailIsInvalid(t *testing.T) {
	in := domain.RegisterInput{Email: "", Password: "password123"}
	err := in.Validate()
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("Validate() error = %v, want apperr.ErrInvalid", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("expected *apperr.Error")
	}
	if _, ok := target.Fields["email"]; !ok {
		t.Errorf("Fields = %v, want key %q", target.Fields, "email")
	}
}

func TestRegisterInput_Validate_MalformedEmailIsInvalid(t *testing.T) {
	cases := []string{"notanemail", "@example.com", "user@"}
	for _, email := range cases {
		in := domain.RegisterInput{Email: email, Password: "password123"}
		err := in.Validate()
		if !errors.Is(err, apperr.ErrInvalid) {
			t.Errorf("email=%q: Validate() error = %v, want apperr.ErrInvalid", email, err)
		}
	}
}

func TestRegisterInput_Validate_ShortPasswordIsInvalid(t *testing.T) {
	in := domain.RegisterInput{Email: "user@example.com", Password: "short"}
	err := in.Validate()
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("Validate() error = %v, want apperr.ErrInvalid", err)
	}

	var target *apperr.Error
	if !errors.As(err, &target) {
		t.Fatalf("expected *apperr.Error")
	}
	if _, ok := target.Fields["password"]; !ok {
		t.Errorf("Fields = %v, want key %q", target.Fields, "password")
	}
}

func TestRegisterInput_NormalizedEmail_LowercasesAndTrims(t *testing.T) {
	in := domain.RegisterInput{Email: "  User@Example.COM  "}
	if got := in.NormalizedEmail(); got != "user@example.com" {
		t.Errorf("NormalizedEmail() = %q, want %q", got, "user@example.com")
	}
}
