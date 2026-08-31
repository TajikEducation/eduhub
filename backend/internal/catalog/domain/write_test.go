package domain_test

import (
	"errors"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

func validCreateInput() domain.CreateInstitutionInput {
	return domain.CreateInstitutionInput{
		Name:   domain.Bilingual{RU: "Сад №1", TG: "Боғчаи №1"},
		Types:  []string{"kindergarten"},
		Region: "dushanbe",
		Lat:    38.5, Lng: 68.7,
	}
}

func TestCreateInstitutionInput_Validate_ValidInputPasses(t *testing.T) {
	if err := validCreateInput().Validate(); err != nil {
		t.Fatalf("Validate() unexpected error: %v", err)
	}
}

func TestCreateInstitutionInput_Validate_MissingNameIsInvalid(t *testing.T) {
	in := validCreateInput()
	in.Name = domain.Bilingual{}
	err := in.Validate()
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("Validate() error = %v, want apperr.ErrInvalid", err)
	}
}

func TestCreateInstitutionInput_Validate_EmptyTypesIsInvalid(t *testing.T) {
	in := validCreateInput()
	in.Types = nil
	err := in.Validate()
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("Validate() error = %v, want apperr.ErrInvalid", err)
	}
}

func TestCreateInstitutionInput_Validate_UnknownTypeIsInvalid(t *testing.T) {
	in := validCreateInput()
	in.Types = []string{"not-a-real-type"}
	err := in.Validate()
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("Validate() error = %v, want apperr.ErrInvalid", err)
	}
}

func TestCreateInstitutionInput_Validate_UnknownRegionIsInvalid(t *testing.T) {
	in := validCreateInput()
	in.Region = "narnia"
	err := in.Validate()
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("Validate() error = %v, want apperr.ErrInvalid", err)
	}
}

func TestCreateInstitutionInput_Validate_OutOfRangeLatLngIsInvalid(t *testing.T) {
	cases := []domain.CreateInstitutionInput{
		func() domain.CreateInstitutionInput { in := validCreateInput(); in.Lat = 91; return in }(),
		func() domain.CreateInstitutionInput { in := validCreateInput(); in.Lng = 181; return in }(),
	}
	for _, in := range cases {
		if err := in.Validate(); !errors.Is(err, apperr.ErrInvalid) {
			t.Errorf("Validate() error = %v, want apperr.ErrInvalid", err)
		}
	}
}
