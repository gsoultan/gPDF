package security

import "gpdf/model"

// Encryptor encrypts PDF strings and streams for a given object reference (used when writing).
type Encryptor interface {
	EncryptString(ref model.Ref, plaintext []byte) ([]byte, error)
	EncryptStream(ref model.Ref, plaintext []byte) ([]byte, error)
}
