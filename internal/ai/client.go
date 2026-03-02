package ai

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Client interface {
	GenerateSearchQuery(ctx context.Context, interests map[string]int) (string, error)
}

type client struct {
	model *genai.GenerativeModel
}

func NewClient(apiKey string) (Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}

	ctx := context.Background()
	c, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}

	// Use gemini-flash-latest which points to the latest stable flash model (currently 1.5)
	// This model is generally available on the free tier.
	model := c.GenerativeModel("gemini-flash-latest")
	model.SetTemperature(0.3) // Lower temperature for more deterministic output

	return &client{model: model}, nil
}

func (c *client) GenerateSearchQuery(ctx context.Context, interests map[string]int) (string, error) {
	if len(interests) == 0 {
		return "technology", nil
	}

	// 1. Sort interests by count (descending) to prioritize top interests
	type kv struct {
		Key   string
		Value int
	}
	var ss []kv
	for k, v := range interests {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	// 2. Take top 10 interests to form the context
	var topInterests []string
	count := 0
	for _, kv := range ss {
		if count >= 10 {
			break
		}
		topInterests = append(topInterests, kv.Key)
		count++
	}

	// 3. Construct the prompt
	prompt := fmt.Sprintf(`
You are an expert search query optimizer for a technical news API.
The user has the following technical interests (derived from GitHub activity):
%s

Task: Generate a SINGLE, optimized search query string suitable for the NewsAPI 'q' parameter.
Rules:
1. Focus on the most significant technical terms (e.g., specific languages, frameworks like "Go", "React", "Kubernetes").
2. Disambiguate common terms (e.g., use "golang" or "go programming" instead of just "go", "rust programming" for "rust").
3. Combine terms using OR operators to broaden the search.
4. Append a safeguard to ensure technical relevance, like " AND (technology OR programming OR software)".
5. The query must be concise (under 500 chars if possible).
6. Return ONLY the raw query string. No markdown, no explanations, no quotes.

Example Output:
(golang OR "rust lang" OR kubernetes) AND (technology OR software)
`, strings.Join(topInterests, ", "))

	// 4. Call Gemini
	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from AI model")
	}

	// 5. Extract text
	part := resp.Candidates[0].Content.Parts[0]
	textPart, ok := part.(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response content type")
	}

	query := string(textPart)
	return strings.TrimSpace(query), nil
}
