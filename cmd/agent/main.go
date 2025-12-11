package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"

	"github.com/VolodyaPopov923/AIBot/config"
	"github.com/VolodyaPopov923/AIBot/internal/agent"
	"github.com/VolodyaPopov923/AIBot/internal/ai"
	"github.com/VolodyaPopov923/AIBot/internal/browser"
)

func main() {
	_ = godotenv.Load()

	ctx := context.Background()

	fmt.Println("🚀 Initializing browser...")
	browserMgr, err := browser.NewManager(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize browser: %v\n", err)
	}
	defer browserMgr.Close(ctx)

	fmt.Println("🤖 Initializing AI client...")
	cfg := config.LoadConfig()
	if cfg.OpenAIAPIKey == "" {
		log.Fatal("OPENAI_API_KEY not available")
	}
	aiClient := ai.NewClient(cfg.OpenAIAPIKey)

	agentInstance := agent.NewAgent(browserMgr, aiClient, true)

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("AI Browser Automation Agent")
	fmt.Println("You can:")
	fmt.Println("  - Type natural language requests (e.g., 'зайди на яндекс карты и найди кремль')")
	fmt.Println("  - Use commands: task <URL> <description>, go <URL>, exit")
	fmt.Println(strings.Repeat("=", 60))

	for {
		fmt.Print("\n> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input: %v\n", err)
			continue
		}
		input = strings.TrimSpace(input)
		input = strings.Trim(input, "\r\n")
		input = strings.TrimPrefix(input, ">")
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}
		command := strings.ToLower(parts[0])

		switch command {
		case "exit", "quit":
			fmt.Println("Goodbye!")
			return

		case "task":
			if len(parts) < 3 {
				fmt.Println("Usage: task <URL> <description>")
				continue
			}
			url := parts[1]
			taskDesc := strings.Join(parts[2:], " ")

			fmt.Printf("\n📋 Executing task: %s\n", taskDesc)
			if err := agentInstance.ExecuteTask(ctx, taskDesc, url); err != nil {
				fmt.Printf("❌ Task failed: %v\n", err)
			} else {
				fmt.Println("✅ Task completed successfully!")
			}

		case "go":
			if len(parts) < 2 {
				fmt.Println("Usage: go <URL>")
				continue
			}
			url := parts[1]
			fmt.Printf("🌐 Navigating to %s...\n", url)
			if err := browserMgr.Navigate(ctx, url); err != nil {
				fmt.Printf("❌ Navigation failed: %v\n", err)
			} else {
				fmt.Println("✅ Navigation successful!")
			}

		default:
			fmt.Printf("🤔 Parsing your request: %s\n", input)
			parsed, err := aiClient.ParseUserRequest(ctx, input)
			if err != nil {
				fmt.Printf("❌ Failed to parse request: %v\n", err)
				continue
			}

			if parsed.NeedsURL && parsed.URL != "" {
				fmt.Printf("🌐 Opening: %s\n", parsed.URL)
				if err := browserMgr.Navigate(ctx, parsed.URL); err != nil {
					if !strings.Contains(err.Error(), "page closed") {
						fmt.Printf("❌ Navigation failed: %v\n", err)
						continue
					}
					fmt.Printf("⚠️  Page closed during navigation (possibly CAPTCHA) - continuing...\n")
				}
				_ = browserMgr.WaitForNavigation(ctx)
			}

			if parsed.Task != "" {
				url := parsed.URL
				if url == "" {
					pageContent, _ := browserMgr.GetPageContent(ctx)
					url = pageContent.URL
				}
				fmt.Printf("📋 Executing task: %s\n", parsed.Task)
				if err := agentInstance.ExecuteTask(ctx, parsed.Task, url); err != nil {
					fmt.Printf("❌ Task failed: %v\n", err)
				} else {
					fmt.Println("✅ Task completed successfully!")
				}
			} else {
				fmt.Printf("ℹ️  %s\n", parsed.Reasoning)
			}
		}
	}
}
