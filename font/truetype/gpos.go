package truetype

import (
	"encoding/binary"
	"math/bits"
)

// gposKern holds the pair adjustments of the GPOS 'kern' feature: explicit
// glyph pairs (PairPos format 1) and class-based pairs (format 2). Modern
// fonts often carry kerning only here, without a legacy 'kern' table.
type gposKern struct {
	lookups [][]pairSubtable // per lookup, subtables in order
}

type pairSubtable struct {
	pairs map[uint32]int16 // format 1: first<<16|second -> x advance of the first glyph

	// format 2: tables are looked up in place (never expanded, so a hostile
	// font cannot make parsing slow)
	cover          coverage
	class1, class2 classDef
	n2             int
	values         []int16
}

// adjust returns the x advance adjustment (font units) applied between the
// glyphs a and b: the first matching subtable of every lookup, summed.
func (k *gposKern) adjust(a, b uint16) int {
	total := 0
	key := uint32(a)<<16 | uint32(b)
	for _, subs := range k.lookups {
		for i := range subs {
			st := &subs[i]
			if st.pairs != nil {
				if v, ok := st.pairs[key]; ok {
					total += int(v)
					break
				}
				continue
			}
			if _, ok := st.cover.index(a); !ok {
				continue
			}
			c1, c2 := st.class1.class(a), st.class2.class(b)
			if i := c1*st.n2 + c2; i < len(st.values) {
				total += int(st.values[i])
			}
			break
		}
	}
	return total
}

// coverage is an OpenType Coverage table read in place.
type coverage []byte

// index returns the coverage index of g.
func (c coverage) index(g uint16) (int, bool) {
	if len(c) < 4 {
		return 0, false
	}
	n := int(binary.BigEndian.Uint16(c[2:]))
	switch binary.BigEndian.Uint16(c) {
	case 1:
		n = min(n, (len(c)-4)/2)
		lo, hi := 0, n
		for lo < hi {
			m := (lo + hi) / 2
			v := binary.BigEndian.Uint16(c[4+m*2:])
			switch {
			case v == g:
				return m, true
			case v < g:
				lo = m + 1
			default:
				hi = m
			}
		}
	case 2:
		n = min(n, (len(c)-4)/6)
		lo, hi := 0, n
		for lo < hi {
			m := (lo + hi) / 2
			r := c[4+m*6:]
			s, e := binary.BigEndian.Uint16(r), binary.BigEndian.Uint16(r[2:])
			switch {
			case g < s:
				hi = m
			case g > e:
				lo = m + 1
			default:
				return int(binary.BigEndian.Uint16(r[4:])) + int(g-s), true
			}
		}
	}
	return 0, false
}

// coverageGlyphs lists the first n glyphs of a coverage table in index
// order (n is bounded by the PairSet count, which the table pays for in bytes).
func coverageGlyphs(c coverage, n int) []uint16 {
	if len(c) < 4 {
		return nil
	}
	out := make([]uint16, 0, min(n, 4096))
	cnt := int(binary.BigEndian.Uint16(c[2:]))
	switch binary.BigEndian.Uint16(c) {
	case 1:
		for i := 0; i < cnt && len(out) < n && 4+i*2+2 <= len(c); i++ {
			out = append(out, binary.BigEndian.Uint16(c[4+i*2:]))
		}
	case 2:
		for i := 0; i < cnt && len(out) < n && 4+i*6+6 <= len(c); i++ {
			s, e := int(binary.BigEndian.Uint16(c[4+i*6:])), int(binary.BigEndian.Uint16(c[6+i*6:]))
			for g := s; g <= e && len(out) < n; g++ {
				out = append(out, uint16(g))
			}
		}
	}
	return out
}

// classDef is an OpenType ClassDef table read in place.
type classDef []byte

// class returns the class of g (0 when g is not listed).
func (d classDef) class(g uint16) int {
	if len(d) < 4 {
		return 0
	}
	switch binary.BigEndian.Uint16(d) {
	case 1:
		if len(d) < 6 {
			return 0
		}
		start := binary.BigEndian.Uint16(d[2:])
		n := int(binary.BigEndian.Uint16(d[4:]))
		if g < start || int(g-start) >= n || 6+int(g-start)*2+2 > len(d) {
			return 0
		}
		return int(binary.BigEndian.Uint16(d[6+int(g-start)*2:]))
	case 2:
		n := min(int(binary.BigEndian.Uint16(d[2:])), (len(d)-4)/6)
		lo, hi := 0, n
		for lo < hi {
			m := (lo + hi) / 2
			r := d[4+m*6:]
			s, e := binary.BigEndian.Uint16(r), binary.BigEndian.Uint16(r[2:])
			switch {
			case g < s:
				hi = m
			case g > e:
				lo = m + 1
			default:
				return int(binary.BigEndian.Uint16(r[4:]))
			}
		}
	}
	return 0
}

// valueRecordSize returns the size in bytes of a ValueRecord with format f.
func valueRecordSize(f uint16) int { return 2 * bits.OnesCount16(f) }

// xAdvance reads the XAdvance field of the ValueRecord at d with format f.
func xAdvance(d []byte, f uint16) (int16, bool) {
	if f&0x4 == 0 {
		return 0, false
	}
	off := 2 * bits.OnesCount16(f&0x3)
	if off+2 > len(d) {
		return 0, false
	}
	return int16(binary.BigEndian.Uint16(d[off:])), true
}

func parsePairPos(d []byte) (pairSubtable, bool) {
	if len(d) < 10 {
		return pairSubtable{}, false
	}
	format := binary.BigEndian.Uint16(d)
	covOff := int(binary.BigEndian.Uint16(d[2:]))
	vf1, vf2 := binary.BigEndian.Uint16(d[4:]), binary.BigEndian.Uint16(d[6:])
	s1, s2 := valueRecordSize(vf1), valueRecordSize(vf2)
	if covOff >= len(d) {
		return pairSubtable{}, false
	}
	cov := coverage(d[covOff:])
	switch format {
	case 1:
		n := int(binary.BigEndian.Uint16(d[8:]))
		first := coverageGlyphs(cov, n)
		st := pairSubtable{pairs: map[uint32]int16{}}
		for i := 0; i < n && i < len(first); i++ {
			p := 10 + i*2
			if p+2 > len(d) {
				break
			}
			set := int(binary.BigEndian.Uint16(d[p:]))
			if set+2 > len(d) {
				continue
			}
			cnt := int(binary.BigEndian.Uint16(d[set:]))
			rec := 2 + s1 + s2
			for j := range cnt {
				r := set + 2 + j*rec
				if r+rec > len(d) {
					break
				}
				if v, ok := xAdvance(d[r+2:], vf1); ok && v != 0 {
					st.pairs[uint32(first[i])<<16|uint32(binary.BigEndian.Uint16(d[r:]))] = v
				}
			}
		}
		return st, len(st.pairs) > 0
	case 2:
		if len(d) < 16 {
			return pairSubtable{}, false
		}
		cd1, cd2 := int(binary.BigEndian.Uint16(d[8:])), int(binary.BigEndian.Uint16(d[10:]))
		n1, n2 := int(binary.BigEndian.Uint16(d[12:])), int(binary.BigEndian.Uint16(d[14:]))
		if cd1 >= len(d) || cd2 >= len(d) || n1 == 0 || n2 == 0 || n1*n2 > 1<<22 {
			return pairSubtable{}, false
		}
		rec := s1 + s2
		if rec == 0 || 16+n1*n2*rec > len(d) {
			return pairSubtable{}, false
		}
		st := pairSubtable{cover: cov, class1: classDef(d[cd1:]), class2: classDef(d[cd2:]), n2: n2,
			values: make([]int16, n1*n2)}
		has := false
		for i := range n1 * n2 {
			r := 16 + i*rec
			if r+rec > len(d) {
				break
			}
			if v, ok := xAdvance(d[r:], vf1); ok && v != 0 {
				st.values[i] = v
				has = true
			}
		}
		return st, has
	}
	return pairSubtable{}, false
}

// parseGPOS reads the pair kerning of the 'kern' feature (all scripts).
func (f *Font) parseGPOS() {
	d, err := f.tableData("GPOS")
	if err != nil || len(d) < 10 {
		return
	}
	scriptListOff := int(binary.BigEndian.Uint16(d[4:6]))
	featureListOff := int(binary.BigEndian.Uint16(d[6:8]))
	lookupListOff := int(binary.BigEndian.Uint16(d[8:10]))
	if lookupListOff+2 > len(d) {
		return
	}
	numLookups := int(binary.BigEndian.Uint16(d[lookupListOff:]))
	order := featureTables(d, featureListOff, allScriptFeatures(d, scriptListOff), "kern")
	k := &gposKern{}
	for _, li := range sortedInts(order) {
		off := lookupListOff + 2 + li*2
		if li < 0 || li >= numLookups || off+2 > len(d) {
			continue
		}
		abs := lookupListOff + int(binary.BigEndian.Uint16(d[off:]))
		if abs+6 > len(d) {
			continue
		}
		var subs []pairSubtable
		forEachLookupSubtable(d, abs, 9, func(t uint16, sub []byte) {
			if t != 2 {
				return
			}
			if st, ok := parsePairPos(sub); ok {
				subs = append(subs, st)
			}
		})
		if len(subs) > 0 {
			k.lookups = append(k.lookups, subs)
		}
	}
	if len(k.lookups) > 0 {
		f.gpos = k
	}
}

// sortedInts returns v sorted ascending (lookups apply in LookupList order).
func sortedInts(v []int) []int {
	out := append([]int(nil), v...)
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// forEachLookupSubtable calls fn for every subtable of the lookup at abs,
// unwrapping extension lookups of type ext (7 in GSUB, 9 in GPOS).
func forEachLookupSubtable(d []byte, abs int, ext uint16, fn func(lookupType uint16, sub []byte)) {
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
		if t == ext {
			if len(sub) < 8 {
				continue
			}
			t = binary.BigEndian.Uint16(sub[2:4])
			e := int(binary.BigEndian.Uint32(sub[4:8]))
			if e >= len(sub) {
				continue
			}
			sub = sub[e:]
		}
		fn(t, sub)
	}
}
