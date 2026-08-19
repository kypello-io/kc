//go:build linux

// Copyright (c) 2015-2023 MinIO, Inc.
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
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

const pipeMaxSizeProcFile = "/proc/sys/fs/pipe-max-size"

func setPipeSize(fd uintptr, size int) error {
	_, err := unix.FcntlInt(fd, unix.F_SETPIPE_SZ, size)
	return err
}

func getConfiguredMaxPipeSize() (int, error) {
	b, err := os.ReadFile(pipeMaxSizeProcFile)
	if err != nil {
		return 0, err
	}
	// Atoi, not ParseInt into an int64 that is then narrowed: the value is only
	// ever used as an int, and this way one too large for the platform's int
	// errors out instead of wrapping to a small or negative pipe size.
	maxSize, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		return 0, fmt.Errorf("error parsing %s content: %v", pipeMaxSizeProcFile, err)
	}
	return maxSize, nil
}

// increasePipeBufferSize attempts to increase the pipe size to the the input value
// or system-max, if the provided size is 0 or less.
func increasePipeBufferSize(f *os.File, desiredPipeSize int) error {
	fd := f.Fd()

	if desiredPipeSize <= 0 {
		maxSize, err := getConfiguredMaxPipeSize()
		if err == nil {
			setPipeSize(fd, maxSize)
			return nil
		}
	}

	return setPipeSize(fd, desiredPipeSize)
}
