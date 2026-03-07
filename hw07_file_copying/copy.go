package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	fr, err := os.OpenFile(fromPath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer fr.Close()

	err = checkFileMode(fr)
	if err != nil {
		return err
	}
	err = checkOffset(offset, fr)
	if err != nil {
		return err
	}
	limit, err = checkAndSetLimit(limit, fr)
	if err != nil {
		return err
	}
	fw, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer fw.Close()

	_, err = fr.Seek(offset, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error Seek: %w", err)
	}
	limitedReader := io.LimitReader(fr, limit)
	total := limit
	if total == 0 {
		fi, _ := fr.Stat()
		total = fi.Size() - offset
	}
	bar := pb.StartNew(int(total))
	bar.Set(pb.Bytes, true) // отображать в байтах
	defer bar.Finish()

	progressReader := bar.NewProxyReader(limitedReader)
	_, err = io.Copy(fw, progressReader)
	if err != nil {
		return fmt.Errorf("error on copy: %w", err)
	}

	return nil
}

func checkFileMode(f *os.File) error {
	fileInfo, err := f.Stat()
	if err != nil {
		return err
	}
	if fileInfo.Mode().IsRegular() {
		return nil
	}
	return ErrUnsupportedFile
}

func checkOffset(offset int64, f *os.File) error {
	fileInfo, err := f.Stat()
	if err != nil {
		return err
	}

	if offset > fileInfo.Size() {
		return ErrOffsetExceedsFileSize
	}
	return nil
}

func checkAndSetLimit(limit int64, f *os.File) (int64, error) {
	fileInfo, err := f.Stat()
	if err != nil {
		return limit, err
	}
	if limit == 0 {
		limit = fileInfo.Size()
	}
	return limit, nil
}
