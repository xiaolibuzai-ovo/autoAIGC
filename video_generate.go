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
	MinDuration      = 3  // 视频最少持续时间 秒为单位

	saveVideoDir    = "./tmp/video"
	saveSubtitleDir = "./tmp/subtitles"
)

const ( // 自动上传相关
	AutoUploadTiktok   = false
	AutoUploadDouYi    = false
	AutoUploadYouTube  = false
	AutoUploadBiliBili = false
)

// GenerateVideo 根据主题生成视频并配文与音频
// TODO 调用时应提交一个异步任务 而不是同步等待
/*
 // 当前存在问题
1. ffmpeg融合有问题 视频的大小分辨率有影响
*/
func GenerateVideo(ctx context.Context) (err error) {
	var (
		subjectText string
		searchTerms []string
		videos      []string
		localVideos []string
		localAudios []string

		combinedAudioUrl string
		subtitleUrl      string
		combinedVideoUrl string
		finalVideoUrl    string

		wg sync.WaitGroup
		mx sync.Mutex
	)
	// 根据主题生成内容
	subjectText, err = GenerateSubjectText(ctx, GPTModel, videoSubject, language)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 根据主题 & 内容 生成视频搜索词
	searchTerms, err = GenerateSearchTermsBySubject(ctx, GPTModel, AmountOfSubjectVideos, videoSubject, subjectText)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 根据关键词检索视频
	for _, term := range searchTerms {
		// 搜索相关视频
		tmpTerm := term
		wg.Add(1)
		go func() {
			defer wg.Done()
			videoUrls, err := SearchVideosInPexels(ctx, tmpTerm, searchVideoLimit, MinDuration)
			if err != nil {
				fmt.Println(err)
				return
			}
			mx.Lock()
			defer mx.Unlock()
			videos = append(videos, videoUrls...)
		}()
	}
	wg.Wait()
	if len(videos) == 0 {
		fmt.Println("videos is empty")
		return
	}
	// video url去重
	videos = RemoveRepetitionString(videos)

	//save the videos locally
	// todo 生成视频的时候就保存到本地 加速存储
	for _, video := range videos {
		localVideoUrl, err := SaveFileLocal(video, saveVideoDir, fmt.Sprintf("%d.mp4", time.Now().UnixNano()))
		if err != nil {
			fmt.Println(err)
			continue
		}
		localVideos = append(localVideos, localVideoUrl)
	}
	// split subject text & generate audio by tts
	subjectText = strings.ReplaceAll(subjectText, "\n", "")
	sentences := strings.Split(subjectText, "。") // TODO 区分中英文符号 英根据subjectText语音来区分
	// remove empty
	var tmpSentences []string
	for _, sentence := range sentences {
		if len(sentence) == 0 {
			continue
		}
		tmpSentences = append(tmpSentences, sentence)
	}
	sentences = tmpSentences
	// TODO tts多线程生成 加速生成速度
	for _, sentence := range sentences {
		//语音转换字幕 srt
		audioUrl, err := TTS(ctx, "", sentence)
		if err != nil {
			fmt.Println(err)
			return err
		}
		localAudios = append(localAudios, audioUrl)
	}
	if len(localAudios) == 0 {
		fmt.Println("localAudios is empty")
		return
	}
	// 合成tts音频
	combinedAudioUrl, err = CombinedAudioByFfmpeg(localAudios)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// 生成字幕
	subtitleUrl, err = GenerateSubtitlesLocally(ctx, sentences, localAudios, saveSubtitleDir)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// 合并视频
	combinedVideoUrl, err = CombinedVideo(ctx, localVideos)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// 融合语音和视频
	finalVideoUrl, err = MixAllInfoForVideo(ctx, combinedVideoUrl, combinedAudioUrl, subtitleUrl)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(finalVideoUrl)
	// 生成视频元数据信息(如标题/分类等)

	// 自动上传相关
	return
}
