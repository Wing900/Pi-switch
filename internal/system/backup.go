package system

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func BackupFile(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil || info.IsDir() {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	backupDir := filepath.Join(home, ".piswitch", "backups")
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return err
	}
	filename := filepath.Base(path)
	target := filepath.Join(backupDir, filename+"-"+time.Now().Format("2006-01-02-150405")+".bak")
	src, err := os.Open(path)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(target)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return trimBackups(backupDir, 20)
}

func trimBackups(dir string, keep int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(entries) <= keep {
		return nil
	}
	sort.Slice(entries, func(i, j int) bool {
		left, _ := entries[i].Info()
		right, _ := entries[j].Info()
		return left.ModTime().After(right.ModTime())
	})
	for _, entry := range entries[keep:] {
		_ = os.Remove(filepath.Join(dir, entry.Name()))
	}
	return nil
}
