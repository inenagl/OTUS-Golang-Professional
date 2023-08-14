package main

import (
	"errors"
	"io"
	"os"
	"strings"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
	MsgEmptyFrom             = "from is empty"
	MsgEmptyTo               = "to is empty"
	MsgNegativeOffset        = "offset less then 0"
	MsgNegativeLimit         = "limit less then 0"
)

// Обёртка для вызовов методов Close у файлов в defer с проверкой на ошибку и распечаткой ошибки.
func runWithPrintError(f func() error) {
	if err := f(); err != nil {
		printError(err)
	}
}

func validateParams(fromPath, toPath string, offset, limit int64) error {
	messages := make([]string, 0, 4)

	if fromPath == "" {
		messages = append(messages, MsgEmptyFrom)
	}
	if toPath == "" {
		messages = append(messages, MsgEmptyTo)
	}
	if offset < 0 {
		messages = append(messages, MsgNegativeOffset)
	}
	if limit < 0 {
		messages = append(messages, MsgNegativeLimit)
	}
	if len(messages) > 0 {
		return errors.New(strings.Join(messages, "\n"))
	}

	return nil
}

func getSourceSize(path string) (int64, error) {
	sourceInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}

	return sourceInfo.Size(), nil
}

// Для корректной работы прогресс-бара приходится точно вычислять, какое число байт будет скопировано.
func getRealLimitBySourceSizeAndOffset(path string, offset, limit int64) (int64, error) {
	sourceSize, err := getSourceSize(path)
	if err != nil {
		return 0, err
	}

	if sourceSize == 0 {
		return 0, ErrUnsupportedFile
	}
	if offset > sourceSize {
		return 0, ErrOffsetExceedsFileSize
	}

	availableSize := sourceSize - offset
	if limit == 0 || limit > availableSize {
		limit = availableSize
	}

	return limit, nil
}

func Copy(fromPath, toPath string, offset, limit int64) error {
	if err := validateParams(fromPath, toPath, offset, limit); err != nil {
		return err
	}

	limit, err := getRealLimitBySourceSizeAndOffset(fromPath, offset, limit)
	if err != nil {
		return err
	}

	sourceFile, err := os.Open(fromPath)
	if err != nil {
		return err
	}

	targetFile, err := os.Create(toPath)
	if err != nil {
		return err
	}
	defer runWithPrintError(targetFile.Close)

	if offset > 0 {
		if _, err = sourceFile.Seek(offset, io.SeekStart); err != nil {
			return err
		}
	}

	bar := pb.Full.Start64(limit)
	barReader := bar.NewProxyReader(sourceFile)
	defer runWithPrintError(barReader.Close)

	if _, err = io.CopyN(targetFile, barReader, limit); err != nil && !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}
