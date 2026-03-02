package news

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"scope/internal/ai"
)

type NewsArticle struct {
	Source      NewsSource `json:"source"`
	Author      string     `json:"author"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	URL         string     `json:"url"`
	URLToImage  string     `json:"urlToImage"`
	PublishedAt time.Time  `json:"publishedAt"`
	Content     string     `json:"content"`
}

type NewsSource struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type newsAPIResponse struct {
	Status       string        `json:"status"`
	TotalResults int           `json:"totalResults"`
	Articles     []NewsArticle `json:"articles"`
}

type Client interface {
	GetTailoredNews(ctx context.Context, interests map[string]int) ([]NewsArticle, error)
}

type client struct {
	apiKey string
	http   *http.Client
	ai     ai.Client
}

func NewClient(apiKey string, aiClient ai.Client) Client {
	return &client{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 10 * time.Second},
		ai:     aiClient,
	}
}

func (c *client) GetTailoredNews(ctx context.Context, interests map[string]int) ([]NewsArticle, error) {
	var query string
	var err error

	if c.ai != nil {
		query, err = c.ai.GenerateSearchQuery(ctx, interests)
		if err != nil {
			fmt.Printf("AI search query generation failed: %v. Falling back to keyword search.\n", err)
			query = "" // Clear query to trigger fallback
		}
	}

	// Fallback to manual query construction if AI fails or is unavailable
	if query == "" {
		if len(interests) == 0 {
			query = "technology"
		} else {
			// Convert map to slice for sorting
			type kv struct {
				Key   string
				Value int
			}
			var ss []kv
			for k, v := range interests {
				ss = append(ss, kv{k, v})
			}

			// Sort by count descending
			sort.Slice(ss, func(i, j int) bool {
				return ss[i].Value > ss[j].Value
			})

			// Take top 5 interests
			var topInterests []string
			for i, kv := range ss {
				if i >= 5 {
					break
				}
				
				key := kv.Key
				var term string
				
				// Handle ambiguous programming terms
				switch strings.ToLower(key) {
				case "go":
					term = "(golang OR \"go programming\")"
				case "c":
					term = "\"c programming\""
				case "rust":
					term = "(rust AND programming)"
				default:
					// Wrap in quotes to handle multi-word interests correctly
					term = fmt.Sprintf("%q", key)
				}
				
				topInterests = append(topInterests, term)
			}
			query = fmt.Sprintf("(%s) AND technology", strings.Join(topInterests, " OR "))
		}
	}

	baseURL := "https://newsapi.org/v2/everything"
	params := url.Values{}
	params.Add("apiKey", c.apiKey)
	params.Add("q", query)
	params.Add("sortBy", "relevancy")
	params.Add("language", "en")
	params.Add("pageSize", "5") // Limit to 5 articles

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s?%s", baseURL, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch news: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("news api returned non-200 status: %d", resp.StatusCode)
	}

	var result newsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Articles, nil
}
