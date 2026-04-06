package security

import (
	"fmt"

	"github.com/gsoultan/gpdf/model"
)

// newStandardR5R6Decryptor builds a decryptor for R=5 (PDF 1.7 ext 3) and R=6 (PDF 2.0).
// Both revisions use AES-256 with a SHA-256 / PBKDF2 based key derivation.
// The existing aes256Decryptor (aes256_decryptor.go) implements the full algorithm.
func newStandardR5R6Decryptor(encryptDict model.Dict, id model.Array, userPassword string, r int) (Decryptor, error) {
	v := int(getInt(encryptDict, "V", 0))
	if !isSupportedV5R5R6(v, r) {
		return nil, fmt.Errorf("unsupported combination V=%d R=%d for AES-256 handler", v, r)
	}
	if err := validateAES256Dict(encryptDict); err != nil {
		return nil, err
	}
	return NewAES256Decryptor(encryptDict, id, userPassword)
}

func isSupportedV5R5R6(v, r int) bool {
	return v == 5 && (r == 5 || r == 6)
}

func validateAES256Dict(encryptDict model.Dict) error {
	filter, _ := encryptDict[model.Name("Filter")].(model.Name)
	if filter != "Standard" {
		return fmt.Errorf("r5/r6: unsupported Filter: %s", filter)
	}
	length := int(getInt(encryptDict, "Length", 256))
	if length != 256 {
		return fmt.Errorf("r5/r6: expected /Length 256, got %d", length)
	}
	return nil
}
