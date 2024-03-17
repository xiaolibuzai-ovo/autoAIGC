package main

import (
	"context"
	"fmt"
	"testing"
)

func TestMergeVideo(t *testing.T) {
	var videos = []string{"a.mp4", "b.mp4"}
	mergeVideoUrl, err := MergeVideo(context.Background(), videos)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(mergeVideoUrl)
	println("Videos merged successfully.")
}
