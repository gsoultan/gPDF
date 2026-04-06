package security

import (
	"crypto/md5"
	"crypto/rc4"
	"encoding/binary"

	"github.com/gsoultan/gpdf/model"
)

// BuildEncryptDictForWrite creates an Encrypt dictionary and an Encryptor for writing (Standard, R=2).
// id is the first element of the trailer /ID array (e.g. 16 bytes); if nil, a zero slice is used for key derivation.
// P is the permission flags (e.g. -4 for allow all).
func BuildEncryptDictForWrite(userPassword, ownerPassword string, id []byte, P int32) (model.Dict, Encryptor, error) {
	const r = 2
	const n = 5
	userPad := padPassword([]byte(userPassword))
	ownerPad := padPassword([]byte(ownerPassword))
	// Algorithm 3.3: O value. owner_key = first n bytes of MD5(owner_pad).
	h := md5.New()
	h.Write(ownerPad)
	ownerKey := h.Sum(nil)[:n]
	// O = RC4(owner_key, user_pad) -> 32 bytes
	o, err := rc4Crypt(ownerKey, userPad)
	if err != nil {
		return nil, nil, err
	}
	if len(o) < 32 {
		return nil, nil, nil
	}
	o = o[:32]
	// Algorithm 3.4: U value. encryption_key = deriveKey(user, O, P, ID)
	pLE := make([]byte, 4)
	binary.LittleEndian.PutUint32(pLE, uint32(P))
	encKey := deriveKey([]byte(userPassword), o, pLE, id, r, n)
	// U = RC4(encryption_key, padding32)
	u, err := rc4Crypt(encKey, padding32)
	if err != nil {
		return nil, nil, err
	}
	u = u[:32]
	dict := model.Dict{
		model.Name("Filter"): model.Name("Standard"),
		model.Name("R"):      model.Integer(r),
		model.Name("O"):      model.String(string(o)),
		model.Name("U"):      model.String(string(u)),
		model.Name("P"):      model.Integer(int64(P)),
	}
	dec := &StandardDecryptor{key: encKey, r: r, n: n}
	return dict, dec, nil
}

func rc4Crypt(key, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	cipher, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(data))
	cipher.XORKeyStream(out, data)
	return out, nil
}
