package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"fmt"
	"hash"

	"github.com/gsoultan/gpdf/model"
)

// AES-256 Standard handler: R 6 (ISO 32000-2 §7.6.4.3.3, Algorithms 2.A, 2.B, 8–13) and
// the deprecated R 5 (Adobe extension level 3, plain SHA-256).

// aes256Password prepares a password: UTF-8, truncated to 127 bytes. (SASLprep
// normalisation is not applied; ASCII and already-normalised passwords are unaffected.)
func aes256Password(password string) []byte {
	b := []byte(password)
	if len(b) > 127 {
		b = b[:127]
	}
	return b
}

// hashR6 is Algorithm 2.B (R 6) or a single SHA-256 (R 5).
func hashR6(password, salt, udata []byte, r int) []byte {
	h := sha256.New()
	h.Write(password)
	h.Write(salt)
	h.Write(udata)
	k := h.Sum(nil)
	if r == 5 {
		return k
	}
	for round := 0; ; round++ {
		seq := make([]byte, 0, len(password)+len(k)+len(udata))
		seq = append(seq, password...)
		seq = append(seq, k...)
		seq = append(seq, udata...)
		k1 := bytes.Repeat(seq, 64)
		e, _ := aesNoPadIV(k[:16], k[16:32], k1)
		sum := 0
		for _, b := range e[:16] {
			sum += int(b)
		}
		var hf hash.Hash
		switch sum % 3 {
		case 0:
			hf = sha256.New()
		case 1:
			hf = sha512.New384()
		default:
			hf = sha512.New()
		}
		hf.Write(e)
		k = hf.Sum(nil)
		if round >= 63 && int(e[len(e)-1]) <= round+1-32 {
			break
		}
	}
	return k[:32]
}

func aesNoPadIV(key, iv, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(data))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, data)
	return out, nil
}

// authenticateAES256 is Algorithm 2.A: it validates the password as owner or user
// password and decrypts the file key from /OE or /UE.
func authenticateAES256(enc model.Dict, password string, r int) ([]byte, bool, error) {
	if v := getInt(enc, "V", 0); v != 5 {
		return nil, false, fmt.Errorf("AES-256 handler requires V 5, got V %d", v)
	}
	o, err := getBytes(enc, "O", 48)
	if err != nil {
		return nil, false, err
	}
	u, err := getBytes(enc, "U", 48)
	if err != nil {
		return nil, false, err
	}
	oe, err := getBytes(enc, "OE", 32)
	if err != nil {
		return nil, false, err
	}
	ue, err := getBytes(enc, "UE", 32)
	if err != nil {
		return nil, false, err
	}
	pw := aes256Password(password)
	u48 := u[:48]
	if subtle.ConstantTimeCompare(hashR6(pw, o[32:40], u48, r), o[:32]) == 1 {
		key, err := aesNoPad(hashR6(pw, o[40:48], u48, r), oe[:32], false)
		return key, true, err
	}
	if subtle.ConstantTimeCompare(hashR6(pw, u[32:40], nil, r), u[:32]) == 1 {
		key, err := aesNoPad(hashR6(pw, u[40:48], nil, r), ue[:32], false)
		return key, false, err
	}
	return nil, false, ErrWrongPassword
}

// ValidatePerms decrypts /Perms with the file key (Algorithm 13) and reports whether it
// is consistent with /P. Viewers may ignore a mismatch, so readers only use this as a check.
func ValidatePerms(enc model.Dict, fileKey []byte) bool {
	perms, err := getBytes(enc, "Perms", 16)
	if err != nil {
		return false
	}
	block, err := aes.NewCipher(fileKey)
	if err != nil {
		return false
	}
	plain := make([]byte, 16)
	block.Decrypt(plain, perms[:16])
	if string(plain[9:12]) != "adb" {
		return false
	}
	p := uint32(int32(getInt(enc, "P", 0)))
	return plain[0] == byte(p) && plain[1] == byte(p>>8) && plain[2] == byte(p>>16) && plain[3] == byte(p>>24)
}
