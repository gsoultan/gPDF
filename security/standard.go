package security

import (
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"
	"fmt"

	"github.com/gsoultan/gpdf/model"
)

// Standard padding string per PDF 1.7 Algorithm 2.
var padding32 = []byte{
	0x28, 0xBF, 0x4E, 0x5E, 0x4E, 0x75, 0x8A, 0x41,
	0x64, 0x00, 0x4E, 0x56, 0xFF, 0xFA, 0x01, 0x08,
	0x2E, 0x2E, 0x00, 0xB6, 0xD0, 0x68, 0x3E, 0x80,
	0x2F, 0x0C, 0xA9, 0xFE, 0x64, 0x53, 0x69, 0x7A,
}

// StandardDecryptor implements Decryptor for PDF Standard security handler (R=2, R=3).
type StandardDecryptor struct {
	key []byte
	r   int
	n   int // key length in bytes: 5 for R=2, 16 for R=3 with Length 128
}

// NewStandardDecryptor builds a decryptor from the Encrypt dict, trailer ID, and user password.
// Supports R=2 and R=3 (RC4), R=4 (AES-128 or RC4 per CF dict), R=5 (AES-256, PDF 1.7 ext 3),
// and R=6 (AES-256, PDF 2.0).
func NewStandardDecryptor(encryptDict model.Dict, id model.Array, userPassword string) (Decryptor, error) {
	if encryptDict == nil {
		return nil, fmt.Errorf("encrypt dict is nil")
	}
	filter, _ := encryptDict[model.Name("Filter")].(model.Name)
	if filter != "Standard" {
		return nil, fmt.Errorf("unsupported Filter: %s", filter)
	}
	r := int(getInt(encryptDict, "R", 0))
	switch {
	case r == 4:
		return newStandardR4Decryptor(encryptDict, id, userPassword)
	case r == 5 || r == 6:
		return newStandardR5R6Decryptor(encryptDict, id, userPassword, r)
	case r == 2 || r == 3:
		return newStandardR2R3Decryptor(encryptDict, id, userPassword, r)
	default:
		return nil, fmt.Errorf("unsupported R: %d", r)
	}
}

// newStandardR2R3Decryptor handles legacy RC4-based R=2 (40-bit) and R=3 (up to 128-bit).
func newStandardR2R3Decryptor(encryptDict model.Dict, id model.Array, userPassword string, r int) (Decryptor, error) {
	length := int(getInt(encryptDict, "Length", 40))
	if length != 40 && length != 128 {
		length = 40
	}
	n := length / 8
	if r == 2 {
		n = 5
	}
	o, err := getBytes32(encryptDict, "O")
	if err != nil {
		return nil, err
	}
	p := getP(encryptDict)
	idBytes := getIdBytes(id)
	key := deriveKey([]byte(userPassword), o, pToBytes(p), idBytes, r, n)
	if key == nil {
		return nil, fmt.Errorf("key derivation failed")
	}
	return &StandardDecryptor{key: key, r: r, n: n}, nil
}

func getInt(d model.Dict, name string, def int64) int64 {
	if v, ok := d[model.Name(name)].(model.Integer); ok {
		return int64(v)
	}
	return def
}

func getP(d model.Dict) uint32 {
	if v, ok := d[model.Name("P")].(model.Integer); ok {
		return uint32(v)
	}
	return 0
}

func getBytes32(d model.Dict, name string) ([]byte, error) {
	v := d[model.Name(name)]
	if v == nil {
		return nil, fmt.Errorf("missing %s", name)
	}
	s, ok := v.(model.String)
	if !ok {
		return nil, fmt.Errorf("%s is not a string", name)
	}
	b := []byte(s)
	if len(b) < 32 {
		return nil, fmt.Errorf("%s too short", name)
	}
	return b[:32], nil
}

func getIdBytes(id model.Array) []byte {
	if id == nil || len(id) == 0 {
		return nil
	}
	s, ok := id[0].(model.String)
	if !ok {
		return nil
	}
	return []byte(s)
}

func padPassword(password []byte) []byte {
	if len(password) >= 32 {
		return password[:32]
	}
	out := make([]byte, 32)
	copy(out, password)
	copy(out[len(password):], padding32[:32-len(password)])
	return out
}

func deriveKey(password, o, pLE, id []byte, r, n int) []byte {
	key := padPassword(password)
	h := md5.New()
	h.Write(key)
	h.Write(o)
	h.Write(pLE)
	if len(id) > 0 {
		h.Write(id)
	}
	digest := h.Sum(nil)
	if r == 2 {
		return digest[:5]
	}
	for range 50 {
		h = md5.New()
		h.Write(digest[:n])
		digest = h.Sum(nil)
	}
	return digest[:n]
}

func pToBytes(p uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, p)
	return b
}

func (d *StandardDecryptor) objectKey(ref model.Ref) ([]byte, error) {
	h := md5.New()
	h.Write(d.key)
	h.Write([]byte{
		byte(ref.ObjectNumber),
		byte(ref.ObjectNumber >> 8),
		byte(ref.ObjectNumber >> 16),
		byte(ref.Generation),
		byte(ref.Generation >> 8),
	})
	digest := h.Sum(nil)
	n := d.n + 5
	if n > len(digest) {
		n = len(digest)
	}
	return digest[:n], nil
}

func (d *StandardDecryptor) DecryptString(ref model.Ref, ciphertext []byte) ([]byte, error) {
	return d.rc4Decrypt(ref, ciphertext)
}

func (d *StandardDecryptor) DecryptStream(ref model.Ref, ciphertext []byte) ([]byte, error) {
	return d.rc4Decrypt(ref, ciphertext)
}

// EncryptString encrypts plaintext (RC4 is symmetric).
func (d *StandardDecryptor) EncryptString(ref model.Ref, plaintext []byte) ([]byte, error) {
	return d.rc4Decrypt(ref, plaintext)
}

// EncryptStream encrypts plaintext (RC4 is symmetric).
func (d *StandardDecryptor) EncryptStream(ref model.Ref, plaintext []byte) ([]byte, error) {
	return d.rc4Decrypt(ref, plaintext)
}

func (d *StandardDecryptor) rc4Decrypt(ref model.Ref, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return ciphertext, nil
	}
	key, err := d.objectKey(ref)
	if err != nil {
		return nil, err
	}
	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(ciphertext))
	cipher.XORKeyStream(out, ciphertext)
	return out, nil
}

var _ Decryptor = (*StandardDecryptor)(nil)
var _ Encryptor = (*StandardDecryptor)(nil)
