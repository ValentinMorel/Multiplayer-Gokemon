package audio

// typedef unsigned char Uint8;
// void OnAudioPlayback(void *userdata, Uint8 *stream, int len);
import "C"
import (
	"log"
	"math"
	"reflect"
	"unsafe"

	"github.com/hajimehoshi/oto"
	"github.com/pokemium/worldwide/pkg/gbc/apu"
	"github.com/veandco/go-sdl2/sdl"
)

var context *oto.Context
var spec *sdl.AudioSpec
var Stream []byte
var enable *bool
var device sdl.AudioDeviceID = 0
var offset int

const (
	toneHz   = 2048
	sampleHz = 44100
	dPhase   = 2 * math.Pi * toneHz / sampleHz
)

//export OnAudioPlayback
func OnAudioPlayback(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	n := int(length)
	hdr := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(stream)), Len: n, Cap: n}
	buf := *(*[]byte)(unsafe.Pointer(&hdr))
	for i := 0; i < n; i++ {
		buf[i] = apu.Buffer[offset]
		offset = (offset + 1) % len(apu.Buffer) // Increase audio offset and loop when it reaches the end
	}
}

func Reset(enablePtr *bool) {
	enable = enablePtr
	if err := sdl.Init(sdl.INIT_AUDIO); err != nil {
		log.Println(err)
		return
	}

	spec = &sdl.AudioSpec{
		Freq:     sampleHz,
		Format:   sdl.AUDIO_U8,
		Channels: 2,
		Samples:  toneHz,
		Callback: nil,
	}
	_device, err := sdl.OpenAudioDevice("", false, spec, nil, 0)
	if err != nil {
		return
	}
	device = _device

}

func Play() {
	sdl.QueueAudio(device, apu.Buffer)
	sdl.PauseAudioDevice(device, false)

}

func SetStream(b []byte) { Stream = b }
