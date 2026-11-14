// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"io"
	"math"
	"slices"

	"github.com/gsoultan/gpdf/model"
	"github.com/gsoultan/gpdf/reader"
	"github.com/gsoultan/gpdf/writer"
)

// PageSource is one input of MergeTo: a document and the pages to take from it.
type PageSource struct {
	Doc Document
	// Pages are 0-based page indices, in output order; a page may repeat. Nil
	// takes every page. See ParsePageRanges for "1-3,5,8-" style selections.
	Pages []int
}

// CopyOptions tune MergeTo and SplitTo.
type CopyOptions struct {
	// UserPassword and OwnerPassword, when either is set, encrypt the output
	// with AES-256 (ISO 32000-2 revision 6). An empty owner password means the
	// user password is also the owner password.
	UserPassword  string
	OwnerPassword string
	// NoObjectStreams writes a classic cross-reference table and leaves every
	// object at top level (larger files, readable by PDF 1.4 tools).
	NoObjectStreams bool
}

// ErrNoPages is returned by MergeTo and SplitTo when the selection is empty.
var ErrNoPages = errors.New("no pages selected")

// MergeTo writes the selected pages of every source, in order, to w as one PDF and
// returns the number of pages written.
//
// Unlike Merge it never builds the merged document in memory: each object is
// written as soon as it is copied, streams are copied in their stored encoding
// (no decode and re-encode, so images and fonts keep their compression and the
// copy is fast), and identical streams (a logo or font file repeated across
// inputs) are written once. Encrypted inputs must be opened with their password;
// the output is encrypted only when opts asks for it. Only the selected pages and
// what they use are copied: outlines, forms (/AcroForm), structure trees and
// links to pages that are not copied are dropped, and /Info comes from the first
// source.
func MergeTo(w io.Writer, opts CopyOptions, sources ...PageSource) (int, error) {
	if len(sources) == 0 {
		return 0, fmt.Errorf("merge: no documents provided")
	}
	wopts := writer.WriterOptions{}
	if opts.NoObjectStreams {
		wopts.ObjectStreams = new(false)
	}
	sw := writer.NewStreamWriter(w, wopts)
	if opts.NoObjectStreams {
		if err := sw.SetVersion("1.4"); err != nil {
			return 0, err
		}
	}
	if opts.UserPassword != "" || opts.OwnerPassword != "" {
		if err := sw.SetAES256Password(opts.UserPassword, opts.OwnerPassword); err != nil {
			return 0, fmt.Errorf("merge: %w", err)
		}
	}
	m := &streamMerger{sw: sw, dedup: make(map[[sha256.Size]byte]int), pagesRef: sw.Alloc()}

	var kids model.Array
	var info model.Object
	for i, s := range sources {
		if s.Doc == nil {
			return 0, fmt.Errorf("merge: document %d is nil", i)
		}
		c := m.forSource(s.Doc)
		if i == 0 {
			if d, err := s.Doc.Info(); err == nil && len(d) > 0 {
				if cloned, err := c.clone(d); err == nil {
					ref := sw.Alloc()
					if err := sw.WriteObject(ref, cloned); err != nil {
						return 0, fmt.Errorf("merge: %w", err)
					}
					info = ref
				}
			}
		}
		refs, err := c.copyPages(s.Pages)
		if err != nil {
			return 0, fmt.Errorf("merge: document %d: %w", i, err)
		}
		kids = append(kids, refs...)
	}
	if len(kids) == 0 {
		return 0, fmt.Errorf("merge: %w", ErrNoPages)
	}
	if err := sw.WriteObject(m.pagesRef, model.Dict{
		model.Name("Type"):  model.Name("Pages"),
		model.Name("Count"): model.Integer(int64(len(kids))),
		model.Name("Kids"):  kids,
	}); err != nil {
		return 0, fmt.Errorf("merge: %w", err)
	}
	catalog := sw.Alloc()
	if err := sw.WriteObject(catalog, model.Dict{
		model.Name("Type"):  model.Name("Catalog"),
		model.Name("Pages"): m.pagesRef,
	}); err != nil {
		return 0, fmt.Errorf("merge: %w", err)
	}
	trailer := model.Dict{model.Name("Root"): catalog}
	if info != nil {
		trailer[model.Name("Info")] = info
	}
	if err := sw.Close(trailer); err != nil {
		return 0, fmt.Errorf("merge: %w", err)
	}
	return len(kids), nil
}

// SplitTo writes one PDF per range of src (0-based page indices, as for Split),
// each to the writer open returns for that part (numbered from 0). The caller
// owns the writers: SplitTo does not close them. Every part is written with
// MergeTo, so parts are produced one at a time and never held in memory.
func SplitTo(src Document, ranges [][]int, opts CopyOptions, open func(part int) (io.Writer, error)) error {
	if len(ranges) == 0 {
		return fmt.Errorf("split: no ranges provided")
	}
	for i, r := range ranges {
		if len(r) == 0 {
			return fmt.Errorf("split: range %d: %w", i, ErrNoPages)
		}
		w, err := open(i)
		if err != nil {
			return fmt.Errorf("split: part %d: %w", i, err)
		}
		if _, err := MergeTo(w, opts, PageSource{Doc: src, Pages: r}); err != nil {
			return fmt.Errorf("split: part %d: %w", i, err)
		}
	}
	return nil
}

// streamMerger holds what is shared by every source of one MergeTo call.
type streamMerger struct {
	sw       *writer.StreamWriter
	pagesRef model.Ref
	// dedup maps a stream's content key to the object number it was written as.
	dedup map[[sha256.Size]byte]int
}

func (m *streamMerger) forSource(src Document) *sourceCopier {
	c := &sourceCopier{m: m, src: src, remapped: make(map[int]int)}
	c.encoded, _ = src.(reader.EncodedResolver)
	return c
}

// sourceCopier copies objects of one source document into the stream writer.
type sourceCopier struct {
	m        *streamMerger
	src      Document
	encoded  reader.EncodedResolver
	remapped map[int]int // source object number -> output object number
}

func (c *sourceCopier) resolve(ref model.Ref) (model.Object, error) {
	if c.encoded != nil {
		return c.encoded.ResolveEncoded(ref)
	}
	return c.src.Resolve(ref)
}

// copyPages writes the pages at indices (all when nil) and returns their output refs.
func (c *sourceCopier) copyPages(indices []int) (model.Array, error) {
	all, err := c.src.Pages()
	if err != nil {
		return nil, fmt.Errorf("read pages: %w", err)
	}
	if indices == nil {
		indices = make([]int, len(all))
		for i := range indices {
			indices[i] = i
		}
	}
	selected := make([]model.Page, len(indices))
	for i, idx := range indices {
		if idx < 0 || idx >= len(all) {
			return nil, fmt.Errorf("page index %d out of range (document has %d pages)", idx, len(all))
		}
		selected[i] = all[idx]
	}
	// Selected pages get their numbers first, so links and annotation /P between
	// copied pages are remapped; any other page reference becomes null.
	refs := make(model.Array, len(selected))
	placed := make(map[int]bool, len(selected))
	for i, p := range selected {
		num := p.Ref.ObjectNumber
		if num > 0 && !placed[num] {
			placed[num] = true
			ref := c.m.sw.Alloc()
			c.remapped[num] = ref.ObjectNumber
			refs[i] = ref
			continue
		}
		refs[i] = c.m.sw.Alloc() // a repeated page is copied again under its own number
	}
	for i, p := range selected {
		pageDict := make(model.Dict, len(p.Dict)+1)
		for k, v := range p.Dict {
			pageDict[k] = v
		}
		for _, k := range pageCopyDroppedKeys {
			delete(pageDict, k)
		}
		cloned, err := c.clone(pageDict)
		if err != nil {
			return nil, fmt.Errorf("copy page %d: %w", indices[i]+1, err)
		}
		d := cloned.(model.Dict)
		d[model.Name("Parent")] = c.m.pagesRef
		if err := c.m.sw.WriteObject(refs[i].(model.Ref), d); err != nil {
			return nil, err
		}
	}
	return refs, nil
}

// clone deep-copies obj, writing every indirect object it reaches.
func (c *sourceCopier) clone(obj model.Object) (model.Object, error) {
	switch v := obj.(type) {
	case nil:
		return nil, nil
	case model.Ref:
		if num, ok := c.remapped[v.ObjectNumber]; ok {
			return model.Ref{ObjectNumber: num}, nil
		}
		resolved, err := c.resolve(v)
		if err != nil {
			return nil, err
		}
		if isPageTreeNode(resolved) {
			return model.Null{}, nil
		}
		var key [sha256.Size]byte
		s, isStream := resolved.(*model.Stream)
		dedupable := isStream && !hasRef(s.Dict)
		if dedupable {
			key = streamKey(s)
			if num, ok := c.m.dedup[key]; ok {
				c.remapped[v.ObjectNumber] = num
				return model.Ref{ObjectNumber: num}, nil
			}
		}
		ref := c.m.sw.Alloc()
		c.remapped[v.ObjectNumber] = ref.ObjectNumber
		cloned, err := c.clone(resolved)
		if err != nil {
			return nil, err
		}
		if err := c.m.sw.WriteObject(ref, cloned); err != nil {
			return nil, err
		}
		if dedupable {
			c.m.dedup[key] = ref.ObjectNumber
		}
		return ref, nil
	case model.Dict:
		out := make(model.Dict, len(v))
		for k, val := range v {
			cv, err := c.clone(val)
			if err != nil {
				return nil, err
			}
			out[k] = cv
		}
		return out, nil
	case model.Array:
		out := make(model.Array, len(v))
		for i, val := range v {
			cv, err := c.clone(val)
			if err != nil {
				return nil, err
			}
			out[i] = cv
		}
		return out, nil
	case *model.Stream:
		d, err := c.clone(v.Dict)
		if err != nil {
			return nil, err
		}
		dict, _ := d.(model.Dict)
		// The content is written straight away and never changed, so it is shared.
		return &model.Stream{Dict: dict, Content: v.Content, Encoded: v.Encoded}, nil
	default:
		return v, nil
	}
}

// hasRef reports whether obj holds an indirect reference anywhere.
func hasRef(obj model.Object) bool {
	switch v := obj.(type) {
	case model.Ref:
		return true
	case model.Dict:
		for _, val := range v {
			if hasRef(val) {
				return true
			}
		}
	case model.Array:
		for _, val := range v {
			if hasRef(val) {
				return true
			}
		}
	}
	return false
}

// streamKey hashes a stream's dictionary (which holds no references) and content.
func streamKey(s *model.Stream) [sha256.Size]byte {
	h := sha256.New()
	if s.Encoded {
		h.Write([]byte{1})
	} else {
		h.Write([]byte{0})
	}
	hashObject(h, s.Dict)
	var n [8]byte
	binary.BigEndian.PutUint64(n[:], uint64(len(s.Content)))
	h.Write(n[:])
	h.Write(s.Content)
	var key [sha256.Size]byte
	h.Sum(key[:0])
	return key
}

// hashObject writes an unambiguous encoding of a reference-free object to h.
func hashObject(h hash.Hash, obj model.Object) {
	var n [8]byte
	writeBytes := func(tag byte, b []byte) {
		binary.BigEndian.PutUint64(n[:], uint64(len(b)))
		h.Write([]byte{tag})
		h.Write(n[:])
		h.Write(b)
	}
	switch v := obj.(type) {
	case model.Dict:
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, string(k))
		}
		slices.Sort(keys)
		binary.BigEndian.PutUint64(n[:], uint64(len(keys)))
		h.Write([]byte{'d'})
		h.Write(n[:])
		for _, k := range keys {
			writeBytes('k', []byte(k))
			hashObject(h, v[model.Name(k)])
		}
	case model.Array:
		binary.BigEndian.PutUint64(n[:], uint64(len(v)))
		h.Write([]byte{'a'})
		h.Write(n[:])
		for _, val := range v {
			hashObject(h, val)
		}
	case model.Name:
		writeBytes('n', []byte(v))
	case model.String:
		writeBytes('s', []byte(v))
	case model.HexString:
		writeBytes('h', []byte(v))
	case model.Integer:
		binary.BigEndian.PutUint64(n[:], uint64(v))
		h.Write([]byte{'i'})
		h.Write(n[:])
	case model.Real:
		binary.BigEndian.PutUint64(n[:], math.Float64bits(float64(v)))
		h.Write([]byte{'r'})
		h.Write(n[:])
	case model.Boolean:
		if v {
			h.Write([]byte{'T'})
		} else {
			h.Write([]byte{'F'})
		}
	default:
		writeBytes('?', []byte(fmt.Sprintf("%T:%v", v, v)))
	}
}
