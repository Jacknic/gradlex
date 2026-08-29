package cmd

import (
	"archive/zip"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// makeTestZip 构造一个最小 Gradle 分发包 zip，包含 gradle/bin/gradle、gradlew 及普通文件。
func makeTestZip(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	entries := map[string]string{
		"gradle-8.5/bin/gradle":     "#!/bin/sh\necho gradle\n",
		"gradle-8.5/bin/gradle.bat": "@echo off\r\n",
		"gradle-8.5/README":         "hello\n",
		"gradlew":                   "#!/bin/sh\necho wrapper\n",
	}
	for name, content := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create entry: %v", err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("write entry: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
}

func TestUnzipSetsExecutablePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable bit is not meaningful on Windows")
	}

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "gradle-8.5-bin.zip")
	makeTestZip(t, zipPath)

	dest := filepath.Join(dir, "out")
	if err := unzip(zipPath, dest); err != nil {
		t.Fatalf("unzip: %v", err)
	}

	gradleBin := filepath.Join(dest, "gradle-8.5", "bin", "gradle")
	gradlew := filepath.Join(dest, "gradlew")
	readme := filepath.Join(dest, "gradle-8.5", "README")

	for _, p := range []string{gradleBin, gradlew, readme} {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("expected file missing: %s: %v", p, err)
		}
	}

	for _, p := range []string{gradleBin, gradlew} {
		info, err := os.Stat(p)
		if err != nil {
			t.Fatalf("stat %s: %v", p, err)
		}
		if info.Mode().Perm()&0111 == 0 {
			t.Errorf("%s is not executable, mode=%v", p, info.Mode().Perm())
		}
	}

	info, err := os.Stat(readme)
	if err != nil {
		t.Fatalf("stat readme: %v", err)
	}
	if info.Mode().Perm()&0111 != 0 {
		t.Errorf("README should not be executable, mode=%v", info.Mode().Perm())
	}
}
