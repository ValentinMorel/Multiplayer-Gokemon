package emulator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

var base_addr = "localhost:8081"

const (
	fps = 60
	// frameMs = 1000 / fps // Duration of a frame in milli seconds
	frameNs = int64(1e3) / fps // Duration of a frame in nano seconds
)

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
