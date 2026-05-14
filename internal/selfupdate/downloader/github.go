package downloader

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type GitHubDownloader struct {
	GitHubDownloaderOptions
}

type GitHubDownloaderOptions struct {
	Owner      string
	Repository string
	Private    bool
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type githubReleaseResponse struct {
	Assets []githubAsset `json:"assets"`
}

type selectedAssets struct {
	BinaryURL    string
	BinaryName   string
	ChecksumsURL string
}

var (
	osMap = map[string]string{
		"linux":   "Linux",
		"darwin":  "Darwin",
		"windows": "Windows",
	}
	archMap = map[string]string{
		"amd64": "x86_64",
		"386":   "i386",
		"arm64": "arm64",
	}
)

func NewGitHubDownloader(opts GitHubDownloaderOptions) Downloader {
	return &GitHubDownloader{
		GitHubDownloaderOptions: opts,
	}
}

func (d *GitHubDownloader) verifyChecksum(checksums []byte, filename string, data []byte) error {
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])

	scanner := bufio.NewScanner(bytes.NewReader(checksums))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}
		want, name := fields[0], fields[1]
		if name == filename {
			if !strings.EqualFold(want, got) {
				return fmt.Errorf("checksum mismatch for %s: got %s, want %s", filename, got, want)
			}

			fmt.Println("Checksum verified for", filename)
			return nil
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return fmt.Errorf("no checksum entry for %s", filename)
}

func (d *GitHubDownloader) DownloadRelease(version string, hostOS string, arch string) ([]byte, error) {
	assets, err := d.fetchReleaseAssets(version)
	if err != nil {
		return nil, err
	}

	selected, err := d.selectAssets(assets, hostOS, arch)
	if err != nil {
		return nil, fmt.Errorf("release %s: %w", version, err)
	}

	binaryData, err := d.httpGetBytes(selected.BinaryURL)
	if err != nil {
		return nil, err
	}
	fmt.Printf("Asset downloaded %s [size: %dMB]\n", selected.BinaryURL, len(binaryData)/1024/1024)

	checksums, err := d.httpGetBytes(selected.ChecksumsURL)
	if err != nil {
		return nil, err
	}

	if err := d.verifyChecksum(checksums, selected.BinaryName, binaryData); err != nil {
		return nil, err
	}

	return binaryData, nil
}

func (d *GitHubDownloader) fetchReleaseAssets(version string) ([]githubAsset, error) {
	req, err := http.NewRequest("GET", d.buildGitHubApiUrl("releases", "tags", version), nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release githubReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	return release.Assets, nil
}

func (d *GitHubDownloader) selectAssets(assets []githubAsset, hostOS, arch string) (selectedAssets, error) {
	wantOS, ok := osMap[hostOS]
	if !ok {
		return selectedAssets{}, fmt.Errorf("unsupported os: %s", hostOS)
	}
	wantArch, ok := archMap[arch]
	if !ok {
		return selectedAssets{}, fmt.Errorf("unsupported arch: %s", arch)
	}

	suffix := fmt.Sprintf("_%s_%s.tar.gz", wantOS, wantArch)
	if hostOS == "windows" {
		suffix = fmt.Sprintf("_%s_%s.zip", wantOS, wantArch)
	}

	var result selectedAssets
	for i := range assets {
		if strings.HasSuffix(assets[i].Name, suffix) {
			result.BinaryURL = assets[i].BrowserDownloadURL
			result.BinaryName = assets[i].Name
		}
		if assets[i].Name == "checksums.txt" {
			result.ChecksumsURL = assets[i].BrowserDownloadURL
		}
	}

	if result.BinaryURL == "" {
		return selectedAssets{}, fmt.Errorf("no asset found for os %s, arch %s", hostOS, arch)
	}
	if result.ChecksumsURL == "" {
		return selectedAssets{}, fmt.Errorf("no checksums.txt found")
	}
	return result, nil
}

func (d *GitHubDownloader) httpGetBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func (d *GitHubDownloader) GetLatestRelease() (string, error) {
	req, err := http.NewRequest("GET", d.buildGitHubApiUrl("releases", "latest"), nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("unexpected status code: %d", resp.StatusCode)
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	type latestReleaseResponse struct {
		TagName string `json:"tag_name"`
	}

	var latestRelease latestReleaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&latestRelease); err != nil {
		return "", err
	}

	return latestRelease.TagName, nil
}

func (d *GitHubDownloader) buildGitHubApiUrl(suffixPath ...string) string {
	u, err := url.JoinPath(fmt.Sprintf("https://api.github.com/repos/%s/%s", d.Owner, d.Repository), suffixPath...)
	if err != nil {
		panic(err)
	}

	return u
}
