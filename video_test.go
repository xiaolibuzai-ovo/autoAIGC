package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestMergeVideo(t *testing.T) {
	var videos = []string{"a.mp4", "b.mp4"}
	mergeVideoUrl, err := MergeVideo(context.Background(), videos)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(mergeVideoUrl)
	println("Videos merged successfully.")
}

func generateVideo(combinedVideoPath string, ttsPath string, subtitlesPath string, threads int, subtitlesPosition string, textColor string) string {
	// Split subtitles position into horizontal and vertical
	positions := strings.Split(subtitlesPosition, ",")
	horizontalSubtitlesPosition := positions[0]
	verticalSubtitlesPosition := positions[1]

	// Build ffmpeg command
	cmd := exec.Command("ffmpeg",
		"-i", combinedVideoPath,
		"-i", ttsPath,
		"-vf", "subtitles="+subtitlesPath+",drawtext=fontfile=/path/to/font.ttf:fontsize=100:fontcolor="+textColor+":x="+horizontalSubtitlesPosition+":y="+verticalSubtitlesPosition,
		"-threads", string(rune(threads)),
		"-c:a", "copy",
		"output.mp4",
	)

	// Execute command
	err := cmd.Run()
	if err != nil {
		panic(err) // Handle error appropriately
	}

	return "output.mp4"
}

func TestGenerateVideo(t *testing.T) {
	//  ffmpeg -i ./tmp/mergeVideo/1710687890048621000.mp4 -i ./tmp/mergeAudio/1711034798001771000.mp3 -vcodec copy -acodec copy output.mp4

	//  ffmpeg -i output.mp4 -strict -2 -vf \
	//subtitles=./tmp/1711031828166829000.srt:force_style='Fontsize=30\,FontName=FZYBKSJW--GB1-0' -qscale:v 3 output2.mp4

	// Example usage
	combinedVideoPath := "./tmp/mergeVideo/1710687890048621000.mp4"
	ttsPath := "./tmp/mergeAudio/1711034798001771000.mp3"
	subtitlesPath := "./tmp/1711031828166829000.srt"
	threads := 2
	subtitlesPosition := "bottom,center"
	textColor := "white"

	outputPath := generateVideo(combinedVideoPath, ttsPath, subtitlesPath, threads, subtitlesPosition, textColor)
	println("Output video path:", outputPath)
}
