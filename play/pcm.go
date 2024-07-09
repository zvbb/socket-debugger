package play

import (
	"github.com/hajimehoshi/oto"
)

// Initialize oto context
const sampleRate = 48000
const numChannels = 1
const bitDepthInBytes = 2

var ctx, _ = oto.NewContext(sampleRate, numChannels, bitDepthInBytes, 8000)

// Create a new player
var player = ctx.NewPlayer()

func Pcm(frame []byte) {
	player.Write(frame)
}
