package main

import (
	"github.com/sashabaranov/go-openai"
	"log"
	"os"
)

var openaiClient *openai.Client

func GetOpenaiClient() *openai.Client {
	return openaiClient
}

func initGptClient() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if len(apiKey) <= 0 {
		log.Panic("OPENAI_API_KEY is not set")
	}
	openaiClient = openai.NewClient(apiKey)
}
