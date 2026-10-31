package metadata

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// XMPInput is what BuildXMP writes into an XMP packet.
type XMPInput struct {
	Title, Author, Subject, Keywords, Creator, Producer string
	Created, Modified                                   time.Time
	// PDFAPart and PDFAConformance fill the pdfaid schema ("2", "B"). Empty leaves it out.
	PDFAPart, PDFAConformance string
}

// BuildXMP returns an XMP metadata packet. The values mirror the Info dictionary, as
// PDF/A requires.
func BuildXMP(in XMPInput) []byte {
	esc := func(s string) string {
		var sb strings.Builder
		xml.EscapeText(&sb, []byte(s))
		return sb.String()
	}
	var b strings.Builder
	b.WriteString("<?xpacket begin=\"\xEF\xBB\xBF\" id=\"W5M0MpCehiHzreSzNTczkc9d\"?>\n")
	b.WriteString("<x:xmpmeta xmlns:x=\"adobe:ns:meta/\">\n")
	b.WriteString(" <rdf:RDF xmlns:rdf=\"http://www.w3.org/1999/02/22-rdf-syntax-ns#\">\n")
	if in.PDFAPart != "" {
		fmt.Fprintf(&b, "  <rdf:Description rdf:about=\"\" xmlns:pdfaid=\"http://www.aiim.org/pdfa/ns/id/\">\n   <pdfaid:part>%s</pdfaid:part>\n   <pdfaid:conformance>%s</pdfaid:conformance>\n  </rdf:Description>\n", esc(in.PDFAPart), esc(in.PDFAConformance))
	}
	b.WriteString("  <rdf:Description rdf:about=\"\" xmlns:dc=\"http://purl.org/dc/elements/1.1/\">\n")
	b.WriteString("   <dc:format>application/pdf</dc:format>\n")
	if in.Title != "" {
		fmt.Fprintf(&b, "   <dc:title><rdf:Alt><rdf:li xml:lang=\"x-default\">%s</rdf:li></rdf:Alt></dc:title>\n", esc(in.Title))
	}
	if in.Author != "" {
		fmt.Fprintf(&b, "   <dc:creator><rdf:Seq><rdf:li>%s</rdf:li></rdf:Seq></dc:creator>\n", esc(in.Author))
	}
	if in.Subject != "" {
		fmt.Fprintf(&b, "   <dc:description><rdf:Alt><rdf:li xml:lang=\"x-default\">%s</rdf:li></rdf:Alt></dc:description>\n", esc(in.Subject))
	}
	b.WriteString("  </rdf:Description>\n")
	if in.Producer != "" || in.Keywords != "" {
		b.WriteString("  <rdf:Description rdf:about=\"\" xmlns:pdf=\"http://ns.adobe.com/pdf/1.3/\">\n")
		if in.Producer != "" {
			fmt.Fprintf(&b, "   <pdf:Producer>%s</pdf:Producer>\n", esc(in.Producer))
		}
		if in.Keywords != "" {
			fmt.Fprintf(&b, "   <pdf:Keywords>%s</pdf:Keywords>\n", esc(in.Keywords))
		}
		b.WriteString("  </rdf:Description>\n")
	}
	b.WriteString("  <rdf:Description rdf:about=\"\" xmlns:xmp=\"http://ns.adobe.com/xap/1.0/\">\n")
	if in.Creator != "" {
		fmt.Fprintf(&b, "   <xmp:CreatorTool>%s</xmp:CreatorTool>\n", esc(in.Creator))
	}
	if !in.Created.IsZero() {
		fmt.Fprintf(&b, "   <xmp:CreateDate>%s</xmp:CreateDate>\n", in.Created.Format(time.RFC3339))
	}
	if !in.Modified.IsZero() {
		fmt.Fprintf(&b, "   <xmp:ModifyDate>%s</xmp:ModifyDate>\n", in.Modified.Format(time.RFC3339))
	}
	b.WriteString("  </rdf:Description>\n")
	b.WriteString(" </rdf:RDF>\n</x:xmpmeta>\n")
	b.WriteString(strings.Repeat(" ", 2048)) // padding, so tools can edit in place
	b.WriteString("\n<?xpacket end=\"w\"?>\n")
	return []byte(b.String())
}

// PDFDate formats t as a PDF date string (ISO 32000-1 §7.9.4).
func PDFDate(t time.Time) string {
	_, off := t.Zone()
	sign := '+'
	if off < 0 {
		sign = '-'
		off = -off
	}
	return fmt.Sprintf("D:%s%c%02d'%02d'", t.Format("20060102150405"), sign, off/3600, off%3600/60)
}
