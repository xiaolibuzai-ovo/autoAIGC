package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// 获取音频时长，结果为秒
func GetWavDuration(filePath string) float64 {
	// 格式转换 ffmpeg -i xxxx  2>&1 | grep 'Duration' | cut -d ' ' -f 4 | sed s/,//
	cmd := fmt.Sprintf("ffmpeg -i %s 2>&1 | grep 'Duration' | cut -d ' ' -f 4 | sed s/,//", filePath)
	command := exec.Command("/bin/bash", "-c", cmd)
	res, err := command.CombinedOutput()
	if err != nil {

		return 0
	}
	body := string(res)
	fmt.Println(body)
	timeArr := strings.Split(body, ",")
	if len(timeArr) != 3 {
		return 0
	}
	// 计算时长，转为秒
	hour, err := strconv.ParseFloat(timeArr[0], 32)
	if err != nil {
		return 0
	}
	min, err := strconv.ParseFloat(timeArr[1], 32)
	if err != nil {
		return 0
	}
	second, err := strconv.ParseFloat(timeArr[2], 32)
	if err != nil {
		return 0
	}

	return 3600*hour + 60*min + second
}

func Test00(t *testing.T) {
}

func Test111(t *testing.T) {
	float, _ := strconv.ParseFloat("0."+"13", 64)
	fmt.Println(float)
}
