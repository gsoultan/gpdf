package system

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gsoultan/gpdf/font"
	"github.com/gsoultan/gpdf/font/truetype"
)

// StandardSubstitute describes the font program found to stand in for one of
// the standard 14 Type 1 fonts.
type StandardSubstitute struct {
	// Font is the parsed, embeddable substitute.
	Font font.EmbeddableFont
	// Path is the file the substitute was loaded from.
	Path string
	// MetricCompatible is true for the families designed to match the
	// standard fonts' advance widths (Liberation, URW Nimbus, Croscore). It
	// is false for a last-resort substitute such as DejaVu, which changes
	// line breaks when a document was laid out with the standard metrics.
	MetricCompatible bool
}

// standardCandidates lists, per standard font, the file basenames (without
// extension, compared case-insensitively) of metric-compatible substitutes
// first and approximate substitutes after the empty marker.
var standardCandidates = map[string][]string{
	"Helvetica":             {"LiberationSans-Regular", "NimbusSans-Regular", "NimbusSanL-Regu", "Arimo-Regular", "Arial", "", "FreeSans", "DejaVuSans"},
	"Helvetica-Bold":        {"LiberationSans-Bold", "NimbusSans-Bold", "NimbusSanL-Bold", "Arimo-Bold", "Arialbd", "Arial-Bold", "", "FreeSansBold", "DejaVuSans-Bold"},
	"Helvetica-Oblique":     {"LiberationSans-Italic", "NimbusSans-Italic", "NimbusSanL-ReguItal", "Arimo-Italic", "Ariali", "Arial-Italic", "", "FreeSansOblique", "DejaVuSans-Oblique"},
	"Helvetica-BoldOblique": {"LiberationSans-BoldItalic", "NimbusSans-BoldItalic", "NimbusSanL-BoldItal", "Arimo-BoldItalic", "Arialbi", "Arial-BoldItalic", "", "FreeSansBoldOblique", "DejaVuSans-BoldOblique"},
	"Times-Roman":           {"LiberationSerif-Regular", "NimbusRoman-Regular", "NimbusRomNo9L-Regu", "Tinos-Regular", "Times", "Times New Roman", "", "FreeSerif", "DejaVuSerif"},
	"Times-Bold":            {"LiberationSerif-Bold", "NimbusRoman-Bold", "NimbusRomNo9L-Medi", "Tinos-Bold", "Timesbd", "Times New Roman Bold", "", "FreeSerifBold", "DejaVuSerif-Bold"},
	"Times-Italic":          {"LiberationSerif-Italic", "NimbusRoman-Italic", "NimbusRomNo9L-ReguItal", "Tinos-Italic", "Timesi", "Times New Roman Italic", "", "FreeSerifItalic", "DejaVuSerif-Italic"},
	"Times-BoldItalic":      {"LiberationSerif-BoldItalic", "NimbusRoman-BoldItalic", "NimbusRomNo9L-MediItal", "Tinos-BoldItalic", "Timesbi", "Times New Roman Bold Italic", "", "FreeSerifBoldItalic", "DejaVuSerif-BoldItalic"},
	"Courier":               {"LiberationMono-Regular", "NimbusMonoPS-Regular", "NimbusMonL-Regu", "Cousine-Regular", "Cour", "Courier New", "", "FreeMono", "DejaVuSansMono"},
	"Courier-Bold":          {"LiberationMono-Bold", "NimbusMonoPS-Bold", "NimbusMonL-Bold", "Cousine-Bold", "Courbd", "Courier New Bold", "", "FreeMonoBold", "DejaVuSansMono-Bold"},
	"Courier-Oblique":       {"LiberationMono-Italic", "NimbusMonoPS-Italic", "NimbusMonL-ReguObli", "Cousine-Italic", "Couri", "Courier New Italic", "", "FreeMonoOblique", "DejaVuSansMono-Oblique"},
	"Courier-BoldOblique":   {"LiberationMono-BoldItalic", "NimbusMonoPS-BoldItalic", "NimbusMonL-BoldObli", "Cousine-BoldItalic", "Courbi", "Courier New Bold Italic", "", "FreeMonoBoldOblique", "DejaVuSansMono-BoldOblique"},
	"Symbol":                {"StandardSymbolsPS", "StandardSymL", "Symbol"},
	"ZapfDingbats":          {"D050000L", "Dingbats", "ZapfDingbats"},
}

// StandardFontNames lists the standard 14 fonts in the order
// StandardSubstitutes reports them.
var StandardFontNames = []string{
	"Helvetica", "Helvetica-Bold", "Helvetica-Oblique", "Helvetica-BoldOblique",
	"Times-Roman", "Times-Bold", "Times-Italic", "Times-BoldItalic",
	"Courier", "Courier-Bold", "Courier-Oblique", "Courier-BoldOblique",
	"Symbol", "ZapfDingbats",
}

// StandardSubstitutes looks through the files the scanner finds for fonts
// that can be embedded in place of the standard 14 fonts: Liberation, URW
// Nimbus and Croscore families match the standard metrics; FreeFont and
// DejaVu are used only when nothing better is installed. Standard fonts with
// no substitute on the system are absent from the result.
func StandardSubstitutes(scanner FontScanner) map[string]StandardSubstitute {
	paths, _ := scanner.Scan()
	byBase := map[string]string{}
	for _, p := range paths {
		base := strings.ToLower(strings.TrimSuffix(filepath.Base(p), filepath.Ext(p)))
		if _, seen := byBase[base]; !seen {
			byBase[base] = p
		}
	}
	out := map[string]StandardSubstitute{}
	for _, name := range StandardFontNames {
		compatible := true
		for _, cand := range standardCandidates[name] {
			if cand == "" {
				compatible = false
				continue
			}
			p, ok := byBase[strings.ToLower(cand)]
			if !ok {
				continue
			}
			f, err := loadFont(p)
			if err != nil {
				continue
			}
			out[name] = StandardSubstitute{Font: f, Path: p, MetricCompatible: compatible}
			break
		}
	}
	return out
}

var (
	defaultSubstitutesOnce sync.Once
	defaultSubstitutes     map[string]StandardSubstitute
)

// DefaultStandardSubstitutes returns StandardSubstitutes for the default
// scanner of this operating system, computed once per process.
func DefaultStandardSubstitutes() map[string]StandardSubstitute {
	defaultSubstitutesOnce.Do(func() {
		defaultSubstitutes = StandardSubstitutes(GetDefaultScanner())
	})
	return defaultSubstitutes
}

// loadFont parses one font file (the first face of a collection).
func loadFont(path string) (font.EmbeddableFont, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if fonts, err := truetype.ParseCollection(data); err == nil && len(fonts) > 0 {
		return fonts[0], nil
	}
	return truetype.Parse(data)
}
