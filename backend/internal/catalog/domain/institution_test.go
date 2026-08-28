package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewInstitution_InitializesCollectionsNonNil(t *testing.T) {
	inst := NewInstitution(uuid.New(), Bilingual{RU: "Сад №1", TG: "Боғчаи №1"}, "dushanbe")

	if inst.Gallery == nil {
		t.Fatal("Gallery должна быть инициализирована non-nil слайсом")
	}
	if len(inst.Gallery) != 0 {
		t.Fatalf("ожидали пустой Gallery, получили len=%d", len(inst.Gallery))
	}
	if inst.Types == nil {
		t.Fatal("Types должен быть инициализирован non-nil слайсом")
	}
	if len(inst.Types) != 0 {
		t.Fatalf("ожидали пустой Types, получили len=%d", len(inst.Types))
	}
}
