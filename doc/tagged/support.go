// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package tagged

import "github.com/gsoultan/gpdf/model"

// Block describes one block-level tagged element (e.g. P, H1..H6) in the structure tree.
// It refers to a single marked-content sequence on a page via MCID.
type Block struct {
	PageIndex int
	MCID      int
	Role      model.Name

	Lang string
	Alt  string
}

// List represents a logical list (/L) on a page.
// Items correspond to /LI elements, each pointing to one marked-content sequence via MCID.
type List struct {
	PageIndex int
	Ordered   bool
	Items     []ListItem
}

// ListItem represents one item in a tagged list. MCID is the item text
// (wrapped in an LBody element); LabelMCID, when HasLabel is set, is the
// bullet or number (an Lbl element).
type ListItem struct {
	MCID      int
	LabelMCID int
	HasLabel  bool
}

// Figure describes a tagged image (Figure) with alternative text for accessibility.
type Figure struct {
	PageIndex int
	MCID      int
	Alt       string
}

// Section groups block, figure, table, and list indices for a Sect structure element.
// Kids keeps the section's elements in the order they were recorded.
type Section struct {
	BlockIndices  []int
	FigureIndices []int
	TableIndices  []int
	ListIndices   []int
	Kids          []ElemRef
}

// ElemKind identifies which Support slice an ElemRef indexes.
type ElemKind byte

const (
	KindBlock   ElemKind = 'b'
	KindFigure  ElemKind = 'f'
	KindTable   ElemKind = 't'
	KindList    ElemKind = 'l'
	KindSection ElemKind = 's'
)

// ElemRef points at one recorded element, so that the structure tree can keep
// the order in which content was added (the reading order).
type ElemRef struct {
	Kind  ElemKind
	Index int
}

// StructCell describes one logical table cell in the structure tree.
type StructCell struct {
	PageIndex int
	MCIDs     []int
	Role      model.Name

	Scope string
	Alt   string
	Lang  string
}

// StructRow describes one logical table row (TR) in the structure tree.
type StructRow struct {
	Cells []StructCell
}

// StructTable describes one logical table (Table) in the structure tree.
type StructTable struct {
	PageIndex  int
	Rows       []StructRow
	RowHeights []float64
}

// Namespace URIs for structure elements (ISO 32000-2 §14.8.6).
const (
	// NamespacePDF2 is the standard structure namespace for PDF 2.0.
	NamespacePDF2 = "http://iso.org/pdf2/ssn"
	// NamespacePDF17 is the standard structure namespace for PDF 1.7, which
	// holds the types PDF 2.0 dropped (Quote, Code, ...).
	NamespacePDF17 = "http://iso.org/pdf/ssn"
)

// Support owns tagged-structure-related state for a document builder.
type Support struct {
	Tables         []StructTable
	Blocks         []Block
	Lists          []List
	Figures        []Figure
	Sections       []Section
	CurrentSection int
	// Order lists the top-level elements (those outside any section, and the
	// sections themselves) in the order they were recorded.
	Order []ElemRef
	// UseNamespaces writes the structure in the PDF 2.0 namespace
	// (StructTreeRoot /Namespaces, /NS on every element), as PDF/UA-2 needs.
	UseNamespaces bool
	// TagAnnotations adds Link structure elements for link annotations and
	// /StructParent entries, as PDF/UA needs.
	TagAnnotations bool
}

// BeginSection opens a Sect element; elements recorded until EndSection
// become its children.
func (s *Support) BeginSection() {
	s.Sections = append(s.Sections, Section{})
	s.CurrentSection = len(s.Sections) - 1
	s.Order = append(s.Order, ElemRef{Kind: KindSection, Index: s.CurrentSection})
}

// EndSection closes the current section.
func (s *Support) EndSection() { s.CurrentSection = -1 }

func (s *Support) record(ref ElemRef) {
	if s.CurrentSection < 0 || s.CurrentSection >= len(s.Sections) {
		s.Order = append(s.Order, ref)
		return
	}
	sec := &s.Sections[s.CurrentSection]
	sec.Kids = append(sec.Kids, ref)
	switch ref.Kind {
	case KindBlock:
		sec.BlockIndices = append(sec.BlockIndices, ref.Index)
	case KindFigure:
		sec.FigureIndices = append(sec.FigureIndices, ref.Index)
	case KindTable:
		sec.TableIndices = append(sec.TableIndices, ref.Index)
	case KindList:
		sec.ListIndices = append(sec.ListIndices, ref.Index)
	}
}

// RecordSectionBlock records a block index in the reading order (inside the current section, if any).
func (s *Support) RecordSectionBlock(blockIndex int) { s.record(ElemRef{KindBlock, blockIndex}) }

// RecordSectionFigure records a figure index in the reading order.
func (s *Support) RecordSectionFigure(figureIndex int) { s.record(ElemRef{KindFigure, figureIndex}) }

// RecordSectionTable records a table index in the reading order.
func (s *Support) RecordSectionTable(tableIndex int) { s.record(ElemRef{KindTable, tableIndex}) }

// RecordSectionList records a list index in the reading order.
func (s *Support) RecordSectionList(listIndex int) { s.record(ElemRef{KindList, listIndex}) }
