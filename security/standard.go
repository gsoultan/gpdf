// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package security

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/rc4"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/gsoultan/gpdf/model"
)

// ErrWrongPassword is returned when neither the user nor the owner password matches.
var ErrWrongPassword = errors.New("security: incorrect password")

// Standard padding string per ISO 32000-1 Algorithm 2.
var padding32 = []byte{
	0x28, 0xBF, 0x4E, 0x5E, 0x4E, 0x75, 0x8A, 0x41,
	0x64, 0x00, 0x4E, 0x56, 0xFF, 0xFA, 0x01, 0x08,
	0x2E, 0x2E, 0x00, 0xB6, 0xD0, 0x68, 0x3E, 0x80,
	0x2F, 0x0C, 0xA9, 0xFE, 0x64, 0x53, 0x69, 0x7A,
}

// cryptMethod is the algorithm of a crypt filter (/CFM).
type cryptMethod int

const (
	methodIdentity cryptMethod = iota
	methodRC4
	methodAESV2 // AES-128, per-object key
	methodAESV3 // AES-256, file key
)

// CryptFilterDecryptor is implemented by decryptors that can decrypt a stream with a named
// crypt filter (a stream whose /Filter starts with /Crypt and names a /CF entry).
type CryptFilterDecryptor interface {
	DecryptStreamWithFilter(filterName string, ref model.Ref, ciphertext []byte) ([]byte, error)
}

// MetadataEncryption reports whether the document's XMP metadata stream is encrypted
// (/EncryptMetadata, default true).
type MetadataEncryption interface {
	EncryptMetadata() bool
}

// StandardDecryptor implements the Standard security handler (ISO 32000-1 §7.6.3 and
// ISO 32000-2 §7.6.4) for R 2–6. It decrypts and, symmetrically, encrypts strings and streams.
type StandardDecryptor struct {
	key             []byte // file encryption key
	r               int
	stm, str        cryptMethod
	filters         map[string]cryptMethod
	encryptMetadata bool
	ownerAccess     bool
	random          io.Reader
}

var (
	_ Decryptor            = (*StandardDecryptor)(nil)
	_ Encryptor            = (*StandardDecryptor)(nil)
	_ CryptFilterDecryptor = (*StandardDecryptor)(nil)
	_ MetadataEncryption   = (*StandardDecryptor)(nil)
)

// NewStandardDecryptor authenticates password against the Encrypt dictionary, first as
// the user password and then as the owner password, and returns a decryptor for the file.
// It returns an error wrapping ErrWrongPassword when neither matches.
func NewStandardDecryptor(encryptDict model.Dict, id model.Array, password string) (Decryptor, error) {
	if encryptDict == nil {
		return nil, fmt.Errorf("encrypt dict is nil")
	}
	filter, _ := encryptDict[model.Name("Filter")].(model.Name)
	if filter != "Standard" {
		return nil, fmt.Errorf("unsupported security handler: %s", filter)
	}
	r := int(getInt(encryptDict, "R", 0))
	v := int(getInt(encryptDict, "V", 0))
	d := &StandardDecryptor{r: r, encryptMetadata: true, random: rand.Reader}
	if b, ok := encryptDict[model.Name("EncryptMetadata")].(model.Boolean); ok {
		d.encryptMetadata = bool(b)
	}
	if err := d.readCryptFilters(encryptDict, v); err != nil {
		return nil, err
	}
	switch r {
	case 2, 3, 4:
		key, owner, err := authenticateLegacy(encryptDict, getIDBytes(id), password, r, d.encryptMetadata)
		if err != nil {
			return nil, err
		}
		d.key, d.ownerAccess = key, owner
	case 5, 6:
		key, owner, err := authenticateAES256(encryptDict, password, r)
		if err != nil {
			return nil, err
		}
		d.key, d.ownerAccess = key, owner
	default:
		return nil, fmt.Errorf("unsupported security handler revision R=%d", r)
	}
	return d, nil
}

// OwnerAccess reports whether the owner password (rather than the user password) was given.
func (d *StandardDecryptor) OwnerAccess() bool { return d.ownerAccess }

// EncryptMetadata reports the /EncryptMetadata flag.
func (d *StandardDecryptor) EncryptMetadata() bool { return d.encryptMetadata }

// readCryptFilters maps /StmF, /StrF and every /CF entry to a crypt method.
func (d *StandardDecryptor) readCryptFilters(enc model.Dict, v int) error {
	d.filters = map[string]cryptMethod{"Identity": methodIdentity}
	if v < 4 {
		d.stm, d.str = methodRC4, methodRC4
		return nil
	}
	if cf, ok := enc[model.Name("CF")].(model.Dict); ok {
		for name, entry := range cf {
			e, ok := entry.(model.Dict)
			if !ok {
				continue
			}
			cfm, _ := e[model.Name("CFM")].(model.Name)
			switch cfm {
			case "V2":
				d.filters[string(name)] = methodRC4
			case "AESV2":
				d.filters[string(name)] = methodAESV2
			case "AESV3":
				d.filters[string(name)] = methodAESV3
			case "None", "":
				d.filters[string(name)] = methodIdentity
			default:
				return fmt.Errorf("unsupported crypt filter method %s", cfm)
			}
		}
	}
	lookup := func(key string) (cryptMethod, error) {
		name, ok := enc[model.Name(key)].(model.Name)
		if !ok {
			return methodIdentity, nil // default is Identity
		}
		m, ok := d.filters[string(name)]
		if !ok {
			return 0, fmt.Errorf("crypt filter %s not defined in /CF", name)
		}
		return m, nil
	}
	var err error
	if d.stm, err = lookup("StmF"); err != nil {
		return err
	}
	d.str, err = lookup("StrF")
	return err
}

func getInt(d model.Dict, name string, def int64) int64 {
	switch v := d[model.Name(name)].(type) {
	case model.Integer:
		return int64(v)
	case model.Real:
		return int64(v)
	}
	return def
}

func getBytes(d model.Dict, name string, minLen int) ([]byte, error) {
	v := d[model.Name(name)]
	if v == nil {
		return nil, fmt.Errorf("encrypt dict: missing /%s", name)
	}
	var b []byte
	switch s := v.(type) {
	case model.String:
		b = []byte(s)
	case model.HexString:
		b = []byte(s)
	default:
		return nil, fmt.Errorf("encrypt dict: /%s is not a string", name)
	}
	if len(b) < minLen {
		return nil, fmt.Errorf("encrypt dict: /%s too short (%d bytes, need %d)", name, len(b), minLen)
	}
	return b, nil
}

func getIDBytes(id model.Array) []byte {
	if len(id) == 0 {
		return nil
	}
	switch s := id[0].(type) {
	case model.String:
		return []byte(s)
	case model.HexString:
		return []byte(s)
	}
	return nil
}

// passwordCandidates returns the byte encodings to try for a legacy (R 2–4) password,
// which is defined in PDFDocEncoding: the raw bytes, and a Latin-1 form of UTF-8 input.
func passwordCandidates(password string) [][]byte {
	out := [][]byte{[]byte(password)}
	if utf8.ValidString(password) {
		var latin []byte
		ok := false
		for _, r := range password {
			if r > 0xFF {
				return out
			}
			if r > 0x7F {
				ok = true
			}
			latin = append(latin, byte(r))
		}
		if ok {
			out = append(out, latin)
		}
	}
	return out
}

func objectKey(fileKey []byte, ref model.Ref, aes bool) []byte {
	h := md5.New()
	h.Write(fileKey)
	h.Write([]byte{
		byte(ref.ObjectNumber), byte(ref.ObjectNumber >> 8), byte(ref.ObjectNumber >> 16),
		byte(ref.Generation), byte(ref.Generation >> 8),
	})
	if aes {
		h.Write([]byte("sAlT"))
	}
	return h.Sum(nil)[:min(len(fileKey)+5, 16)]
}

func (d *StandardDecryptor) crypt(m cryptMethod, ref model.Ref, data []byte, encrypt bool) ([]byte, error) {
	switch m {
	case methodIdentity:
		return data, nil
	case methodRC4:
		return rc4Crypt(objectKey(d.key, ref, false), data)
	case methodAESV2:
		key := objectKey(d.key, ref, true)
		if encrypt {
			return aesEncrypt(key, data, d.random)
		}
		return aesDecrypt(key, data)
	case methodAESV3:
		if encrypt {
			return aesEncrypt(d.key, data, d.random)
		}
		return aesDecrypt(d.key, data)
	}
	return nil, fmt.Errorf("unknown crypt method %d", m)
}

// DecryptString decrypts a string of the indirect object ref.
func (d *StandardDecryptor) DecryptString(ref model.Ref, ciphertext []byte) ([]byte, error) {
	return d.crypt(d.str, ref, ciphertext, false)
}

// DecryptStream decrypts stream data of the indirect object ref.
func (d *StandardDecryptor) DecryptStream(ref model.Ref, ciphertext []byte) ([]byte, error) {
	return d.crypt(d.stm, ref, ciphertext, false)
}

// DecryptStreamWithFilter decrypts stream data with the named crypt filter.
func (d *StandardDecryptor) DecryptStreamWithFilter(filterName string, ref model.Ref, ciphertext []byte) ([]byte, error) {
	m, ok := d.filters[filterName]
	if !ok {
		return nil, fmt.Errorf("crypt filter %s not defined in /CF", filterName)
	}
	return d.crypt(m, ref, ciphertext, false)
}

// EncryptString encrypts a string of the indirect object ref.
func (d *StandardDecryptor) EncryptString(ref model.Ref, plaintext []byte) ([]byte, error) {
	return d.crypt(d.str, ref, plaintext, true)
}

// EncryptStream encrypts stream data of the indirect object ref.
func (d *StandardDecryptor) EncryptStream(ref model.Ref, plaintext []byte) ([]byte, error) {
	return d.crypt(d.stm, ref, plaintext, true)
}

func rc4Crypt(key, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	c, err := rc4.NewCipher(key)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(data))
	c.XORKeyStream(out, data)
	return out, nil
}

// rc4Rounds applies RC4 with key, then 19 more times with each key byte XORed with i
// (Algorithms 3 and 5 for R ≥ 3). reverse runs i from 19 down to 0 (Algorithm 7).
func rc4Rounds(key, data []byte, reverse bool) []byte {
	out := bytes.Clone(data)
	k := make([]byte, len(key))
	for step := range 20 {
		i := step
		if reverse {
			i = 19 - step
		}
		for j := range key {
			k[j] = key[j] ^ byte(i)
		}
		out, _ = rc4Crypt(k, out)
	}
	return out
}

// aesEncrypt encrypts with AES-CBC, a random IV prefix and PKCS#5 padding.
func aesEncrypt(key, plaintext []byte, random io.Reader) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if random == nil {
		random = rand.Reader
	}
	padded := pkcs7Pad(plaintext, aes.BlockSize)
	out := make([]byte, aes.BlockSize+len(padded))
	if _, err := io.ReadFull(random, out[:aes.BlockSize]); err != nil {
		return nil, err
	}
	cipher.NewCBCEncrypter(block, out[:aes.BlockSize]).CryptBlocks(out[aes.BlockSize:], padded)
	return out, nil
}

// aesDecrypt decrypts IV-prefixed AES-CBC data. It is lenient the way viewers are: a
// truncated final block is ignored and invalid padding leaves the data unpadded as is.
func aesDecrypt(key, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	if len(data) < 2*aes.BlockSize {
		if len(data) == aes.BlockSize {
			return nil, nil // IV only: empty string
		}
		return nil, fmt.Errorf("aes: ciphertext too short (%d bytes)", len(data))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	body := data[aes.BlockSize:]
	body = body[:len(body)-len(body)%aes.BlockSize]
	out := make([]byte, len(body))
	cipher.NewCBCDecrypter(block, data[:aes.BlockSize]).CryptBlocks(out, body)
	if unpadded, err := pkcs7Unpad(out, aes.BlockSize); err == nil {
		return unpadded, nil
	}
	return out, nil
}

// aesNoPad encrypts or decrypts whole blocks with AES-CBC and a zero IV (used for /UE, /OE).
func aesNoPad(key, data []byte, encrypt bool) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("aes: data is not block aligned")
	}
	iv := make([]byte, aes.BlockSize)
	out := make([]byte, len(data))
	if encrypt {
		cipher.NewCBCEncrypter(block, iv).CryptBlocks(out, data)
	} else {
		cipher.NewCBCDecrypter(block, iv).CryptBlocks(out, data)
	}
	return out, nil
}

func pkcs7Pad(b []byte, blockSize int) []byte {
	padLen := blockSize - len(b)%blockSize
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
	if padLen == 0 || padLen > blockSize || padLen > len(b) {
		return nil, fmt.Errorf("invalid padding length")
	}
	for i := range padLen {
		if b[len(b)-1-i] != byte(padLen) {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}
	return b[:len(b)-padLen], nil
}
