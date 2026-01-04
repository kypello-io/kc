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
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/kypello-io/kc/pkg/probe"
	"github.com/kypello-io/pkg/v3/console"
	"github.com/minio/cli"
	json "github.com/minio/colorjson"
	"github.com/minio/madmin-go/v3"
	"github.com/minio/madmin-go/v3/estream"
)

var supportInspectFlags = append(supportGlobalFlags,
	cli.BoolFlag{
		Name:  "legacy",
		Usage: "use the older inspect format",
	},
)

var supportInspectCmd = cli.Command{
	Name:            "inspect",
	Usage:           "upload raw object contents for analysis",
	Action:          mainSupportInspect,
	OnUsageError:    onUsageError,
	Before:          setGlobalsFromContext,
	Flags:           supportInspectFlags,
	HideHelpCommand: true,
	CustomHelpTemplate: `NAME:
  {{.HelpName}} - {{.Usage}}

USAGE:
  {{.HelpName}} [FLAGS] TARGET

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
EXAMPLES:
  1. Upload 'xl.meta' of a specific object from all the drives
     {{.Prompt}} {{.HelpName}} myminio/bucket/test*/xl.meta

  2. Upload recursively all objects at a prefix. NOTE: This can be an expensive operation use it with caution.
     {{.Prompt}} {{.HelpName}} myminio/bucket/test/**

  3. Download 'xl.meta' of a specific object from all the drives locally, and upload to SUBNET manually
     {{.Prompt}} {{.HelpName}} myminio/bucket/test*/xl.meta --airgap
`,
}

type inspectMessage struct {
	Status     string `json:"status"`
	AliasedURL string `json:"aliasedURL,omitempty"`
	File       string `json:"file,omitempty"`
	Key        string `json:"key,omitempty"`
}

// Colorized message for console printing.
func (t inspectMessage) String() string {
	var msg string
	if globalAirgapped || t.File != "" {
		if t.Key == "" {
			msg = fmt.Sprintf("File data successfully downloaded as %s", console.Colorize("File", t.File))
		} else {
			msg = fmt.Sprintf("Encrypted file data successfully downloaded as %s\n", console.Colorize("File", t.File))
			msg += fmt.Sprintf("Decryption key: %s\n\n", console.Colorize("Key", t.Key))

			msg += "The decryption key will ONLY be shown here. It cannot be recovered.\n"
			msg += "The encrypted file can safely be shared without the decryption key.\n"
			msg += "Even with the decryption key, data stored with encryption cannot be accessed."
		}
	} else {
		msg = fmt.Sprintf("Object inspection data for '%s' uploaded to SUBNET successfully", t.AliasedURL)
	}
	return console.Colorize(supportSuccessMsgTag, msg)
}

func (t inspectMessage) JSON() string {
	t.Status = "success"
	jsonMessageBytes, e := json.MarshalIndent(t, "", " ")
	fatalIf(probe.NewError(e), "Unable to marshal into JSON.")
	return string(jsonMessageBytes)
}

func checkSupportInspectSyntax(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		showCommandHelpAndExit(ctx, 1) // last argument is exit code
	}
}

// mainSupportInspect - the entry function of inspect command
func mainSupportInspect(ctx *cli.Context) error {
	// Check for command syntax
	checkSupportInspectSyntax(ctx)

	setSuccessMessageColor()

	// Get the alias parameter from cli
	args := ctx.Args()
	aliasedURL := args.Get(0)

	console.SetColor("File", color.New(color.FgWhite, color.Bold))
	console.SetColor("Key", color.New(color.FgHiRed, color.Bold))

	// Create a new MinIO Admin Client
	client, err := newAdminClient(aliasedURL)
	if err != nil {
		fatalIf(err.Trace(aliasedURL), "Unable to initialize admin client.")
		return nil
	}

	// Compute bucket and object from the aliased URL
	aliasedURL = filepath.ToSlash(aliasedURL)
	splits := splitStr(aliasedURL, "/", 3)
	bucket, prefix := splits[1], splits[2]

	shellName, _ := getShellName()
	if runtime.GOOS != "windows" && shellName != "bash" && strings.Contains(prefix, "*") {
		console.Infoln("Your shell is auto determined as '" + shellName + "', wildcard patterns are only supported with 'bash' SHELL.")
	}

	key, r, e := client.Inspect(context.Background(), madmin.InspectOptions{
		Volume: bucket,
		File:   prefix,
	})
	fatalIf(probe.NewError(e).Trace(aliasedURL), "Unable to inspect file.")

	// Download the inspect data in a temporary file first
	tmpFile, e := os.CreateTemp("", "mc-inspect-")
	fatalIf(probe.NewError(e), "Unable to download file data.")
	copied := false

	// Validate stream, report errors on stream.
	if len(key) == 0 {
		pr, pw := io.Pipe()
		copied = true
		var aErr error
		var wg sync.WaitGroup
		// Validate stream...
		wg.Go(func() {
			er, err := estream.NewReader(pr)
			if err != nil {
				// Ignore if header is non-parsable...
				io.Copy(io.Discard, pr)
				return
			}
			// We don't care about decrypting at this point.
			er.SkipEncrypted(true)
			for {
				var st *estream.Stream
				st, aErr = er.NextStream()
				if errors.Is(aErr, io.EOF) {
					aErr = nil
					pr.CloseWithError(nil)
					return
				}
				if aErr != nil {
					pr.CloseWithError(aErr)
					return
				}
				aErr = st.Skip()
				if aErr != nil {
					fmt.Println("Skip:", aErr)
					pr.CloseWithError(aErr)
					return
				}
			}
		})
		_, e = io.Copy(io.MultiWriter(tmpFile, pw), r)
		pw.CloseWithError(e)
		wg.Wait()
		if e == nil && aErr != nil {
			e = aErr
		}
	}
	if !copied {
		_, e = io.Copy(tmpFile, r)
	}
	fatalIf(probe.NewError(e), "Unable to download file data.")
	r.Close()
	tmpFile.Close()
	wantFileName := "inspect-" + conservativeFileName(strings.Join(splits, "_")) + ".enc"

	// globalAirgapped
	saveInspectDataFile(wantFileName, key, tmpFile)
	return nil
}

func saveInspectDataFile(dstFileName string, key []byte, tmpFile *os.File) {
	var keyHex string

	// Choose a name and move the inspect data to its final destination
	if key != nil {
		// Create an id that is also crc.
		var id [4]byte
		binary.LittleEndian.PutUint32(id[:], crc32.ChecksumIEEE(key[:]))
		// We use 4 bytes of the 32 bytes to identify they file.
		dstFileName = fmt.Sprintf("inspect-data.%s.enc", hex.EncodeToString(id[:]))
		keyHex = hex.EncodeToString(id[:]) + hex.EncodeToString(key[:])
	}

	fi, e := os.Stat(dstFileName)
	if e == nil && !fi.IsDir() {
		e = moveFile(dstFileName, dstFileName+"."+time.Now().Format(dateTimeFormatFilename))
		fatalIf(probe.NewError(e), "Unable to create a backup of "+dstFileName)
	} else {
		if !os.IsNotExist(e) {
			fatal(probe.NewError(e), "Unable to download file data")
		}
	}

	fatalIf(probe.NewError(moveFile(tmpFile.Name(), dstFileName)), "Unable to rename downloaded data, file exists at %s", tmpFile.Name())

	printMsg(inspectMessage{
		File: dstFileName,
		Key:  keyHex,
	})
}
