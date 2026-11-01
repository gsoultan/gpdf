// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package sign

import (
	"crypto"
	"time"
)

// SubFilter names the signature encoding written to the signature dictionary.
type SubFilter string

const (
	// SubFilterCAdES is ETSI.CAdES.detached: a PAdES baseline signature (ISO 32000-2
	// §12.8.3.4, ETSI EN 319 142-1). This is the default.
	SubFilterCAdES SubFilter = "ETSI.CAdES.detached"
	// SubFilterPKCS7 is adbe.pkcs7.detached, the PDF 1.x CMS signature.
	SubFilterPKCS7 SubFilter = "adbe.pkcs7.detached"
)

// SignOptions controls PDF-level signature metadata and optional appearance.
type SignOptions struct {
	Reason   string
	Location string
	Contact  string
	// FieldName is the signature field's /T. Default "Signature1", "Signature2", ...
	FieldName string
	// PageIndex and Rect place the signature widget. A zero Rect gives an invisible
	// signature; otherwise a simple appearance (signer name, date, reason) is drawn.
	PageIndex int
	Rect      [4]float64
	// Hash is the digest of the byte range (default SHA-256).
	Hash crypto.Hash
	// SigningTime is written as /M (default: now).
	SigningTime time.Time
	// SubFilter defaults to SubFilterCAdES.
	SubFilter SubFilter
	// Certify makes this a certification signature with a DocMDP transform (/P level:
	// 1 no changes, 2 form filling and signing, 3 plus annotations). 0 for an
	// ordinary approval signature. Only the first signature of a document can certify.
	Certify int
	// ContentsSize reserves room (in bytes) for the CMS container. Default 16384,
	// enough for a certificate chain and a time-stamp token.
	ContentsSize int

	signerName string // filled from a CMSSigner's certificate for the appearance
}
