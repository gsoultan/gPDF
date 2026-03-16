package security

import "gpdf/model"

// Decryptor decrypts PDF strings and streams for a given object reference.
type Decryptor interface {
	// DecryptString decrypts an encrypted string. ref is the indirect object reference (object number, generation).
	DecryptString(ref model.Ref, ciphertext []byte) ([]byte, error)
	// DecryptStream decrypts an encrypted stream. ref is the indirect object reference.
	DecryptStream(ref model.Ref, ciphertext []byte) ([]byte, error)
}
