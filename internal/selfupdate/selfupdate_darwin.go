package selfupdate

import "os"

func replaceBinary(newPath string) error {
	self, _ := os.Executable()
	if err := os.Chmod(newPath, 0755); err != nil {
		return err
	}

	return os.Rename(newPath, self)
}
