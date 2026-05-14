package downloader

type NoOpDownloader struct {
}

func NewNoOpDownloader() Downloader {
	return &NoOpDownloader{}
}

func (d *NoOpDownloader) DownloadRelease(version string, os string, arch string) ([]byte, error) {
	return []byte{}, nil
}

func (d *NoOpDownloader) GetLatestRelease() (string, error) {
	return "", nil
}
