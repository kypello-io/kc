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

	"github.com/kypello-io/kc/pkg/probe"
	"github.com/minio/cli"
	json "github.com/minio/colorjson"
)

// releasesURL is where kc releases are published. kc does not replace its own
// binary: upstream mc self-updated from dl.min.io, which for this fork would
// have downloaded MinIO's kc and installed it over kc.
const releasesURL = "https://github.com/kypello-io/kc/releases/latest"

// Report where to obtain new releases.
var updateCmd = cli.Command{
	Name:         "update",
	Usage:        "show the installed version and where to get newer ones",
	Action:       mainUpdate,
	OnUsageError: onUsageError,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "json",
			Usage: "enable JSON lines formatted output",
		},
	},
	CustomHelpTemplate: `Name:
   {{.HelpName}} - {{.Usage}}

USAGE:
   {{.HelpName}}{{if .VisibleFlags}} [FLAGS]{{end}}
{{if .VisibleFlags}}
FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}{{end}}
EXIT STATUS:
  0 - version information was printed

EXAMPLES:
  1. Show the installed version and the download location:
     {{.Prompt}} {{.HelpName}}
`,
}

type updateMessage struct {
	Status   string `json:"status"`
	Version  string `json:"version"`
	Releases string `json:"releases"`
	Message  string `json:"message"`
}

// String colorized update message.
func (s updateMessage) String() string {
	return s.Message
}

// JSON jsonified update message.
func (s updateMessage) JSON() string {
	s.Status = "success"
	updateJSONBytes, e := json.MarshalIndent(s, "", " ")
	fatalIf(probe.NewError(e), "Unable to marshal into JSON.")

	return string(updateJSONBytes)
}

func mainUpdate(ctx *cli.Context) {
	if len(ctx.Args()) > 0 {
		showCommandHelpAndExit(ctx, -1)
	}

	globalQuiet = ctx.Bool("quiet") || ctx.GlobalBool("quiet")
	globalJSON = ctx.Bool("json") || ctx.GlobalBool("json")

	msg := fmt.Sprintf("You are running %s version %s.\n", ctx.App.Name, ReleaseTag)
	msg += fmt.Sprintf("kc does not update itself. Download a newer release from %s\n", colorCyanBold(releasesURL))
	msg += "If you installed kc from a deb, rpm or apk package, or run it as a container image,\n"
	msg += "update it through that channel instead."

	printMsg(updateMessage{
		Status:   "success",
		Version:  ReleaseTag,
		Releases: releasesURL,
		Message:  msg,
	})
}
