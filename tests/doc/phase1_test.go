// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"
	"testing"

	"github.com/gsoultan/gpdf/doc"
	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/reader"
)

func TestPortableCollection(t *testing.T) {
	b := doc.New().PageSize(200, 100).AddPage()
	b.AttachFile(doc.Attachment{Name: "b.txt", Data: []byte("b"), CollectionItem: map[string]any{"owner": "Ümit", "pages": 3}})
	b.AttachFile(doc.Attachment{Name: "a.txt", Data: []byte("a"), CollectionItem: map[string]any{"owner": "Ann", "pages": 1}})
	b.Collection(doc.CollectionOptions{
		View: doc.CollectionDetails,
		Fields: []doc.CollectionField{
			{Key: "owner", Name: "Owner", Type: doc.CollectionText},
			{Key: "pages", Name: "Pages", Type: doc.CollectionNumber},
			{Key: "size", Name: "Size", Type: doc.CollectionSize, Hidden: true},
		},
		SortKey: "owner", Initial: "a.txt",
	})
	d := reopen(t, savePDF(t, b))
	cat, _ := d.Catalog()
	colObj, err := d.Resolve(*cat.CollectionRef())
	if err != nil {
		t.Fatal(err)
	}
	col := colObj.(model.Dict)
	schema := col[model.Name("Schema")].(model.Dict)
	owner := schema[model.Name("owner")].(model.Dict)
	if owner[model.Name("Subtype")] != model.Name("S") || owner[model.Name("O")] != model.Integer(1) {
		t.Errorf("owner field = %v", owner)
	}
	if size := schema[model.Name("size")].(model.Dict); size[model.Name("V")] != model.Boolean(false) {
		t.Errorf("hidden field = %v", size)
	}
	if sort := col[model.Name("Sort")].(model.Dict); sort[model.Name("S")] != model.Name("owner") {
		t.Errorf("sort = %v", sort)
	}
	if col[model.Name("D")].(model.String).Text() != "a.txt" {
		t.Errorf("initial = %v", col[model.Name("D")])
	}
	// Each attachment carries its collection item; the name tree is sorted.
	names, _ := d.Resolve(cat.Dict[model.Name("Names")].(model.Dict)[model.Name("EmbeddedFiles")].(model.Ref))
	arr := names.(model.Dict)[model.Name("Names")].(model.Array)
	if arr[0].(model.String) != "a.txt" || arr[2].(model.String) != "b.txt" {
		t.Errorf("name tree = %v", arr)
	}
	spec, _ := d.Resolve(arr[3].(model.Ref))
	ci := spec.(model.Dict)[model.Name("CI")].(model.Dict)
	if ci[model.Name("owner")].(model.String).Text() != "Ümit" || ci[model.Name("pages")] != model.Integer(3) {
		t.Errorf("CI = %v", ci)
	}
	popplerCheck(t, "pdfinfo", savePDF(t, b))
}

func TestGeoViewportRoundTrip(t *testing.T) {
	b := doc.New().PageSize(400, 300).AddPage()
	p := b.CurrentPage()
	// A map of a 1° x 1° area: lower-left at 6°S 106°E (Jakarta-ish), rectangle 50..350 x 50..250.
	p.AddGeoViewport(doc.GeoViewport{
		Name: "Map", Rect: [4]float64{50, 50, 350, 250},
		GPTS:         [4][2]float64{{-7, 106}, {-6, 106}, {-6, 107}, {-7, 107}},
		DisplayUnits: [3]string{"KM", "SQKM", "DEG"},
	})
	d := reopen(t, savePDF(t, b))
	pages, _ := d.Pages()
	vps, err := reader.ExtractGeoViewports(d, pages[0])
	if err != nil || len(vps) != 1 {
		t.Fatalf("%v %+v", err, vps)
	}
	g := vps[0]
	if g.Name != "Map" || g.EPSG != 4326 || len(g.GPTS) != 4 {
		t.Errorf("%+v", g)
	}
	lat, lon := g.ToGeo(200, 150) // centre
	if math.Abs(lat+6.5) > 1e-9 || math.Abs(lon-106.5) > 1e-9 {
		t.Errorf("centre = %v, %v", lat, lon)
	}
	lat, lon = g.ToGeo(350, 50) // lower-right corner
	if math.Abs(lat+7) > 1e-9 || math.Abs(lon-107) > 1e-9 {
		t.Errorf("lower right = %v, %v", lat, lon)
	}
	measure := pages[0].Viewports()[0].(model.Dict)[model.Name("Measure")].(model.Dict)
	if pdu := measure[model.Name("PDU")].(model.Array); pdu[0] != model.Name("KM") {
		t.Errorf("PDU = %v", pdu)
	}
}

func TestAdd3DAnnotation(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for i := range 16 {
		img.Set(i%4, i/4, color.NRGBA{R: 200, G: 30, B: 30, A: 255})
	}
	var poster bytes.Buffer
	png.Encode(&poster, img)
	model3d := []byte("U3D\x00fake model bytes")
	b := doc.New().PageSize(300, 300).AddPage()
	b.CurrentPage().Add3D([4]float64{20, 20, 280, 280}, doc.Model3D{Format: "U3D", Data: model3d, Name: "Part", Poster: poster.Bytes(), ActivateOnOpen: true})
	pdf := savePDF(t, b)
	d := reopen(t, pdf)
	pages, _ := d.Pages()
	annots := pages[0].Dict[model.Name("Annots")].(model.Array)
	if len(annots) != 1 {
		t.Fatalf("annots = %v", annots)
	}
	a, _ := d.Resolve(annots[0].(model.Ref))
	ad := a.(model.Dict)
	if ad[model.Name("Subtype")] != model.Name("3D") || ad[model.Name("3DA")].(model.Dict)[model.Name("A")] != model.Name("PO") {
		t.Errorf("annotation = %v", ad)
	}
	st, _ := d.Resolve(ad[model.Name("3DD")].(model.Ref))
	s := st.(*model.Stream)
	if s.Dict[model.Name("Subtype")] != model.Name("U3D") || !bytes.Equal(s.Content, model3d) {
		t.Errorf("3D stream = %v (%d bytes)", s.Dict, len(s.Content))
	}
	if _, ok := ad[model.Name("AP")]; !ok {
		t.Error("no poster appearance")
	}
	if _, err := doc.New().AddPage().CurrentPage().Add3D([4]float64{0, 0, 1, 1}, doc.Model3D{Format: "OBJ", Data: []byte("x")}).Builder().Build(); err == nil || !strings.Contains(err.Error(), "U3D or PRC") {
		t.Errorf("bad format accepted: %v", err)
	}
	popplerCheck(t, "pdftoppm", pdf, "-r", "20", "-png", "-singlefile", "-o")
}
