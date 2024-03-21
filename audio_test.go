package main

import (
	"fmt"
	"testing"
)

func TestMergeAudioByFfmpeg(t *testing.T) {
	audio, err := MergeAudioByFfmpeg([]string{"/Users/limingzhi/go/src/autoAIGC/tmp/audio/16k16bit.mp3", "/Users/limingzhi/go/src/autoAIGC/tmp/audio/16k16bit.mp3"})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(audio)
}
