package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/authorizerdev/authorizer/server/constants"
	"github.com/authorizerdev/authorizer/server/envstore"
)

// EncryptAES encrypts data using AES algorithm
func EncryptAES(text []byte) ([]byte, error) {
	key := []byte(envstore.EnvStoreObj.GetStringStoreEnvVariable(constants.EnvKeyEncryptionKey))
	c, err := aes.NewCipher(key)
	var res []byte
	if err != nil {
		return res, err
	}

	// gcm or Galois/Counter Mode, is a mode of operation
	// for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return res, err
	}

	// creates a new byte array the size of the nonce
	// which must be passed to Seal
	nonce := make([]byte, gcm.NonceSize())
	// populates our nonce with a cryptographically secure
	// random sequence
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return res, err
	}

	// here we encrypt our text using the Seal function
	// Seal encrypts and authenticates plaintext, authenticates the
	// additional data and appends the result to dst, returning the updated
	// slice. The nonce must be NonceSize() bytes long and unique for all
	// time, for a given key.
	return gcm.Seal(nonce, nonce, text, nil), nil
}

// DecryptAES decrypts data using AES algorithm
func DecryptAES(ciphertext []byte) ([]byte, error) {
	key := []byte(envstore.EnvStoreObj.GetStringStoreEnvVariable(constants.EnvKeyEncryptionKey))
	c, err := aes.NewCipher(key)
	var res []byte
	if err != nil {
		return res, err
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return res, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return res, err
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return res, err
	}

	return plaintext, nil
}