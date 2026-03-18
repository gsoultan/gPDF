package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/pbkdf2"

	"gpdf/model"
)

// aes256Encryptor implements Encryptor using AES-256-CBC with a per-object IV.
type aes256Encryptor struct {
	key []byte
}

// BuildAES256EncryptDictForWrite creates an Encrypt dictionary and Encryptor for AES-256.
func BuildAES256EncryptDictForWrite(userPassword, ownerPassword string, id []byte, P int32) (model.Dict, Encryptor, error) {
	if id == nil || len(id) == 0 {
		id = make([]byte, 16)
		if _, err := rand.Read(id); err != nil {
			return nil, nil, err
		}
	}
	salt := sha256.Sum256(id)
	fileKey := pbkdf2.Key([]byte(userPassword), salt[:], 100_000, 32, sha256.New)

	uHash := sha256.Sum256(append([]byte(userPassword), id...))
	oHash := sha256.Sum256(append([]byte(ownerPassword), id...))

	dict := model.Dict{
		model.Name("Filter"):          model.Name("Standard"),
		model.Name("V"):               model.Integer(5),
		model.Name("R"):               model.Integer(6),
		model.Name("Length"):          model.Integer(256),
		model.Name("O"):               model.String(string(oHash[:])),
		model.Name("U"):               model.String(string(uHash[:])),
		model.Name("P"):               model.Integer(int64(P)),
		model.Name("EncryptMetadata"): model.Boolean(true),
	}
	return dict, &aes256Encryptor{key: fileKey}, nil
}

func (e *aes256Encryptor) EncryptString(ref model.Ref, plaintext []byte) ([]byte, error) {
	return e.encrypt(ref, plaintext)
}

func (e *aes256Encryptor) EncryptStream(ref model.Ref, plaintext []byte) ([]byte, error) {
	return e.encrypt(ref, plaintext)
}

func (e *aes256Encryptor) encrypt(_ model.Ref, plaintext []byte) ([]byte, error) {
	if len(plaintext) == 0 {
		return plaintext, nil
	}
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, err
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}
	padded := pkcs7Pad(plaintext, aes.BlockSize)
	ciphertext := make([]byte, len(iv)+len(padded))
	copy(ciphertext, iv)
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[len(iv):], padded)
	return ciphertext, nil
}

var _ Encryptor = (*aes256Encryptor)(nil)
