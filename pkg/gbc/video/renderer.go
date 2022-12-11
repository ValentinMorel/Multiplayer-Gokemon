package video

import (
	"github.com/pokemium/worldwide/pkg/util"
)

// GBVideoRenderer & GBVideoSoftwareRenderer
type Renderer struct {
	// GBVideoRenderer
	g                                 *Video
	disableBG, disableOBJ, disableWIN bool
	highlightBG                       bool
	highlightOBJ                      [MAX_OBJ]bool
	highlightWIN                      bool
	highlightColor                    uint16
	highlightAmount                   byte

	// GBVideoSoftwareRenderer
	// Renderer.row[i] -> Renderer.Lookup[i] -> Renderer.palette[i] -> outputBuffer
	OutputBuffer       [256 * 256]Color
	outputBufferStride int

	// each element represents palette color
	// x = 0 or 1 or 2 or 3
	// x -> BGP or BGP0
	// 1*4 + x -> BGP1
	// 2*4 + x -> BGP2
	// ...
	// 7*4 + x -> BGP7
	// 8*4 + x -> OBP0
	// 9*4 + x -> OBP1
	row [HORIZONTAL_PIXELS + 8]uint16

	Palette [64 * 3]Color

	// palette color(Renderer.row element) -> Renderer.palette index
	Lookup [64 * 3]byte

	wy, wx, currentWy, currentWx byte
	lastY, lastX                 int
	hasWindow                    bool

	lastHighlightAmount byte
	Model               util.GBModel
	obj                 [MAX_LINE_OBJ]Sprite
	objMax              int

	sgbBorders    bool
	sgbRenderMode int
	sgbAttributes []byte
}

func NewRenderer(g *Video) *Renderer {
	r := &Renderer{
		g:                  g,
		highlightColor:     0x7fff,
		outputBufferStride: 160,
		lastY:              VERTICAL_PIXELS,
	}

	for i := byte(0); i < 192; i++ {
		r.Lookup[i] = i
	}

	return r
}

// GBVideoSoftwareRendererUpdateWindow
func (r *Renderer) updateWindow(before, after bool, oldWy byte) {
	if r.lastY >= VERTICAL_PIXELS || !(after || before) {
		return
	}
	if !r.hasWindow && r.lastX == HORIZONTAL_PIXELS {
		return
	}
	if r.lastY >= int(oldWy) {
		if !after {
			r.currentWy = byte(int(r.currentWy) - r.lastY)
			r.hasWindow = true
		} else if !before {
			if !r.hasWindow {
				r.currentWy = byte(r.lastY - int(r.wy))
				if r.lastY >= int(r.wy) && r.lastX > int(r.wx) {
					r.currentWy++
				}
			} else {
				r.currentWy += byte(r.lastY)
			}
		} else if r.wy != oldWy {
			r.currentWy += oldWy - r.wy
			r.hasWindow = true
		}
	}
}

// writeVideoRegister / GBVideoSoftwareRendererWriteVideoRegister
// this is called from GBIOWrite/GBVideoWritePalette/etc...
func (r *Renderer) WriteVideoRegister(offset byte, value byte) byte {
	wasWindow := r.inWindow()
	wy := r.wy

	switch offset {
	case GB_REG_LCDC:
		r.g.LCDC = value
		r.updateWindow(wasWindow, r.inWindow(), wy)
	case GB_REG_WY:
		r.wy = value
		r.updateWindow(wasWindow, r.inWindow(), wy)
	case GB_REG_WX:
		r.wx = value
		r.updateWindow(wasWindow, r.inWindow(), wy)
	case GB_REG_BGP:
		r.Lookup[0] = value & 3
		r.Lookup[1] = (value >> 2) & 3
		r.Lookup[2] = (value >> 4) & 3
		r.Lookup[3] = (value >> 6) & 3
		r.Lookup[PAL_HIGHLIGHT_BG+0] = PAL_HIGHLIGHT + (value & 3)
		r.Lookup[PAL_HIGHLIGHT_BG+1] = PAL_HIGHLIGHT + ((value >> 2) & 3)
		r.Lookup[PAL_HIGHLIGHT_BG+2] = PAL_HIGHLIGHT + ((value >> 4) & 3)
		r.Lookup[PAL_HIGHLIGHT_BG+3] = PAL_HIGHLIGHT + ((value >> 6) & 3)
	case GB_REG_OBP0:
		r.Lookup[PAL_OBJ+0] = value & 3
		r.Lookup[PAL_OBJ+1] = (value >> 2) & 3
		r.Lookup[PAL_OBJ+2] = (value >> 4) & 3
		r.Lookup[PAL_OBJ+3] = (value >> 6) & 3
		r.Lookup[PAL_HIGHLIGHT_OBJ+0] = PAL_HIGHLIGHT + (value & 3)
		r.Lookup[PAL_HIGHLIGHT_OBJ+1] = PAL_HIGHLIGHT + ((value >> 2) & 3)
		r.Lookup[PAL_HIGHLIGHT_OBJ+2] = PAL_HIGHLIGHT + ((value >> 4) & 3)
		r.Lookup[PAL_HIGHLIGHT_OBJ+3] = PAL_HIGHLIGHT + ((value >> 6) & 3)
	case GB_REG_OBP1:
		r.Lookup[PAL_OBJ+4] = value & 3
		r.Lookup[PAL_OBJ+5] = (value >> 2) & 3
		r.Lookup[PAL_OBJ+6] = (value >> 4) & 3
		r.Lookup[PAL_OBJ+7] = (value >> 6) & 3
		r.Lookup[PAL_HIGHLIGHT_OBJ+4] = PAL_HIGHLIGHT + (value & 3)
		r.Lookup[PAL_HIGHLIGHT_OBJ+5] = PAL_HIGHLIGHT + ((value >> 2) & 3)
		r.Lookup[PAL_HIGHLIGHT_OBJ+6] = PAL_HIGHLIGHT + ((value >> 4) & 3)
		r.Lookup[PAL_HIGHLIGHT_OBJ+7] = PAL_HIGHLIGHT + ((value >> 6) & 3)
	}

	return value
}

// writePalette / GBVideoSoftwareRendererWritePalette
// GBVideoWritePalette calls this
func (r *Renderer) writePalette(index int, value Color) {
	color := value
	if r.Model&util.GB_MODEL_SGB != 0 {
		if index < 0x10 && index != 0 && (index&3) == 0 {
			color = r.Palette[0]
		} else if index >= PAL_SGB_BORDER && (index&0xf) == 0 {
			color = r.Palette[0]
		} else if index > PAL_HIGHLIGHT && index < PAL_HIGHLIGHT_OBJ && (index&3) == 0 {
			color = r.Palette[PAL_HIGHLIGHT_BG]
		}
	}

	r.Palette[index] = color
	if index < PAL_SGB_BORDER && (index < PAL_OBJ || (index&3) != 0) {
		r.Palette[index+PAL_HIGHLIGHT] = color
	}
}

// drawRange / GBVideoSoftwareRendererDrawRange
// by row
func (r *Renderer) drawRange(startX, endX, y int) {
	r.lastY, r.lastX = y, endX
	if startX >= endX {
		return
	}

	mapIdx := GB_BASE_MAP // 0x9800
	if util.Bit(r.g.LCDC, TileMap) {
		mapIdx += GB_SIZE_MAP // 0x9c00
	}

	if r.disableBG {
		for x := startX; x < endX; x++ {
			r.row[x] = 0
		}
	}

	if util.Bit(r.g.LCDC, BgEnable) || r.Model >= util.GB_MODEL_CGB {
		wy, wx := int(r.wy)+int(r.currentWy), int(r.wx)+int(r.currentWx)-7
		if util.Bit(r.g.LCDC, Window) && wy == y && wx <= endX {
			r.hasWindow = true
		}
		scx, scy := int(r.g.io[GB_REG_SCX]), int(r.g.io[GB_REG_SCY])
		if util.Bit(r.g.LCDC, Window) && r.hasWindow && wx <= endX && !r.disableWIN {
			if wx > 0 && !r.disableBG {
				// bg -> window
				r.drawBackground(mapIdx, startX, wx, scx, scy+y, r.highlightBG)

				// fallthrough and draw window
			}

			mapIdx = GB_BASE_MAP
			if util.Bit(r.g.LCDC, WindowTileMap) {
				mapIdx += GB_SIZE_MAP // 0x9c00
			}
			r.drawBackground(mapIdx, wx, endX, -wx, y-wy, r.highlightWIN)
		} else if !r.disableBG {
			r.drawBackground(mapIdx, startX, endX, scx, scy+y, r.highlightBG)
		}
	} else if !r.disableBG {
		for x := startX; x < endX; x++ {
			r.row[x] = 0
		}
	}

	if startX == 0 {
		r.cleanOAM(y)
	}
	if util.Bit(r.g.LCDC, ObjEnable) && !r.disableOBJ {
		for i := 0; i < r.objMax; i++ {
			r.drawObj(r.obj[i], startX, endX, y)
		}
	}

	highlightAmount := (r.highlightAmount + 6) >> 4
	if r.lastHighlightAmount != highlightAmount {
		r.lastHighlightAmount = highlightAmount
		for i := 0; i < PAL_SGB_BORDER; i++ {
			if i >= PAL_OBJ && i&3 == 0 {
				continue
			}
			r.Palette[i+PAL_HIGHLIGHT] = r.Palette[i]
		}
	}

	sgbOffset := 0
	if (r.Model&util.GB_MODEL_SGB != 0) && r.sgbBorders {
		sgbOffset = r.outputBufferStride*40 + 48
	}

	row := r.OutputBuffer[r.outputBufferStride*y+sgbOffset:]
	x, p := startX, 0
	switch r.sgbRenderMode {
	case 0:
		if r.Model&util.GB_MODEL_SGB != 0 {
			p = int(r.sgbAttributes[(startX>>5)+5*(y>>3)])
			p >>= 6 - ((x / 4) & 0x6)
			p &= 3
			p <<= 2
		}
		for ; x < ((startX+7) & ^7) && x < endX; x++ {
			row[x] = r.Palette[p|int(r.Lookup[r.row[x]&OBJ_PRIO_MASK])]
		}

		for ; x+7 < (endX & ^7); x += 8 {
			if (r.Model & util.GB_MODEL_SGB) != 0 {
				p = int(r.sgbAttributes[(x>>5)+5*(y>>3)])
				p >>= 6 - ((x / 4) & 0x6)
				p &= 3
				p <<= 2
			}
			row[x+0] = r.Palette[p|int(r.Lookup[r.row[x]&OBJ_PRIO_MASK])]
			row[x+1] = r.Palette[p|int(r.Lookup[r.row[x+1]&OBJ_PRIO_MASK])]
			row[x+2] = r.Palette[p|int(r.Lookup[r.row[x+2]&OBJ_PRIO_MASK])]
			row[x+3] = r.Palette[p|int(r.Lookup[r.row[x+3]&OBJ_PRIO_MASK])]
			row[x+4] = r.Palette[p|int(r.Lookup[r.row[x+4]&OBJ_PRIO_MASK])]
			row[x+5] = r.Palette[p|int(r.Lookup[r.row[x+5]&OBJ_PRIO_MASK])]
			row[x+6] = r.Palette[p|int(r.Lookup[r.row[x+6]&OBJ_PRIO_MASK])]
			row[x+7] = r.Palette[p|int(r.Lookup[r.row[x+7]&OBJ_PRIO_MASK])]
		}
		if (r.Model & util.GB_MODEL_SGB) != 0 {
			p = int(r.sgbAttributes[(x>>5)+5*(y>>3)])
			p >>= 6 - ((x / 4) & 0x6)
			p &= 3
			p <<= 2
		}
		for ; x < endX; x++ {
			row[x] = r.Palette[p|int(r.Lookup[r.row[x]&OBJ_PRIO_MASK])]
		}
	case 2:
		for ; x < ((startX+7) & ^7) && x < endX; x++ {
			row[x] = 0
		}
		for ; x+7 < (endX & ^7); x += 8 {
			row[x] = 0
			row[x+1] = 0
			row[x+2] = 0
			row[x+3] = 0
			row[x+4] = 0
			row[x+5] = 0
			row[x+6] = 0
			row[x+7] = 0
		}
		for ; x < endX; x++ {
			row[x] = 0
		}
	case 3:
		for ; x < ((startX+7) & ^7) && x < endX; x++ {
			row[x] = r.Palette[0]
		}
		for ; x+7 < (endX & ^7); x += 8 {
			row[x] = r.Palette[0]
			row[x+1] = r.Palette[0]
			row[x+2] = r.Palette[0]
			row[x+3] = r.Palette[0]
			row[x+4] = r.Palette[0]
			row[x+5] = r.Palette[0]
			row[x+6] = r.Palette[0]
			row[x+7] = r.Palette[0]
		}
		for ; x < endX; x++ {
			row[x] = r.Palette[0]
		}
	}
}

// finishScanline / GBVideoSoftwareRendererFinishScanline
func (r *Renderer) finishScanline(y int) {
	r.lastX, r.currentWx = 0, 0
}

// finishFrame / GBVideoSoftwareRendererFinishFrame
func (r *Renderer) finishFrame() {
	if !util.Bit(r.g.LCDC, Enable) {
		r.clearScreen()
	}
	r.lastY, r.lastX = VERTICAL_PIXELS, 0
	r.currentWy, r.currentWx = 0, 0
	r.hasWindow = false
}

// GBVideoSoftwareRendererDrawBackground
// by row
func (r *Renderer) drawBackground(mapIdx, startX, endX, sx, sy int, highlight bool) {
	vramIdx := 0
	attrIdx := mapIdx + GB_SIZE_VRAM_BANK0 // for CGB
	if !util.Bit(r.g.LCDC, TileData) {
		vramIdx += 0x1000
	}

	topY := ((sy >> 3) & 0x1F) * 0x20
	bottomY := sy & 7
	if startX < 0 {
		startX = 0
	}

	x := 0
	if ((startX + sx) & 7) > 0 {
		startX2 := startX + 8 - ((startX + sx) & 7)
		for x := startX; x < startX2; x++ {
			localData := vramIdx
			localY := bottomY
			topX, bottomX := ((x+sx)>>3)&0x1F, 7-((x+sx)&7)
			bgTile := 0
			if util.Bit(r.g.LCDC, TileData) {
				// 0x8000-0x8800 [0, 255]
				bgTile = int(r.g.VRAM.Buffer[mapIdx+topX+topY])
			} else {
				// 0x8800-0x97ff [-128, 127]
				bgTile = int(int8(r.g.VRAM.Buffer[mapIdx+topX+topY]))
			}

			p := uint16(0)
			if highlight {
				p = PAL_HIGHLIGHT_BG
			}

			if r.Model >= util.GB_MODEL_CGB {
				attrs := r.g.VRAM.Buffer[attrIdx+topX+topY]
				p |= uint16(attrs&0x7) * 4
				if util.Bit(attrs, ObjAttrPriority) && util.Bit(r.g.LCDC, BgEnable) {
					p |= OBJ_PRIORITY
				}
				if util.Bit(attrs, ObjAttrBank) {
					localData += GB_SIZE_VRAM_BANK0
				}
				if util.Bit(attrs, ObjAttrYFlip) {
					localY = 7 - bottomY
				}
				if util.Bit(attrs, ObjAttrXFlip) {
					bottomX = 7 - bottomX
				}
			}
			tileDataLower := r.g.VRAM.Buffer[localData+(bgTile*8+localY)*2]
			tileDataUpper := r.g.VRAM.Buffer[localData+(bgTile*8+localY)*2+1]
			tileDataUpper >>= bottomX
			tileDataLower >>= bottomX
			r.row[x] = p | uint16((tileDataUpper&1)<<1) | uint16(tileDataLower&1)
		}
		startX = startX2
	}

	// by tile row
	for x = startX; x < endX; x += 8 {
		localData := vramIdx
		localY := bottomY
		topX := ((x + sx) >> 3) & 0x1F

		bgTile := 0
		if util.Bit(r.g.LCDC, TileData) {
			// 0x8000-0x8800 [0, 255]
			bgTile = int(r.g.VRAM.Buffer[mapIdx+topX+topY])
		} else {
			// 0x8800-0x97ff [-128, 127]
			bgTile = int(int8(r.g.VRAM.Buffer[mapIdx+topX+topY]))
		}

		p := uint16(PAL_BG)
		if highlight {
			p = PAL_HIGHLIGHT_BG
		}

		if r.Model >= util.GB_MODEL_CGB {
			attrs := r.g.VRAM.Buffer[attrIdx+topX+topY]
			p |= uint16(attrs&0x7) * 4
			if util.Bit(attrs, ObjAttrPriority) && util.Bit(r.g.LCDC, BgEnable) {
				p |= OBJ_PRIORITY
			}
			if util.Bit(attrs, ObjAttrBank) {
				localData += GB_SIZE_VRAM_BANK0
			}
			if util.Bit(attrs, ObjAttrYFlip) {
				localY = 7 - bottomY
			}
			if util.Bit(attrs, ObjAttrXFlip) {
				tileDataLower := r.g.VRAM.Buffer[localData+(bgTile*8+localY)*2]
				tileDataUpper := r.g.VRAM.Buffer[localData+(bgTile*8+localY)*2+1]
				r.row[x+0] = p | uint16((tileDataUpper&1)<<1) | uint16(tileDataLower&1)
				r.row[x+1] = p | uint16(tileDataUpper&2) | uint16((tileDataLower&2)>>1)
				r.row[x+2] = p | uint16((tileDataUpper&4)>>1) | uint16((tileDataLower&4)>>2)
				r.row[x+3] = p | uint16((tileDataUpper&8)>>2) | uint16((tileDataLower&8)>>3)
				r.row[x+4] = p | uint16((tileDataUpper&16)>>3) | uint16((tileDataLower&16)>>4)
				r.row[x+5] = p | uint16((tileDataUpper&32)>>4) | uint16((tileDataLower&32)>>5)
				r.row[x+6] = p | uint16((tileDataUpper&64)>>5) | uint16((tileDataLower&64)>>6)
				r.row[x+7] = p | uint16((tileDataUpper&128)>>6) | uint16((tileDataLower&128)>>7)
				continue
			}
		}

		tileDataLower := r.g.VRAM.Buffer[localData+(bgTile*8+localY)*2]
		tileDataUpper := r.g.VRAM.Buffer[localData+(bgTile*8+localY)*2+1]
		r.row[x+7] = p | uint16((tileDataUpper&1)<<1) | uint16(tileDataLower&1)
		r.row[x+6] = p | uint16(tileDataUpper&2) | uint16((tileDataLower&2)>>1)
		r.row[x+5] = p | uint16((tileDataUpper&4)>>1) | uint16((tileDataLower&4)>>2)
		r.row[x+4] = p | uint16((tileDataUpper&8)>>2) | uint16((tileDataLower&8)>>3)
		r.row[x+3] = p | uint16((tileDataUpper&16)>>3) | uint16((tileDataLower&16)>>4)
		r.row[x+2] = p | uint16((tileDataUpper&32)>>4) | uint16((tileDataLower&32)>>5)
		r.row[x+1] = p | uint16((tileDataUpper&64)>>5) | uint16((tileDataLower&64)>>6)
		r.row[x+0] = p | uint16((tileDataUpper&128)>>6) | uint16((tileDataLower&128)>>7)
	}
}

// GBVideoSoftwareRendererDrawObj
func (r *Renderer) drawObj(obj Sprite, startX, endX, y int) {
	objX := int(obj.obj.x)
	ix := objX - 8
	if endX < ix || startX >= ix+8 {
		return
	}
	if objX < endX {
		endX = objX
	}
	if objX-8 > startX {
		startX = objX - 8
	}
	if startX < 0 {
		startX = 0
	}

	vramIdx := 0x0
	tileOffset, bottomY := 0, 0
	objY := int(obj.obj.y)
	if util.Bit(obj.obj.attr, ObjAttrYFlip) {
		bottomY = 7 - ((y - objY - 16) & 7)
		if util.Bit(r.g.LCDC, ObjSize) && y-objY < -8 {
			tileOffset++
		}
	} else {
		bottomY = (y - objY - 16) & 7
		if util.Bit(r.g.LCDC, ObjSize) && y-objY >= -8 {
			tileOffset++
		}
	}
	if util.Bit(r.g.LCDC, ObjSize) && obj.obj.tile&1 == 1 {
		tileOffset--
	}

	mask, mask2 := uint(0x60), uint(OBJ_PRIORITY/3)
	if util.Bit(obj.obj.attr, ObjAttrPriority) {
		mask, mask2 = 0x63, 0
	}

	p := uint16(PAL_OBJ)
	if r.highlightOBJ[obj.index] {
		p = PAL_HIGHLIGHT_OBJ
	}

	if r.Model >= util.GB_MODEL_CGB {
		p |= uint16(obj.obj.attr&0x07) * 4
		if util.Bit(obj.obj.attr, ObjAttrBank) {
			vramIdx += GB_SIZE_VRAM_BANK0
		}
		if !util.Bit(r.g.LCDC, BgEnable) {
			mask, mask2 = 0x60, OBJ_PRIORITY/3
		}
	} else {
		p |= (uint16((obj.obj.attr>>ObjAttrPalette)&1) + 8) * 4 // 8x4 or 9x4
	}

	bottomX, x, objTile := 0, startX, int(obj.obj.tile)+tileOffset
	if (x-objX)&7 != 0 {
		for ; x < endX; x++ {
			if util.Bit(obj.obj.attr, ObjAttrXFlip) {
				bottomX = (x - objX) & 7
			} else {
				bottomX = 7 - ((x - objX) & 7)
			}
			tileDataLower := r.g.VRAM.Buffer[vramIdx+(objTile*8+bottomY)*2]
			tileDataUpper := r.g.VRAM.Buffer[vramIdx+(objTile*8+bottomY)*2+1]
			tileDataUpper >>= bottomX
			tileDataLower >>= bottomX
			current := r.row[x]
			if ((tileDataUpper|tileDataLower)&1 > 0) && (uint(current)&mask == 0) && (uint(current)&mask2) <= OBJ_PRIORITY {
				r.row[x] = p | uint16((tileDataUpper&1)<<1) | uint16(tileDataLower&1)
			}
		}
	} else if util.Bit(obj.obj.attr, ObjAttrXFlip) {
		tileDataLower := r.g.VRAM.Buffer[vramIdx+(objTile*8+bottomY)*2]
		tileDataUpper := r.g.VRAM.Buffer[vramIdx+(objTile*8+bottomY)*2+1]
		current := r.row[x]
		if ((tileDataUpper|tileDataLower)&1) != 0 && (uint(current)&mask == 0) && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x] = p | uint16((tileDataUpper&1)<<1) | uint16(tileDataLower&1)
		}
		current = r.row[x+1]
		if ((tileDataUpper|tileDataLower)&2) != 0 && (uint(current)&mask == 0) && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x+1] = p | uint16(tileDataUpper&2) | uint16((tileDataLower&2)>>1)
		}
		current = r.row[x+2]
		if ((tileDataUpper|tileDataLower)&4) != 0 && (uint(current)&mask == 0) && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x+2] = p | uint16((tileDataUpper&4)>>1) | uint16((tileDataLower&4)>>2)
		}
		current = r.row[x+3]
		if ((tileDataUpper|tileDataLower)&8) != 0 && (uint(current)&mask == 0) && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x+3] = p | uint16((tileDataUpper&8)>>2) | uint16((tileDataLower&8)>>3)
		}
		current = r.row[x+4]
		if ((tileDataUpper|tileDataLower)&16) != 0 && (uint(current)&mask == 0) && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x+4] = p | uint16((tileDataUpper&16)>>3) | uint16((tileDataLower&16)>>4)
		}
		current = r.row[x+5]
		if ((tileDataUpper|tileDataLower)&32) != 0 && (uint(current)&mask == 0) && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x+5] = p | uint16((tileDataUpper&32)>>4) | uint16((tileDataLower&32)>>5)
		}
		current = r.row[x+6]
		if ((tileDataUpper|tileDataLower)&64) != 0 && (uint(current)&mask == 0) && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x+6] = p | uint16((tileDataUpper&64)>>5) | uint16((tileDataLower&64)>>6)
		}
		current = r.row[x+7]
		if ((tileDataUpper|tileDataLower)&128) != 0 && (uint(current)&mask == 0) && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x+7] = p | uint16((tileDataUpper&128)>>6) | uint16((tileDataLower&128)>>7)
		}
	} else {
		tileDataLower := r.g.VRAM.Buffer[vramIdx+(objTile*8+bottomY)*2]
		tileDataUpper := r.g.VRAM.Buffer[vramIdx+(objTile*8+bottomY)*2+1]
		current := r.row[x+7]
		if ((tileDataUpper|tileDataLower)&1) != 0 && (uint(current)&mask) == 0 && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x+7] = p | uint16((tileDataUpper&1)<<1) | uint16(tileDataLower&1)
		}
		current = r.row[x+6]
		if ((tileDataUpper|tileDataLower)&2) != 0 && (uint(current)&mask) == 0 && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x+6] = p | uint16(tileDataUpper&2) | uint16((tileDataLower&2)>>1)
		}
		current = r.row[x+5]
		if ((tileDataUpper|tileDataLower)&4) != 0 && (uint(current)&mask) == 0 && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x+5] = p | uint16((tileDataUpper&4)>>1) | uint16((tileDataLower&4)>>2)
		}
		current = r.row[x+4]
		if ((tileDataUpper|tileDataLower)&8) != 0 && (uint(current)&mask) == 0 && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x+4] = p | uint16((tileDataUpper&8)>>2) | uint16((tileDataLower&8)>>3)
		}
		current = r.row[x+3]
		if ((tileDataUpper|tileDataLower)&16) != 0 && (uint(current)&mask) == 0 && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x+3] = p | uint16((tileDataUpper&16)>>3) | uint16((tileDataLower&16)>>4)
		}
		current = r.row[x+2]
		if ((tileDataUpper|tileDataLower)&32) != 0 && (uint(current)&mask) == 0 && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x+2] = p | uint16((tileDataUpper&32)>>4) | uint16((tileDataLower&32)>>5)
		}
		current = r.row[x+1]
		if ((tileDataUpper|tileDataLower)&64) != 0 && (uint(current)&mask) == 0 && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x+1] = p | uint16((tileDataUpper&64)>>5) | uint16((tileDataLower&64)>>6)
		}
		current = r.row[x]
		if ((tileDataUpper|tileDataLower)&128) != 0 && (uint(current)&mask) == 0 && (uint(current)&mask2) <= OBJ_PRIORITY {
			r.row[x] = p | uint16((tileDataUpper&128)>>6) | uint16((tileDataLower&128)>>7)
		}
	}
}

// _cleanOAM
func (r *Renderer) cleanOAM(y int) {
	spriteHeight := 8
	if util.Bit(r.g.LCDC, ObjSize) {
		spriteHeight = 16
	}

	o := 0
	for i := 0; i < MAX_OBJ && o < MAX_LINE_OBJ; i++ {
		oy := int(r.g.Oam.Objs[i].y)
		if y < oy-16 || y >= oy-16+spriteHeight {
			continue
		}

		r.obj[o].obj = *r.g.Oam.Objs[i]
		r.obj[o].index = int8(i)
		o++
		if o == 10 {
			break
		}
	}
	r.objMax = o
}

// _inWindow
func (r *Renderer) inWindow() bool {
	return util.Bit(r.g.LCDC, Window) && HORIZONTAL_PIXELS+7 > r.wx
}

// _clearScreen
func (r *Renderer) clearScreen() {
	if r.Model&util.GB_MODEL_SGB != 0 {
		return
	}

	for y := 0; y < VERTICAL_PIXELS; y++ {
		row := r.OutputBuffer[r.outputBufferStride*y:]
		for x := 0; x < HORIZONTAL_PIXELS; x += 4 {
			row[x+0] = r.Palette[0]
			row[x+1] = r.Palette[0]
			row[x+2] = r.Palette[0]
			row[x+3] = r.Palette[0]
		}
	}
}
