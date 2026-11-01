package reader

import (
	"fmt"
	"time"

	"github.com/gsoultan/gpdf/model"
)

// EmbeddedFile is a file embedded in a document (ISO 32000-2 §7.11.4), found through
// the EmbeddedFiles name tree or an /AF array.
type EmbeddedFile struct {
	// Name is the file name (/UF, or /F when there is no Unicode name).
	Name        string
	Description string
	// MIME is the media type from the embedded file stream's /Subtype, if any.
	MIME string
	// Relationship is the /AFRelationship for associated files (PDF 2.0), else "".
	Relationship string
	// Modified is the /Params /ModDate when present.
	Modified time.Time
	Data     []byte
}

// Resolver resolves indirect references.
type Resolver interface {
	Resolve(ref model.Ref) (model.Object, error)
}

// ExtractEmbeddedFiles lists the files in the catalog's EmbeddedFiles name tree and
// the document-level /AF array (duplicates between the two are returned once).
func ExtractEmbeddedFiles(r Resolver, cat *model.Catalog) ([]EmbeddedFile, error) {
	if cat == nil || cat.Dict == nil {
		return nil, nil
	}
	deref := func(o model.Object) model.Object {
		if ref, ok := o.(model.Ref); ok {
			v, err := r.Resolve(ref)
			if err != nil {
				return nil
			}
			return v
		}
		return o
	}
	var out []EmbeddedFile
	seen := map[model.Ref]bool{}
	add := func(specObj model.Object) error {
		if ref, ok := specObj.(model.Ref); ok {
			if seen[ref] {
				return nil
			}
			seen[ref] = true
		}
		spec, ok := deref(specObj).(model.Dict)
		if !ok {
			return nil
		}
		f, err := embeddedFileFromSpec(spec, deref)
		if err != nil {
			return err
		}
		if f != nil {
			out = append(out, *f)
		}
		return nil
	}
	if names, ok := deref(cat.Dict[model.Name("Names")]).(model.Dict); ok {
		if tree, ok := deref(names[model.Name("EmbeddedFiles")]).(model.Dict); ok {
			if err := walkNameTree(tree, deref, 0, add); err != nil {
				return out, err
			}
		}
	}
	if af, ok := deref(cat.Dict[model.Name("AF")]).(model.Array); ok {
		for _, o := range af {
			if err := add(o); err != nil {
				return out, err
			}
		}
	}
	return out, nil
}

func walkNameTree(node model.Dict, deref func(model.Object) model.Object, depth int, fn func(model.Object) error) error {
	if depth > 32 {
		return fmt.Errorf("name tree too deep")
	}
	if names, ok := deref(node[model.Name("Names")]).(model.Array); ok {
		for i := 1; i < len(names); i += 2 {
			if err := fn(names[i]); err != nil {
				return err
			}
		}
	}
	if kids, ok := deref(node[model.Name("Kids")]).(model.Array); ok {
		for _, k := range kids {
			if kid, ok := deref(k).(model.Dict); ok {
				if err := walkNameTree(kid, deref, depth+1, fn); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func textOf(o model.Object) string {
	switch v := o.(type) {
	case model.String:
		return v.Text()
	case model.HexString:
		return v.Text()
	}
	return ""
}

func embeddedFileFromSpec(spec model.Dict, deref func(model.Object) model.Object) (*EmbeddedFile, error) {
	f := &EmbeddedFile{Name: textOf(spec[model.Name("UF")]), Description: textOf(spec[model.Name("Desc")])}
	if f.Name == "" {
		f.Name = textOf(spec[model.Name("F")])
	}
	if rel, ok := spec[model.Name("AFRelationship")].(model.Name); ok {
		f.Relationship = string(rel)
	}
	ef, ok := deref(spec[model.Name("EF")]).(model.Dict)
	if !ok {
		return nil, nil
	}
	streamObj := ef[model.Name("UF")]
	if streamObj == nil {
		streamObj = ef[model.Name("F")]
	}
	st, ok := deref(streamObj).(*model.Stream)
	if !ok {
		return nil, fmt.Errorf("embedded file %q: missing stream", f.Name)
	}
	f.Data = st.Content
	if mime, ok := st.Dict[model.Name("Subtype")].(model.Name); ok {
		f.MIME = string(mime)
	}
	if params, ok := deref(st.Dict[model.Name("Params")]).(model.Dict); ok {
		if d := textOf(params[model.Name("ModDate")]); d != "" {
			if t, err := model.ParseDate(d); err == nil {
				f.Modified = t
			}
		}
	}
	return f, nil
}
