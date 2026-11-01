package truetype

import (
	"encoding/binary"
	"fmt"
)

// GSUB holds the glyph substitutions the shaper applies: ligatures from every
// ligature lookup in the font, and the single substitutions of the Arabic
// contextual-form features (isol, init, medi, fina) and the required-ligature
// feature of the Arabic script.
type GSUB struct {
	// Ligatures maps a packed sequence of 2-byte glyph IDs (2 to 3 glyphs) to
	// the ligature glyph.
	Ligatures map[string]uint16
	// Forms maps an OpenType feature tag (init, medi, fina, isol) to the
	// single substitutions it performs for the Arabic script.
	Forms map[string]map[uint16]uint16
}

// HasArabicForms reports whether the font carries contextual-form features
// for the Arabic script, so that joining can use real glyph substitutions
// instead of Unicode presentation forms.
func (g *GSUB) HasArabicForms() bool {
	return g != nil && (len(g.Forms["init"]) > 0 || len(g.Forms["fina"]) > 0 || len(g.Forms["medi"]) > 0)
}

func (f *Font) parseGSUB() error {
	d, err := f.tableData("GSUB")
	if err != nil {
		return nil // Optional
	}
	if len(d) < 10 {
		return fmt.Errorf("truetype: GSUB table too short")
	}
	scriptListOff := int(binary.BigEndian.Uint16(d[4:6]))
	featureListOff := int(binary.BigEndian.Uint16(d[6:8]))
	lookupListOff := int(binary.BigEndian.Uint16(d[8:10]))
	if lookupListOff+2 > len(d) {
		return fmt.Errorf("truetype: GSUB lookup list offset out of bounds")
	}

	numLookups := int(binary.BigEndian.Uint16(d[lookupListOff : lookupListOff+2]))
	f.gsub = &GSUB{Ligatures: make(map[string]uint16), Forms: map[string]map[uint16]uint16{}}

	lookupOffset := func(i int) (int, bool) {
		off := lookupListOff + 2 + i*2
		if i < 0 || i >= numLookups || off+2 > len(d) {
			return 0, false
		}
		abs := lookupListOff + int(binary.BigEndian.Uint16(d[off:off+2]))
		if abs+6 > len(d) {
			return 0, false
		}
		return abs, true
	}

	// Ligatures come from the features that are on by default (liga, clig,
	// rlig) of every script; discretionary and stylistic ligatures stay off.
	// Arabic contextual forms come from the init/medi/fina/isol features of
	// the 'arab' script (falling back to 'DFLT').
	ligatureLookups := map[int]bool{}
	for _, fi := range allScriptFeatures(d, scriptListOff) {
		tag, lookups := featureLookups(d, featureListOff, fi)
		switch tag {
		case "liga", "clig", "rlig":
			for _, li := range lookups {
				ligatureLookups[li] = true
			}
		}
	}
	for li := range ligatureLookups {
		abs, ok := lookupOffset(li)
		if !ok {
			continue
		}
		f.forEachSubtable(d, abs, func(lookupType uint16, sub []byte) {
			if lookupType == 4 {
				f.parseGSUBType4(sub)
			}
		})
	}

	featureIdx := scriptFeatures(d, scriptListOff, "arab")
	if featureIdx == nil {
		featureIdx = scriptFeatures(d, scriptListOff, "DFLT")
	}
	for _, fi := range featureIdx {
		tag, lookups := featureLookups(d, featureListOff, fi)
		switch tag {
		case "init", "medi", "fina", "isol":
		default:
			continue
		}
		m := f.gsub.Forms[tag]
		if m == nil {
			m = map[uint16]uint16{}
			f.gsub.Forms[tag] = m
		}
		for _, li := range lookups {
			abs, ok := lookupOffset(li)
			if !ok {
				continue
			}
			f.forEachSubtable(d, abs, func(lookupType uint16, sub []byte) {
				if lookupType == 1 {
					parseGSUBType1(sub, m)
				}
			})
		}
	}
	return nil
}

// allScriptFeatures returns the feature indices of the default language
// system of every script in the ScriptList.
func allScriptFeatures(d []byte, scriptListOff int) []int {
	if scriptListOff+2 > len(d) {
		return nil
	}
	count := int(binary.BigEndian.Uint16(d[scriptListOff:]))
	var out []int
	for i := range count {
		rec := scriptListOff + 2 + i*6
		if rec+6 > len(d) {
			break
		}
		out = append(out, scriptFeatures(d, scriptListOff, string(d[rec:rec+4]))...)
	}
	return out
}

// forEachSubtable calls fn for every subtable of the lookup at abs, unwrapping
// extension (type 7) lookups.
func (f *Font) forEachSubtable(d []byte, abs int, fn func(lookupType uint16, sub []byte)) {
	lookupType := binary.BigEndian.Uint16(d[abs : abs+2])
	numSubtables := int(binary.BigEndian.Uint16(d[abs+4 : abs+6]))
	for j := range numSubtables {
		p := abs + 6 + j*2
		if p+2 > len(d) {
			return
		}
		subOff := abs + int(binary.BigEndian.Uint16(d[p:p+2]))
		if subOff >= len(d) {
			continue
		}
		sub := d[subOff:]
		t := lookupType
		if t == 7 {
			// Extension: format, extensionLookupType, extensionOffset (32-bit).
			if len(sub) < 8 {
				continue
			}
			t = binary.BigEndian.Uint16(sub[2:4])
			ext := int(binary.BigEndian.Uint32(sub[4:8]))
			if ext >= len(sub) {
				continue
			}
			sub = sub[ext:]
		}
		fn(t, sub)
	}
}

// scriptFeatures returns the feature indices of the default language system
// of the script with the given tag, or nil when the script is absent.
func scriptFeatures(d []byte, scriptListOff int, tag string) []int {
	if scriptListOff+2 > len(d) {
		return nil
	}
	count := int(binary.BigEndian.Uint16(d[scriptListOff:]))
	for i := range count {
		rec := scriptListOff + 2 + i*6
		if rec+6 > len(d) {
			return nil
		}
		if string(d[rec:rec+4]) != tag {
			continue
		}
		script := scriptListOff + int(binary.BigEndian.Uint16(d[rec+4:]))
		if script+4 > len(d) {
			return nil
		}
		langSysOff := int(binary.BigEndian.Uint16(d[script:]))
		if langSysOff == 0 {
			// No default language system: use the first listed one.
			n := int(binary.BigEndian.Uint16(d[script+2:]))
			if n == 0 || script+10 > len(d) {
				return nil
			}
			langSysOff = int(binary.BigEndian.Uint16(d[script+8:]))
		}
		ls := script + langSysOff
		if ls+6 > len(d) {
			return nil
		}
		req := int(binary.BigEndian.Uint16(d[ls+2:]))
		n := int(binary.BigEndian.Uint16(d[ls+4:]))
		var out []int
		if req != 0xFFFF {
			out = append(out, req)
		}
		for k := range n {
			p := ls + 6 + k*2
			if p+2 > len(d) {
				break
			}
			out = append(out, int(binary.BigEndian.Uint16(d[p:])))
		}
		return out
	}
	return nil
}

// featureLookups returns the tag and lookup indices of feature number idx.
func featureLookups(d []byte, featureListOff, idx int) (string, []int) {
	if featureListOff+2 > len(d) {
		return "", nil
	}
	count := int(binary.BigEndian.Uint16(d[featureListOff:]))
	if idx < 0 || idx >= count {
		return "", nil
	}
	rec := featureListOff + 2 + idx*6
	if rec+6 > len(d) {
		return "", nil
	}
	tag := string(d[rec : rec+4])
	feat := featureListOff + int(binary.BigEndian.Uint16(d[rec+4:]))
	if feat+4 > len(d) {
		return tag, nil
	}
	n := int(binary.BigEndian.Uint16(d[feat+2:]))
	var out []int
	for k := range n {
		p := feat + 4 + k*2
		if p+2 > len(d) {
			break
		}
		out = append(out, int(binary.BigEndian.Uint16(d[p:])))
	}
	return tag, out
}

// parseGSUBType1 adds the single substitutions of one subtable to m.
func parseGSUBType1(d []byte, m map[uint16]uint16) {
	if len(d) < 6 {
		return
	}
	format := binary.BigEndian.Uint16(d[0:2])
	coverageOff := int(binary.BigEndian.Uint16(d[2:4]))
	if coverageOff >= len(d) {
		return
	}
	covered := parseCoverageTable(d[coverageOff:])
	switch format {
	case 1:
		delta := int16(binary.BigEndian.Uint16(d[4:6]))
		for _, g := range covered {
			m[g] = uint16(int(g) + int(delta))
		}
	case 2:
		n := int(binary.BigEndian.Uint16(d[4:6]))
		if len(d) < 6+n*2 {
			return
		}
		for i, g := range covered {
			if i >= n {
				break
			}
			m[g] = binary.BigEndian.Uint16(d[6+i*2:])
		}
	}
}

func (f *Font) parseGSUBType4(d []byte) {
	if len(d) < 6 {
		return
	}
	// format := binary.BigEndian.Uint16(d[0:2])
	coverageOff := binary.BigEndian.Uint16(d[2:4])
	numSets := int(binary.BigEndian.Uint16(d[4:6]))
	if len(d) < 6+numSets*2 || int(coverageOff) >= len(d) {
		return
	}

	// Coverage table tells us which glyph IDs start a ligature
	starts := parseCoverageTable(d[coverageOff:])
	if len(starts) != numSets {
		return
	}

	for i, startGID := range starts {
		setOff := int(binary.BigEndian.Uint16(d[6+i*2 : 8+i*2]))
		if setOff+2 > len(d) {
			continue
		}
		numLigs := int(binary.BigEndian.Uint16(d[setOff : setOff+2]))
		for k := range numLigs {
			p := setOff + 2 + k*2
			if p+2 > len(d) {
				break
			}
			ligOff := int(binary.BigEndian.Uint16(d[p : p+2]))
			if setOff+ligOff >= len(d) {
				continue
			}
			f.parseLigature(d[setOff+ligOff:], startGID)
		}
	}
}

// parseCoverageTable returns the covered glyphs in coverage-index order.
func parseCoverageTable(d []byte) []uint16 {
	if len(d) < 4 {
		return nil
	}
	format := binary.BigEndian.Uint16(d[0:2])
	if format == 1 {
		numGlyphs := int(binary.BigEndian.Uint16(d[2:4]))
		if len(d) < 4+numGlyphs*2 {
			return nil
		}
		res := make([]uint16, numGlyphs)
		for i := range numGlyphs {
			res[i] = binary.BigEndian.Uint16(d[4+i*2 : 6+i*2])
		}
		return res
	} else if format == 2 {
		numRanges := int(binary.BigEndian.Uint16(d[2:4]))
		if len(d) < 4+numRanges*6 {
			return nil
		}
		var res []uint16
		for i := range numRanges {
			start := binary.BigEndian.Uint16(d[4+i*6 : 6+i*6])
			end := binary.BigEndian.Uint16(d[6+i*6 : 8+i*6])
			if end < start || int(end)-int(start) > 65535 {
				return nil
			}
			for g := int(start); g <= int(end); g++ {
				res = append(res, uint16(g))
			}
		}
		return res
	}
	return nil
}

func (f *Font) parseLigature(d []byte, startGID uint16) {
	if len(d) < 4 {
		return
	}
	ligGlyph := binary.BigEndian.Uint16(d[0:2])
	compCount := int(binary.BigEndian.Uint16(d[2:4]))
	if compCount < 1 || len(d) < 4+(compCount-1)*2 {
		return
	}
	key := make([]byte, compCount*2)
	binary.BigEndian.PutUint16(key[0:2], startGID)
	for i := range compCount - 1 {
		gid := binary.BigEndian.Uint16(d[4+i*2 : 6+i*2])
		binary.BigEndian.PutUint16(key[2+i*2:4+i*2], gid)
	}
	f.gsub.Ligatures[string(key)] = ligGlyph
}
