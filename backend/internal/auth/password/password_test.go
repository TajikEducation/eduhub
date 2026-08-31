package password_test

import (
	"strings"
	"testing"

	"github.com/abdulhalim/eduhub/backend/internal/auth/password"
)

// testParams — заниженные параметры стоимости ради скорости юнит-тестов, не для прода.
var testParams = password.Params{MemoryKiB: 8 * 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}

func TestHashVerify_CorrectPassword_ReturnsTrue(t *testing.T) {
	hasher := password.New(testParams)

	hash, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() вернул ошибку: %v", err)
	}

	ok, err := password.Verify(hash, "correct horse battery staple")
	if err != nil {
		t.Fatalf("Verify() вернул ошибку: %v", err)
	}
	if !ok {
		t.Error("Verify() = false для правильного пароля, want true")
	}
}

func TestHashVerify_WrongPassword_ReturnsFalseNoError(t *testing.T) {
	hasher := password.New(testParams)

	hash, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("Hash() вернул ошибку: %v", err)
	}

	ok, err := password.Verify(hash, "неверный пароль")
	if err != nil {
		t.Fatalf("Verify() с неверным паролем вернул ошибку %v, ожидали (false, nil)", err)
	}
	if ok {
		t.Error("Verify() = true для неверного пароля, want false")
	}
}

func TestHash_SamePasswordTwice_ProducesDifferentHashes(t *testing.T) {
	hasher := password.New(testParams)

	hash1, err := hasher.Hash("тот же пароль")
	if err != nil {
		t.Fatalf("Hash() (1) вернул ошибку: %v", err)
	}
	hash2, err := hasher.Hash("тот же пароль")
	if err != nil {
		t.Fatalf("Hash() (2) вернул ошибку: %v", err)
	}

	if hash1 == hash2 {
		t.Error("два хеша одного пароля идентичны — соль не рандомизируется")
	}

	// Оба хеша всё равно должны проходить верификацию исходным паролем.
	for _, h := range []string{hash1, hash2} {
		ok, err := password.Verify(h, "тот же пароль")
		if err != nil {
			t.Fatalf("Verify(%q) вернул ошибку: %v", h, err)
		}
		if !ok {
			t.Errorf("Verify(%q) = false, want true", h)
		}
	}
}

func TestVerify_MalformedHash_ReturnsErrorNotPanic(t *testing.T) {
	cases := []string{
		"",
		"не-phc-строка-вообще",
		"$argon2id$v=19$m=8192,t=1,p=1$только-соль-без-хеша",
		"$bcrypt$v=1$m=8192,t=1,p=1$c2FsdA$aGFzaA", // не argon2id
	}

	for _, malformed := range cases {
		_, err := password.Verify(malformed, "любой-пароль")
		if err == nil {
			t.Errorf("Verify(%q) вернул nil-ошибку для битого хеша, ожидали ошибку", malformed)
		}
	}
}

func TestVerify_HashParamsHigherThanCurrentDefaults_StillVerifies(t *testing.T) {
	// Ключевое требование плана (E2.2): "параметры можно поднять без слома существующих
	// хешей" — PHC-строка несёт свои параметры, Verify их не берёт из текущего конфига.
	oldParams := password.Params{MemoryKiB: 8 * 1024, Iterations: 1, Parallelism: 1, SaltLength: 16, KeyLength: 32}
	oldHasher := password.New(oldParams)

	hash, err := oldHasher.Hash("пароль до апгрейда параметров")
	if err != nil {
		t.Fatalf("Hash() вернул ошибку: %v", err)
	}

	// "Поднятые" параметры для НОВЫХ хешей — Verify старого хеша не должен их использовать.
	newParams := password.Params{MemoryKiB: 64 * 1024, Iterations: 3, Parallelism: 4, SaltLength: 16, KeyLength: 32}
	_ = password.New(newParams) // новый Hasher существует, но Verify — свободная функция

	ok, err := password.Verify(hash, "пароль до апгрейда параметров")
	if err != nil {
		t.Fatalf("Verify() старого хеша после апгрейда параметров вернул ошибку: %v", err)
	}
	if !ok {
		t.Error("Verify() старого хеша после апгрейда параметров = false, want true — хеш должен остаться валиден")
	}
}

func TestHash_ProducesPHCFormatWithArgon2idTag(t *testing.T) {
	hasher := password.New(testParams)

	hash, err := hasher.Hash("проверка формата")
	if err != nil {
		t.Fatalf("Hash() вернул ошибку: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$v=") {
		t.Errorf("Hash() = %q, ожидали префикс PHC-формата $argon2id$v=", hash)
	}
	if strings.Contains(hash, "проверка формата") {
		t.Error("хеш содержит plaintext пароля буквально — недопустимо")
	}
}
