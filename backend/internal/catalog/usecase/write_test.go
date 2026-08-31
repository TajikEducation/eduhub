package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/catalog/usecase"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

func validCreateInput() domain.CreateInstitutionInput {
	return domain.CreateInstitutionInput{
		Name:   domain.Bilingual{RU: "Сад №1", TG: "Боғчаи №1"},
		Types:  []string{"kindergarten"},
		Region: "dushanbe",
	}
}

func TestService_Register_SetsPendingStatusAndOwner(t *testing.T) {
	fake := &fakeRepo{}
	svc := usecase.New(fake)
	ownerID := uuid.New()

	created, err := svc.Register(context.Background(), ownerID, validCreateInput())
	if err != nil {
		t.Fatalf("Register() unexpected error: %v", err)
	}

	if created.ModerationStatus != "pending" {
		t.Errorf("ModerationStatus = %q, want %q", created.ModerationStatus, "pending")
	}
	if fake.createOwnerID != ownerID {
		t.Errorf("repo received ownerID = %v, want %v", fake.createOwnerID, ownerID)
	}
	if fake.createInst.Name.RU != "Сад №1" {
		t.Errorf("repo received Name.RU = %q, want %q", fake.createInst.Name.RU, "Сад №1")
	}
}

func TestService_Register_InvalidInputRejectedBeforeRepo(t *testing.T) {
	fake := &fakeRepo{}
	svc := usecase.New(fake)

	in := validCreateInput()
	in.Region = "not-a-region"

	_, err := svc.Register(context.Background(), uuid.New(), in)
	if !errors.Is(err, apperr.ErrInvalid) {
		t.Fatalf("Register() error = %v, want apperr.ErrInvalid", err)
	}
	if fake.createInst.ID != uuid.Nil {
		t.Error("Register() with invalid input must not touch the repo")
	}
}

func TestService_Update_DelegatesToRepo(t *testing.T) {
	fake := &fakeRepo{}
	svc := usecase.New(fake)
	id := uuid.New()
	price := 500

	updated, err := svc.Update(context.Background(), id, domain.UpdateInstitutionInput{Price: &price})
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if updated.ID != id {
		t.Errorf("updated.ID = %v, want %v", updated.ID, id)
	}
	if fake.updateID != id {
		t.Errorf("repo received id = %v, want %v", fake.updateID, id)
	}
	if fake.updatePatch.Price == nil || *fake.updatePatch.Price != price {
		t.Errorf("repo received patch.Price = %v, want %d", fake.updatePatch.Price, price)
	}
}

func TestService_IsOwner_MatchingUserReturnsTrue(t *testing.T) {
	userID := uuid.New()
	fake := &fakeRepo{ownerID: userID}
	svc := usecase.New(fake)

	ok, err := svc.IsOwner(context.Background(), uuid.New(), userID)
	if err != nil {
		t.Fatalf("IsOwner() unexpected error: %v", err)
	}
	if !ok {
		t.Error("IsOwner() = false, want true for matching user")
	}
}

func TestService_IsOwner_DifferentUserReturnsFalse(t *testing.T) {
	fake := &fakeRepo{ownerID: uuid.New()}
	svc := usecase.New(fake)

	ok, err := svc.IsOwner(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("IsOwner() unexpected error: %v", err)
	}
	if ok {
		t.Error("IsOwner() = true, want false for different user")
	}
}

func TestService_SetModerationStatus_DelegatesToRepo(t *testing.T) {
	fake := &fakeRepo{}
	svc := usecase.New(fake)
	id := uuid.New()

	if err := svc.SetModerationStatus(context.Background(), id, "approved"); err != nil {
		t.Fatalf("SetModerationStatus() unexpected error: %v", err)
	}
	if fake.statusID != id || fake.statusValue != "approved" {
		t.Errorf("repo received (%v, %q), want (%v, %q)", fake.statusID, fake.statusValue, id, "approved")
	}
}
