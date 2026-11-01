package doc

import "github.com/gsoultan/gpdf/model"

// textStringKeys lists dictionary entries whose value is a PDF text string
// (ISO 32000-2 §7.9.2.2), which the builder writes from Go (UTF-8) strings.
var textStringKeys = map[model.Name]bool{
	"Title": true, "Author": true, "Subject": true, "Keywords": true, "Creator": true, "Producer": true,
	"T": true, "TU": true, "TM": true, "V": true, "DV": true, "Alt": true, "ActualText": true, "E": true,
	"Lang": true, "Contents": true, "Desc": true, "UF": true, "Subj": true, "NM": true, "Name": true,
	"OutputConditionIdentifier": true, "OutputCondition": true, "RegistryName": true, "Info": true,
}

// encodeTextStrings rewrites every text string in objs that is not plain ASCII as
// UTF-8 (PDF 2.0) or UTF-16BE with a byte order mark, so that readers decode it
// correctly instead of reading the UTF-8 bytes as PDFDocEncoding. Strings that
// already start with a byte order mark are left alone.
func encodeTextStrings(objs map[int]model.Object, utf8BOM bool) {
	for _, o := range objs {
		walkBuilt(o, 0, func(o model.Object) {
			d, ok := o.(model.Dict)
			if !ok {
				return
			}
			for k, v := range d {
				s, ok := v.(model.String)
				if !ok || !textStringKeys[k] || model.IsTextStringBOM([]byte(s)) {
					continue
				}
				if k == "Name" {
					if t, _ := d[model.Name("Type")].(model.Name); t != "OCG" {
						continue
					}
				}
				d[k] = model.EncodeTextString(string(s), utf8BOM)
			}
		})
	}
}
