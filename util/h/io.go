package h

import (
	"embed"
	"io/fs"
	"os"
)

func CheckFsFolder(dir embed.FS, folderName string) (bool, fs.FS) {
  dirFS, err := fs.Sub(dir, folderName)
	if err != nil {
		if os.IsNotExist(err) {
      return false, nil // The folder does not exist
		}
		// Handle other errors if needed
    return false, nil
	}
	return true, dirFS // The folder exists
}
