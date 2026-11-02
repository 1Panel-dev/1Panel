package dockerhub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type SearchItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SearchResult struct {
	Items         []SearchItem `json:"items"`
	Total         int64        `json:"total"`
	ProxyRequired bool         `json:"proxyRequired"`
}

// Search uses the panel HTTP client, not the Docker daemon's proxy settings.
func Search(ctx context.Context, client *http.Client, query string, page, pageSize int) (SearchResult, error) {
	result := SearchResult{Items: []SearchItem{}}
	query = strings.TrimSpace(query)
	if query == "" || len([]rune(query)) > 256 || page < 1 || pageSize < 1 || pageSize > 100 {
		return result, fmt.Errorf("invalid Docker Hub search parameters")
	}
	params := url.Values{"query": {query}, "page": {strconv.Itoa(page)}, "page_size": {strconv.Itoa(pageSize)}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://hub.docker.com/v2/search/repositories/?"+params.Encode(), nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		// Transport errors can contain proxy credentials; do not expose their raw text.
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		return result, fmt.Errorf("Docker Hub connection failed; check network and proxy settings")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("Docker Hub search: HTTP %d", resp.StatusCode)
	}
	var body struct {
		Count   *int64 `json:"count"`
		Results []struct {
			Name        string `json:"repo_name"`
			Description string `json:"short_description"`
		} `json:"results"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&body); err != nil {
		return result, fmt.Errorf("invalid Docker Hub search response")
	}
	if body.Count == nil || *body.Count < 0 {
		return result, fmt.Errorf("invalid Docker Hub search response")
	}
	result.Total = *body.Count
	for _, item := range body.Results {
		if item.Name == "" {
			return result, fmt.Errorf("invalid Docker Hub repository name")
		}
		result.Items = append(result.Items, SearchItem{Name: item.Name, Description: item.Description})
	}
	return result, nil
}
