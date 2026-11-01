package doc

import "github.com/gsoultan/gpdf/model"

// ObjectEditor changes a document's objects in place. On a document opened from
// a file the changes are tracked, so SaveIncremental appends exactly them.
type ObjectEditor interface {
	// SetObject replaces (or adds) the indirect object ref with obj.
	SetObject(ref model.Ref, obj model.Object)
	// NewObject adds obj under a fresh object number and returns its reference.
	NewObject(obj model.Object) model.Ref
	// SetInfo sets one entry of the document information dictionary, such as
	// "Title" or "Author", creating the dictionary when the document has none.
	SetInfo(key string, value model.Object) error
}
