package crypto

import (
	"testing"
)

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name       string
		plaintext  string
		passphrase string
	}{
		{"simple", "hello world", "secret"},
		{"empty", "", "secret"},
		{"unicode", "entry with @tags and emojis 📓", "p@ss!"},
		{"multiline", "# 2026-03-29 Saturday\n\n## [05:13 PM]\n\nJournal entry.\n", "pass"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := Encrypt([]byte(tt.plaintext), tt.passphrase)
			if err != nil {
				t.Fatalf("Encrypt: %v", err)
			}

			if tt.plaintext != "" && string(enc) == tt.plaintext {
				t.Fatal("ciphertext should differ from plaintext")
			}

			dec, err := Decrypt(enc, tt.passphrase)
			if err != nil {
				t.Fatalf("Decrypt: %v", err)
			}

			if string(dec) != tt.plaintext {
				t.Errorf("got %q, want %q", dec, tt.plaintext)
			}
		})
	}
}

func TestDecryptWrongPassphrase(t *testing.T) {
	enc, err := Encrypt([]byte("secret data"), "correct")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	_, err = Decrypt(enc, "wrong")
	if err == nil {
		t.Fatal("expected error decrypting with wrong passphrase")
	}
}
