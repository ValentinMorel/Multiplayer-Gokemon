package emulator

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pokemium/worldwide/pkg/emulator/offsets"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/pokemium/worldwide/pkg/emulator/audio"
	"github.com/pokemium/worldwide/pkg/emulator/debug"
	"github.com/pokemium/worldwide/pkg/emulator/joypad"
	"github.com/pokemium/worldwide/pkg/gbc"
)

var (
	second = time.NewTicker(time.Second)

	Width, Height int32 = 160, 144
	cache         []byte
)
var IsConnected bool

type wsEvent struct {
	//The json package only accesses the exported fields of struct types
	Event    string `json:"event"`
	Id       string `json:"id"`
	MapId    int    `json:"mapId"`
	X        int32  `json:"x"`
	Y        int32  `json:"y"`
	SpriteId int32  `json:"spriteId"`
}
type PlayerData struct {
	Event    string `json:"event"`
	Id       string `json:"id"`
	MapId    int    `json:"mapId"`
	MapX     int32  `json:"x"`
	MapY     int32  `json:"y"`
	SpriteId int32  `json:"spriteId"`
}
type SpriteData struct {
	X        int32
	Y        int32
	SpriteId int32
}
type MovementData struct {
	MapId       uint8
	MapX        uint8
	MapY        uint8
	Direction   int
	WalkCounter uint8
}

func NewMovementData() *MovementData {
	return &MovementData{
		MapId:       0,
		MapX:        0,
		MapY:        0,
		Direction:   0,
		WalkCounter: 0,
	}
}

func NewPlayerData(e *Emulator) *PlayerData {
	return &PlayerData{
		Id:       "",
		MapId:    0,
		MapX:     0,
		MapY:     0,
		SpriteId: 0,
	}
}

func DrawOtherPlayers(e *Emulator, others []*PlayerData, self *PlayerData) {
	for _, player := range others {
		if player.MapId != self.MapId {
			continue
		}
		x, y := GetPlayerDrawPosition(self, player)
		spriteData := SpriteData{
			X:        x,
			Y:        y,
			SpriteId: player.SpriteId,
		}

		RenderSprite(e, player, &spriteData)

	}
}
func GetPlayerDrawPosition(self *PlayerData, other *PlayerData) (int32, int32) {
	baseX := int32(Width/2 - 16)
	baseY := int32(Height/2 - 12)
	selfX, selfY := GetPlayerPosition(self)
	otherX, otherY := GetPlayerPosition(other)
	return otherX - selfX + baseX, otherY - selfY + baseY

}
func GetPlayerPosition(player *PlayerData) (int32, int32) {
	x := int32(player.MapX * 16)
	y := int32(player.MapY * 16)
	// TODO: use walk counter to determine offset on each direction
	return x, y
}

func drawSprite(e *Emulator, offset int, spriteData *SpriteData, flip bool) {

	for i := 0; i < 4; i++ {
		addr := offset + 16*i
		//addr = 16 * i
		bank := 0
		var currentPos int32

		for y := int32(0); y < 8; y++ {
			tileAddr := addr + 2*int(y) + 0x2000*bank
			tileDataLower, tileDataUpper := e.GBC.Video.VRAM.Buffer[tileAddr], e.GBC.Video.VRAM.Buffer[tileAddr+1]
			if i == 0 {
				currentPos = (int32(spriteData.Y)+y)*Width + int32(spriteData.X)
				//log.Println("currentPos 0: ", currentPos)
			} else if i == 1 {
				currentPos = (int32(spriteData.Y)+y)*Width + int32(spriteData.X) + int32(8)
				//log.Println("currentPos 1: ", currentPos)
			} else if i == 2 {
				currentPos = (int32(spriteData.Y)+y+int32(8))*Width + int32(spriteData.X)
				//log.Println("currentPos 2: ", currentPos)
			} else if i == 3 {
				currentPos = (int32(spriteData.Y)+y+int32(8))*Width + int32(spriteData.X) + int32(8)
				//log.Println("currentPos 3: ", currentPos)
			}
			if (currentPos*4)+3 > 160*144*4 || (currentPos*4)+3 < 0 {
				continue
			}
			if flip {
				for x := 0; x < 8; x++ {
					b := (7 - uint(x))
					upperColor := (tileDataUpper >> b) & 0x01
					lowerColor := (tileDataLower >> b) & 0x01
					palIdx := (upperColor << 1) | lowerColor // 0 or 1 or 2 or 3
					if palIdx > 0 {
						p := e.GBC.Video.Palette[e.GBC.Video.Renderer.Lookup[palIdx]]
						red, green, blue := byte((p&0b11111)*8), byte(((p>>5)&0b11111)*8), byte(((p>>10)&0b11111)*8)
						//bufferIdx := i*64*4 + int(y)*8*4 + x*4
						//buffer[bufferIdx] = red
						//buffer[bufferIdx+1] = green
						//buffer[bufferIdx+2] = blue
						//buffer[bufferIdx+3] = 0xff
						//m.SetRGBA(i*8+x, i*8+y, color.RGBA{red, green, blue, 0xff})
						if i == 0 || i == 2 {
							cache[(int(currentPos)+(16-x-1))*4] = red
							cache[(int(currentPos)+(16-x-1))*4+1] = green
							cache[(int(currentPos)+(16-x-1))*4+2] = blue
							cache[(int(currentPos)+(16-x-1))*4+3] = 0xff

						} else if i == 1 || i == 3 {
							cache[(int(currentPos)-x-1)*4] = red
							cache[(int(currentPos)-x-1)*4+1] = green
							cache[(int(currentPos)-x-1)*4+2] = blue
							cache[(int(currentPos)-x-1)*4+3] = 0xff
						}

						//currentPos += 1
					} else {
						//currentPos += 1
					}
				}

			} else if !flip {
				for x := 0; x < 8; x++ {
					b := (7 - uint(x))
					upperColor := (tileDataUpper >> b) & 0x01
					lowerColor := (tileDataLower >> b) & 0x01
					palIdx := (upperColor << 1) | lowerColor // 0 or 1 or 2 or 3
					if palIdx > 0 {
						p := e.GBC.Video.Palette[e.GBC.Video.Renderer.Lookup[palIdx]]
						red, green, blue := byte((p&0b11111)*8), byte(((p>>5)&0b11111)*8), byte(((p>>10)&0b11111)*8)
						//bufferIdx := i*64*4 + int(y)*8*4 + x*4
						//buffer[bufferIdx] = red
						//buffer[bufferIdx+1] = green
						//buffer[bufferIdx+2] = blue
						//buffer[bufferIdx+3] = 0xff
						//m.SetRGBA(i*8+x, i*8+y, color.RGBA{red, green, blue, 0xff})
						cache[currentPos*4] = red
						cache[currentPos*4+1] = green
						cache[currentPos*4+2] = blue
						cache[currentPos*4+3] = 0xff

						currentPos += 1
					} else {
						currentPos += 1
					}
				}
			}
		}
	}
}

func RenderSprite(e *Emulator, playerData *PlayerData, spriteData *SpriteData) bool {

	const SPRITE_HEIGHT uint32 = 16
	const SPRITE_WIDTH uint32 = 16

	// Check if the sprite actually appears on the screen
	if spriteData.Y >= int32(Height) || spriteData.Y+int32(SPRITE_HEIGHT) <= 0 || spriteData.X >= int32(Width) || spriteData.X+int32(SPRITE_WIDTH) <= 0 {
		return false
	}
	//m := image.NewRGBA(image.Rect(0, 0, 8*8, 8*8))

	// Check if the sprite is hidden under a menu or a tile
	//var buffer [64 * 4 * 4]byte
	//spriteId := e.GBC.Load8(0xC102)

	// downward :     0 to 3
	// upward :       4 to 7
	// to the left :  8 to 11
	// to the right: 12 to 15

	if spriteData.SpriteId == 0 {
		drawSprite(e, 0, spriteData, false)
		return true
	} else if spriteData.SpriteId == 1 {
		drawSprite(e, (16 * 16 * 8), spriteData, false)
		return true
	} else if spriteData.SpriteId == 2 {
		drawSprite(e, 0, spriteData, false)
		return true
	} else if spriteData.SpriteId == 3 {
		drawSprite(e, (16 * 16 * 8), spriteData, true)
		return true
	} else if spriteData.SpriteId == 4 {
		drawSprite(e, (16 * 4), spriteData, false)
		return true
	} else if spriteData.SpriteId == 5 {
		drawSprite(e, (16*16*8 + 4*16), spriteData, false)
		return true
	} else if spriteData.SpriteId == 6 {
		drawSprite(e, (16 * 4), spriteData, false)
		return true
	} else if spriteData.SpriteId == 7 {
		drawSprite(e, (16*16*8 + 4*16), spriteData, false)
		return true
	} else if spriteData.SpriteId == 8 {
		drawSprite(e, (16 * 8), spriteData, false)
		return true
	} else if spriteData.SpriteId == 9 {
		drawSprite(e, (16*16*8 + 8*16), spriteData, false)
		return true
	} else if spriteData.SpriteId == 10 {
		drawSprite(e, (16 * 8), spriteData, false)
		return true
	} else if spriteData.SpriteId == 11 {
		drawSprite(e, (16*16*8 + 8*16), spriteData, false)
		return true
	} else if spriteData.SpriteId == 12 {
		drawSprite(e, (16 * 8), spriteData, true)
		return true
	} else if spriteData.SpriteId == 13 {
		drawSprite(e, (16*16*8 + 8*16), spriteData, true)
		return true
	} else if spriteData.SpriteId == 14 {
		drawSprite(e, (16 * 8), spriteData, true)
		return true
	} else if spriteData.SpriteId == 15 {
		drawSprite(e, (16*16*8 + 8*16), spriteData, true)
		return true
	}
	return true

	/*
		f, _ := os.Create("img1.jpg")

		if err := jpeg.Encode(f, m, nil); err != nil {
			log.Println("unable to encode image.")
		}
		log.Println(buffer)
	*/

}

func GetTileIdAddr(x uint8, y uint8) uint16 {
	yOffset := uint16(((y + 4) & 0xF0) >> 3)
	xOffset := uint16((x >> 3) + 0x14)

	return offsets.TILE_MAP + 20*yOffset + xOffset
}

type Emulator struct {
	sync.Mutex
	GBC      *gbc.GBC
	Rom      []byte
	RomDir   string
	debugger *debug.Debugger
	pause    bool
	reset    bool
	quit     bool
	Self     *PlayerData
	Others   []*PlayerData
	Ws       *websocket.Conn
}

func New(romData []byte, romDir string) *Emulator {
	g := gbc.New(romData, joypad.Handler, audio.SetStream)
	audioBool := false
	audio.Reset(&audioBool)

	ebiten.SetWindowResizable(true)
	ebiten.SetWindowTitle("60fps")
	ebiten.SetWindowSize(160*2, 144*2)

	e := &Emulator{
		GBC:    g,
		Rom:    romData,
		RomDir: romDir,
	}

	e.debugger = debug.New(g, &e.pause)
	e.setupCloseHandler()
	e.Self = NewPlayerData(e)
	e.loadSav()
	return e
}

func (e *Emulator) ResetGBC() {
	e.writeSav()

	oldCallbacks := e.GBC.Callbacks
	e.GBC = gbc.New(e.Rom, joypad.Handler, audio.SetStream)
	e.GBC.Callbacks = oldCallbacks

	e.debugger.Reset(e.GBC)
	e.loadSav()

	e.reset = false
}

func (e *Emulator) Update() error {
	if e.quit {
		return errors.New("quit")
	}
	if e.reset {
		e.ResetGBC()
		return nil
	}
	if e.pause {
		return nil
	}

	defer e.GBC.PanicHandler("update", true)
	/*
		log.Println("0xFF47", e.GBC.Load8(0xFF47))
		log.Println("0xFF48", e.GBC.Load8(0xFF48))
		log.Println("0xFF49", e.GBC.Load8(0xFF49))
		log.Println("0xFF69", e.GBC.Load8(0xff69))
		log.Println("0xFF6B", e.GBC.Load8(0xff6b))
		log.Println("color palette: ", e.GBC.Video.Palette)
	*/

	e.GBC.Update()
	if e.pause {
		return nil
	}

	audio.Play()

	select {
	case <-second.C:
		e.GBC.RTC.IncrementSecond()
		ebiten.SetWindowTitle(fmt.Sprintf("%dfps", int(ebiten.CurrentTPS())))
	default:
	}

	return nil
}

func (e *Emulator) Draw(screen *ebiten.Image) {
	if IsConnected {
		defer e.GBC.PanicHandler("draw", true)

		cache = e.GBC.Draw()

		DrawOtherPlayers(e, e.Others, e.Self)

		screen.ReplacePixels(cache)

	} else if !IsConnected {
		defer e.GBC.PanicHandler("draw", true)
		cache = e.GBC.Draw()
		screen.ReplacePixels(cache)
	}
	if e.pause {
		screen.ReplacePixels(cache)
		return
	}

}

func (e *Emulator) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 160, 144
}

func (e *Emulator) Exit() {
	e.writeSav()
}

func (e *Emulator) setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		e.Exit()
		os.Exit(0)
	}()
}
