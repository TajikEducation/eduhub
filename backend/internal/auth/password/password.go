// Package password — хеширование паролей argon2id в PHC-формате (E2.2, веха 2).
// Параметры стоимости хранятся в самой PHC-строке хеша, поэтому смену параметров в конфиге
// (например при увеличении стоимости со временем) не требует перехеширования существующих
// паролей — Verify всегда использует параметры, зашитые в проверяемый хеш, не текущий конфиг.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Params — параметры стоимости argon2id.
type Params struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultParams — параметры по умолчанию (OWASP Password Storage Cheat Sheet, вариант
// "высокая стойкость"). Настраиваются через конфиг — см. internal/platform/config.
var DefaultParams = Params{
	MemoryKiB:   64 * 1024, // 64 MiB
	Iterations:  3,
	Parallelism: 4,
	SaltLength:  16,
	KeyLength:   32,
}

// ErrMalformedHash — переданная строка не является валидным PHC-хешем argon2id.
var ErrMalformedHash = errors.New("password: malformed argon2id hash")

// Hasher хеширует пароли с фиксированным набором параметров стоимости.
type Hasher struct {
	params Params
}

// New создаёт Hasher с заданными параметрами стоимости.
func New(params Params) *Hasher {
	return &Hasher{params: params}
}

// Hash хеширует plaintext-пароль в PHC-строку argon2id со свежей случайной солью.
func (h *Hasher) Hash(plaintext string) (string, error) {
	salt := make([]byte, h.params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("password: generate salt: %w", err)
	}

	key := argon2.IDKey([]byte(plaintext), salt, h.params.Iterations, h.params.MemoryKiB, h.params.Parallelism, h.params.KeyLength)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.params.MemoryKiB, h.params.Iterations, h.params.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// Verify сверяет plaintext-пароль с PHC-хешем argon2id. Параметры стоимости и соль берутся
// ИЗ САМОГО хеша, не из текущего конфига — старые хеши остаются проверяемыми после смены
// параметров по умолчанию. Свободная функция, не метод Hasher — намеренно: Hasher.params
// здесь не участвует.
func Verify(phc, plaintext string) (bool, error) {
	parts := strings.Split(phc, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" {
		return false, ErrMalformedHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, ErrMalformedHash
	}
	if version != argon2.Version {
		return false, ErrMalformedHash
	}

	var memoryKiB, iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memoryKiB, &iterations, &parallelism); err != nil {
		return false, ErrMalformedHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, ErrMalformedHash
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, ErrMalformedHash
	}

	//nolint:gosec // len(want) — длина декодированного хеша (десятки байт), не пользовательский
	// ввод произвольного размера; переполнение uint32 физически недостижимо.
	got := argon2.IDKey([]byte(plaintext), salt, iterations, memoryKiB, parallelism, uint32(len(want)))

	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
