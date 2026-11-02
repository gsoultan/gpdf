// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package security

import (
	"bytes"
	"errors"
	"testing"

	"github.com/gsoultan/gpdf/model"
)

func TestNewStandardDecryptor_InvalidFilter(t *testing.T) {
	d := model.Dict{
		model.Name("Filter"): model.Name("Unknown"),
		model.Name("R"):      model.Integer(2),
		model.Name("O"):      model.String(string(make([]byte, 32))),
		model.Name("U"):      model.String(string(make([]byte, 32))),
		model.Name("P"):      model.Integer(-4),
	}
	if _, err := NewStandardDecryptor(d, nil, "user"); err == nil {
		t.Error("expected error for unsupported Filter")
	}
}

func TestNewStandardDecryptor_UnsupportedR(t *testing.T) {
	d := model.Dict{
		model.Name("Filter"): model.Name("Standard"),
		model.Name("R"):      model.Integer(7),
		model.Name("O"):      model.String(string(make([]byte, 32))),
		model.Name("U"):      model.String(string(make([]byte, 32))),
		model.Name("P"):      model.Integer(-4),
	}
	if _, err := NewStandardDecryptor(d, nil, "user"); err == nil {
		t.Error("expected error for unsupported R")
	}
}

func TestNewStandardDecryptor_MissingO(t *testing.T) {
	d := model.Dict{
		model.Name("Filter"): model.Name("Standard"),
		model.Name("R"):      model.Integer(2),
		model.Name("P"):      model.Integer(-4),
	}
	if _, err := NewStandardDecryptor(d, nil, "user"); err == nil {
		t.Error("expected error for missing O")
	}
}

var testID = []byte("0123456789abcdef")

// Regression: R 2–4 never checked /U, so any password "worked" and produced garbage.
func TestRC4WrongPasswordRejectedAndOwnerAccepted(t *testing.T) {
	dict, enc, err := BuildEncryptDictForWrite("user", "owner", testID, -4)
	if err != nil {
		t.Fatal(err)
	}
	id := model.Array{model.String(testID), model.String(testID)}
	ref := model.Ref{ObjectNumber: 7}
	cipher, _ := enc.EncryptString(ref, []byte("secret text"))

	if _, err := NewStandardDecryptor(dict, id, "WRONG"); !errors.Is(err, ErrWrongPassword) {
		t.Fatalf("wrong password: err = %v, want ErrWrongPassword", err)
	}
	for _, pw := range []string{"user", "owner"} {
		dec, err := NewStandardDecryptor(dict, id, pw)
		if err != nil {
			t.Fatalf("%s password: %v", pw, err)
		}
		plain, err := dec.DecryptString(ref, cipher)
		if err != nil || string(plain) != "secret text" {
			t.Fatalf("%s password: got %q, %v", pw, plain, err)
		}
		if got := dec.(*StandardDecryptor).OwnerAccess(); got != (pw == "owner") {
			t.Fatalf("%s password: OwnerAccess = %v", pw, got)
		}
	}
}

// Regression: AES-256 used a private PBKDF2 scheme. This checks the ISO 32000-2 R 6
// algorithms round-trip for user and owner passwords and reject a wrong one.
func TestAES256R6UserOwnerAndWrongPassword(t *testing.T) {
	dict, enc, err := BuildAES256EncryptDictForWrite("pässwörd", "owner", testID, -4)
	if err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"U", "O"} {
		if n := len(dict[model.Name(k)].(model.HexString)); n != 48 {
			t.Fatalf("/%s length %d, want 48", k, n)
		}
	}
	for _, k := range []string{"UE", "OE"} {
		if n := len(dict[model.Name(k)].(model.HexString)); n != 32 {
			t.Fatalf("/%s length %d, want 32", k, n)
		}
	}
	ref := model.Ref{ObjectNumber: 3}
	cipher, err := enc.EncryptStream(ref, []byte("stream body"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewStandardDecryptor(dict, nil, "nope"); !errors.Is(err, ErrWrongPassword) {
		t.Fatalf("wrong password: %v", err)
	}
	for _, pw := range []string{"pässwörd", "owner"} {
		dec, err := NewStandardDecryptor(dict, nil, pw)
		if err != nil {
			t.Fatalf("%q: %v", pw, err)
		}
		plain, err := dec.DecryptStream(ref, cipher)
		if err != nil || !bytes.Equal(plain, []byte("stream body")) {
			t.Fatalf("%q: got %q, %v", pw, plain, err)
		}
		if !ValidatePerms(dict, dec.(*StandardDecryptor).key) {
			t.Fatalf("%q: /Perms does not validate", pw)
		}
	}
}

// Algorithm 2.B must be deterministic, 32 bytes, and differ from the R 5 hash.
func TestHashR6Deterministic(t *testing.T) {
	a := hashR6([]byte("x"), make([]byte, 8), nil, 6)
	b := hashR6([]byte("x"), make([]byte, 8), nil, 6)
	if !bytes.Equal(a, b) || len(a) != 32 {
		t.Fatal("hashR6 not deterministic or wrong length")
	}
	if bytes.Equal(a, hashR6([]byte("x"), make([]byte, 8), nil, 5)) {
		t.Fatal("R 6 hash must differ from the R 5 single SHA-256")
	}
}

// Regression: R 4 ignored /EncryptMetadata false, which changes the key derivation.
func TestR4EncryptMetadataFalseChangesKey(t *testing.T) {
	o := computeO([]byte("o"), []byte("u"), 4, 16)
	k1 := computeFileKey([]byte("u"), o, uint32(0xFFFFFFFC), testID, 4, 16, true)
	k2 := computeFileKey([]byte("u"), o, uint32(0xFFFFFFFC), testID, 4, 16, false)
	if bytes.Equal(k1, k2) {
		t.Fatal("EncryptMetadata false must change the R 4 file key")
	}
	u := computeU(k2, testID, 4)
	dict := model.Dict{
		"Filter": model.Name("Standard"), "V": model.Integer(4), "R": model.Integer(4), "Length": model.Integer(128),
		"O": model.String(o), "U": model.String(u), "P": model.Integer(-4), "EncryptMetadata": model.Boolean(false),
		"CF":   model.Dict{"StdCF": model.Dict{"CFM": model.Name("AESV2")}},
		"StmF": model.Name("StdCF"), "StrF": model.Name("StdCF"),
	}
	dec, err := NewStandardDecryptor(dict, model.Array{model.String(testID)}, "u")
	if err != nil {
		t.Fatal(err)
	}
	if dec.(MetadataEncryption).EncryptMetadata() {
		t.Fatal("EncryptMetadata should be false")
	}
	ref := model.Ref{ObjectNumber: 9}
	c, _ := dec.(Encryptor).EncryptString(ref, []byte("aes-128"))
	p, err := dec.DecryptString(ref, c)
	if err != nil || string(p) != "aes-128" {
		t.Fatalf("AESV2 round trip: %q %v", p, err)
	}
}
