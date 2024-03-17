package main

import (
	"context"
	"fmt"
	"github.com/asticode/go-astisub"
	"os"
	"path/filepath"
	"time"
)

func GenerateSubtitlesLocally(ctx context.Context, sentences []string, audios []string, outputPath string) (string, error) {
	var subs astisub.Subtitles
	startTime := 0 * time.Second

	for i, sentence := range sentences {
		duration := GetAudioDuration(audios[i])
		endTime := startTime + duration

		item := astisub.Item{
			StartAt: startTime,
			EndAt:   endTime,
			Lines: []astisub.Line{
				{
					Items: []astisub.LineItem{
						{
							Text: sentence,
						},
					},
				},
			},
		}

		subs.Items = append(subs.Items, &item)

		startTime = endTime
	}

	fileUrl := fmt.Sprintf("%d.srt", time.Now().UnixNano())
	videoPath := filepath.Join(outputPath, fileUrl)

	// 创建文件所在的文件夹路径
	err := os.MkdirAll(filepath.Dir(videoPath), os.ModePerm)
	if err != nil {
		return "", err
	}

	// 将字幕写入到本地文件
	outputFile, err := os.Create(videoPath)
	if err != nil {
		return "", err
	}
	defer outputFile.Close()

	err = subs.WriteToSRT(outputFile)
	if err != nil {
		return "", err
	}

	return videoPath, nil
}
