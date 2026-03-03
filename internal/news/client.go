package news

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"regexp"
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

// XML Structs for Google News RSS
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Items       []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	PubDate     string `xml:"pubDate"`
	Description string `xml:"description"`
	Source      Source `xml:"source"`
	GUID        string `xml:"guid"`
}

type Source struct {
	Name string `xml:",chardata"`
	URL  string `xml:"url,attr"`
}

type Client interface {
	GetTailoredNews(ctx context.Context, interests map[string]int) ([]NewsArticle, error)
}

type client struct {
	http *http.Client
	ai   ai.Client
}

func NewClient(aiClient ai.Client) Client {
	return &client{
		http: &http.Client{Timeout: 10 * time.Second},
		ai:   aiClient,
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

	// Google News RSS URL
	baseURL := "https://news.google.com/rss/search"

	// Filter by last month (using 'after:YYYY-MM-DD' syntax supported by Google News)
	oneMonthAgo := time.Now().AddDate(0, -1, 0).Format("2006-01-02")
	finalQuery := fmt.Sprintf("%s after:%s", query, oneMonthAgo)

	params := url.Values{}
	params.Add("q", finalQuery)
	params.Add("hl", "en-US")
	params.Add("gl", "US")
	params.Add("ceid", "US:en")

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
		return nil, fmt.Errorf("google news rss returned non-200 status: %d", resp.StatusCode)
	}

	var rss RSS
	if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
		return nil, fmt.Errorf("failed to decode xml response: %w", err)
	}

	// Convert RSS items to NewsArticle
	var articles []NewsArticle

	// Limit to 5 articles
	maxArticles := 5
	for i, item := range rss.Channel.Items {
		if i >= maxArticles {
			break
		}

		pubDate, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 GMT", item.PubDate)
		if pubDate.IsZero() {
			// Try without GMT or other formats if needed, but Google usually sends GMT
			pubDate = time.Now()
		}

		articles = append(articles, NewsArticle{
			Source: NewsSource{
				ID:   "google-news",
				Name: item.Source.Name,
			},
			Author:      item.Source.Name,
			Title:       item.Title,
			Description: cleanDescription(item.Description),
			URL:         item.Link,
			URLToImage:  "", // Not available in standard RSS
			PublishedAt: pubDate,
			Content:     "", // Not available in standard RSS
		})
	}

	return articles, nil
}

// cleanDescription removes HTML tags and decodes entities
func cleanDescription(htmlStr string) string {
	// Simple regex to strip tags
	re := regexp.MustCompile(`<[^>]*>`)
	stripped := re.ReplaceAllString(htmlStr, "")
	return html.UnescapeString(stripped)
}
