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
	"github.com/fatih/color"
	"github.com/kypello-io/kc/pkg/probe"
	"github.com/kypello-io/pkg/v3/console"
	"github.com/minio/cli"
	json "github.com/minio/colorjson"
)

const (
	supportSuccessMsgTag = "SupportSuccessMessage"
	supportErrorMsgTag   = "SupportErrorMessage"
)

var supportGlobalFlags = append(globalFlags,
	cli.BoolFlag{
		Name:   "dev",
		Usage:  "Development mode",
		Hidden: true,
	},
	cli.BoolFlag{
		Name:  "airgap",
		Usage: "use in environments without network access to SUBNET (e.g. airgapped, firewalled, etc.)",
	},
)

var supportSubcommands = []cli.Command{
	supportDiagCmd,
	supportPerfCmd,
	supportInspectCmd,
	supportProfileCmd,
	supportTopCmd,
}

var supportCmd = cli.Command{
	Name:            "support",
	Usage:           "support related commands",
	Action:          mainSupport,
	Before:          setGlobalsFromContext,
	Flags:           globalFlags,
	Subcommands:     supportSubcommands,
	HideHelpCommand: true,
}

func setSuccessMessageColor() {
	console.SetColor(supportSuccessMsgTag, color.New(color.FgGreen, color.Bold))
}

func setErrorMessageColor() {
	console.SetColor(supportErrorMsgTag, color.New(color.FgYellow, color.Italic))
}

func toJSON(obj any) string {
	jsonBytes, e := json.MarshalIndent(obj, "", " ")
	fatalIf(probe.NewError(e), "Unable to marshal into JSON.")

	return string(jsonBytes)
}

// mainSupport is the handle for "mc support" command.
func mainSupport(ctx *cli.Context) error {
	commandNotFound(ctx, supportSubcommands)
	return nil
	// Sub-commands like "register", "callhome", "diagnostics" have their own main.
}
