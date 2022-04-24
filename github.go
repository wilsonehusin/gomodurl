package gomodurl

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"

	"github.com/rs/xid"
)

type GitHubSource struct {
	Host   string `json:"host"`
	Owner  string `json:"owner"`
	Vanity string `json:"vanity"`
}

func ParseGitHubSource(b []byte) (*GitHubSource, error) {
	s := &GitHubSource{Host: "https://api.github.com"}
	if err := json.Unmarshal(b, s); err != nil {
		return nil, err
	}

	s.Vanity = strings.TrimSuffix(s.Vanity, "/")

	return s, nil
}

type GithubRepositoryResponse struct {
	Name   string `json:"name"`
	Branch string `json:"default_branch"`
	URL    string `json:"html_url"`
}

var githubRepositoryListURL = template.Must(
	template.New("url").Parse(`{{ .Host }}/users/{{ .Owner }}/repos`))

func (s *GitHubSource) Repositories(ctx context.Context, c *http.Client) ([]*GoPackage, error) {
	var urlStringer strings.Builder
	if err := githubRepositoryListURL.Execute(&urlStringer, s); err != nil {
		return nil, fmt.Errorf("generating url: %w", err)
	}

	urlStr := urlStringer.String()
	baseReq, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("generating request: %w", err)
	}

	packages := []*GoPackage{}

	page := 0
	query := url.Values{}
	query.Set("per_page", "100")
	for {
		id := xid.New().String()

		page++
		query.Set("page", strconv.Itoa(page))

		req := baseReq.Clone(ctx)
		req.URL.RawQuery = query.Encode()
		logger.Printf("[%s] outgoing %s: %s", id, req.Method, req.URL.String())

		resp, err := c.Do(req)
		if err != nil {
			logger.Printf("[%s] error: %v", id, err.Error())
			break
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			logger.Printf("[%s] received HTTP %d", id, resp.StatusCode)
			break
		}

		result := []GithubRepositoryResponse{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			logger.Printf("[%s] error parsing json: %v", id, err.Error())
			break
		}

		if len(result) == 0 {
			break
		}

		for _, r := range result {
			display := fmt.Sprintf("%v %v/tree/master{/dir} %v/blob/master{/dir}/{file}#L{line}", r.URL, r.URL, r.URL)
			packages = append(packages, &GoPackage{
				Import:         s.Vanity + "/" + r.Name,
				VersionControl: "git",
				Repository:     r.URL,
				Display:        display,
				Host:           s.Vanity,
				Branch:         r.Branch,
			})
		}

		if len(result) < 99 {
			break
		}
	}

	return packages, nil
}
