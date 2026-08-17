package domain

import (
	"errors"
	"sort"
)

// ErrUnsupportedFormat is returned when a label format code is unknown.
var ErrUnsupportedFormat = errors.New("unsupported label format")

// Format describes a printable label sheet template. Dimensions are in inches
// and map to CSS values used by the print sheet renderer.
type Format struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Width       string `json:"width"`
	Height      string `json:"height"`
	Columns     int    `json:"columns"`
	Rows        int    `json:"rows"`
	FontSizePx  int    `json:"font_size_px"`
	CellPadding string `json:"cell_padding"`
}

// LabelsPerSheet returns the number of labels on one sheet.
func (f Format) LabelsPerSheet() int {
	return f.Columns * f.Rows
}

// DefaultLabelFormat is the template used when none is selected.
const DefaultLabelFormat = "5160"

// PrintScript is the inline script embedded in the printable sheet. It is
// referenced both by the renderer and by the sheet handler so the CSP can
// allow it via a sha256 hash instead of unsafe-inline.
const PrintScript = `document.addEventListener('DOMContentLoaded', function () {
  var btn = document.getElementById('print-btn');
  if (btn) { btn.addEventListener('click', function () { window.print(); }); }
});`

var labelFormats = map[string]Format{
	"5160": {
		Code:        "5160",
		Name:        "Avery 5160 · Address (1\" x 2 5/8\")",
		Width:       "2.625in",
		Height:      "1in",
		Columns:     3,
		Rows:        10,
		FontSizePx:  11,
		CellPadding: "0.12in 0.15in",
	},
	"8160": {
		Code:        "8160",
		Name:        "Avery 8160 · Address (1\" x 2 5/8\")",
		Width:       "2.625in",
		Height:      "1in",
		Columns:     3,
		Rows:        10,
		FontSizePx:  11,
		CellPadding: "0.12in 0.15in",
	},
	"5162": {
		Code:        "5162",
		Name:        "Avery 5162 · Large Address (1.33\" x 4\")",
		Width:       "4in",
		Height:      "1.33in",
		Columns:     2,
		Rows:        7,
		FontSizePx:  12,
		CellPadding: "0.16in 0.2in",
	},
	"5163": {
		Code:        "5163",
		Name:        "Avery 5163 · Shipping (2\" x 4\")",
		Width:       "4in",
		Height:      "2in",
		Columns:     2,
		Rows:        5,
		FontSizePx:  14,
		CellPadding: "0.2in 0.22in",
	},
	"6871": {
		Code:        "6871",
		Name:        "Avery 6871 · Print-to-edge (1.25\" x 2 3/8\")",
		Width:       "2.375in",
		Height:      "1.25in",
		Columns:     3,
		Rows:        6,
		FontSizePx:  12,
		CellPadding: "0.15in 0.15in",
	},
}

// SupportedFormats returns all available label templates sorted by code.
func SupportedFormats() []Format {
	codes := make([]string, 0, len(labelFormats))
	for code := range labelFormats {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	formats := make([]Format, 0, len(codes))
	for _, code := range codes {
		formats = append(formats, labelFormats[code])
	}
	return formats
}

// FormatByCode resolves a format code, falling back to the default when empty.
// Unknown codes return ErrUnsupportedFormat.
func FormatByCode(code string) (Format, error) {
	if code == "" {
		code = DefaultLabelFormat
	}
	f, ok := labelFormats[code]
	if !ok {
		return Format{}, ErrUnsupportedFormat
	}
	return f, nil
}
