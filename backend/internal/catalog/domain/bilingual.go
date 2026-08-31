// Package domain — чистая доменная модель каталога учреждений EduHub: без зависимостей
// от транспорта (HTTP) или хранилища (БД), см. .claude/rules/architecture.md.
package domain

// Bilingual — двуязычное поле (ru/tg), зеркалит JSONB {ru,tg} в БД
// (docs/EduHub_Database_Schema.md, catalog.institutions и сателлиты).
type Bilingual struct {
	RU string
	TG string
}
