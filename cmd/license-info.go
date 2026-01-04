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
	"time"

	"github.com/kypello-io/kc/pkg/probe"
	"github.com/kypello-io/pkg/v3/console"
	"github.com/minio/cli"
	json "github.com/minio/colorjson"
)

var licenseInfoCmd = cli.Command{
	Name:         "info",
	Usage:        "display license information",
	OnUsageError: onUsageError,
	Action:       mainLicenseInfo,
	Before:       setGlobalsFromContext,
	Flags:        supportGlobalFlags,
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} ALIAS

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}

EXAMPLES:
  1. Display license configuration for cluster with alias 'play'
     {{.Prompt}} {{.HelpName}} play
`,
}

const (
	licInfoMsgTag   = "licenseInfoMessage"
	licInfoErrTag   = "licenseInfoError"
	licInfoFieldTag = "licenseInfoField"
	licInfoValTag   = "licenseValueField"
)

type licInfoMessage struct {
	Status string  `json:"status"`
	Info   licInfo `json:"info,omitzero"`
	Error  string  `json:"error,omitempty"`
}

type licInfo struct {
	LicenseID    string     `json:"license_id,omitempty"`    // Unique ID of the license
	Organization string     `json:"org,omitempty"`           // Subnet organization name
	Plan         string     `json:"plan,omitempty"`          // Subnet plan
	IssuedAt     *time.Time `json:"issued_at,omitempty"`     // Time of license issue
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`    // Time of license expiry
	DeploymentID string     `json:"deployment_id,omitempty"` // Cluster deployment ID
	Message      string     `json:"message,omitempty"`       // Message to be displayed
	APIKey       string     `json:"api_key,omitempty"`       // API Key of the org account
}

func licInfoMsg(s string) string {
	return console.Colorize(licInfoMsgTag, s)
}

func licInfoErr(s string) string {
	return console.Colorize(licInfoErrTag, s)
}

// String colorized license info
func (li licInfoMessage) String() string {
	if len(li.Error) > 0 {
		return licInfoErr(li.Error)
	}

	return licInfoMsg(li.Info.Message)
}

// JSON jsonified license info
func (li licInfoMessage) JSON() string {
	jsonBytes, e := json.MarshalIndent(li, "", " ")
	fatalIf(probe.NewError(e), "Unable to marshal into JSON.")

	return string(jsonBytes)
}

func getAGPLMessage() string {
	return `License: GNU AGPL v3 <https://www.gnu.org/licenses/agpl-3.0.txt>
If you are distributing or hosting Kypello along with your proprietary application as combined works, you may require to switch to a commercial license included in the MinIO AIStor Subscriptions. (https://min.io/signup?ref=mc)`
}

func mainLicenseInfo(_ *cli.Context) error {
	printMsg(licInfoMessage{
		Status: "success",
		Info: licInfo{
			Plan:    "AGPLv3",
			Message: getAGPLMessage(),
		},
	})
	return nil
}
