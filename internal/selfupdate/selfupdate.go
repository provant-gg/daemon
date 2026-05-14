package selfupdate

import (
	"fmt"
	"os"
	"provantgg-daemon/internal/selfupdate/downloader"
	"runtime"

	"golang.org/x/mod/semver"
)

type Release struct {
	Version string `json:"version"`
}

type SelfUpdater struct {
	currentVersion string
	opts           *Options
}

type Options struct {
	Downloader downloader.Downloader
}

func NewSelfUpdater(currentVersion string, opts *Options) *SelfUpdater {
	if ok := semver.IsValid(currentVersion); !ok {
		panic("invalid version: " + currentVersion)
	}

	if opts == nil {
		opts = &Options{
			Downloader: downloader.NewNoOpDownloader(),
		}
	}

	return &SelfUpdater{
		currentVersion: currentVersion,
		opts:           opts,
	}
}

func (su *SelfUpdater) CheckForUpdate() (*Release, error) {
	latestVersion, err := su.opts.Downloader.GetLatestRelease()
	if err != nil {
		return nil, err
	}

	if semver.Compare(latestVersion, su.currentVersion) <= 0 {
		return nil, nil
	}

	return &Release{
		Version: latestVersion,
	}, nil
}

func (su *SelfUpdater) ApplyUpdate(release *Release) error {
	binaryData, err := su.opts.Downloader.DownloadRelease(release.Version, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}

	tmpFile, err := os.CreateTemp(os.TempDir(), "daemon-update-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.Write(binaryData)
	if err != nil {
		return err
	}

	fmt.Printf("Downloaded new version to %s\n", tmpFile.Name())
	return replaceBinary(tmpFile.Name())
}
