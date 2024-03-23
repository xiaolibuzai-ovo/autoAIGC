package main

import (
	"context"
	"fmt"
	"testing"
)

func TestGenerateSubtitlesLocally(t *testing.T) {
	sentences := []string{
		"这个视频将带您领略大自然的壮丽风景",
		"在这片美丽的土地上，您将看到壮观的山脉、宁静的湖泊以及茂密的森林",
		"这些景色让人心旷神怡，让人感受到大自然的力量和美丽",
		"无论是日出时分还是傍晚时刻，每一个景色都让人流连忘返",
		"在这个视频中，您将看到壮丽的瀑布从高山倾泻而下，水花飞溅，如梦如幻",
		"您还将欣赏到广袤的草原上悠闲自在的动物，它们在自然环境中自由奔跑、觅食 这些景象让人感叹大自然的神奇和无穷魅力",
		"除了自然风光，这个视频还将带您领略不同季节的景色变化",
		"春天的万物复苏、夏天的绿荫蔽日、秋天的金黄色调以及冬天的银装素裹，每一个季节都有独特的美丽之处",
		"不同季节所展现的景色交相辉映，让人沉浸在大自然的魅力中无法自拔",
	}

	audios := []string{"./tmp/mergeAudio/1711209179090926000.mp3"}

	url, err := GenerateSubtitlesLocally(context.Background(), sentences, audios, "./tmp/subtitles/")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(url)
}
