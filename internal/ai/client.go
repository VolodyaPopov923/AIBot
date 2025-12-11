package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Client struct {
	openaiClient *openai.Client
	model        string
	maxTokens    int
}

func NewClient(apiKey string) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	return &Client{
		openaiClient: openai.NewClient(apiKey),
		model:        "gpt-4-turbo-preview",
		maxTokens:    3000,
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type DecisionResponse struct {
	Action       string `json:"action"`
	Selector     string `json:"selector,omitempty"`
	Text         string `json:"text,omitempty"`
	URL          string `json:"url,omitempty"`
	Reasoning    string `json:"reasoning"`
	IsComplete   bool   `json:"is_complete"`
	NextStep     string `json:"next_step,omitempty"`
	NeedsConfirm bool   `json:"needs_confirm"`
}

type UserRequestParsed struct {
	Task      string `json:"task"`
	URL       string `json:"url,omitempty"`
	NeedsURL  bool   `json:"needs_url"`
	Reasoning string `json:"reasoning"`
}

func (c *Client) MakeDecision(ctx context.Context, systemPrompt, userInput string) (DecisionResponse, error) {
	resp, err := c.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       c.model,
		Temperature: 0.7,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userInput},
		},
	})
	if err != nil {
		return DecisionResponse{}, fmt.Errorf("failed to call OpenAI: %w", err)
	}
	if len(resp.Choices) == 0 {
		return DecisionResponse{}, fmt.Errorf("empty response from OpenAI")
	}

	raw := resp.Choices[0].Message.Content
	content := strings.TrimSpace(raw)
	if strings.HasPrefix(content, "```") {
		parts := strings.SplitN(content, "\n", 2)
		if len(parts) == 2 {
			content = strings.TrimSpace(parts[1])
			if idx := strings.LastIndex(content, "```"); idx != -1 {
				content = strings.TrimSpace(content[:idx])
			}
		}
	}

	var decision DecisionResponse
	if err := json.Unmarshal([]byte(content), &decision); err != nil {
		return DecisionResponse{
			Action:     "error",
			Reasoning:  raw,
			IsComplete: false,
		}, fmt.Errorf("failed to parse decision JSON: %w", err)
	}

	return decision, nil
}

func (c *Client) GetAnalysis(ctx context.Context, pageContent string, task string) (string, error) {
	condensed, err := c.CondenseForAnalysis(ctx, pageContent, task)
	if err != nil {
		return "", fmt.Errorf("failed to condense content: %w", err)
	}

	resp, err := c.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       c.model,
		Temperature: 0.7,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "You are an intelligent web automation agent."},
			{Role: openai.ChatMessageRoleUser, Content: fmt.Sprintf("Task: %s\n\nRelevant page content (condensed):\n%s", task, condensed)},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to call OpenAI: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response from OpenAI")
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *Client) CondenseForAnalysis(ctx context.Context, content string, task string) (string, error) {
	if approxTokens(content) <= c.maxTokens {
		return content, nil
	}

	chunkTokenLimit := int(float64(c.maxTokens) * 0.35)
	if chunkTokenLimit < 200 {
		chunkTokenLimit = 200
	}

	chunks := chunkTextByTokens(content, chunkTokenLimit)

	var summaries []string
	for _, ch := range chunks {
		prompt := fmt.Sprintf("Summarize the following page segment into concise bullets focused on the task '%s'. Keep only information useful for accomplishing the task.\n\nSegment:\n%s", task, ch)
		resp, err := c.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       c.model,
			Temperature: 0.0,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: "You are a concise summarizer that preserves task-relevant facts."},
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			MaxTokens: 400,
		})
		if err != nil {
			return "", fmt.Errorf("failed to summarize chunk: %w", err)
		}
		if len(resp.Choices) == 0 {
			continue
		}
		summaries = append(summaries, resp.Choices[0].Message.Content)
	}

	combined := strings.Join(summaries, "\n\n")
	if approxTokens(combined) > c.maxTokens {
		prompt := fmt.Sprintf("The following are summaries of segments from a page. Please further condense into a short list of facts strictly relevant to the task '%s'. Prioritize actionable information and key findings.\n\nSummaries:\n%s", task, combined)
		resp, err := c.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       c.model,
			Temperature: 0.0,
			Messages: []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleSystem, Content: "You are a concise summarizer that preserves task-relevant facts."},
				{Role: openai.ChatMessageRoleUser, Content: prompt},
			},
			MaxTokens: 600,
		})
		if err != nil {
			return "", fmt.Errorf("failed to summarize combined summaries: %w", err)
		}
		if len(resp.Choices) > 0 {
			combined = resp.Choices[0].Message.Content
		}
	}

	return combined, nil
}

func (c *Client) ParseUserRequest(ctx context.Context, userInput string) (UserRequestParsed, error) {
	systemPrompt := `You are a request parser for a web automation agent. Parse the user's request and extract:
1. Whether a URL is needed or should be extracted
2. The actual task to perform
3. Any URLs mentioned
4. Your reasoning

Respond as valid JSON with: {"task": "...", "url": "...", "needs_url": boolean, "reasoning": "..."}`

	resp, err := c.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       c.model,
		Temperature: 0.0,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: userInput},
		},
	})
	if err != nil {
		return UserRequestParsed{}, fmt.Errorf("failed to parse request: %w", err)
	}
	if len(resp.Choices) == 0 {
		return UserRequestParsed{}, fmt.Errorf("empty response from OpenAI")
	}

	raw := resp.Choices[0].Message.Content
	content := strings.TrimSpace(raw)
	if strings.HasPrefix(content, "```") {
		parts := strings.SplitN(content, "\n", 2)
		if len(parts) == 2 {
			content = strings.TrimSpace(parts[1])
			if idx := strings.LastIndex(content, "```"); idx != -1 {
				content = strings.TrimSpace(content[:idx])
			}
		}
	}

	var parsed UserRequestParsed
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return UserRequestParsed{
			Task:      userInput,
			Reasoning: "Could not parse, treating as direct task",
		}, nil
	}

	return parsed, nil
}

func (c *Client) PlanTask(ctx context.Context, task string, pageContext string) ([]string, error) {
	prompt := fmt.Sprintf(`You are a planner for a web automation agent.
Given the high-level task: "%s"
and the current page context (brief):
%s

Break the task into a concise, ordered list of concrete steps that an automated agent can perform in sequence. Each step should be a single short sentence or instruction. Return the result as a JSON array of strings only. Example:
["Open the images tab", "Click the first image", "Save image URL"]
`, task, pageContext)

	resp, err := c.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       c.model,
		Temperature: 0.0,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: "You convert user tasks into step-by-step actionable plans for a browser automation agent."},
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		MaxTokens: 800,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI for planning: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty planning response from OpenAI")
	}

	raw := strings.TrimSpace(resp.Choices[0].Message.Content)
	if strings.HasPrefix(raw, "```") {
		parts := strings.SplitN(raw, "\n", 2)
		if len(parts) == 2 {
			raw = strings.TrimSpace(parts[1])
			if idx := strings.LastIndex(raw, "```"); idx != -1 {
				raw = strings.TrimSpace(raw[:idx])
			}
		}
	}

	var steps []string
	if err := json.Unmarshal([]byte(raw), &steps); err != nil {
		lines := strings.Split(raw, "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" {
				continue
			}
			l = strings.TrimPrefix(l, "- ")
			l = strings.TrimPrefix(l, "*")
			if len(l) > 2 && l[1] == '.' && l[0] >= '0' && l[0] <= '9' {
				l = strings.TrimSpace(l[2:])
			}
			steps = append(steps, l)
		}
		if len(steps) == 0 {
			return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
		}
	}

	return steps, nil
}
