package main

import (
	"fmt"
	"testing"
	"time"
)

func TestSaveFileLocal(t *testing.T) {
	local, err := SaveFileLocal("https://player.vimeo.com/external/291648067.sd.mp4?s=7f9ee1f8ec1e5376027e4a6d1d05d5738b2fbb29&profile_id=164&oauth2_token_id=57447761", "./tmp/", fmt.Sprintf("%d.mp4", time.Now().UnixNano()))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(local)
}

func TestCreateTxtFileWithDynamicContent(t *testing.T) {
	// 创建带有时间戳的文本文件并写入内容
	path, err := CreateTxtFileWithDynamicContent([]string{"file 'a.mp3'", "file 'b.mp3'"})
	if err != nil {
		panic(err)
	}
	// 输出文件的路径
	fmt.Printf("文件已创建于: %s\n", path)
}
