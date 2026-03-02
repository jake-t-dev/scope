package github

import (
	"context"

	"github.com/google/go-github/v84/github"
)

type Client interface {
	GetUserInterests(ctx context.Context, username string) (map[string]int, error)
}

type client struct {
	gh *github.Client
}

func NewClient() Client {
	return &client{
		gh: github.NewClient(nil),
	}
}

func (c *client) GetUserInterests(ctx context.Context, username string) (map[string]int, error) {
	opt := &github.ActivityListStarredOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	starredRepos, _, err := c.gh.Activity.ListStarred(ctx, username, opt)
	if err != nil {
		return nil, err
	}

	interests := make(map[string]int)
	for _, result := range starredRepos {
		repo := result.Repository
		if repo != nil {
			for _, topic := range repo.Topics {
				interests[topic]++
			}
			if repo.Language != nil && *repo.Language != "" {
				interests[*repo.Language]++
			}
		}
	}

	// Also fetch repositories the user owns or contributes to
	repoOpt := &github.RepositoryListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	userRepos, _, err := c.gh.Repositories.List(ctx, username, repoOpt)
	if err != nil {
		// Log error but continue with starred interests?
		// For now, let's just return the error or maybe log it.
		// Since the function signature returns error, let's return it.
		// However, partial data (starred) might be better than nothing.
		// But to keep it simple and robust, let's try to proceed or just return error.
		return nil, err
	}

	for _, repo := range userRepos {
		if repo != nil {
			for _, topic := range repo.Topics {
				interests[topic]++
			}
			if repo.Language != nil && *repo.Language != "" {
				interests[*repo.Language]++
			}
		}
	}

	return interests, nil
}
