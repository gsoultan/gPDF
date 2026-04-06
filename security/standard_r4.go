package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rc4"
	"fmt"

	"github.com/gsoultan/gpdf/model"
)

// r4Method names the crypt method for a given CF entry.
type r4Method int

const (
	r4MethodNone r4Method = iota
	r4MethodRC4
	r4MethodAES128
)

// standardR4Decryptor implements Decryptor for PDF Standard security handler R=4 (V=4).
// R=4 supports AES-128 or RC4 per the /CF (crypt filter) dictionary, with
// /StmF and /StrF selecting the filter for streams and strings respectively.
type standardR4Decryptor struct {
	key  []byte // 16-byte file encryption key
	stmF r4Method
	strF r4Method
}

// newStandardR4Decryptor derives the file key and reads the CF dict for R=4.
func newStandardR4Decryptor(encryptDict model.Dict, id model.Array, userPassword string) (Decryptor, error) {
	length := int(getInt(encryptDict, "Length", 128))
	if length < 40 || length > 128 {
		length = 128
	}
	n := length / 8

	o, err := getBytes32(encryptDict, "O")
	if err != nil {
		return nil, err
	}
	p := getP(encryptDict)
	idBytes := getIdBytes(id)

	key := deriveKeyR4([]byte(userPassword), o, pToBytes(p), idBytes, n)

	stmF := resolveCFMethod(encryptDict, "StmF")
	strF := resolveCFMethod(encryptDict, "StrF")

	return &standardR4Decryptor{key: key, stmF: stmF, strF: strF}, nil
}

// deriveKeyR4 computes the R=4 encryption key using MD5 with 50 rounds.
func deriveKeyR4(password, o, pLE, id []byte, n int) []byte {
	padded := padPassword(password)
	h := md5.New()
	h.Write(padded)
	h.Write(o)
	h.Write(pLE)
	if len(id) > 0 {
		h.Write(id)
	}
	digest := h.Sum(nil)
	for range 50 {
		h = md5.New()
		h.Write(digest[:n])
		digest = h.Sum(nil)
	}
	return digest[:n]
}

// resolveCFMethod looks up the /StmF or /StrF name in the CF dict and maps it to a method.
func resolveCFMethod(encryptDict model.Dict, filterField string) r4Method {
	name, ok := encryptDict[model.Name(filterField)].(model.Name)
	if !ok || name == "Identity" {
		return r4MethodNone
	}
	cf, ok := encryptDict[model.Name("CF")].(model.Dict)
	if !ok {
		return r4MethodRC4
	}
	entry, ok := cf[model.Name(string(name))].(model.Dict)
	if !ok {
		return r4MethodRC4
	}
	cfm, ok := entry[model.Name("CFM")].(model.Name)
	if !ok {
		return r4MethodRC4
	}
	switch cfm {
	case "AESV2":
		return r4MethodAES128
	case "V2":
		return r4MethodRC4
	default:
		return r4MethodNone
	}
}

func (d *standardR4Decryptor) DecryptString(ref model.Ref, ciphertext []byte) ([]byte, error) {
	return d.decryptWith(ref, ciphertext, d.strF)
}

func (d *standardR4Decryptor) DecryptStream(ref model.Ref, ciphertext []byte) ([]byte, error) {
	return d.decryptWith(ref, ciphertext, d.stmF)
}

func (d *standardR4Decryptor) decryptWith(ref model.Ref, ciphertext []byte, method r4Method) ([]byte, error) {
	switch method {
	case r4MethodNone:
		return ciphertext, nil
	case r4MethodAES128:
		return d.aes128Decrypt(ref, ciphertext)
	default:
		return d.rc4Decrypt(ref, ciphertext)
	}
}

func (d *standardR4Decryptor) objectKeyR4(ref model.Ref) []byte {
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
	n := len(d.key) + 5
	if n > 16 {
		n = 16
	}
	return digest[:n]
}

func (d *standardR4Decryptor) objectKeyAES(ref model.Ref) []byte {
	h := md5.New()
	h.Write(d.key)
	h.Write([]byte{
		byte(ref.ObjectNumber),
		byte(ref.ObjectNumber >> 8),
		byte(ref.ObjectNumber >> 16),
		byte(ref.Generation),
		byte(ref.Generation >> 8),
		0x73, 0x41, 0x6C, 0x54, // "sAlT"
	})
	digest := h.Sum(nil)
	n := len(d.key) + 5
	if n > 16 {
		n = 16
	}
	return digest[:n]
}

func (d *standardR4Decryptor) rc4Decrypt(ref model.Ref, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return ciphertext, nil
	}
	key := d.objectKeyR4(ref)
	c, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(ciphertext))
	c.XORKeyStream(out, ciphertext)
	return out, nil
}

func (d *standardR4Decryptor) aes128Decrypt(ref model.Ref, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return ciphertext, nil
	}
	if len(ciphertext) < aes.BlockSize {
		return nil, fmt.Errorf("r4 aes128: ciphertext too short")
	}
	key := d.objectKeyAES(ref)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := ciphertext[:aes.BlockSize]
	body := ciphertext[aes.BlockSize:]
	if len(body)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("r4 aes128: ciphertext not block-aligned")
	}
	out := make([]byte, len(body))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, body)
	return pkcs7Unpad(out, aes.BlockSize)
}

var _ Decryptor = (*standardR4Decryptor)(nil)
