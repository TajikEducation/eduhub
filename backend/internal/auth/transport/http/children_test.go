package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/abdulhalim/eduhub/backend/internal/auth/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
	"github.com/abdulhalim/eduhub/backend/internal/platform/httpx"
)

// fakeChildService — тестовый двойник childService.
type fakeChildService struct {
	createChildResult domain.Child
	createChildErr    error
	createChildCalls  []createChildCall

	listPendingResult []domain.Child
	listPendingErr    error

	confirmResult domain.Child
	confirmErr    error
	confirmCalls  []confirmChildCall

	rejectResult domain.Child
	rejectErr    error
	rejectCalls  []rejectChildCall
}

// *Call-типы захватывают аргументы вызова — тесты проверяют, что хендлер берёт владельца/актора
// из Principal (контекст), а не из тела запроса, и не путает местами childID/actorID (оба
// uuid.UUID — компилятор такую подмену не поймает).
type createChildCall struct {
	userID, institutionID uuid.UUID
	ageGroup, status      string
}

type confirmChildCall struct {
	childID, actorID uuid.UUID
	actorRole, reqID string
}

type rejectChildCall struct {
	childID, actorID             uuid.UUID
	actorRole, reasonCode, reqID string
	reasonText                   *string
}

func (f *fakeChildService) CreateChild(_ context.Context, userID, institutionID uuid.UUID, ageGroup, status string) (domain.Child, error) {
	f.createChildCalls = append(f.createChildCalls, createChildCall{userID: userID, institutionID: institutionID, ageGroup: ageGroup, status: status})
	return f.createChildResult, f.createChildErr
}

func (f *fakeChildService) ListPendingByInstitution(_ context.Context, _ uuid.UUID) ([]domain.Child, error) {
	return f.listPendingResult, f.listPendingErr
}

func (f *fakeChildService) ConfirmChild(_ context.Context, childID, actorID uuid.UUID, actorRole, requestID string) (domain.Child, error) {
	f.confirmCalls = append(f.confirmCalls, confirmChildCall{childID: childID, actorID: actorID, actorRole: actorRole, reqID: requestID})
	return f.confirmResult, f.confirmErr
}

func (f *fakeChildService) RejectChild(_ context.Context, childID, actorID uuid.UUID, actorRole, reasonCode string, reasonText *string, requestID string) (domain.Child, error) {
	f.rejectCalls = append(f.rejectCalls, rejectChildCall{
		childID: childID, actorID: actorID, actorRole: actorRole,
		reasonCode: reasonCode, reasonText: reasonText, reqID: requestID,
	})
	return f.rejectResult, f.rejectErr
}

// withRequestID кладёt в контекст запроса request_id через httpx.WithRequestID (не выполняя
// никакого реального хендлера) — чтобы тест мог сравнить конкретное значение request_id,
// дошедшее до usecase-слоя.
func withRequestID(req *http.Request, id string) *http.Request {
	req.Header.Set("X-Request-ID", id)
	var captured *http.Request
	httpx.WithRequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		captured = r
	})).ServeHTTP(httptest.NewRecorder(), req)
	return captured
}

// withPrincipal кладёт Principal в контекст запроса — по образцу handler_test.go (MeHandler).
func withPrincipal(r *http.Request, userID uuid.UUID, role string) *http.Request {
	ctx := context.WithValue(r.Context(), principalContextKey{}, Principal{UserID: userID, Role: role})
	return r.WithContext(ctx)
}

func TestCreateChildHandler(t *testing.T) {
	t.Run("happy path возвращает 201", func(t *testing.T) {
		userID := uuid.New()
		institutionID := uuid.New()
		want := domain.Child{
			ID: uuid.New(), UserID: userID, InstitutionID: institutionID,
			AgeGroup: "primary", Status: "current", ConfirmationStatus: "pending",
			CreatedAt: time.Now(),
		}
		svc := &fakeChildService{createChildResult: want}
		req := httptest.NewRequest(http.MethodPost, "/auth/children", jsonBody(t, createChildRequest{
			InstitutionID: institutionID, AgeGroup: "primary", Status: "current",
		}))
		req = withPrincipal(req, userID, "user")
		rec := httptest.NewRecorder()

		CreateChildHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusCreated, rec.Body.String())
		}
		var got childResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		if got.ID != want.ID {
			t.Errorf("ID = %v, want %v", got.ID, want.ID)
		}

		if len(svc.createChildCalls) != 1 {
			t.Fatalf("CreateChild() вызван %d раз, want 1", len(svc.createChildCalls))
		}
		call := svc.createChildCalls[0]
		if call.userID != userID {
			t.Errorf("userID = %v, want %v (Principal.UserID, не поле тела запроса)", call.userID, userID)
		}
		if call.institutionID != institutionID {
			t.Errorf("institutionID = %v, want %v", call.institutionID, institutionID)
		}
	})

	t.Run("неизвестное поле name возвращает 400 unknown_field", func(t *testing.T) {
		svc := &fakeChildService{}
		body := `{"institution_id":"` + uuid.New().String() + `","age_group":"primary","status":"current","name":"Аня"}`
		req := httptest.NewRequest(http.MethodPost, "/auth/children", strings.NewReader(body))
		req = withPrincipal(req, uuid.New(), "user")
		rec := httptest.NewRecorder()

		CreateChildHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		var respBody map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &respBody); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		errObj := respBody["error"].(map[string]any)
		fields := errObj["fields"].(map[string]any)
		if fields["body"] != "unknown_field" {
			t.Errorf("fields[body] = %v, want unknown_field", fields["body"])
		}
	})

	t.Run("невалидный age_group возвращает 400 с полем age_group", func(t *testing.T) {
		svc := &fakeChildService{}
		req := httptest.NewRequest(http.MethodPost, "/auth/children", jsonBody(t, createChildRequest{
			InstitutionID: uuid.New(), AgeGroup: "not-a-real-group", Status: "current",
		}))
		req = withPrincipal(req, uuid.New(), "user")
		rec := httptest.NewRecorder()

		CreateChildHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		var respBody map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &respBody); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		errObj := respBody["error"].(map[string]any)
		fields := errObj["fields"].(map[string]any)
		if _, ok := fields["age_group"]; !ok {
			t.Errorf("fields не содержит age_group: %v", fields)
		}
	})

	t.Run("невалидный status возвращает 400 с полем status", func(t *testing.T) {
		svc := &fakeChildService{}
		req := httptest.NewRequest(http.MethodPost, "/auth/children", jsonBody(t, createChildRequest{
			InstitutionID: uuid.New(), AgeGroup: "primary", Status: "not-a-real-status",
		}))
		req = withPrincipal(req, uuid.New(), "user")
		rec := httptest.NewRecorder()

		CreateChildHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		var respBody map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &respBody); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		errObj := respBody["error"].(map[string]any)
		fields := errObj["fields"].(map[string]any)
		if _, ok := fields["status"]; !ok {
			t.Errorf("fields не содержит status: %v", fields)
		}
	})

	t.Run("отсутствующий institution_id возвращает 400", func(t *testing.T) {
		svc := &fakeChildService{}
		req := httptest.NewRequest(http.MethodPost, "/auth/children", jsonBody(t, createChildRequest{
			AgeGroup: "primary", Status: "current",
		}))
		req = withPrincipal(req, uuid.New(), "user")
		rec := httptest.NewRecorder()

		CreateChildHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})

	t.Run("конфликт от usecase транслируется в 409", func(t *testing.T) {
		svc := &fakeChildService{createChildErr: apperr.ConflictCode("institution_not_approved", "учреждение не прошло модерацию")}
		req := httptest.NewRequest(http.MethodPost, "/auth/children", jsonBody(t, createChildRequest{
			InstitutionID: uuid.New(), AgeGroup: "primary", Status: "current",
		}))
		req = withPrincipal(req, uuid.New(), "user")
		rec := httptest.NewRecorder()

		CreateChildHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusConflict, rec.Body.String())
		}
	})
}

func TestListPendingChildrenHandler(t *testing.T) {
	t.Run("happy path возвращает 200 со списком", func(t *testing.T) {
		institutionID := uuid.New()
		want := []domain.Child{
			{ID: uuid.New(), InstitutionID: institutionID, ConfirmationStatus: "pending"},
		}
		svc := &fakeChildService{listPendingResult: want}
		req := httptest.NewRequest(http.MethodGet, "/institutions/"+institutionID.String()+"/children/pending", nil)
		req.SetPathValue("id", institutionID.String())
		rec := httptest.NewRecorder()

		ListPendingChildrenHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
		var got []childResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		wantID := want[0].ID
		if len(got) != 1 || got[0].ID != wantID {
			t.Errorf("got = %v, want %v", got, want)
		}
	})

	t.Run("невалидный id в пути возвращает 400", func(t *testing.T) {
		svc := &fakeChildService{}
		req := httptest.NewRequest(http.MethodGet, "/institutions/not-a-uuid/children/pending", nil)
		req.SetPathValue("id", "not-a-uuid")
		rec := httptest.NewRecorder()

		ListPendingChildrenHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
	})
}

func TestConfirmChildHandler(t *testing.T) {
	t.Run("happy path возвращает 200", func(t *testing.T) {
		childID := uuid.New()
		actorID := uuid.New()
		want := domain.Child{ID: childID, ConfirmationStatus: "confirmed"}
		svc := &fakeChildService{confirmResult: want}
		req := httptest.NewRequest(http.MethodPost, "/children/"+childID.String()+"/confirm", nil)
		req.SetPathValue("id", childID.String())
		req = withPrincipal(req, actorID, "moderator")
		req = withRequestID(req, "req-confirm-1")
		rec := httptest.NewRecorder()

		ConfirmChildHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
		var got childResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		if got.ConfirmationStatus != "confirmed" {
			t.Errorf("ConfirmationStatus = %q, want %q", got.ConfirmationStatus, "confirmed")
		}

		if len(svc.confirmCalls) != 1 {
			t.Fatalf("ConfirmChild() вызван %d раз, want 1", len(svc.confirmCalls))
		}
		call := svc.confirmCalls[0]
		if call.childID != childID {
			t.Errorf("childID = %v, want %v", call.childID, childID)
		}
		if call.actorID != actorID {
			t.Errorf("actorID = %v, want %v (Principal.UserID)", call.actorID, actorID)
		}
		if call.actorRole != "moderator" {
			t.Errorf("actorRole = %q, want %q (Principal.Role)", call.actorRole, "moderator")
		}
		if call.reqID != "req-confirm-1" {
			t.Errorf("requestID = %q, want %q", call.reqID, "req-confirm-1")
		}
	})

	t.Run("несуществующий id возвращает 404", func(t *testing.T) {
		svc := &fakeChildService{confirmErr: apperr.NotFound("child", "irrelevant")}
		childID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/children/"+childID.String()+"/confirm", nil)
		req.SetPathValue("id", childID.String())
		req = withPrincipal(req, uuid.New(), "moderator")
		rec := httptest.NewRecorder()

		ConfirmChildHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
		}
	})

	t.Run("уже не pending возвращает 409", func(t *testing.T) {
		svc := &fakeChildService{confirmErr: apperr.ConflictCode("child_not_pending", "привязка уже обработана")}
		childID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/children/"+childID.String()+"/confirm", nil)
		req.SetPathValue("id", childID.String())
		req = withPrincipal(req, uuid.New(), "moderator")
		rec := httptest.NewRecorder()

		ConfirmChildHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusConflict, rec.Body.String())
		}
	})
}

func TestRejectChildHandler(t *testing.T) {
	t.Run("happy path возвращает 200", func(t *testing.T) {
		childID := uuid.New()
		actorID := uuid.New()
		want := domain.Child{ID: childID, ConfirmationStatus: "rejected"}
		svc := &fakeChildService{rejectResult: want}
		reasonText := "документ не подтверждает обучение"
		req := httptest.NewRequest(http.MethodPost, "/children/"+childID.String()+"/reject", jsonBody(t, rejectChildRequest{
			ReasonCode: "invalid_document", ReasonText: &reasonText,
		}))
		req.SetPathValue("id", childID.String())
		req = withPrincipal(req, actorID, "admin")
		req = withRequestID(req, "req-reject-1")
		rec := httptest.NewRecorder()

		RejectChildHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
		}
		var got childResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		if got.ConfirmationStatus != "rejected" {
			t.Errorf("ConfirmationStatus = %q, want %q", got.ConfirmationStatus, "rejected")
		}

		if len(svc.rejectCalls) != 1 {
			t.Fatalf("RejectChild() вызван %d раз, want 1", len(svc.rejectCalls))
		}
		call := svc.rejectCalls[0]
		if call.childID != childID {
			t.Errorf("childID = %v, want %v", call.childID, childID)
		}
		if call.actorID != actorID {
			t.Errorf("actorID = %v, want %v (Principal.UserID)", call.actorID, actorID)
		}
		if call.actorRole != "admin" {
			t.Errorf("actorRole = %q, want %q (Principal.Role)", call.actorRole, "admin")
		}
		if call.reasonCode != "invalid_document" {
			t.Errorf("reasonCode = %q, want %q", call.reasonCode, "invalid_document")
		}
		if call.reasonText == nil || *call.reasonText != reasonText {
			t.Errorf("reasonText = %v, want %q", call.reasonText, reasonText)
		}
		if call.reqID != "req-reject-1" {
			t.Errorf("requestID = %q, want %q", call.reqID, "req-reject-1")
		}
	})

	t.Run("пустой reason_code возвращает 400", func(t *testing.T) {
		svc := &fakeChildService{}
		childID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/children/"+childID.String()+"/reject", jsonBody(t, rejectChildRequest{}))
		req.SetPathValue("id", childID.String())
		req = withPrincipal(req, uuid.New(), "moderator")
		rec := httptest.NewRecorder()

		RejectChildHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
		}
		var respBody map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &respBody); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		errObj := respBody["error"].(map[string]any)
		fields := errObj["fields"].(map[string]any)
		if _, ok := fields["reason_code"]; !ok {
			t.Errorf("fields не содержит reason_code: %v", fields)
		}
	})

	t.Run("несуществующий id возвращает 404", func(t *testing.T) {
		svc := &fakeChildService{rejectErr: apperr.NotFound("child", "irrelevant")}
		childID := uuid.New()
		req := httptest.NewRequest(http.MethodPost, "/children/"+childID.String()+"/reject", jsonBody(t, rejectChildRequest{
			ReasonCode: "invalid_document",
		}))
		req.SetPathValue("id", childID.String())
		req = withPrincipal(req, uuid.New(), "moderator")
		rec := httptest.NewRecorder()

		RejectChildHandler(svc, testLogger()).ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
		}
	})
}
