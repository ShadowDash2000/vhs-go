package webvtt

import (
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
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

func CreateFromSheets(sheetPaths []string, outFile string, videoDuration, frameDuration, imgCount, cols, rows, thumbWidth, thumbHeight int) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(outFile), 0755)
	if err != nil {
		return nil, err
	}

	file, err := os.Create(outFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	file.Write([]byte("WEBVTT\n\n"))

	second := 0
	col := 0
	row := 0
	sheetNum := 0
	sheetPath := sheetPaths[sheetNum]
	for i := 0; i < imgCount; i++ {
		frameStart := time.Time{}
		frameStart = frameStart.Add(time.Duration(second) * time.Second)
		frameEnd := time.Time{}
		frameEnd = frameEnd.Add(time.Duration(second+frameDuration) * time.Second)
		timeFormat := "15:04:05.000"

		timeString := frameStart.Format(timeFormat) + " --> " + frameEnd.Format(timeFormat)
		file.Write([]byte(timeString + "\n"))

		spritePath := fmt.Sprintf("%s#xywh=%d,%d,%d,%d", strings.ReplaceAll(sheetPath, "\\", "/"), col*thumbWidth, row*thumbHeight, thumbWidth, thumbHeight)
		file.Write([]byte(spritePath + "\n\n"))

		second = second + frameDuration

		if second >= videoDuration {
			break
		}

		col++
		if col >= cols {
			col = 0
			row++
		}
		if row >= rows {
			row = 0
			sheetNum++

			if sheetNum >= len(sheetPaths) {
				break
			}
			sheetPath = sheetPaths[sheetNum]
		}
	}

	return file, nil
}

func CreateSpriteSheet(filePaths []string, outFile string, cols, rows, thumbWidth, thumbHeight int) error {
	err := os.MkdirAll(filepath.Dir(outFile), 0755)
	if err != nil {
		return err
	}

	outImg := image.NewRGBA(image.Rect(0, 0, cols*thumbWidth, rows*thumbHeight))
	i := 0

RowsLoop:
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			if i >= len(filePaths) {
				break RowsLoop
			}

			f, err := os.Open(filePaths[i])
			if err != nil {
				break
			}

			img, err := jpeg.Decode(f)
			f.Close()
			if err != nil {
				return err
			}

			r := image.Rect(x*thumbWidth, y*thumbHeight, (x+1)*thumbWidth, (y+1)*thumbHeight)
			draw.Draw(outImg, r, img, image.Point{}, draw.Src)
			i++
		}
	}

	of, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer of.Close()

	return jpeg.Encode(of, outImg, &jpeg.Options{Quality: 80})
}
