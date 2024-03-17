package main

import (
	"context"
	"fmt"
	"testing"
)

func TestGenerateSubtitlesLocally(t *testing.T) {
	sentences := []string{"华为致力于把数字世界带入每个,每个家庭,每个组织,构建万物互联的智能世界"}
	audios := []string{"16k16bit.mp3"}

	url, err := GenerateSubtitlesLocally(context.Background(), sentences, audios, "./tmp")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(url)
}
