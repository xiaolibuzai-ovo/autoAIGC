package main

import (
	"context"
	"fmt"
	"github.com/bytedance/sonic"
	"github.com/sashabaranov/go-openai"
)

/*
GenerateSubjectText 生成主题文字
aiModel: ai模型
subject: 生成内容的主题
language: 语言

return: 根据主题生成的内容
*/
func GenerateSubjectText(ctx context.Context, aiModel string, subject string, language string) (string, error) {
	client := GetOpenaiClient()
	request := openai.ChatCompletionRequest{
		Model:       aiModel,
		Temperature: 0.8,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				Content: fmt.Sprintf(`
			As a subject development master, you will need to expand your content around the topics I provide			

			Generate a text for a video, depending on the subject of the video.

            The text is to be returned as a string with the specified number of paragraphs.

            Here is an example of a string:
            "This is an example string."

            Do not under any circumstance reference this prompt in your response.

            Get straight to the point, don't start with unnecessary things like, "welcome to this video".

            Obviously, the text should be related to the subject 

            YOU MUST NOT INCLUDE ANY TYPE OF MARKDOWN OR FORMATTING IN THE TEXT, NEVER USE A TITLE.
            YOU MUST WRITE THE TEXT IN THE LANGUAGE SPECIFIED IN [LANGUAGE].
            ONLY RETURN THE RAW CONTENT OF THE TEXT. DO NOT INCLUDE "VOICEOVER", "NARRATOR" OR SIMILAR INDICATORS OF WHAT SHOULD BE SPOKEN AT THE BEGINNING OF EACH PARAGRAPH OR LINE. YOU MUST NOT MENTION THE PROMPT, OR ANYTHING ABOUT THE TEXT ITSELF. ALSO, NEVER TALK ABOUT THE AMOUNT OF PARAGRAPHS OR LINES. JUST WRITE THE TEXT.
			
			SUBJECT is %s
			LANGUAGE is %s
`, subject, language),
			},
		},
	}
	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", err
	} else if len(response.Choices) < 1 {
		return "", fmt.Errorf("no response")
	}
	text := response.Choices[0].Message.Content
	return text, nil
}

/*
GenerateSearchTermsBySubject 根据主题文字生成视频搜索词
aiModel: ai模型
amount: 搜索词数量
subject: 主题
subjectText: 生成的主题内容

return: 主题视频搜索词slices
*/
func GenerateSearchTermsBySubject(ctx context.Context, aiModel string, amount int64, subject string, subjectText string) ([]string, error) {
	client := GetOpenaiClient()
	request := openai.ChatCompletionRequest{
		Model:       aiModel,
		Temperature: 0.8,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleUser,
				Content: fmt.Sprintf(`
			Generate {%d} search terms for stock videos,
			depending on the subject of a video.
			Subject: {%s}
		
			The search terms are to be returned as
			a JSON-Array of strings.
		
			Each search term should consist of 1-3 words,
			always add the main subject of the video.
			
			YOU MUST ONLY RETURN THE JSON-ARRAY OF texts.
			YOU MUST NOT RETURN ANYTHING ELSE. 
			YOU MUST NOT RETURN THE text.
			
			The search terms must be related to the subject of the video.
			Here is an example of a JSON-Array of strings:
			["search term 1", "search term 2", "search term 3"]
		
			For context, here is the full text:
			{%s}
		`, amount, subject, subjectText),
			},
		},
	}
	response, err := client.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, err
	} else if len(response.Choices) < 1 {
		return nil, fmt.Errorf("no response")
	}

	var (
		tmpSearchTerms []string
		searchTerms    []string
	)
	err = sonic.Unmarshal([]byte(response.Choices[0].Message.Content), &tmpSearchTerms)
	if err != nil {
		return nil, err
	}
	for _, term := range tmpSearchTerms {
		if !ContainsString(term, searchTerms) {
			searchTerms = append(searchTerms, term)
		}
	}
	return searchTerms, nil
}
