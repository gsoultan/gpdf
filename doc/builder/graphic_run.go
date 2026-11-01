// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package builder

import (
	"github.com/gsoultan/gpdf/content"
	"github.com/gsoultan/gpdf/model"
)

// GraphicRun holds pre-built content stream operators for one vector drawing operation.
type GraphicRun struct {
	Ops        []content.Op
	ExtGStates map[model.Name]model.Dict
	// Seq is the run's position among all runs on its page (see Page.AddGraphic).
	Seq int
}
