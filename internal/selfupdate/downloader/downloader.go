package downloader

type Downloader interface {
	DownloadRelease(version string, os string, arch string) ([]byte, error)
	GetLatestRelease() (string, error)
}
