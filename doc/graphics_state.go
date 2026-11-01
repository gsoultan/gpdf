// Copyright (c) 2026 Gembit Soultan Shirazi <gembit.soultan@gmail.com>

package doc

import "github.com/gsoultan/gpdf/doc/style"

type BlendMode = style.BlendMode
type GraphicsState = style.GraphicsState

const (
	BlendNormal     = style.BlendNormal
	BlendMultiply   = style.BlendMultiply
	BlendScreen     = style.BlendScreen
	BlendOverlay    = style.BlendOverlay
	BlendDarken     = style.BlendDarken
	BlendLighten    = style.BlendLighten
	BlendColorDodge = style.BlendColorDodge
	BlendColorBurn  = style.BlendColorBurn
	BlendHardLight  = style.BlendHardLight
	BlendSoftLight  = style.BlendSoftLight
	BlendDifference = style.BlendDifference
	BlendExclusion  = style.BlendExclusion
)

// BlackPointCompensation is the ExtGState /UseBlackPtComp value (PDF 2.0).
type BlackPointCompensation = style.BlackPointCompensation

const (
	BlackPointDefault = style.BlackPointDefault
	BlackPointOn      = style.BlackPointOn
	BlackPointOff     = style.BlackPointOff
)
