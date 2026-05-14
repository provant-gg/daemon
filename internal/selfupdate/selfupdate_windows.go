package selfupdate

import "os"

func replaceBinary(newPath string) error {
	self, _ := os.Executable()
	old := self + ".old"
	os.Remove(old)

	if err := os.Rename(self, old); err != nil {
		return err
	}

	return os.Rename(newPath, self)
}
