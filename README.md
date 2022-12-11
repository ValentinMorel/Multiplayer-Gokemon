![logo](./logo.png)

# 🌏 worldwide
![Go](https://github.com/pokemium/worldwide/workflows/Go/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/pokemium/worldwide)](https://goreportcard.com/report/github.com/pokemium/worldwide)
[![GitHub stars](https://img.shields.io/github/stars/pokemium/worldwide)](https://github.com/pokemium/worldwide/stargazers)
[![GitHub license](https://img.shields.io/github/license/pokemium/worldwide)](https://github.com/pokemium/worldwide/blob/master/LICENSE)

日本語のドキュメントは[こちら](./README.ja.md)

GameBoyColor emulator written in golang.  

This emulator can play a lot of ROMs work without problems and has many features.

<img src="https://imgur.com/ZlrXAW9.png" width="320px"> <img src="https://imgur.com/xVqjkrk.png" width="320px"><br/>
<img src="https://imgur.com/E7oob9c.png" width="320px"> <img src="https://imgur.com/nYpkH95.png" width="320px">

## 🚩 Features & TODO list
- [x] 60fps
- [x] Pass [cpu_instrs](https://github.com/retrio/gb-test-roms/tree/master/cpu_instrs) and [instr_timing](https://github.com/retrio/gb-test-roms/tree/master/instr_timing)
- [x] Low CPU consumption
- [x] Sound(ported from goboy)
- [x] GameBoy Color ROM support
- [x] Multi-platform support
- [x] MBC1, MBC2, MBC3, MBC5 support
- [x] RTC
- [x] SRAM save
- [x] Resizable window
- [x] HTTP server API
- [ ] Plugins support
- [ ] [Libretro](https://docs.libretro.com/) support
- [ ] Netplay in local network
- [ ] Netplay in global network
- [ ] SGB support
- [ ] Shader support

## 🎮 Usage

Download binary from [here](https://github.com/pokemium/worldwide/releases).

```sh
./worldwide "***.gb" # or ***.gbc
```

## 🐛 HTTP Server

`worldwide` contains an HTTP server, and the user can give various instructions to it through HTTP requests.

Please read [Server Document](./server/README.md).

## 🔨 Build

For those who want to build from source code.

Requirements
- Go 1.16
- make

```sh
make build                              # If you use Windows, `make build-windows`
./build/darwin-amd64/worldwide "***.gb" # If you use Windows, `./build/windows-amd64/worldwide.exe "***.gb"`
```

## 📄 Command 

| keyboard             | game pad      |
| -------------------- | ------------- |
| <kbd>&larr;</kbd>    | &larr; button |
| <kbd>&uarr;</kbd>    | &uarr; button |
| <kbd>&darr;</kbd>    | &darr; button |
| <kbd>&rarr;</kbd>    | &rarr; button |
| <kbd>X</kbd>         | A button      |
| <kbd>Z</kbd>         | B button      |
| <kbd>Enter</kbd>     | Start button  |
| <kbd>Backspace</kbd> | Select button |


## Modifications

emulator.go :

```go 
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

```
server.go : 

```go 
var flag int
var (
	mapId    int
	mapX     int32
	mapY     int32
	spriteId int32
)

func (e *Emulator) ConnectToServer() {
	quit := make(chan bool)

	u := url.URL{Scheme: "ws", Host: base_addr, Path: "/socket"}
	c, _, _ := websocket.DefaultDialer.Dial(u.String(), nil)
	e.Ws = c
	IsConnected = true
	player := &PlayerData{}
	e.Others = append(e.Others, player)
	go func() {
		for {
			w, _ := e.Ws.NextWriter(websocket.TextMessage)
			event := fmt.Sprintf(`{"event":"update","mapId": %v, "x": %v, "y": %v, "spriteId": %v}`, int(e.GBC.Load8(0xD35E)), int32(e.GBC.Load8(0xD362)), int32(e.GBC.Load8(0xD361)), int32(e.GBC.Load8(0xC102)))
			w.Write([]byte(event))
			e.Self.MapId = int(e.GBC.Load8(0xD35E))
			e.Self.MapX = int32(e.GBC.Load8(0xD362))
			e.Self.MapY = int32(e.GBC.Load8(0xD361))
			e.Self.SpriteId = int32(e.GBC.Load8(0xC102))

		}
	}()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		var tmpJson map[string]interface{}
		err = json.Unmarshal(message, &tmpJson)
		if err != nil {
			log.Println("unmarshal: ", err)
		}

		if tmpJson["event"].(string) == "createPlayer" {
			e.Self.Id = tmpJson["event"].(string)
		} else {
			e.Others[0].MapId = int(tmpJson["mapId"].(float64))
			e.Others[0].MapX = int32(tmpJson["x"].(float64))
			e.Others[0].MapY = int32(tmpJson["y"].(float64))
			e.Others[0].SpriteId = int32(tmpJson["spriteId"].(float64))

		}

		/* time.Sleep takes a duration for an argument. time.Nanosecond
		   is of type Duration */

	}
	<-quit
}
```