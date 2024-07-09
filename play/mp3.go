package play

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
)


func Play(data []byte) {
	reader := io.NopCloser(bytes.NewReader(data))
	// 解码音频流
	streamer, format, err := wav.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(streamer)
}

func PlayFrame(frame []byte) {
	reader := bytes.NewReader(frame)

	streamer, format, err := wav.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/60))

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)
	streamer.Close()

	for {
		fmt.Print("Press [ENTER] to fire a gunshot! ")
		fmt.Scanln()

		shot := buffer.Streamer(0, buffer.Len())
		speaker.Play(shot)
	}
}
