package main

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io"
	"os"
	"time"
)

func TTS(ctx context.Context, voice string, content string) (voiceUrl string, err error) {
	// TODO 先直接生成tts 后续对长句子切分生成
	client := GetOpenaiClient()
	speech, err := client.CreateSpeech(ctx, openai.CreateSpeechRequest{
		Model: openai.TTSModel1,
		Input: content,
		Voice: openai.VoiceAlloy,
		//ResponseFormat: openai.SpeechResponseFormatMp3, // default mp3
	})
	if err != nil {
		return
	}
	defer speech.Close()

	buf, err := io.ReadAll(speech)
	if err != nil {
		return
	}
	name := fmt.Sprintf("./tmp/voice/%d.mp3", time.Now().Unix())
	// save buf to file as mp3
	err = os.WriteFile(name, buf, 0644)
	if err != nil {
		return
	}
	return name, nil
}
