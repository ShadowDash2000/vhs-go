package webvtt

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

func CreateFromFilePaths(filePaths []string, outputPath string, videoDuration, frameDuration int) (*os.File, error) {
	err := os.MkdirAll(outputPath, 0755)
	if err != nil {
		return nil, err
	}

	file, err := os.Create(filepath.Join(outputPath, "thumbs.vtt"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	file.Write([]byte("WEBVTT\n\n"))

	second := 0
	for _, filePath := range filePaths {
		frameStart := time.Time{}
		frameStart = frameStart.Add(time.Duration(second) * time.Second)
		frameEnd := time.Time{}
		frameEnd = frameEnd.Add(time.Duration(second+frameDuration) * time.Second)
		timeFormat := "15:04:05.000"

		timeString := frameStart.Format(timeFormat) + " --> " + frameEnd.Format(timeFormat)

		file.Write([]byte(timeString + "\n"))
		file.Write([]byte(strings.ReplaceAll(filePath, "\\", "/") + "\n\n"))

		second = second + frameDuration

		if second >= videoDuration {
			break
		}
	}

	return file, nil
}
