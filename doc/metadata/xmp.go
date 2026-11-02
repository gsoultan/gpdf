// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package metadata

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/gsoultan/gpdf/model"
)

// XMPInput is what BuildXMP writes into an XMP packet.
type XMPInput struct {
	Title, Author, Subject, Keywords, Creator, Producer string
	Created, Modified                                   time.Time
	// PDFAPart and PDFAConformance fill the pdfaid schema ("2", "B"). Empty leaves it out.
	// PDFARev is the pdfaid:rev year for PDF/A-4 ("2020"), which has no conformance level.
	PDFAPart, PDFAConformance, PDFARev string
	// PDFUAPart fills the pdfuaid schema ("1" or "2"); PDFUARev is the pdfuaid:rev
	// year for PDF/UA-2 ("2024").
	PDFUAPart, PDFUARev string
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
		fmt.Fprintf(&b, "  <rdf:Description rdf:about=\"\" xmlns:pdfaid=\"http://www.aiim.org/pdfa/ns/id/\">\n   <pdfaid:part>%s</pdfaid:part>\n", esc(in.PDFAPart))
		if in.PDFARev != "" {
			fmt.Fprintf(&b, "   <pdfaid:rev>%s</pdfaid:rev>\n", esc(in.PDFARev))
		}
		if in.PDFAConformance != "" {
			fmt.Fprintf(&b, "   <pdfaid:conformance>%s</pdfaid:conformance>\n", esc(in.PDFAConformance))
		}
		b.WriteString("  </rdf:Description>\n")
	}
	if in.PDFUAPart != "" {
		fmt.Fprintf(&b, "  <rdf:Description rdf:about=\"\" xmlns:pdfuaid=\"http://www.aiim.org/pdfua/ns/id/\">\n   <pdfuaid:part>%s</pdfuaid:part>\n", esc(in.PDFUAPart))
		if in.PDFUARev != "" {
			fmt.Fprintf(&b, "   <pdfuaid:rev>%s</pdfuaid:rev>\n", esc(in.PDFUARev))
		}
		b.WriteString("  </rdf:Description>\n")
		if in.PDFAPart != "" {
			writePDFUAExtensionSchema(&b, in.PDFUARev != "")
		}
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

// writePDFUAExtensionSchema describes the pdfuaid schema in the PDF/A extension
// schema container (ISO 19005-2 §6.6.2.3.2): PDF/A predefines no PDF/UA
// identification schema, so a file claiming both must declare it.
func writePDFUAExtensionSchema(b *strings.Builder, withRev bool) {
	b.WriteString(`  <rdf:Description rdf:about="" xmlns:pdfaExtension="http://www.aiim.org/pdfa/ns/extension/" xmlns:pdfaSchema="http://www.aiim.org/pdfa/ns/schema#" xmlns:pdfaProperty="http://www.aiim.org/pdfa/ns/property#">
   <pdfaExtension:schemas>
    <rdf:Bag>
     <rdf:li rdf:parseType="Resource">
      <pdfaSchema:schema>PDF/UA identification schema</pdfaSchema:schema>
      <pdfaSchema:namespaceURI>http://www.aiim.org/pdfua/ns/id/</pdfaSchema:namespaceURI>
      <pdfaSchema:prefix>pdfuaid</pdfaSchema:prefix>
      <pdfaSchema:property>
       <rdf:Seq>
        <rdf:li rdf:parseType="Resource">
         <pdfaProperty:name>part</pdfaProperty:name>
         <pdfaProperty:valueType>Integer</pdfaProperty:valueType>
         <pdfaProperty:category>internal</pdfaProperty:category>
         <pdfaProperty:description>PDF/UA version identifier</pdfaProperty:description>
        </rdf:li>
`)
	if withRev {
		b.WriteString(`        <rdf:li rdf:parseType="Resource">
         <pdfaProperty:name>rev</pdfaProperty:name>
         <pdfaProperty:valueType>Integer</pdfaProperty:valueType>
         <pdfaProperty:category>internal</pdfaProperty:category>
         <pdfaProperty:description>PDF/UA revision year</pdfaProperty:description>
        </rdf:li>
`)
	}
	b.WriteString(`       </rdf:Seq>
      </pdfaSchema:property>
     </rdf:li>
    </rdf:Bag>
   </pdfaExtension:schemas>
  </rdf:Description>
`)
}

// PDFDate formats t as a PDF 1.x date string. See model.FormatDate for the PDF 2.0 form.
func PDFDate(t time.Time) string { return model.FormatDate(t, false) }
