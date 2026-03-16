package updater

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CheckLatestRelease queries the GitHub Releases API and returns the latest
// tag name (e.g. "v0.5.0"). Returns ("", err) on any failure — callers should
// treat errors as "no update available" and not surface them to the user.
func CheckLatestRelease(owner, repo string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.TagName == "" {
		return "", fmt.Errorf("tag_name is empty")
	}
	return payload.TagName, nil
}

// IsNewer reports whether latest is a higher semver than current.
// Leading "v" prefixes are stripped before comparison.
func IsNewer(current, latest string) bool {
	cv := parseSemver(current)
	lv := parseSemver(latest)
	for i := 0; i < 3; i++ {
		if lv[i] > cv[i] {
			return true
		}
		if lv[i] < cv[i] {
			return false
		}
	}
	return false // equal
}

// parseSemver splits a "vX.Y.Z" or "X.Y.Z" string into [major, minor, patch].
// Missing components default to 0.
func parseSemver(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	var out [3]int
	for i, p := range parts {
		if i >= 3 {
			break
		}
		out[i], _ = strconv.Atoi(p)
	}
	return out
}
