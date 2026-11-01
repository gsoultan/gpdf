package sign

import (
	"errors"
	"fmt"
	"io"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/reader"
	"github.com/gsoultan/gpdf/writer"
)

// ValidationData is the material a verifier needs to check a signature long after it
// was made (PAdES-B-LT, ISO 32000-2 §12.8.4.3): DER certificates, DER CRLs and DER
// OCSP responses.
type ValidationData struct {
	Certs [][]byte
	CRLs  [][]byte
	OCSPs [][]byte
}

// AddValidationData appends an incremental update that stores data in the
// document security store (catalog /DSS). Earlier signatures stay valid because
// their bytes are untouched. The input must not be encrypted.
func AddValidationData(r io.ReaderAt, size int64, out io.Writer, data ValidationData) error {
	if len(data.Certs)+len(data.CRLs)+len(data.OCSPs) == 0 {
		return errors.New("sign: no validation data")
	}
	doc, err := reader.NewPDFReader().ReadDocument(r, size)
	if err != nil {
		return fmt.Errorf("sign: read: %w", err)
	}
	if doc.Trailer().Dict[model.Name("Encrypt")] != nil {
		return ErrEncrypted
	}
	cat, err := doc.Catalog()
	if err != nil {
		return err
	}
	root := doc.Trailer().Root()
	trailer := doc.Trailer().Dict
	sz, _ := trailer[model.Name("Size")].(model.Integer)
	next := max(int(sz), 1)
	u := &update{objects: map[int]model.Object{}, gens: map[int]int{}}
	alloc := func() int { n := next; next++; return n }

	dss, _ := resolveDict(doc, cat.Dict[model.Name("DSS")])
	dss = copyDict(dss)
	addAll := func(key model.Name, items [][]byte) {
		if len(items) == 0 {
			return
		}
		arr, _ := resolveArray(doc, dss[key])
		arr = append(model.Array{}, arr...)
		for _, b := range items {
			n := alloc()
			u.objects[n] = &model.Stream{
				Dict:    model.Dict{model.Name("Length"): model.Integer(len(b)), model.Name("Filter"): model.Name("FlateDecode")},
				Content: b,
			}
			arr = append(arr, model.Ref{ObjectNumber: n})
		}
		dss[key] = arr
	}
	addAll("Certs", data.Certs)
	addAll("CRLs", data.CRLs)
	addAll("OCSPs", data.OCSPs)
	dssNum := alloc()
	u.objects[dssNum] = dss

	catalog := copyDict(cat.Dict)
	catalog[model.Name("DSS")] = model.Ref{ObjectNumber: dssNum}
	// PDF 1.x readers learn about /DSS through the ESIC extension.
	if v, _ := doc.PDFVersion().Major, 0; v < 2 {
		ext, _ := resolveDict(doc, catalog[model.Name("Extensions")])
		ext = copyDict(ext)
		ext[model.Name("ESIC")] = model.Dict{model.Name("BaseVersion"): model.Name("1.7"), model.Name("ExtensionLevel"): model.Integer(5)}
		catalog[model.Name("Extensions")] = ext
	}
	u.objects[root.ObjectNumber] = catalog
	u.gens[root.ObjectNumber] = root.Generation

	td := model.Dict{model.Name("Root"): *root, model.Name("Size"): model.Integer(next)}
	if info := doc.Trailer().Info(); info != nil {
		td[model.Name("Info")] = *info
	}
	if id, ok := trailer[model.Name("ID")]; ok {
		td[model.Name("ID")] = id
	}
	u.trailer = model.Trailer{Dict: td}

	if _, err := io.Copy(out, io.NewSectionReader(r, 0, size)); err != nil {
		return err
	}
	pw := writer.NewPDFWriterWithOptions(writer.WriterOptions{ValidateGraph: new(false)})
	return pw.WriteIncremental(out, size, doc.StartXRefOffset(), u)
}
