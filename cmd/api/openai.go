package main

import (
	"context"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func askOpenAI(ctx context.Context, prompt string, fallback []string) []string {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fallback
	}
	client := openai.NewClient(apiKey)
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		}},
		MaxTokens: 300,
	})
	if err != nil || len(resp.Choices) == 0 {
		return fallback
	}
	return []string{resp.Choices[0].Message.Content}
}
