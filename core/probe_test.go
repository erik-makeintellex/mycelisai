package main

import (
	"context"
	"fmt"
	"log"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	config := openai.DefaultConfig("ollama")
	config.BaseURL = "http://127.0.0.1:11434/v1"

	fmt.Printf("Testing connection to: %s\n", config.BaseURL)

	client := openai.NewClientWithConfig(config)

	models, err := client.ListModels(context.Background())
	if err != nil {
		log.Fatalf("Probe failed: %v", err)
	}

	fmt.Printf("Success! Found %d models.\n", len(models.Models))
	for _, m := range models.Models {
		fmt.Printf("- %s\n", m.ID)
	}
}
