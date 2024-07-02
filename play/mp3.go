package play

import (
	"bytes"
	"io"
	"log"
	"time"

	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)



func Play(mp3Data []byte) {
	reader := io.NopCloser(bytes.NewReader(mp3Data))
	// 解码音频流
	streamer, format, err := mp3.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	speaker.Play(streamer)
}
