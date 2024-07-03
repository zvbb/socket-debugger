package play

import (
	"bytes"
	"io"
	"log"
	"time"

	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)

func Play(mp3Data []byte) {
	reader := io.NopCloser(bytes.NewReader(mp3Data))
	// 解码音频流
	streamer, format, err := wav.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	speaker.Play(streamer)
}
