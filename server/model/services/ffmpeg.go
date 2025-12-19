package services

import (
	"os/exec"
)

type Engine struct {
	ffmpegPath  string
	ffprobePath string
}

var ffmpegEngine Engine

func InitializeFFMPEG() {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		panic("ffmpeg not found in PATH")
	}
	probePath, err := exec.LookPath("ffprobe")
	if err != nil {
		panic("ffprobe not found in PATH")
	}
	ffmpegEngine = Engine{
		ffmpegPath:  ffmpegPath,
		ffprobePath: probePath,
	}
}
