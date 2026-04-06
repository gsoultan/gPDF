package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/pbkdf2"

	"github.com/gsoultan/gpdf/model"
)

// aes256Decryptor mirrors aes256Encryptor for the read path.
type aes256Decryptor struct {
	key []byte
}

// NewAES256Decryptor constructs a decryptor from the Encrypt dict and trailer ID using
// the same PBKDF2-SHA256 scheme as BuildAES256EncryptDictForWrite.
func NewAES256Decryptor(encryptDict model.Dict, id model.Array, userPassword string) (Decryptor, error) {
	if encryptDict == nil {
		return nil, fmt.Errorf("encrypt dict is nil")
	}
	v := getInt(encryptDict, "V", 0)
	r := getInt(encryptDict, "R", 0)
	if v != 5 || r != 6 {
		return nil, fmt.Errorf("unsupported AES256 handler V=%d R=%d", v, r)
	}
	if id == nil || len(id) == 0 {
		return nil, fmt.Errorf("missing trailer ID for AES256 decryptor")
	}
	s, ok := id[0].(model.String)
	if !ok {
		return nil, fmt.Errorf("trailer ID[0] is not a string")
	}
	idBytes := []byte(s)
	salt := sha256.Sum256(idBytes)
	fileKey := pbkdf2.Key([]byte(userPassword), salt[:], 100_000, 32, sha256.New)
	return &aes256Decryptor{key: fileKey}, nil
}

func (d *aes256Decryptor) DecryptString(ref model.Ref, ciphertext []byte) ([]byte, error) {
	return d.decrypt(ref, ciphertext)
}

func (d *aes256Decryptor) DecryptStream(ref model.Ref, ciphertext []byte) ([]byte, error) {
	return d.decrypt(ref, ciphertext)
}

func (d *aes256Decryptor) decrypt(_ model.Ref, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return ciphertext, nil
	}
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	body := ciphertext[aes.BlockSize:]
	if len(body)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("ciphertext length not multiple of block size")
	}
	block, err := aes.NewCipher(d.key)
	if err != nil {
		return nil, err
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plain := make([]byte, len(body))
	mode.CryptBlocks(plain, body)
	return pkcs7Unpad(plain, aes.BlockSize)
}

var _ Decryptor = (*aes256Decryptor)(nil)
