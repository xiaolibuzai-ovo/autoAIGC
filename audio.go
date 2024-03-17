package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
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
	err = os.WriteFile(name, buf, 777)
	if err != nil {
		return
	}
	return name, nil
}

type MergeAudioResponse struct {
	Data struct {
		FileRequestId string `json:"FileRequestId"`
	} `json:"Data"`
	Code      int         `json:"Code"`
	Message   interface{} `json:"Message"`
	Action    interface{} `json:"Action"`
	SessionId interface{} `json:"SessionId"`
}

type HandleStatusResponse struct {
	Data struct {
		Status       int         `json:"Status"`
		Message      interface{} `json:"Message"`
		FileName     string      `json:"FileName"`
		FolderName   string      `json:"FolderName"`
		DownloadLink string      `json:"DownloadLink"`
	} `json:"Data"`
	Code      int         `json:"Code"`
	Message   interface{} `json:"Message"`
	Action    interface{} `json:"Action"`
	SessionId interface{} `json:"SessionId"`
}

// MergeAudio https://products.aspose.app/audio/zh-cn/merger/api
func MergeAudio(ctx context.Context, audioUrls []string) (audioUrl string, err error) {
	var (
		header = make(http.Header)
		param  = url.Values{}

		bodyIdx = 1
	)
	header.Set("authority", "api.products.aspose.app")
	param.Set("audioFormat", "mp3")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, audioUrl := range audioUrls {
		// 打开要上传的文件
		file, err := os.Open(audioUrl)
		if err != nil {
			fmt.Println(err)
			return
		}

		// 创建一个文件表单字段并将文件内容写入其中
		part, err := writer.CreateFormFile(fmt.Sprint(bodyIdx), file.Name())
		if err != nil {
			fmt.Println(err)
			return
		}
		if _, err = io.Copy(part, file); err != nil {
			fmt.Println(err)
			return
		}

		file.Close()
		bodyIdx++
	}

	// 关闭multipart writer
	if err = writer.Close(); err != nil {
		fmt.Println(err)
		return
	}
	request, err := BuildRequest(ctx, http.MethodGet, "https://api.products.aspose.app/audio/merger/api/merger", body, header, param)
	if err != nil {
		return "", err
	}
	var resp MergeAudioResponse
	err = SendRequest(request, &http.Client{}, &resp)
	if err != nil {
		return "", err
	}
	requestId := resp.Data.FileRequestId
	param = url.Values{}
	param.Set("fileRequestId", requestId)
	for {
		req, err := BuildRequest(ctx, http.MethodGet, "https://api.products.aspose.app/audio/merger/api/merger/HandleStatus", nil, header, param)
		if err != nil {
			return "", err
		}
		var resp0 HandleStatusResponse
		err = SendRequest(req, &http.Client{}, &resp0)
		if err != nil {
			return "", err
		}
		if resp0.Data.Status == 0 {
			// 转换完成
			link := resp0.Data.DownloadLink
			// 下载本地
			localAudio, err := SaveFileLocal(link, "./tmp/audio/", fmt.Sprintf("%d.mp3", time.Now().Unix()))
			if err != nil {
				return "", err
			}
			return localAudio, nil
		}
	}
}
