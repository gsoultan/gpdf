// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc

import (
	"fmt"

	"github.com/gsoultan/gpdf/model"
)

// GeoViewport georeferences a rectangle of the page (ISO 32000-2 §12.10): the
// corners of Rect correspond to the geographic points in GPTS, so a viewer can show
// coordinates and measure distances.
type GeoViewport struct {
	Name string
	// Rect is the page area, in PDF units.
	Rect [4]float64
	// GPTS lists the geographic coordinates (latitude, longitude pairs) of the
	// rectangle's corners in the order lower-left, upper-left, upper-right,
	// lower-right, as /LPTS [0 0 0 1 1 1 1 0] expects.
	GPTS [4][2]float64
	// EPSG is the coordinate system code of the GPTS values (default 4326, WGS 84).
	// WKT may be given instead for a projected system.
	EPSG int
	WKT  string
	// DisplayUnits are the /PDU preferences: linear, area and angular units, such
	// as "M", "SQM", "DEG". Optional.
	DisplayUnits [3]string
}

// AddGeoViewport adds a georeferenced viewport to the page's /VP array.
func (p *Page) AddGeoViewport(v GeoViewport) *Page {
	if v.Rect[2] <= v.Rect[0] || v.Rect[3] <= v.Rect[1] {
		p.builder.setErr(fmt.Errorf("geo viewport needs a non-empty rectangle"))
		return p
	}
	gcs := model.Dict{model.Name("Type"): model.Name("GEOGCS")}
	switch {
	case v.WKT != "":
		gcs[model.Name("Type")] = model.Name("PROJCS")
		gcs[model.Name("WKT")] = model.String(v.WKT)
	default:
		epsg := v.EPSG
		if epsg == 0 {
			epsg = 4326
		}
		gcs[model.Name("EPSG")] = model.Integer(epsg)
	}
	gpts := make(model.Array, 0, 8)
	for _, pt := range v.GPTS {
		gpts = append(gpts, model.Real(pt[0]), model.Real(pt[1]))
	}
	measure := model.Dict{
		model.Name("Type"):    model.Name("Measure"),
		model.Name("Subtype"): model.Name("GEO"),
		model.Name("Bounds"):  model.Array{model.Integer(0), model.Integer(0), model.Integer(0), model.Integer(1), model.Integer(1), model.Integer(1), model.Integer(1), model.Integer(0)},
		model.Name("GPTS"):    gpts,
		model.Name("LPTS"):    model.Array{model.Integer(0), model.Integer(0), model.Integer(0), model.Integer(1), model.Integer(1), model.Integer(1), model.Integer(1), model.Integer(0)},
		model.Name("GCS"):     gcs,
	}
	if v.DisplayUnits != [3]string{} {
		measure[model.Name("PDU")] = model.Array{model.Name(v.DisplayUnits[0]), model.Name(v.DisplayUnits[1]), model.Name(v.DisplayUnits[2])}
	}
	vp := model.Dict{
		model.Name("Type"):    model.Name("Viewport"),
		model.Name("BBox"):    rectArray(v.Rect),
		model.Name("Measure"): measure,
	}
	if v.Name != "" {
		vp[model.Name("Name")] = model.String(v.Name)
	}
	ps := &p.builder.pc.pages[p.pageIndex]
	if ps.Dict == nil {
		ps.Dict = model.Dict{}
	}
	arr, _ := ps.Dict[model.Name("VP")].(model.Array)
	ps.Dict[model.Name("VP")] = append(arr, vp)
	return p
}
