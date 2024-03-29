package symmetric

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"

	mbase "github.com/multiformats/go-multibase"
)

const (
	// NonceBytes is the length of GCM nonce.
	NonceBytes = 12

	// KeyBytes is the length of GCM key.
	KeyBytes = 32
)

var (
	cipherTextError = errors.New("malformed cipher text")
	invalidKeyError = errors.New("invalid key")
)

// Key is a wrapper for a symmetric key.
type Key struct {
	raw []byte
}

// NewRandom returns a random key.
func NewRandom() (*Key, error) {
	raw := make([]byte, KeyBytes)
	if _, err := rand.Read(raw); err != nil {
		return nil, err
	}
	return &Key{raw: raw}, nil
}

// New returns Key if err is nil and panics otherwise.
func New() *Key {
	k, err := NewRandom()
	if err != nil {
		panic(err)
	}
	return k
}

// FromBytes returns a key by decoding bytes.
func FromBytes(k []byte) (*Key, error) {
	if len(k) != KeyBytes {
		return nil, invalidKeyError
	}
	return &Key{raw: k}, nil
}

// FromString returns a key by decoding a base32-encoded string.
func FromString(k string) (*Key, error) {
	_, b, err := mbase.Decode(k)
	if err != nil {
		return nil, err
	}
	return FromBytes(b)
}

// Bytes returns raw key bytes.
func (k *Key) Bytes() []byte {
	return k.raw
}

// MarshalBinary implements BinaryMarshaler.
func (k *Key) MarshalBinary() ([]byte, error) {
	return k.raw, nil
}

// String returns the base32-encoded string representation of raw key bytes.
func (k *Key) String() string {
	str, err := mbase.Encode(mbase.Base32, k.raw)
	if err != nil {
		panic("should not error with hardcoded mbase: " + err.Error())
	}
	return str
}

// Encrypt performs AES-256 GCM encryption on plaintext.
func (k *Key) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(k.raw[:KeyBytes])
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, NonceBytes)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	ciphertext = append(nonce[:], ciphertext...)
	return ciphertext, nil
}

// Decrypt uses key to perform AES-256 GCM decryption on ciphertext.
func (k *Key) Decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(k.raw[:KeyBytes])
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < NonceBytes {
		return nil, cipherTextError
	}
	nonce := ciphertext[:NonceBytes]
	plain, err := aesgcm.Open(nil, nonce, ciphertext[NonceBytes:], nil)
	if err != nil {
		return nil, err
	}
	return plain, nil
}
