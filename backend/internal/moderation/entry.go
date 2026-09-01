// Package moderation — журнал решений модерации (moderation.audit_log), инфраструктура
// заранее для будущих сервисов модерации (веха 3+) — тот же прецедент, что RequireRole/E2.2/E2.3.
package moderation

import (
	"github.com/google/uuid"
)

// Entry — одна запись журнала модерации. ActorID/ActorRole nil для системных действий
// (actor_type='system'); ReasonCode/ReasonText nil, если решение не требует причины (approve).
type Entry struct {
	ActorID     *uuid.UUID
	ActorType   string
	ActorRole   *string
	Action      string
	TargetType  string
	TargetID    uuid.UUID
	ReasonCode  *string
	ReasonText  *string
	PayloadDiff []byte
	RequestID   string
}
