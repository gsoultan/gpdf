// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package style

import (
	"encoding/binary"
	"math"
)

// SRGBICCProfile returns an sRGB profile with an embedded ICC v2 display profile, as
// PDF/A output intents need. The profile has the D50-adapted sRGB primaries, the D50
// white point and the sRGB transfer curve as a 1024-point table.
func SRGBICCProfile() ICCProfile {
	return ICCProfile{Data: buildSRGBICC(), N: 3, Alternate: "DeviceRGB"}
}

func s15Fixed16(v float64) uint32 {
	return uint32(int32(math.Round(v * 65536)))
}

func iccXYZ(x, y, z float64) []byte {
	b := make([]byte, 20)
	copy(b, "XYZ ")
	binary.BigEndian.PutUint32(b[8:], s15Fixed16(x))
	binary.BigEndian.PutUint32(b[12:], s15Fixed16(y))
	binary.BigEndian.PutUint32(b[16:], s15Fixed16(z))
	return b
}

func iccCurve() []byte {
	const n = 1024
	b := make([]byte, 12+2*n)
	copy(b, "curv")
	binary.BigEndian.PutUint32(b[8:], n)
	for i := range n {
		x := float64(i) / (n - 1)
		var y float64
		if x <= 0.04045 {
			y = x / 12.92
		} else {
			y = math.Pow((x+0.055)/1.055, 2.4)
		}
		binary.BigEndian.PutUint16(b[12+2*i:], uint16(math.Round(y*65535)))
	}
	return b
}

func iccDesc(s string) []byte {
	// textDescriptionType: ASCII (with NUL), empty Unicode and ScriptCode parts.
	b := make([]byte, 0, 8+4+len(s)+1+4+4+2+1+67)
	b = append(b, "desc\x00\x00\x00\x00"...)
	b = binary.BigEndian.AppendUint32(b, uint32(len(s)+1))
	b = append(b, s...)
	b = append(b, 0)
	b = append(b, make([]byte, 4+4+2+1+67)...)
	return b
}

func iccText(s string) []byte {
	b := append([]byte("text\x00\x00\x00\x00"), s...)
	return append(b, 0)
}

func buildSRGBICC() []byte {
	type tag struct {
		sig  string
		data []byte
	}
	curve := iccCurve()
	tags := []tag{
		{"desc", iccDesc("sRGB IEC61966-2.1")},
		{"cprt", iccText("Public domain")},
		{"wtpt", iccXYZ(0.9642, 1.0, 0.8249)},
		{"rXYZ", iccXYZ(0.43607, 0.22249, 0.01392)},
		{"gXYZ", iccXYZ(0.38515, 0.71687, 0.09708)},
		{"bXYZ", iccXYZ(0.14307, 0.06061, 0.71410)},
		{"rTRC", curve},
		{"gTRC", curve},
		{"bTRC", curve},
	}
	header := make([]byte, 128)
	binary.BigEndian.PutUint32(header[8:], 0x02100000) // version 2.1
	copy(header[12:], "mntr")
	copy(header[16:], "RGB ")
	copy(header[20:], "XYZ ")
	binary.BigEndian.PutUint16(header[24:], 2026) // creation date
	binary.BigEndian.PutUint16(header[26:], 1)
	binary.BigEndian.PutUint16(header[28:], 1)
	copy(header[36:], "acsp")
	binary.BigEndian.PutUint32(header[68:], s15Fixed16(0.9642)) // D50 illuminant
	binary.BigEndian.PutUint32(header[72:], s15Fixed16(1.0))
	binary.BigEndian.PutUint32(header[76:], s15Fixed16(0.8249))

	table := binary.BigEndian.AppendUint32(nil, uint32(len(tags)))
	offset := 128 + 4 + 12*len(tags)
	var body []byte
	for _, t := range tags {
		table = append(table, t.sig...)
		table = binary.BigEndian.AppendUint32(table, uint32(offset+len(body)))
		table = binary.BigEndian.AppendUint32(table, uint32(len(t.data)))
		body = append(body, t.data...)
		for len(body)%4 != 0 {
			body = append(body, 0)
		}
	}
	out := append(append(header, table...), body...)
	binary.BigEndian.PutUint32(out[0:], uint32(len(out)))
	return out
}
