package auth

import (
	"strings"
	"testing"
)

func TestHashAndVerifyPassword(t *testing.T) {
	encodedHash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}

	valid, err := VerifyPassword("correct horse battery staple", encodedHash)
	if err != nil || !valid {
		t.Fatalf("correct password was rejected: valid=%v err=%v", valid, err)
	}

	valid, err = VerifyPassword("wrong password", encodedHash)
	if err != nil {
		t.Fatal(err)
	}
	if valid {
		t.Fatal("wrong password was accepted")
	}
}

func TestHashPasswordUsesRandomSalt(t *testing.T) {
	first, err := HashPassword("same password")
	if err != nil {
		t.Fatal(err)
	}
	second, err := HashPassword("same password")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("two password hashes unexpectedly matched")
	}
}

func TestPasswordHashValidation(t *testing.T) {
	if _, err := HashPassword(""); err == nil {
		t.Fatal("empty password was accepted")
	}

	invalidHashes := []string{
		"",
		"not-a-hash",
		"$argon2i$v=19$m=19456,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"$argon2id$v=999$m=19456,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"$argon2id$v=19$m=0,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"$argon2id$v=19$m=999999999,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"$argon2id$v=19$m=19456,t=2,p=1$invalid!$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"$argon2id$v=19$m=19456,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$invalid!",
	}
	for _, encodedHash := range invalidHashes {
		t.Run(strings.ReplaceAll(encodedHash, "/", "_"), func(t *testing.T) {
			if _, err := VerifyPassword("password", encodedHash); err == nil {
				t.Fatalf("accepted invalid hash %q", encodedHash)
			}
		})
	}
}
