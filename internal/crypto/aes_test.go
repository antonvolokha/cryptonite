package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	testCases := []struct {
		name     string
		data     []byte
		password string
	}{
		{
			name:     "Simple string",
			data:     []byte("Hello, World!"),
			password: "test123",
		},
		{
			name:     "Empty data",
			data:     []byte{},
			password: "test123",
		},
		{
			name:     "Binary data",
			data:     []byte{0x00, 0x01, 0x02, 0x03},
			password: "test123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := Encrypt(tc.data, tc.password)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			decrypted, err := Decrypt(encrypted, tc.password)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if !bytes.Equal(decrypted, tc.data) {
				t.Errorf("Decrypted data doesn't match original.\nGot: %v\nWant: %v", decrypted, tc.data)
			}
		})
	}

	// Test wrong password
	t.Run("Wrong password", func(t *testing.T) {
		data := []byte("test data")
		encrypted, err := Encrypt(data, "correct")
		if err != nil {
			t.Fatalf("Encrypt failed: %v", err)
		}

		_, err = Decrypt(encrypted, "wrong")
		if err == nil {
			t.Error("Expected error with wrong password, got nil")
		}
	})
}
