package main

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"strings"
	"sync"
	"time"
)

const ( // 参数先写死
	paragraphNum      = 1                    // 段落数
	GPTModel          = openai.GPT3Dot5Turbo // ai模型
	SubtitlesPosition = 1                    // 字幕在视频中的位置
	BgMusic           = ""                   // 背景音乐
	TmpFolder         = "./tmp/"             // 存储临时文件夹

	// TODO 简单指定测试 后续上游数据处理 llm分析后得到主题
	videoSubject          = "scenery" // 主题
	language              = "Chinese" // 语言
	AmountOfSubjectVideos = 3         // 生成的主题视频数量

	searchVideoLimit = 10 // 搜索的视频数量
	MinDuration      = 10 // 视频最少持续时间 秒为单位

	saveVideoDir = "./tmp/video"
)

const ( // 自动上传相关
	AutoUploadTiktok   = false
	AutoUploadDouYi    = false
	AutoUploadYouTube  = false
	AutoUploadBiliBili = false
)

func GenerateVideo(ctx context.Context) (err error) {
	var (
		subjectText string
		searchTerms []string
		videos      []string
		localVideos []string
		localAudios []string

		mergeAudioUrl    string
		subtitleUrl      string
		combinedVideoUrl string

		wg sync.WaitGroup
		mx sync.Mutex
	)
	// 根据主题生成内容
	subjectText, err = GenerateSubjectText(ctx, GPTModel, videoSubject, language)
	if err != nil {
		return
	}

	// 根据主题 & 内容 生成视频搜索词
	searchTerms, err = GenerateSearchTermsBySubject(ctx, GPTModel, AmountOfSubjectVideos, videoSubject, subjectText)
	if err != nil {
		return
	}

	// 根据关键词检索视频
	for _, term := range searchTerms {
		// 搜索相关视频
		tmpTerm := term
		go func() {
			wg.Add(1)
			defer wg.Done()
			videoUrls, err := SearchVideosInPexels(ctx, tmpTerm, searchVideoLimit, MinDuration)
			if err != nil {
				return
			}
			mx.Lock()
			defer mx.Unlock()
			videos = append(videos, videoUrls...)
		}()
	}
	wg.Wait()
	if len(videos) == 0 {
		return
	}
	// video url去重
	videos = RemoveRepetitionString(videos)

	// video保存本地
	for _, video := range videos {
		localVideoUrl, err := SaveFileLocal(video, saveVideoDir, fmt.Sprintf("%d.mp4", time.Now().UnixNano()))
		if err != nil {
			continue
		}
		localVideos = append(localVideos, localVideoUrl)
	}
	// 分割主题内容,生成tts
	sentences := strings.Split(subjectText, ". ")
	for _, sentence := range sentences {
		if len(sentence) == 0 {
			// remove empty
			continue
		}
		// 语音转换字幕 srt
		ttsUrl, err := TTS(ctx, "", sentence)
		if err != nil {
			return err
		}
		localAudios = append(localAudios, ttsUrl)
	}
	// 合成tts音频
	mergeAudioUrl, err = MergeAudioByFfmpeg(localAudios)
	if err != nil {
		return err
	}
	// 生成字幕
	subtitleUrl, err = GenerateSubtitlesLocally(ctx, sentences, localVideos, "./tmp/subtitles/")
	if err != nil {
		return err
	}
	// 合并视频
	combinedVideoUrl, err = MergeVideo(ctx, localVideos)
	if err != nil {
		return err
	}

	_ = mergeAudioUrl
	_ = subtitleUrl
	_ = combinedVideoUrl
	// 融合语音和视频

	// 生成视频元数据信息(如标题/分类等)

	// 自动上传相关
	return
}
