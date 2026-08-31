package usecase

// RatingSync — порт синхронизации агрегированных метрик рейтинга в
// catalog.institution_metrics (веха 4: reviews.institution_rating_agg → RatingSync).
// Пока не используется нигде в Service — только объявление точки стыковки,
// реализация и метод(ы) появятся в вехе 4, когда будет реальный вызывающий и тест.
type RatingSync interface {
}
