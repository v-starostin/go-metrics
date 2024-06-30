package crypto_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/v-starostin/go-metrics/internal/crypto"
)

func TestLoadPrivateKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	t.Run("good case", func(t *testing.T) {
		tempFile := createTempFile(t, privateKeyPEM)
		defer os.Remove(tempFile.Name())

		_, err := crypto.LoadPrivateKey(tempFile.Name())
		if err != nil {
			t.Fatalf("Expected nil err, got %v", err)
		}
	})

	t.Run("attempt to load invalid file", func(t *testing.T) {
		_, err := crypto.LoadPrivateKey("nonexistent-file.pem")
		if err == nil {
			t.Fatal("Expected error when loading nonexistent file, but got nil")
		}

		expectedErrMsg := "open nonexistent-file.pem: no such file or directory"
		if err.Error() != expectedErrMsg {
			t.Errorf("Expected error %s, got %s", expectedErrMsg, err.Error())
		}
	})

	t.Run("attempt to load invalid PEM", func(t *testing.T) {
		tempFileInvalid := createTempFile(t, []byte("invalid-pem-data"))
		defer os.Remove(tempFileInvalid.Name())

		_, err := crypto.LoadPrivateKey(tempFileInvalid.Name())
		if err == nil {
			t.Fatal("Expected error, but got nil")
		}
		expectedErrMsg := "failed to decode PEM block containing private key"
		if err.Error() != expectedErrMsg {
			t.Errorf("Expected error message %s, got %s", expectedErrMsg, err.Error())
		}
	})
}

func TestLoadPublicKey(t *testing.T) {
	t.Run("good case", func(t *testing.T) {
		tempFile := createTempFile(t, []byte(validPublicKeyPEM))
		defer os.Remove(tempFile.Name())

		_, err := crypto.LoadPublicKey(tempFile.Name())
		if err != nil {
			t.Fatalf("Expected nil err, but got %v", err)
		}
	})

	t.Run("attempt to load invalid file", func(t *testing.T) {
		_, err := crypto.LoadPublicKey("nonexistent-file.pem")
		if err == nil {
			t.Fatal("Expected error, but got nil")
		}
		expectedErrMsg := "open nonexistent-file.pem: no such file or directory"
		if err.Error() != expectedErrMsg {
			t.Errorf("Expected error %s, got %s", expectedErrMsg, err.Error())
		}
	})

	t.Run("attempt to load invalid PEM", func(t *testing.T) {
		tempFileInvalid := createTempFile(t, []byte("invalid-pem-data"))
		defer os.Remove(tempFileInvalid.Name())

		_, err := crypto.LoadPublicKey(tempFileInvalid.Name())
		if err == nil {
			t.Fatal("Expected error, but got nil")
		}
		expectedErrMsg := "failed to decode PEM block containing public key"
		if err.Error() != expectedErrMsg {
			t.Errorf("Expected error %s, got %s", expectedErrMsg, err.Error())
		}
	})
}

// Helper function to create a temporary file for testing
func createTempFile(t *testing.T, content []byte) *os.File {
	tempFile, err := os.CreateTemp("", "test-private-key-*.pem")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer tempFile.Close()

	if _, err := tempFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temporary file: %v", err)
	}

	return tempFile
}

const validPublicKeyPEM = `
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAvlV4wMy3Lrziw2K3+jVt
0dd/UvU8zJ6B2FZJqvLuwVX8lWmT9bJH8GHF//d4Q5TNT+R37s1Z7PjYMFZz5lmL
fAKSZt44V7J2kcObmI5ZqYslJLKDGay1F6Fo8z6dSLTy5x6T5vW0aMnQJtDq0U1X
zm7V/MW3MPjToFP4TZLD4Zhw39KnDk2D5DF+BlyyWZytw5HFF9S2xI6LKKFgcsVa
URkxlRz56i5qVt0DqflhdCh/5hOp9yxAhE4hwtnXb2nRtS3WxTy6RBR7fN/cKrRL
/v6+rd/MYcH71eIgMb5X5C0CVFKaZRsFjKu6b/fZKqkMogohp2m6YOmzC/5DfQwT
2QIDAQAB
-----END PUBLIC KEY-----
`
