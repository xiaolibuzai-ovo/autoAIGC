package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	initOpenaiClient()
	err := GenerateVideo(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
}
