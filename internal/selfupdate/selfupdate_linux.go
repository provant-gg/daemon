package selfupdate

import (
	"fmt"
	"os"
)

func replaceBinary(newPath string) error {
	self, _ := os.Executable()
	fmt.Println("Replacing binary at", self, "with", newPath)
	if err := os.Chmod(newPath, 0755); err != nil {
		return err
	}

	return os.Rename(newPath, self)
}
