// Copyright (c) 2015-2022 MinIO, Inc.
// Copyright (c) 2026 Kypello
//
// This file is part of the Kypello Client (kc), a fork of MinIO Client (mc).
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

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

var (
	// Check if we stderr, stdout are dumb terminals, we do not apply
	// ansi coloring on dumb terminals.
	isTerminal = func() bool {
		return isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stderr.Fd())
	}

	colorCyanBold = func() func(a ...any) string {
		if isTerminal() {
			color.New(color.FgCyan, color.Bold).SprintFunc()
		}
		return fmt.Sprint
	}()
)
