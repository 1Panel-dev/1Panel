// Package dockerhub checks Docker Hub metadata without downloading image layers.
package dockerhub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/distribution/reference"
)

const registryURL = "https://registry-1.docker.io"

// Reference only accepts tagged Docker Hub references. Digest pins are intentional.
func Reference(raw string) (string, string, bool) {
	if strings.HasPrefix(raw, "sha256:") {
		return "", "", false
	}
	ref, err := reference.ParseNormalizedNamed(raw)
	if err != nil || reference.Domain(ref) != "docker.io" {
		return "", "", false
	}
	if _, pinned := ref.(reference.Digested); pinned {
		return "", "", false
	}
	tagged := reference.TagNameOnly(ref).(reference.NamedTagged)
	return reference.Path(ref), tagged.Tag(), true
}

type Client struct {
	HTTP       *http.Client
	repository string
	token      string
}

type Platform struct{ OS, Architecture, Variant string }
type Target struct{ Tag, Digest, ConfigDigest string }

func New(ctx context.Context, repository, username, password string) (*Client, error) {
	c := &Client{HTTP: &http.Client{Timeout: 30 * time.Second}, repository: repository}
	endpoint := "https://auth.docker.io/token?service=registry.docker.io&scope=" + url.QueryEscape("repository:"+repository+":pull")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if username != "" {
		req.SetBasicAuth(username, password)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Docker Hub authentication: HTTP %d", resp.StatusCode)
	}
	var body struct {
		Token       string `json:"token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return nil, err
	}
	c.token = body.Token
	if c.token == "" {
		c.token = body.AccessToken
	}
	if c.token == "" {
		return nil, fmt.Errorf("Docker Hub returned an empty token")
	}
	return c, nil
}

func (c *Client) get(ctx context.Context, endpoint string, result interface{}) (http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.oci.image.index.v1+json, application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.v2+json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Docker Hub metadata: HTTP %d", resp.StatusCode)
	}
	err = json.NewDecoder(io.LimitReader(resp.Body, 16<<20)).Decode(result)
	return resp.Header, err
}

func (c *Client) Tags(ctx context.Context) ([]string, error) {
	endpoint := registryURL + "/v2/" + c.repository + "/tags/list?n=1000"
	var tags []string
	for page := 0; page < 100; page++ {
		var body struct {
			Tags []string `json:"tags"`
		}
		headers, err := c.get(ctx, endpoint, &body)
		if err != nil {
			return nil, err
		}
		tags = append(tags, body.Tags...)
		link := headers.Get("Link")
		if link == "" {
			return tags, nil
		}
		start, end := strings.Index(link, "<"), strings.Index(link, ">")
		if start < 0 || end <= start {
			return nil, fmt.Errorf("invalid Docker Hub pagination")
		}
		base, _ := url.Parse(registryURL)
		next, err := base.Parse(link[start+1 : end])
		if err != nil || next.Scheme != "https" || next.Host != base.Host || next.Path != "/v2/"+c.repository+"/tags/list" {
			return nil, fmt.Errorf("invalid Docker Hub pagination URL")
		}
		endpoint = next.String()
	}
	return nil, fmt.Errorf("Docker Hub tag pagination limit exceeded")
}

type manifest struct {
	Config struct {
		Digest string `json:"digest"`
	} `json:"config"`
	Manifests []struct {
		Digest   string `json:"digest"`
		Platform struct {
			OS           string `json:"os"`
			Architecture string `json:"architecture"`
			Variant      string `json:"variant"`
		} `json:"platform"`
	} `json:"manifests"`
}

func (c *Client) Resolve(ctx context.Context, tag string, platform Platform) (Target, bool, error) {
	var m manifest
	headers, err := c.get(ctx, registryURL+"/v2/"+c.repository+"/manifests/"+url.PathEscape(tag), &m)
	if err != nil {
		return Target{}, false, err
	}
	digest := headers.Get("Docker-Content-Digest")
	if len(m.Manifests) > 0 {
		digest = ""
		for _, item := range m.Manifests {
			if item.Platform.OS == platform.OS && item.Platform.Architecture == platform.Architecture && normalizedVariant(platform.Architecture, item.Platform.Variant) == normalizedVariant(platform.Architecture, platform.Variant) {
				digest = item.Digest
				break
			}
		}
		if digest == "" {
			return Target{}, false, nil
		}
		m = manifest{}
		if _, err := c.get(ctx, registryURL+"/v2/"+c.repository+"/manifests/"+url.PathEscape(digest), &m); err != nil {
			return Target{}, false, err
		}
	}
	if m.Config.Digest == "" {
		return Target{}, false, fmt.Errorf("unsupported Docker Hub manifest")
	}
	// Validate single-platform manifests too; the config digest is the Docker Image ID.
	var config struct {
		OS           string `json:"os"`
		Architecture string `json:"architecture"`
		Variant      string `json:"variant"`
	}
	if _, err := c.get(ctx, registryURL+"/v2/"+c.repository+"/blobs/"+url.PathEscape(m.Config.Digest), &config); err != nil {
		return Target{}, false, err
	}
	if config.OS != platform.OS || config.Architecture != platform.Architecture || (platform.Variant != "" && config.Variant != "" && config.Variant != platform.Variant) {
		return Target{}, false, nil
	}
	return Target{Tag: tag, Digest: digest, ConfigDigest: m.Config.Digest}, true, nil
}

var versionPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)(?:\.(\d+))?(-[a-zA-Z][a-zA-Z0-9_.-]*)?$`)
var prereleasePattern = regexp.MustCompile(`(?i)(^|[-_.])(alpha|beta|rc|dev|nightly|edge|test|unstable|preview)([0-9_.-]|$)`)

type version struct {
	numbers [3]uint64
	parts   int
	variant string
}

func parseVersion(tag string) (version, bool) {
	m := versionPattern.FindStringSubmatch(tag)
	if m == nil || prereleasePattern.MatchString(m[4]) {
		return version{}, false
	}
	v := version{parts: 2, variant: m[4]}
	for i := 0; i < 3; i++ {
		if m[i+1] != "" {
			n, err := strconv.ParseUint(m[i+1], 10, 64)
			if err != nil {
				return version{}, false
			}
			v.numbers[i] = n
		}
	}
	if m[3] != "" {
		v.parts = 3
	}
	return v, true
}
func greater(a, b version) bool {
	for i := 0; i < 3; i++ {
		if a.numbers[i] != b.numbers[i] {
			return a.numbers[i] > b.numbers[i]
		}
	}
	return false
}

// Candidates preserves the current tag's variant and precision. Unknown/floating
// tags are checked only for content changes, never guessed from upload dates.
func Candidates(current string, tags []string) []string {
	baseline, ok := parseVersion(current)
	if !ok {
		return []string{current}
	}
	type candidate struct {
		tag     string
		version version
	}
	candidates := []candidate{}
	seen := map[string]bool{}
	for _, tag := range tags {
		v, valid := parseVersion(tag)
		if valid && v.parts == baseline.parts && v.variant == baseline.variant && greater(v, baseline) && !seen[tag] {
			candidates = append(candidates, candidate{tag: tag, version: v})
			seen[tag] = true
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.version.numbers == b.version.numbers {
			return a.tag < b.tag
		}
		return greater(a.version, b.version)
	})
	result := make([]string, 0, len(candidates)+1)
	for _, candidate := range candidates {
		result = append(result, candidate.tag)
	}

	return append(result, current)
}

func NeedsTags(tag string) bool { _, ok := parseVersion(tag); return ok }

func normalizedVariant(architecture, variant string) string {
	if architecture == "arm64" && variant == "" {
		return "v8"
	}
	return variant
}
