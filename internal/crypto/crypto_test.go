package crypto_test

import (
	"testing"

	"github.com/alrayyes/pipeline-analytics/internal/crypto"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecrypt(t *testing.T) {
	t.Parallel()

	t.Run("round trip recovers the plaintext", func(t *testing.T) {
		t.Parallel()

		key := make([]byte, 32)
		plaintext := []byte("ghp_supersecrettoken")

		ciphertext, err := crypto.Encrypt(key, plaintext)
		require.NoError(t, err)
		require.NotEqual(t, plaintext, ciphertext)

		got, err := crypto.Decrypt(key, ciphertext)
		require.NoError(t, err)
		require.Equal(t, plaintext, got)
	})

	t.Run("rejects a key of the wrong size", func(t *testing.T) {
		t.Parallel()

		_, err := crypto.Encrypt(make([]byte, 16), []byte("data"))
		require.ErrorIs(t, err, crypto.ErrInvalidKeySize)
	})

	t.Run("decrypt fails under the wrong key", func(t *testing.T) {
		t.Parallel()

		key := make([]byte, 32)
		wrongKey := make([]byte, 32)
		wrongKey[0] = 1

		ciphertext, err := crypto.Encrypt(key, []byte("data"))
		require.NoError(t, err)

		_, err = crypto.Decrypt(wrongKey, ciphertext)
		require.Error(t, err)
	})
}

func TestRandomHex(t *testing.T) {
	t.Parallel()

	a, err := crypto.RandomHex(32)
	require.NoError(t, err)
	require.Len(t, a, 64)

	b, err := crypto.RandomHex(32)
	require.NoError(t, err)
	require.NotEqual(t, a, b)
}
