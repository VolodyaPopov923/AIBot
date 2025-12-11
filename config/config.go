package config

import (
	"os"
	"strconv"
)

const testOpenAIKey = ""

type Config struct {
	OpenAIAPIKey  string
	BrowserPath   string
	Debug         bool
	MaxTokens     int
	MaxIterations int
}

func LoadConfig() Config {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = testOpenAIKey
	}

	return Config{
		OpenAIAPIKey:  apiKey,
		BrowserPath:   os.Getenv("BROWSER_PATH"),
		Debug:         debug,
		MaxTokens:     8000,
		MaxIterations: 20,
	}
}
