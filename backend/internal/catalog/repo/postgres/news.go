package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/abdulhalim/eduhub/backend/internal/catalog/domain"
	"github.com/abdulhalim/eduhub/backend/internal/platform/apperr"
)

func bilingualSliceJSONOf(items []domain.Bilingual) []byte {
	wire := make([]bilingualJSON, len(items))
	for i, b := range items {
		wire[i] = bilingualJSON{RU: b.RU, TG: b.TG}
	}
	raw, _ := json.Marshal(wire)
	return raw
}

const newsColumns = `id, institution_id, title, category, cover_s3_key, video_url, content, tags, status, views_count, created_at, updated_at`

// scanNewsRow разбирает одну строку catalog.news_articles (newsColumns) — общий код для
// ListNews/ListPublishedNews (pgx.Rows) и GetPublishedNewsByID (pgx.Row), обе реализуют Scan.
func scanNewsRow(row interface{ Scan(dest ...any) error }) (domain.NewsArticle, error) {
	var n domain.NewsArticle
	var titleRaw, categoryRaw, contentRaw, tagsRaw []byte
	if err := row.Scan(&n.ID, &n.InstitutionID, &titleRaw, &categoryRaw, &n.CoverS3Key, &n.VideoURL, &contentRaw, &tagsRaw, &n.Status, &n.ViewsCount, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return domain.NewsArticle{}, err
	}
	title, err := scanBilingual(titleRaw)
	if err != nil {
		return domain.NewsArticle{}, fmt.Errorf("unmarshal news title: %w", err)
	}
	n.Title = title
	category, err := scanBilingualPtr(categoryRaw)
	if err != nil {
		return domain.NewsArticle{}, fmt.Errorf("unmarshal news category: %w", err)
	}
	n.Category = category
	content, err := scanBilingual(contentRaw)
	if err != nil {
		return domain.NewsArticle{}, fmt.Errorf("unmarshal news content: %w", err)
	}
	n.Content = content
	if len(tagsRaw) > 0 {
		var tagsWire []bilingualJSON
		if err := json.Unmarshal(tagsRaw, &tagsWire); err != nil {
			return domain.NewsArticle{}, fmt.Errorf("unmarshal news tags: %w", err)
		}
		n.Tags = make([]domain.Bilingual, len(tagsWire))
		for i, tg := range tagsWire {
			n.Tags[i] = domain.Bilingual{RU: tg.RU, TG: tg.TG}
		}
	}
	return n, nil
}

func (r *InstitutionRepo) listNewsQuery(ctx context.Context, query string, args ...any) ([]domain.NewsArticle, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("postgres: list news: %w", err)
	}
	defer rows.Close()

	var out []domain.NewsArticle
	for rows.Next() {
		n, err := scanNewsRow(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres: scan news: %w", err)
		}
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: list news rows: %w", err)
	}
	return out, nil
}

// ListNews возвращает новости учреждения (все статусы — для кабинета учреждения, не для
// публичной страницы, которая по FR должна фильтровать status='published').
func (r *InstitutionRepo) ListNews(ctx context.Context, institutionID uuid.UUID) ([]domain.NewsArticle, error) {
	return r.listNewsQuery(ctx, `SELECT `+newsColumns+` FROM catalog.news_articles WHERE institution_id = $1 ORDER BY created_at DESC`, institutionID)
}

// ListPublishedNews возвращает только опубликованные новости учреждения — для публичной
// вкладки профиля (FR-24).
func (r *InstitutionRepo) ListPublishedNews(ctx context.Context, institutionID uuid.UUID) ([]domain.NewsArticle, error) {
	return r.listNewsQuery(ctx, `SELECT `+newsColumns+` FROM catalog.news_articles WHERE institution_id = $1 AND status = 'published' ORDER BY created_at DESC`, institutionID)
}

// GetPublishedNewsByID возвращает одну опубликованную новость и атомарно увеличивает
// views_count — для публичной страницы /news/{id}.
func (r *InstitutionRepo) GetPublishedNewsByID(ctx context.Context, id uuid.UUID) (domain.NewsArticle, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE catalog.news_articles SET views_count = views_count + 1
		WHERE id = $1 AND status = 'published'
		RETURNING `+newsColumns, id)
	n, err := scanNewsRow(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NewsArticle{}, apperr.NotFound("news", id.String())
		}
		return domain.NewsArticle{}, fmt.Errorf("postgres: get published news: %w", err)
	}
	return n, nil
}

// CreateNews вставляет новость учреждения.
func (r *InstitutionRepo) CreateNews(ctx context.Context, institutionID uuid.UUID, in domain.CreateNewsInput) (domain.NewsArticle, error) {
	row := r.db.QueryRow(ctx, `
		INSERT INTO catalog.news_articles (institution_id, title, category, cover_s3_key, video_url, content, tags, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`, institutionID, bilingualJSONOf(in.Title), bilingualJSONPtrOf(in.Category), in.CoverS3Key, in.VideoURL, bilingualJSONOf(in.Content), bilingualSliceJSONOf(in.Tags), in.Status)

	n := domain.NewsArticle{InstitutionID: institutionID, Title: in.Title, Category: in.Category, CoverS3Key: in.CoverS3Key, VideoURL: in.VideoURL, Content: in.Content, Tags: in.Tags, Status: in.Status}
	if err := row.Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt); err != nil {
		return domain.NewsArticle{}, fmt.Errorf("postgres: insert news: %w", err)
	}
	return n, nil
}

// UpdateNews заменяет данные новости (полная замена).
func (r *InstitutionRepo) UpdateNews(ctx context.Context, id uuid.UUID, in domain.CreateNewsInput) (domain.NewsArticle, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE catalog.news_articles SET
			title = $2, category = $3, cover_s3_key = $4, video_url = $5, content = $6, tags = $7, status = $8, updated_at = now()
		WHERE id = $1
		RETURNING created_at, updated_at
	`, id, bilingualJSONOf(in.Title), bilingualJSONPtrOf(in.Category), in.CoverS3Key, in.VideoURL, bilingualJSONOf(in.Content), bilingualSliceJSONOf(in.Tags), in.Status)

	n := domain.NewsArticle{ID: id, Title: in.Title, Category: in.Category, CoverS3Key: in.CoverS3Key, VideoURL: in.VideoURL, Content: in.Content, Tags: in.Tags, Status: in.Status}
	if err := row.Scan(&n.CreatedAt, &n.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.NewsArticle{}, apperr.NotFound("news", id.String())
		}
		return domain.NewsArticle{}, fmt.Errorf("postgres: update news: %w", err)
	}
	return n, nil
}

// DeleteNews удаляет новость.
func (r *InstitutionRepo) DeleteNews(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM catalog.news_articles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("postgres: delete news: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return apperr.NotFound("news", id.String())
	}
	return nil
}

// GetNewsInstitutionID возвращает institution_id владеющей записи.
func (r *InstitutionRepo) GetNewsInstitutionID(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	row := r.db.QueryRow(ctx, `SELECT institution_id FROM catalog.news_articles WHERE id = $1`, id)
	var instID uuid.UUID
	if err := row.Scan(&instID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, apperr.NotFound("news", id.String())
		}
		return uuid.Nil, fmt.Errorf("postgres: get news institution id: %w", err)
	}
	return instID, nil
}
