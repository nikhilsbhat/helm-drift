package pkg

import "github.com/schollz/progressbar/v3"

const (
	defaultWidth = 80
)

func NewProgress(length int, message string) *progressbar.ProgressBar {
	progressBar := progressbar.NewOptions(length,
		progressbar.OptionSetWidth(defaultWidth),
		progressbar.OptionSetDescription(message),
		progressbar.OptionSetVisibility(true),
		progressbar.OptionShowBytes(true),
	)

	return progressBar
}
