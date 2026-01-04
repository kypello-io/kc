// Copyright (c) 2015-2022 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/fatih/color"
	"github.com/minio/pkg/v3/console"
)

// progress extender.
type progressBar struct {
	*pb.ProgressBar
	proxyReader *pb.Reader
}

func newPB(total int64) *pb.ProgressBar {
	// Progress bar specific theme customization.
	console.SetColor("Bar", color.New(color.FgGreen, color.Bold))

	// get the new original progress bar.
	bar := pb.New64(total)

	// Set new human friendly print units.
	bar.Set(pb.Bytes, true)
	// Refresh rate for progress bar is set to 125 milliseconds.
	bar.SetRefreshRate(time.Millisecond * 125)

	// Use different unicodes for Linux, OS X and Windows.
	var template string
	switch runtime.GOOS {
	case "linux":
		template = `{{string . "prefix"}} {{counters . }} {{bar . "┃" "▓" "█" "░" "┃"}} {{speed . | green}} {{percent . }}`
	case "darwin":
		template = `{{string . "prefix"}} {{counters . }} {{bar . " " "▓" " " "░" " "}} {{speed . | green}} {{percent . }}`
	default:
		template = `{{string . "prefix"}} {{counters . }} {{bar . "[" "=" ">" " " "]"}} {{speed . | green}} {{percent . }}`
	}
	bar.SetTemplateString(template)

	// Start the progress bar.
	return bar.Start()
}

func newProgressReader(r io.Reader, caption string, total int64) *pb.Reader {
	bar := newPB(total)

	if caption != "" {
		bar.Set("prefix", caption)
	}

	return bar.NewProxyReader(r)
}

// newProgressBar - instantiate a progress bar.
func newProgressBar(total int64) *progressBar {
	bar := newPB(total)

	// Create a proxy reader for nil reader to support Read() method
	reader := bar.NewProxyReader(nil)

	// Return new progress bar here.
	return &progressBar{ProgressBar: bar, proxyReader: reader}
}

// Set caption.
func (p *progressBar) SetCaption(caption string) *progressBar {
	caption = fixateBarCaption(caption, getFixedWidth(p.Width(), 18))
	p.ProgressBar.Set("prefix", caption)
	return p
}

func (p *progressBar) Finish() {
	p.ProgressBar.Finish()
}

func (p *progressBar) Set64(length int64) *progressBar {
	p.SetCurrent(length)
	return p
}

func (p *progressBar) Get() int64 {
	return p.Current()
}

func (p *progressBar) Read(buf []byte) (n int, err error) {
	n, err = p.proxyReader.Read(buf)
	if err != nil {
		return
	}

	// Upload retry can read one object twice; Avoid read to be greater than Total
	if t := p.Total(); t > 0 && int64(n) > t {
		p.SetCurrent(t)
	}

	return
}

func (p *progressBar) SetTotal(total int64) {
	p.ProgressBar.SetTotal(total)
}

// Write implements io.Writer interface for compatibility with v1 API
func (p *progressBar) Write(buf []byte) (n int, err error) {
	n = len(buf)
	p.Add64(int64(n))
	return
}

// Update is a no-op in v3, kept for compatibility
func (p *progressBar) Update() {
	// v3 updates automatically through templates
}

// Set sets the current progress (compatibility with v1)
func (p *progressBar) Set(current int64) {
	p.SetCurrent(current)
}

// cursorAnimate - returns a animated rune through read channel for every read.
func cursorAnimate() <-chan string {
	cursorCh := make(chan string)
	var cursors []string

	switch runtime.GOOS {
	case "linux":
		// cursors = "➩➪➫➬➭➮➯➱"
		// cursors = "▁▃▄▅▆▇█▇▆▅▄▃"
		cursors = []string{"◐", "◓", "◑", "◒"}
		// cursors = "←↖↑↗→↘↓↙"
		// cursors = "◴◷◶◵"
		// cursors = "◰◳◲◱"
		// cursors = "⣾⣽⣻⢿⡿⣟⣯⣷"
	case "darwin":
		cursors = []string{"◐", "◓", "◑", "◒"}
	default:
		cursors = []string{"|", "/", "-", "\\"}
	}
	go func() {
		for {
			for _, cursor := range cursors {
				cursorCh <- cursor
			}
		}
	}()
	return cursorCh
}

// fixateBarCaption - fancify bar caption based on the terminal width.
func fixateBarCaption(caption string, width int) string {
	switch {
	case len(caption) > width:
		// Trim caption to fit within the screen
		trimSize := len(caption) - width + 3
		if trimSize < len(caption) {
			caption = "..." + caption[trimSize:]
		}
	case len(caption) < width:
		caption += strings.Repeat(" ", width-len(caption))
	}
	return caption
}

// getFixedWidth - get a fixed width based for a given percentage.
func getFixedWidth(width, percent int) int {
	return width * percent / 100
}
