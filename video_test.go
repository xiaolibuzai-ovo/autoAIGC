package main

import (
	"context"
	"fmt"
	"testing"
)

func TestCombinedVideo(t *testing.T) {
	var videos = []string{"a.mp4", "b.mp4"}
	CombinedVideoUrl, err := CombinedVideo(context.Background(), videos)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(CombinedVideoUrl)
	println("Videos merged successfully.")
}

func TestMixAllInfoForVideo(t *testing.T) {
	combinedVideoPath := "./tmp/mergeVideo/1710687890048621000.mp4"
	ttsPath := "./tmp/mergeAudio/1711034798001771000.mp3"
	subtitlesPath := "./tmp/1711031828166829000.srt"
	finalVideo, err := MixAllInfoForVideo(context.Background(), combinedVideoPath, ttsPath, subtitlesPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(finalVideo)
}
