package security

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/binary"
	"fmt"

	"github.com/gsoultan/gpdf/model"
)

// Legacy Standard handler, R 2–4 (ISO 32000-1 §7.6.3.3, Algorithms 2–7).

func padPassword(password []byte) []byte {
	out := make([]byte, 32)
	n := copy(out, password)
	copy(out[n:], padding32)
	return out
}

func legacyKeyLength(enc model.Dict, r int) int {
	if r == 2 {
		return 5
	}
	bits := int(getInt(enc, "Length", 128))
	if bits < 40 || bits > 128 || bits%8 != 0 {
		bits = 128
	}
	return bits / 8
}

// computeFileKey is Algorithm 2.
func computeFileKey(password, o []byte, p uint32, id []byte, r, n int, encryptMetadata bool) []byte {
	h := md5.New()
	h.Write(padPassword(password))
	h.Write(o[:32])
	var pb [4]byte
	binary.LittleEndian.PutUint32(pb[:], p)
	h.Write(pb[:])
	h.Write(id)
	if r >= 4 && !encryptMetadata {
		h.Write([]byte{0xFF, 0xFF, 0xFF, 0xFF})
	}
	digest := h.Sum(nil)
	if r >= 3 {
		for range 50 {
			sum := md5.Sum(digest[:n])
			digest = sum[:]
		}
	}
	return digest[:n]
}

// computeU is Algorithm 4 (R 2) and Algorithm 5 (R ≥ 3). For R ≥ 3 only the first 16
// bytes are significant; the rest is arbitrary padding.
func computeU(key, id []byte, r int) []byte {
	if r == 2 {
		u, _ := rc4Crypt(key, padding32)
		return u
	}
	h := md5.New()
	h.Write(padding32)
	h.Write(id)
	u := rc4Rounds(key, h.Sum(nil), false)
	return append(u, make([]byte, 16)...)
}

// ownerKey is steps a–d of Algorithm 3: the RC4 key derived from the owner password.
func ownerKey(ownerPassword []byte, r, n int) []byte {
	sum := md5.Sum(padPassword(ownerPassword))
	digest := sum[:]
	if r >= 3 {
		for range 50 {
			s := md5.Sum(digest)
			digest = s[:]
		}
	}
	return digest[:n]
}

// computeO is Algorithm 3.
func computeO(ownerPassword, userPassword []byte, r, n int) []byte {
	if len(ownerPassword) == 0 {
		ownerPassword = userPassword
	}
	key := ownerKey(ownerPassword, r, n)
	if r == 2 {
		o, _ := rc4Crypt(key, padPassword(userPassword))
		return o
	}
	return rc4Rounds(key, padPassword(userPassword), false)
}

func checkUser(password, o, u []byte, p uint32, id []byte, r, n int, encryptMetadata bool) ([]byte, bool) {
	key := computeFileKey(password, o, p, id, r, n, encryptMetadata)
	want := computeU(key, id, r)
	cmpLen := 32
	if r >= 3 {
		cmpLen = 16
	}
	return key, subtle.ConstantTimeCompare(want[:cmpLen], u[:cmpLen]) == 1
}

// authenticateLegacy tries the password as user password (Algorithm 6) and then as owner
// password (Algorithm 7), and returns the file key.
func authenticateLegacy(enc model.Dict, id []byte, password string, r int, encryptMetadata bool) ([]byte, bool, error) {
	o, err := getBytes(enc, "O", 32)
	if err != nil {
		return nil, false, err
	}
	u, err := getBytes(enc, "U", 32)
	if err != nil {
		return nil, false, err
	}
	if _, ok := enc[model.Name("P")]; !ok {
		return nil, false, fmt.Errorf("encrypt dict: missing /P")
	}
	p := uint32(int32(getInt(enc, "P", 0)))
	n := legacyKeyLength(enc, r)
	for _, pw := range passwordCandidates(password) {
		if key, ok := checkUser(pw, o, u, p, id, r, n, encryptMetadata); ok {
			return key, false, nil
		}
	}
	for _, pw := range passwordCandidates(password) {
		key := ownerKey(pw, r, n)
		var userPw []byte
		if r == 2 {
			userPw, _ = rc4Crypt(key, o[:32])
		} else {
			userPw = rc4Rounds(key, o[:32], true)
		}
		if fileKey, ok := checkUser(userPw, o, u, p, id, r, n, encryptMetadata); ok {
			return fileKey, true, nil
		}
	}
	return nil, false, ErrWrongPassword
}
