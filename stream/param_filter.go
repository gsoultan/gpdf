package stream

import (
	"io"

	"github.com/gsoultan/gpdf/model"
)

// ParamFilter is implemented by filters whose decoding depends on /DecodeParms
// (for example LZWDecode /EarlyChange). The pipeline calls DecodeWithParams when
// the stream has parameters for the filter; predictors are applied by the pipeline itself.
type ParamFilter interface {
	DecodeWithParams(dst io.Writer, src io.Reader, name string, params model.Dict) error
}
