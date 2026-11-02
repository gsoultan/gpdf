// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package reader

import (
	"fmt"

	"github.com/gsoultan/gpdf/model"
)

// GeoViewport is a georeferenced page area read from a page's /VP entry with a
// /Measure of subtype GEO.
type GeoViewport struct {
	Name string
	// BBox is the page area in PDF units.
	BBox [4]float64
	// GPTS are the geographic points (latitude, longitude, or the projected
	// system's coordinates) and LPTS the matching points in the unit square of BBox.
	GPTS [][2]float64
	LPTS [][2]float64
	EPSG int
	WKT  string
}

// ExtractGeoViewports returns the georeferenced viewports of a page.
func ExtractGeoViewports(r Resolver, page model.Page) ([]GeoViewport, error) {
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
	var out []GeoViewport
	for _, item := range page.Viewports() {
		vp, ok := deref(item).(model.Dict)
		if !ok {
			continue
		}
		m, ok := deref(vp[model.Name("Measure")]).(model.Dict)
		if !ok {
			continue
		}
		if st, _ := m[model.Name("Subtype")].(model.Name); st != "GEO" {
			continue
		}
		g := GeoViewport{Name: textOf(vp[model.Name("Name")])}
		bbox, _ := deref(vp[model.Name("BBox")]).(model.Array)
		if len(bbox) != 4 {
			return out, fmt.Errorf("geo viewport: /BBox needs four numbers")
		}
		for i := range 4 {
			g.BBox[i] = numberOf(bbox[i])
		}
		g.GPTS = pairs(deref(m[model.Name("GPTS")]))
		g.LPTS = pairs(deref(m[model.Name("LPTS")]))
		if len(g.LPTS) == 0 {
			g.LPTS = [][2]float64{{0, 0}, {0, 1}, {1, 1}, {1, 0}}
		}
		if gcs, ok := deref(m[model.Name("GCS")]).(model.Dict); ok {
			if e, ok := gcs[model.Name("EPSG")].(model.Integer); ok {
				g.EPSG = int(e)
			}
			g.WKT = textOf(gcs[model.Name("WKT")])
		}
		if len(g.GPTS) < 3 || len(g.GPTS) != len(g.LPTS) {
			return out, fmt.Errorf("geo viewport: GPTS and LPTS need the same number (at least three) of points")
		}
		out = append(out, g)
	}
	return out, nil
}

// ToGeo maps a page point (PDF units) inside the viewport to geographic
// coordinates by the affine transform fitted to the first three control points
// (viewports written by gPDF are rectangles, so the fit is exact).
func (g GeoViewport) ToGeo(x, y float64) (lat, lon float64) {
	u := (x - g.BBox[0]) / (g.BBox[2] - g.BBox[0])
	v := (y - g.BBox[1]) / (g.BBox[3] - g.BBox[1])
	// Solve geo = A*[u v 1] with three points.
	p, q := g.LPTS, g.GPTS
	det := (p[1][0]-p[0][0])*(p[2][1]-p[0][1]) - (p[2][0]-p[0][0])*(p[1][1]-p[0][1])
	if det == 0 {
		return q[0][0], q[0][1]
	}
	// Barycentric weights of (u, v) in the triangle p0 p1 p2.
	w1 := ((u-p[0][0])*(p[2][1]-p[0][1]) - (p[2][0]-p[0][0])*(v-p[0][1])) / det
	w2 := ((p[1][0]-p[0][0])*(v-p[0][1]) - (u-p[0][0])*(p[1][1]-p[0][1])) / det
	w0 := 1 - w1 - w2
	lat = w0*q[0][0] + w1*q[1][0] + w2*q[2][0]
	lon = w0*q[0][1] + w1*q[1][1] + w2*q[2][1]
	return lat, lon
}

func pairs(o model.Object) [][2]float64 {
	arr, ok := o.(model.Array)
	if !ok {
		return nil
	}
	out := make([][2]float64, 0, len(arr)/2)
	for i := 0; i+1 < len(arr); i += 2 {
		out = append(out, [2]float64{numberOf(arr[i]), numberOf(arr[i+1])})
	}
	return out
}

func numberOf(o model.Object) float64 {
	switch v := o.(type) {
	case model.Integer:
		return float64(v)
	case model.Real:
		return float64(v)
	}
	return 0
}
