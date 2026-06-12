package cmd

import (
	"io"
	"os"
	"path/filepath"
)

// copyDirectory recursively copies a directory from src to dst.
// It creates the destination directory if it does not already exist.
func copyDirectory(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDirectory(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

// copyFile copies a single file from src to dst.
// It ensures the destination directory exists before writing the file.
func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// copyFiles copies multiple files given a slice of source/destination pairs.
func copyFiles(filePairs [][2]string) error {
	for _, pair := range filePairs {
		src := pair[0]
		dst := pair[1]
		if err := copyFile(src, dst); err != nil {
			return err
		}
	}
	return nil
}

// createMarkerFiles creates .lck and .ok marker files for a Gradle zip package.
func createMarkerFiles(targetDir, zipFileName string) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}

	lckPath := filepath.Join(targetDir, zipFileName+".lck")
	if _, err := os.Create(lckPath); err != nil {
		return err
	}

	okPath := filepath.Join(targetDir, zipFileName+".ok")
	if _, err := os.Create(okPath); err != nil {
		return err
	}

	return nil
}
