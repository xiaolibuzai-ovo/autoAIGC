package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type PexelsVideoResponse struct {
	Page         int    `json:"page"`
	PerPage      int    `json:"per_page"`
	TotalResults int    `json:"total_results"`
	Url          string `json:"url"`
	Videos       []struct {
		Id       int    `json:"id"`
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		Url      string `json:"url"`
		Image    string `json:"image"`
		Duration int    `json:"duration"`
		User     struct {
			Id   int    `json:"id"`
			Name string `json:"name"`
			Url  string `json:"url"`
		} `json:"user"`
		VideoFiles []struct {
			Id       int    `json:"id"`
			Quality  string `json:"quality"`
			FileType string `json:"file_type"`
			Width    int    `json:"width"`
			Height   int    `json:"height"`
			Link     string `json:"link"`
		} `json:"video_files"`
		VideoPictures []struct {
			Id      int    `json:"id"`
			Picture string `json:"picture"`
			Nr      int    `json:"nr"`
		} `json:"video_pictures"`
	} `json:"videos"`
}

/*
SearchVideosInPexels 检索视频

term: 检索条件
limit: 检索数量
minDuration: 视频最短时间

return
videoUrls: 检索到的视频url slices
*/
func SearchVideosInPexels(ctx context.Context, term string, limit int, minDuration int) (videoUrls []string, err error) {
	var (
		param  = url.Values{}
		header = make(http.Header)

		response   PexelsVideoResponse
		videoRatio int
	)
	param.Add("query", term)
	param.Add("per_page", fmt.Sprint(limit))
	token := os.Getenv("PEXELS_API_KEY")
	if len(token) < 0 {
		return nil, fmt.Errorf("PEXELS_API_KEY not set")
	}
	header.Add("Authorization", token)
	request, err := BuildRequest(ctx, http.MethodGet, "https://api.pexels.com/videos/search", nil, header, param)
	if err != nil {
		return
	}
	err = SendRequest(request, &http.Client{}, &response)
	if err != nil {
		return
	}

	for _, video := range response.Videos {
		if video.Duration < minDuration {
			// 过滤不满足时长的视频
			continue
		}
		tmpLinkUrl := ""
		for _, linkFile := range video.VideoFiles {
			// 检查是否满足下赞路径
			if !strings.Contains(linkFile.Link, ".com/external") {
				continue
			}
			// 只保留分辨率更高的视频
			if linkFile.Width*linkFile.Height < videoRatio {
				continue
			}

			tmpLinkUrl = linkFile.Link
			videoRatio = linkFile.Width * linkFile.Height

			if tmpLinkUrl != "" {
				videoUrls = append(videoUrls, tmpLinkUrl)
			}
		}

	}
	return
}

/*
CombinedVideo 使用ffmpeg将多个视频合成为一个

videos: 要合成的视频的本地地址

return
combinedVideoUrl 合成后视频的本地地址

example:
ffmpeg -i a.mp4 -c copy -bsf:v h264_mp4toannexb -f mpegts 1.ts
ffmpeg -i b.mp4 -c copy -bsf:v h264_mp4toannexb -f mpegts 2.ts
ffmpeg -i "concat:1.ts|2.ts" -c copy -bsf:a aac_adtstoasc -movflags +faststart ts.mp4
*/
func CombinedVideo(ctx context.Context, videos []string) (combinedVideoUrl string, err error) {
	//mp4->ts merge ts ts->mp4
	var (
		tsDir            = "./tmp/ts/"
		combinedVideoDir = "./tmp/combinedVideo/"

		tsList []string
	)
	for _, video := range videos {
		tmpTs := fmt.Sprintf("%s%d.ts", tsDir, time.Now().UnixNano())
		err = os.MkdirAll(filepath.Dir(tmpTs), os.ModePerm)
		if err != nil {
			return
		}
		// mp4转ts
		cmd := fmt.Sprintf("ffmpeg -i %s -c copy -bsf:v h264_mp4toannexb -f mpegts %s", video, tmpTs)
		command := exec.Command("/bin/bash", "-c", cmd)
		_, err = command.CombinedOutput()
		if err != nil {
			return
		}
		tsList = append(tsList, tmpTs)
	}

	// 合并所有的ts 并转为mp4
	allTs := strings.Join(tsList, "|")
	// mp4转ts
	combinedVideoUrl = fmt.Sprintf("%s%d.mp4", combinedVideoDir, time.Now().UnixNano())
	err = os.MkdirAll(filepath.Dir(combinedVideoUrl), os.ModePerm)
	if err != nil {
		return
	}
	cmd := fmt.Sprintf(`ffmpeg -i "concat:%s" -c copy -bsf:a aac_adtstoasc -movflags +faststart %s`, allTs, combinedVideoUrl)
	command := exec.Command("/bin/bash", "-c", cmd)
	_, err = command.CombinedOutput()
	if err != nil {
		return
	}
	return
}

/*
MixAllInfoForVideo 将音频字幕视频融合

CombinedVideo 合并视频路径
mergeAudio 合并音频路径
subtitle 字幕路径

return
finalVideo 最终视频
*/
func MixAllInfoForVideo(ctx context.Context, CombinedVideo string, mergeAudio string, subtitle string) (finalVideo string, err error) {
	var (
		tmpVideoWithAudio = "./tmp/videoWithAudio/"
		finalVideoPath    = "./tmp/finalVideo/"
	)
	// 合并视频和音频
	CombinedVideoAndAudio := fmt.Sprintf("%s%d.mp4", tmpVideoWithAudio, time.Now().UnixNano())
	err = os.MkdirAll(filepath.Dir(CombinedVideoAndAudio), os.ModePerm)
	if err != nil {
		return
	}
	cmd := fmt.Sprintf(`ffmpeg -i %s -i %s -vcodec copy -acodec copy %s`, CombinedVideo, mergeAudio, CombinedVideoAndAudio)
	command := exec.Command("/bin/bash", "-c", cmd)
	_, err = command.CombinedOutput()
	if err != nil {
		return
	}
	// 添加字幕
	finalVideo = fmt.Sprintf("%s%d.mp4", finalVideoPath, time.Now().UnixNano())
	err = os.MkdirAll(filepath.Dir(finalVideo), os.ModePerm)
	if err != nil {
		return
	}
	cmd = fmt.Sprintf(`ffmpeg -i %s -strict -2 -vf \
subtitles=%s:force_style='Fontsize=15\,FontName=FZYBKSJW--GB1-0' -qscale:v 3 %s`, CombinedVideoAndAudio, subtitle, finalVideo)
	command = exec.Command("/bin/bash", "-c", cmd)
	_, err = command.CombinedOutput()
	if err != nil {
		return
	}
	return
}
