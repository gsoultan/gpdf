// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package security

import (
	"crypto/aes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/gsoultan/gpdf/model"
)

// BuildEncryptDictForWrite creates an Encrypt dictionary and Encryptor for the Standard
// security handler with 128-bit RC4 (V 2, R 3). id must be the first element of the
// trailer /ID that is written with the file. P is the permission flags (-4 allows all).
func BuildEncryptDictForWrite(userPassword, ownerPassword string, id []byte, P int32) (model.Dict, Encryptor, error) {
	const r, n = 3, 16
	user := []byte(userPassword)
	o := computeO([]byte(ownerPassword), user, r, n)
	key := computeFileKey(user, o, uint32(P), id, r, n, true)
	u := computeU(key, id, r)
	dict := model.Dict{
		model.Name("Filter"): model.Name("Standard"),
		model.Name("V"):      model.Integer(2),
		model.Name("R"):      model.Integer(r),
		model.Name("Length"): model.Integer(n * 8),
		model.Name("O"):      model.HexString(o),
		model.Name("U"):      model.HexString(u),
		model.Name("P"):      model.Integer(int64(P)),
	}
	enc := &StandardDecryptor{key: key, r: r, stm: methodRC4, str: methodRC4, encryptMetadata: true, random: rand.Reader}
	return dict, enc, nil
}

// BuildAES256EncryptDictForWrite creates an Encrypt dictionary and Encryptor for AES-256
// (ISO 32000-2 Standard security handler, V 5, R 6), using crypto/rand.
func BuildAES256EncryptDictForWrite(userPassword, ownerPassword string, id []byte, P int32) (model.Dict, Encryptor, error) {
	return BuildAES256EncryptDictForWriteWithRandom(userPassword, ownerPassword, id, P, rand.Reader)
}

// BuildAES256EncryptDictForWriteWithRandom is BuildAES256EncryptDictForWrite with an explicit
// source for the file key and salts. Per-string IVs always come from crypto/rand.
// An empty owner password is replaced by the user password.
func BuildAES256EncryptDictForWriteWithRandom(userPassword, ownerPassword string, _ []byte, P int32, random io.Reader) (model.Dict, Encryptor, error) {
	if random == nil {
		random = rand.Reader
	}
	if ownerPassword == "" {
		ownerPassword = userPassword
	}
	const r = 6
	var seed [32 + 4*8 + 4]byte
	if _, err := io.ReadFull(random, seed[:]); err != nil {
		return nil, nil, fmt.Errorf("aes-256: random: %w", err)
	}
	fileKey := seed[:32]
	uvs, uks, ovs, oks := seed[32:40], seed[40:48], seed[48:56], seed[56:64]
	userPw, err := aes256Password(userPassword)
	if err != nil {
		return nil, nil, err
	}
	ownerPw, err := aes256Password(ownerPassword)
	if err != nil {
		return nil, nil, err
	}

	// Algorithm 8: U and UE.
	u := append(append(hashR6(userPw, uvs, nil, r), uvs...), uks...)
	ue, err := aesNoPad(hashR6(userPw, uks, nil, r), fileKey, true)
	if err != nil {
		return nil, nil, err
	}
	// Algorithm 9: O and OE (salted with the 48-byte U).
	o := append(append(hashR6(ownerPw, ovs, u, r), ovs...), oks...)
	oe, err := aesNoPad(hashR6(ownerPw, oks, u, r), fileKey, true)
	if err != nil {
		return nil, nil, err
	}
	// Algorithm 10: Perms.
	var perms [16]byte
	binary.LittleEndian.PutUint32(perms[0:4], uint32(P))
	copy(perms[4:8], []byte{0xFF, 0xFF, 0xFF, 0xFF})
	perms[8] = 'T'
	copy(perms[9:12], "adb")
	copy(perms[12:16], seed[64:68])
	block, err := aes.NewCipher(fileKey)
	if err != nil {
		return nil, nil, err
	}
	block.Encrypt(perms[:], perms[:])

	dict := model.Dict{
		model.Name("Filter"): model.Name("Standard"),
		model.Name("V"):      model.Integer(5),
		model.Name("R"):      model.Integer(r),
		model.Name("Length"): model.Integer(256),
		model.Name("CF"): model.Dict{
			model.Name("StdCF"): model.Dict{
				model.Name("AuthEvent"): model.Name("DocOpen"),
				model.Name("CFM"):       model.Name("AESV3"),
				model.Name("Length"):    model.Integer(32),
			},
		},
		model.Name("StmF"):            model.Name("StdCF"),
		model.Name("StrF"):            model.Name("StdCF"),
		model.Name("O"):               model.HexString(o),
		model.Name("U"):               model.HexString(u),
		model.Name("OE"):              model.HexString(oe),
		model.Name("UE"):              model.HexString(ue),
		model.Name("P"):               model.Integer(int64(P)),
		model.Name("Perms"):           model.HexString(perms[:]),
		model.Name("EncryptMetadata"): model.Boolean(true),
	}
	enc := &StandardDecryptor{
		key: append([]byte(nil), fileKey...), r: r, stm: methodAESV3, str: methodAESV3,
		filters: map[string]cryptMethod{"Identity": methodIdentity, "StdCF": methodAESV3}, encryptMetadata: true, random: rand.Reader,
	}
	return dict, enc, nil
}

// NewAES256Decryptor returns a decryptor for an AES-256 (V 5, R 5 or 6) Encrypt dictionary.
// It is kept for compatibility; NewStandardDecryptor handles every revision.
func NewAES256Decryptor(encryptDict model.Dict, id model.Array, password string) (Decryptor, error) {
	r := getInt(encryptDict, "R", 0)
	if getInt(encryptDict, "V", 0) != 5 || (r != 5 && r != 6) {
		return nil, fmt.Errorf("unsupported AES256 handler V=%d R=%d", getInt(encryptDict, "V", 0), r)
	}
	return NewStandardDecryptor(encryptDict, id, password)
}
