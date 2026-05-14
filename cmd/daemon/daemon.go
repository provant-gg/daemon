package main

import (
	"fmt"
	"rl-stats-daemon/internal/selfupdate"
	"rl-stats-daemon/internal/selfupdate/downloader"
)

var (
	version = "v0.1.3"
)

func main() {
	selfUpdater := selfupdate.NewSelfUpdater(version, &selfupdate.Options{
		Downloader: downloader.NewGitHubDownloader(downloader.GitHubDownloaderOptions{
			Owner:      "provant-gg",
			Repository: "daemon",
			Private:    false,
		}),
	})

	update, err := selfUpdater.CheckForUpdate()
	if err != nil {
		panic(err)
	}

	if update != nil {
		fmt.Printf("New version available: %s\n", update.Version)
		if err := selfUpdater.ApplyUpdate(update); err != nil {
			panic(err)
		}
	}

	fmt.Printf("Current version: %s\n", version)
}
