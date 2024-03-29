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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func TTS(ctx context.Context, voice string, content string) (audioUrl string, err error) {
	// TODO 先直接生成tts 后续对长句子切分生成
	client := GetOpenaiClient()
	speech, err := client.CreateSpeech(ctx, openai.CreateSpeechRequest{
		Model: openai.TTSModel1,
		Input: content,
		Voice: openai.VoiceAlloy,
		//ResponseFormat: openai.SpeechResponseFormatMp3, // default mp3
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	defer speech.Close()
	buf, err := io.ReadAll(speech)
	if err != nil {
		return
	}
	audioUrl = fmt.Sprintf("./tmp/audio/%d.mp3", time.Now().UnixNano())
	err = os.MkdirAll(filepath.Dir(audioUrl), os.ModePerm)
	if err != nil {
		return
	}
	// save buf to file as mp3
	err = os.WriteFile(audioUrl, buf, 777)
	if err != nil {
		return
	}
	return audioUrl, nil
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
	for _, audio := range audioUrls {
		// 打开要上传的文件
		file, err := os.Open(audio)
		if err != nil {
			fmt.Println(err)
			return "", err
		}

		// 创建一个文件表单字段并将文件内容写入其中
		part, err := writer.CreateFormFile(fmt.Sprint(bodyIdx), file.Name())
		if err != nil {
			fmt.Println(err)
			return "", err

		}
		if _, err = io.Copy(part, file); err != nil {
			fmt.Println(err)
			return "", err
		}

		file.Close()
		bodyIdx++
	}
	// 关闭multipart writer
	if err = writer.Close(); err != nil {
		return
	}
	request, err := BuildRequest(ctx, http.MethodGet, "https://api.products.aspose.app/audio/merger/api/merger", body, header, param)
	if err != nil {
		return "", err
	}
	var resp MergeAudioResponse
	err = SendRequest(request, &http.Client{}, &resp)
	if err != nil {
		fmt.Println(err)
		fmt.Println(111)
		return "", err
	}
	requestId := resp.Data.FileRequestId
	param = url.Values{}
	param.Set("fileRequestId", requestId)
	for {
		req, err := BuildRequest(ctx, http.MethodGet, "https://api.products.aspose.app/audio/merger/api/merger/HandleStatus", nil, header, param)
		if err != nil {
			fmt.Println(err)
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
			localAudio, err := SaveFileLocal(link, "./tmp/audio/", fmt.Sprintf("%d.mp3", time.Now().UnixNano()))
			if err != nil {
				return "", err
			}
			return localAudio, nil
		}
	}
}

func CombinedAudioByFfmpeg(audioUrls []string) (mergeAudioUrl string, err error) {
	var (
		mergeAudioDir = "./tmp/mergeAudio/"
	)
	// 给url加上绝对路径
	pwd, err := os.Getwd()
	if err != nil {
		return
	}
	for i, audioUrl := range audioUrls {
		audioUrl = audioUrl[1:] // 去掉相对路径的点
		audioUrls[i] = fmt.Sprintf("%s%s", pwd, audioUrl)
	}

	// 合并所有文件
	txtFile, err := CreateTxtFileWithDynamicContent(audioUrls)
	if err != nil {
		return
	}
	mergeAudioUrl = fmt.Sprintf("%s%d.mp3", mergeAudioDir, time.Now().UnixNano())
	err = os.MkdirAll(filepath.Dir(mergeAudioUrl), os.ModePerm)
	if err != nil {
		return
	}
	cmd := fmt.Sprintf(`ffmpeg -f concat -safe 0 -i %s -c copy %s`, txtFile, mergeAudioUrl)
	command := exec.Command("/bin/bash", "-c", cmd)
	_, err = command.CombinedOutput()
	if err != nil {
		return
	}
	return
}

/*
GetAudioDuration 利用ffmpeg获取音频持续时间
*/
func GetAudioDuration(audio string) time.Duration {
	// 格式转换 ffmpeg -i xxxx  2>&1 | grep 'Duration' | cut -d ' ' -f 4 | sed s/,//
	cmd := fmt.Sprintf("ffmpeg -i %s 2>&1 | grep 'Duration' | cut -d ' ' -f 4 | sed s/,//", audio)
	command := exec.Command("/bin/bash", "-c", cmd)
	res, err := command.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	body := string(res)

	parts := strings.Split(body, ":")
	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])

	secondsAndMilliseconds := strings.Split(parts[2], ".")

	seconds, _ := strconv.Atoi(secondsAndMilliseconds[0])
	milliseconds, _ := strconv.ParseFloat("0."+strings.TrimSpace(secondsAndMilliseconds[1]), 64)
	totalSeconds := float64(hours*3600+minutes*60+seconds) + milliseconds
	duration := time.Duration(totalSeconds * float64(time.Second))

	return duration
}
