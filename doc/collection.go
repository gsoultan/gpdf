// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc

import (
	"fmt"
	"sort"

	"github.com/gsoultan/gpdf/model"
)

// CollectionView is how a viewer presents a portable collection.
type CollectionView string

const (
	CollectionDetails CollectionView = "D" // table of the schema fields (default)
	CollectionTile    CollectionView = "T" // tiles with thumbnails
	CollectionHidden  CollectionView = "H"
)

// CollectionFieldType is the data type of a collection schema field.
type CollectionFieldType string

const (
	CollectionText         CollectionFieldType = "S"
	CollectionDate         CollectionFieldType = "D"
	CollectionNumber       CollectionFieldType = "N"
	CollectionFileName     CollectionFieldType = "F"
	CollectionDescription  CollectionFieldType = "Desc"
	CollectionModDate      CollectionFieldType = "ModDate"
	CollectionCreationDate CollectionFieldType = "CreationDate"
	CollectionSize         CollectionFieldType = "Size"
)

// CollectionField is one column of a portable collection's schema.
type CollectionField struct {
	// Key is the field's name in each attachment's CollectionItem map.
	Key string
	// Name is the column title shown to the user.
	Name string
	Type CollectionFieldType
	// Visible hides the column when false. Default visible.
	Hidden bool
}

// CollectionOptions turns the document into a portable collection (ISO 32000-2
// §12.3.5, "PDF portfolio"): the attachments are presented with the schema's
// columns, and each Attachment's CollectionItem supplies the values.
type CollectionOptions struct {
	View   CollectionView
	Fields []CollectionField
	// SortKey is the schema field to sort by; SortDescending reverses the order.
	SortKey        string
	SortDescending bool
	// Initial is the name of the attachment opened first (/D).
	Initial string
}

// Collection makes the document a portable collection of its attachments.
func (b *DocumentBuilder) Collection(opts CollectionOptions) *DocumentBuilder {
	for _, f := range opts.Fields {
		if f.Key == "" {
			b.setErr(fmt.Errorf("collection field needs a key"))
			return b
		}
	}
	b.collection = &opts
	return b
}

func (b *DocumentBuilder) buildCollection(objs map[int]model.Object, catalogDict model.Dict, nextNum int) int {
	if b.collection == nil {
		return nextNum
	}
	o := b.collection
	schema := model.Dict{model.Name("Type"): model.Name("CollectionSchema")}
	for i, f := range o.Fields {
		t := f.Type
		if t == "" {
			t = CollectionText
		}
		fd := model.Dict{
			model.Name("Type"):    model.Name("CollectionField"),
			model.Name("Subtype"): model.Name(t),
			model.Name("N"):       model.String(f.Name),
			model.Name("O"):       model.Integer(i + 1),
		}
		if f.Hidden {
			fd[model.Name("V")] = model.Boolean(false)
		}
		schema[model.Name(f.Key)] = fd
	}
	col := model.Dict{
		model.Name("Type"):   model.Name("Collection"),
		model.Name("Schema"): schema,
	}
	if o.View != "" {
		col[model.Name("View")] = model.Name(o.View)
	}
	if o.SortKey != "" {
		col[model.Name("Sort")] = model.Dict{model.Name("S"): model.Name(o.SortKey), model.Name("A"): model.Boolean(!o.SortDescending)}
	}
	if o.Initial != "" {
		col[model.Name("D")] = model.String(o.Initial)
	}
	num := nextNum
	nextNum++
	objs[num] = col
	catalogDict[model.Name("Collection")] = model.Ref{ObjectNumber: num}
	return nextNum
}

// collectionItem builds the /CI dictionary of a file specification.
func collectionItem(values map[string]any, pdf20 bool) model.Dict {
	if len(values) == 0 {
		return nil
	}
	ci := model.Dict{model.Name("Type"): model.Name("CollectionItem")}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		switch v := values[k].(type) {
		case string:
			ci[model.Name(k)] = model.EncodeTextString(v, pdf20)
		case int:
			ci[model.Name(k)] = model.Integer(v)
		case int64:
			ci[model.Name(k)] = model.Integer(v)
		case float64:
			ci[model.Name(k)] = model.Real(v)
		case model.Object:
			ci[model.Name(k)] = v
		default:
			ci[model.Name(k)] = model.EncodeTextString(fmt.Sprint(v), pdf20)
		}
	}
	return ci
}
