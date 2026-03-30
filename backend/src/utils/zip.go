package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ExtractZip(data []byte, dest string) error {
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("zip error: %w", err)
	}

	for _, file := range reader.File {
		name := file.Name

		// Security: skip path traversal attempts
		if strings.Contains(name, "..") {
			continue
		}

		outPath := filepath.Join(dest, name)

		// Security: ensure resolved path stays within dest
		if !strings.HasPrefix(filepath.Clean(outPath), filepath.Clean(dest)) {
			continue
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(outPath, os.ModePerm)
			continue
		}

		if parent := filepath.Dir(outPath); parent != "" {
			os.MkdirAll(parent, os.ModePerm)
		}

		rc, err := file.Open()
		if err != nil {
			return fmt.Errorf("open file in zip: %w", err)
		}

		outFile, err := os.Create(outPath)
		if err != nil {
			rc.Close()
			return fmt.Errorf("create file: %w", err)
		}

		// Limit extraction size to prevent zip bombs (1GB max per file)
		_, err = io.Copy(outFile, io.LimitReader(rc, 1<<30))
		outFile.Close()
		rc.Close()
		if err != nil {
			return fmt.Errorf("write file: %w", err)
		}
	}

	return nil
}

func CreateZip(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil || info == nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name, _ = filepath.Rel(filepath.Dir(source), path)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})
}
