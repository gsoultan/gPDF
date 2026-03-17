package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"

	"gpdf/model"
)

// aes256Encryptor implements Encryptor using AES-256-CBC with a per-object IV.
// This is a simplified AES-based handler intended to provide stronger crypto
// than RC4 while keeping the API surface small and pluggable.
type aes256Encryptor struct {
	key []byte
}

// aes256Decryptor mirrors aes256Encryptor for the read path.
type aes256Decryptor struct {
	key []byte
}

// BuildAES256EncryptDictForWrite creates an Encrypt dictionary and Encryptor for AES-256.
// This uses a PBKDF2-SHA256 based key derivation for the file encryption key.
func BuildAES256EncryptDictForWrite(userPassword, ownerPassword string, id []byte, P int32) (model.Dict, Encryptor, error) {
	if id == nil || len(id) == 0 {
		id = make([]byte, 16)
		if _, err := rand.Read(id); err != nil {
			return nil, nil, err
		}
	}
	// Derive a 32-byte file key from the user password and file ID.
	salt := sha256.Sum256(id)
	fileKey := pbkdf2.Key([]byte(userPassword), salt[:], 100_000, 32, sha256.New)

	// For simplicity, store validation hashes for user/owner instead of full PDF 2.0 O/U/OE/UE fields.
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

func (e *aes256Encryptor) EncryptString(ref model.Ref, plaintext []byte) ([]byte, error) {
	return e.encrypt(ref, plaintext)
}

func (e *aes256Encryptor) EncryptStream(ref model.Ref, plaintext []byte) ([]byte, error) {
	return e.encrypt(ref, plaintext)
}

func (d *aes256Decryptor) DecryptString(ref model.Ref, ciphertext []byte) ([]byte, error) {
	return d.decrypt(ref, ciphertext)
}

func (d *aes256Decryptor) DecryptStream(ref model.Ref, ciphertext []byte) ([]byte, error) {
	return d.decrypt(ref, ciphertext)
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

func pkcs7Pad(b []byte, blockSize int) []byte {
	if blockSize <= 0 {
		return b
	}
	padLen := blockSize - (len(b) % blockSize)
	if padLen == 0 {
		padLen = blockSize
	}
	out := make([]byte, len(b)+padLen)
	copy(out, b)
	for i := len(b); i < len(out); i++ {
		out[i] = byte(padLen)
	}
	return out
}

func pkcs7Unpad(b []byte, blockSize int) ([]byte, error) {
	if len(b) == 0 || len(b)%blockSize != 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	padLen := int(b[len(b)-1])
	if padLen == 0 || padLen > len(b) {
		return nil, fmt.Errorf("invalid padding length")
	}
	for i := 0; i < padLen; i++ {
		if b[len(b)-1-i] != byte(padLen) {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}
	return b[:len(b)-padLen], nil
}

var _ Encryptor = (*aes256Encryptor)(nil)
var _ Decryptor = (*aes256Decryptor)(nil)
