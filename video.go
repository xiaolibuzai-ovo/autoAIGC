package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
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
