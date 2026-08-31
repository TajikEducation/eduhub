// Package domain — доменная модель модуля moderation: запись журнала решений модерации
// (moderation.audit_log). Не общий лог всех действий пользователей — только модераторские
// мутации (см. SRS Веха 3, критерии готовности).
package domain

import "github.com/google/uuid"

// Entry — одна запись moderation.audit_log.
type Entry struct {
	ActorType  string // "user" | "system"
	ActorID    *uuid.UUID
	ActorRole  string
	Action     string // например "approve_institution", "reject_institution"
	TargetType string // например "institution"
	TargetID   uuid.UUID
	ReasonCode *string
	ReasonText *string
	RequestID  string
}
