package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
)

func ContainsString(v string, vv []string) bool {
	for _, s := range vv {
		if v == s {
			return true
		}
	}
	return false
}

func RemoveRepetitionString(vv []string) []string {
	var (
		result    []string
		recordMap = make(map[string]bool)
	)
	for _, v := range vv {
		if recordMap[v] {
			continue
		}
		recordMap[v] = true
		result = append(result, v)
	}
	return result
}

func SendRequest(req *http.Request, client *http.Client, v any) error {
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return handleErrorResp(res)
	}

	if v == nil {
		return nil
	}
	if result, ok := v.(*string); ok {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		*result = string(b)
		return nil
	}
	return json.NewDecoder(res.Body).Decode(v)
}

func BuildRequest(
	ctx context.Context,
	method string,
	url string,
	body any,
	header http.Header,
	params url.Values,
) (req *http.Request, err error) {
	var bodyReader io.Reader
	if body != nil {
		if v, ok := body.(io.Reader); ok {
			bodyReader = v
		} else {
			var reqBytes []byte
			reqBytes, err = json.Marshal(body)
			if err != nil {
				return
			}
			bodyReader = bytes.NewBuffer(reqBytes)
		}
	}
	if params != nil && len(params) > 0 {
		url += "?" + params.Encode()
	}
	req, err = http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return
	}
	if header != nil {
		req.Header = header
	}
	return
}

type errRes struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func handleErrorResp(resp *http.Response) error {
	var apiError errRes
	err := json.NewDecoder(resp.Body).Decode(&apiError)
	if err != nil {
		return err
	}
	apiError = errRes{
		Code:    fmt.Sprint(http.StatusInternalServerError),
		Message: "http fail",
	}
	return fmt.Errorf("%v", apiError)
}

func SaveFileLocal(fileUrl string, directory string, fileName string) (string, error) {
	// Create full video path
	videoPath := filepath.Join(directory, fileName)

	// Make GET request to fetch video content
	response, err := http.Get(fileUrl)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	// Read video content
	videoContent, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	// Create video file and write content
	err = ioutil.WriteFile(videoPath, videoContent, 0644)
	if err != nil {
		return "", err
	}

	return videoPath, nil
}
