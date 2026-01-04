package twx

import (
	"io"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

// NewTable is a standart table UI.
func NewTable(w io.Writer) *tablewriter.Table {
	// Set table header
	return tablewriter.NewTable(
		w,
		tablewriter.WithRendition(tw.Rendition{
			Borders: tw.Border{
				Left:      tw.Off,
				Right:     tw.Off,
				Top:       tw.Off,
				Bottom:    tw.Off,
				Overwrite: false,
			},
			Symbols: tw.NewSymbols(tw.StyleNone),
			Settings: tw.Settings{
				Separators: tw.Separators{},
				Lines: tw.Lines{
					ShowTop:        0,
					ShowBottom:     0,
					ShowHeaderLine: tw.Off,
					ShowFooterLine: 0,
				},
				CompactMode: 0,
			},
			Streaming: false,
		}),
		tablewriter.WithTrimSpace(tw.Off),
	).Configure(func(cfg *tablewriter.Config) {
		cfg.Header.Alignment.Global = tw.AlignLeft
		cfg.Row.Alignment.Global = tw.AlignLeft
		cfg.Header.Formatting.AutoFormat = tw.On
		cfg.Row.Formatting.AutoWrap = tw.WrapNormal
		cfg.Row.Padding.Global.Left = "\t"
		cfg.Row.Padding.Global.Right = "\t"
		cfg.Header.Padding.Global.Left = "\t"
		cfg.Header.Padding.Global.Right = "\t"
		cfg.Footer.Padding.Global.Left = "\t"
		cfg.Footer.Padding.Global.Right = "\t"
	})
}
