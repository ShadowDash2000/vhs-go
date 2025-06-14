package ffhelp

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"os"
)

type Probe struct {
	Streams []Stream `json:"streams" mapstructure:"streams"`
	Format  Format   `json:"format" mapstructure:"format"`
}

type Format struct {
	Filename       string            `json:"filename" mapstructure:"filename"`
	NbStreams      int               `json:"nb_streams" mapstructure:"nb_streams"`
	NbPrograms     int               `json:"nb_programs" mapstructure:"nb_programs"`
	FormatName     string            `json:"format_name" mapstructure:"format_name"`
	FormatLongName string            `json:"format_long_name" mapstructure:"format_long_name"`
	StartTime      float64           `json:"start_time" mapstructure:"start_time"`
	Duration       float64           `json:"duration" mapstructure:"duration"`
	Size           string            `json:"size" mapstructure:"size"`
	Bitrate        string            `json:"bit_rate" mapstructure:"bit_rate"`
	ProbeScore     int               `json:"probe_score" mapstructure:"probe_score"`
	Tags           map[string]string `json:"tags" mapstructure:"tags"`
}

type Stream struct {
	Index          int     `json:"index" mapstructure:"index"`
	CodecName      string  `json:"codec_name" mapstructure:"codec_name"`
	CodecLongName  string  `json:"codec_long_name" mapstructure:"codec_long_name"`
	Profile        string  `json:"profile" mapstructure:"profile"`
	CodecType      string  `json:"codec_type" mapstructure:"codec_type"`
	CodecTagString string  `json:"codec_tag_string" mapstructure:"codec_tag_string"`
	CodecTag       string  `json:"codec_tag" mapstructure:"codec_tag"`
	Width          int     `json:"width" mapstructure:"width"`
	Height         int     `json:"height" mapstructure:"height"`
	CodecWidth     int     `json:"codec_width" mapstructure:"codec_width"`
	CodecHeight    int     `json:"codec_height" mapstructure:"codec_height"`
	DurationTs     int     `json:"duration_ts" mapstructure:"duration_ts"`
	Duration       float64 `json:"duration" mapstructure:"duration"`
	RFrameRate     string  `json:"r_frame_rate" mapstructure:"r_frame_rate"`
	Bitrate        string  `json:"bit_rate" mapstructure:"bit_rate"`
}

type FFHelp struct {
	stream   *ffmpeg.Stream
	p        *Probe
	filename string
}

func Input(filename string) *FFHelp {
	return &FFHelp{
		stream:   ffmpeg.Input(filename),
		p:        nil,
		filename: filename,
	}
}

func (ff *FFHelp) probe() (*Probe, error) {
	if ff.p != nil {
		return ff.p, nil
	}

	metaJson, err := ffmpeg.Probe(ff.filename)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	err = json.Unmarshal([]byte(metaJson), &m)
	if err != nil {
		return nil, err
	}

	err = mapstructure.WeakDecode(m, &ff.p)
	if err != nil {
		return nil, err
	}

	return ff.p, nil
}

func (ff *FFHelp) SplitVideoToThumbnails(output string, frameDuration float64) error {
	err := os.MkdirAll(output, os.ModePerm)
	if err != nil {
		return err
	}

	duration := ff.GetVideoDuration()
	i := 0
	for second := 1.0; second < duration; second = second + frameDuration {
		imagePath := fmt.Sprintf("%s/img%06d.jpg", output, i)

		ffmpeg.
			Input(ff.filename, ffmpeg.KwArgs{"ss": second}).
			Output(imagePath, ffmpeg.KwArgs{"vframes": "1", "q:v": frameDuration}).
			Silent(true).
			Run()

		i = i + 1
	}

	return nil
}

func (ff *FFHelp) GetVideoDuration() float64 {
	probe, err := ff.probe()
	if err != nil {
		return 0
	}

	var duration float64
	for _, stream := range probe.Streams {
		if stream.CodecType == "video" {
			duration = stream.Duration
			break
		}
	}

	return duration
}

func (ff *FFHelp) Probe() *Probe {
	probe, err := ff.probe()
	if err != nil {
		return &Probe{}
	}

	return probe
}
