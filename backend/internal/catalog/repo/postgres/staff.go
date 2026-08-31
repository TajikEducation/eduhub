package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

// CreateStaff вставляет запись сотрудника.
func (r *InstitutionRepo) CreateStaff(ctx context.Context, institutionID uuid.UUID, in domain.CreateStaffInput) (domain.StaffMember, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO catalog.institution_staff (institution_id, name, role_type, role_label, subject, photo_url, exp, bio, email, phone)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`, institutionID, bilingualJSONOf(in.Name), in.RoleType, bilingualJSONOf(in.RoleLabel), bilingualJSONPtrOf(in.Subject), in.PhotoURL, in.Exp, bilingualJSONPtrOf(in.Bio), in.Email, in.Phone)

	m := domain.StaffMember{Name: in.Name, RoleType: in.RoleType, RoleLabel: in.RoleLabel, Subject: in.Subject, PhotoURL: in.PhotoURL, Exp: in.Exp, Bio: in.Bio, Email: in.Email, Phone: in.Phone}
	if err := row.Scan(&m.ID, &m.CreatedAt); err != nil {
		return domain.StaffMember{}, fmt.Errorf("postgres: insert staff: %w", err)
	}
	return m, nil
}

// UpdateStaff заменяет данные сотрудника (полная замена, не partial-patch).
func (r *InstitutionRepo) UpdateStaff(ctx context.Context, id uuid.UUID, in domain.CreateStaffInput) (domain.StaffMember, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE catalog.institution_staff SET
			name = $2, role_type = $3, role_label = $4, subject = $5, photo_url = $6, exp = $7, bio = $8, email = $9, phone = $10
		WHERE id = $1
	`, id, bilingualJSONOf(in.Name), in.RoleType, bilingualJSONOf(in.RoleLabel), bilingualJSONPtrOf(in.Subject), in.PhotoURL, in.Exp, bilingualJSONPtrOf(in.Bio), in.Email, in.Phone)
	if err != nil {
		return domain.StaffMember{}, fmt.Errorf("postgres: update staff: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.StaffMember{}, apperr.NotFound("staff", id.String())
	}
	return domain.StaffMember{ID: id, Name: in.Name, RoleType: in.RoleType, RoleLabel: in.RoleLabel, Subject: in.Subject, PhotoURL: in.PhotoURL, Exp: in.Exp, Bio: in.Bio, Email: in.Email, Phone: in.Phone}, nil
}

// DeleteStaff удаляет запись сотрудника.
func (r *InstitutionRepo) DeleteStaff(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM catalog.institution_staff WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("postgres: delete staff: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound("staff", id.String())
	}
	return nil
}

// GetStaffInstitutionID возвращает institution_id владеющей записи (для проверки прав доступа
// в usecase — тот же паттерн, что GetOwnerID для институций).
func (r *InstitutionRepo) GetStaffInstitutionID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	row := r.db.QueryRow(ctx, `SELECT institution_id FROM catalog.institution_staff WHERE id = $1`, id)
	var instID uuid.UUID
	if err := row.Scan(&instID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, apperr.NotFound("staff", id.String())
		}
		return uuid.Nil, fmt.Errorf("postgres: get staff institution id: %w", err)
	}
	return instID, nil
}

// GetPublicStaffByID возвращает сотрудника по id — только если его институция approved (та же
// видимость, что Service.Get для самой институции). JOIN в пределах одной схемы (catalog),
// физический FK institution_staff.institution_id → institutions(id) это разрешает.
func (r *InstitutionRepo) GetPublicStaffByID(ctx context.Context, id uuid.UUID) (domain.StaffMember, error) {
	row := r.db.QueryRow(ctx, `
		SELECT s.id, s.institution_id, s.name, s.role_type, s.role_label, s.subject, s.photo_url, s.exp, s.bio, s.email, s.phone, s.created_at
		FROM catalog.institution_staff s
		JOIN catalog.institutions i ON i.id = s.institution_id
		WHERE s.id = $1 AND i.moderation_status = 'approved'
	`, id)

	var m domain.StaffMember
	var nameRaw, roleLabelRaw, subjectRaw, bioRaw []byte
	if err := row.Scan(&m.ID, &m.InstitutionID, &nameRaw, &m.RoleType, &roleLabelRaw, &subjectRaw, &m.PhotoURL, &m.Exp, &bioRaw, &m.Email, &m.Phone, &m.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.StaffMember{}, apperr.NotFound("staff", id.String())
		}
		return domain.StaffMember{}, fmt.Errorf("postgres: get public staff: %w", err)
	}
	name, err := scanBilingual(nameRaw)
	if err != nil {
		return domain.StaffMember{}, fmt.Errorf("postgres: unmarshal staff name: %w", err)
	}
	m.Name = name
	roleLabel, err := scanBilingual(roleLabelRaw)
	if err != nil {
		return domain.StaffMember{}, fmt.Errorf("postgres: unmarshal staff role_label: %w", err)
	}
	m.RoleLabel = roleLabel
	subject, err := scanBilingualPtr(subjectRaw)
	if err != nil {
		return domain.StaffMember{}, fmt.Errorf("postgres: unmarshal staff subject: %w", err)
	}
	m.Subject = subject
	bio, err := scanBilingualPtr(bioRaw)
	if err != nil {
		return domain.StaffMember{}, fmt.Errorf("postgres: unmarshal staff bio: %w", err)
	}
	m.Bio = bio
	return m, nil
}
