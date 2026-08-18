package application

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strings"

	billingDomain "github.com/sraj/addressbook/internal/billing/domain"
	"github.com/sraj/addressbook/internal/features/contact/domain"
	"github.com/xuri/excelize/v2"
)

// ImportRecord is one parsed row from a CSV or XLSX file.
type ImportRecord struct {
	Name    string
	Email   string
	Phone   string
	Line1   string
	Line2   string
	City    string
	State   string
	Zip     string
	Country string
	Notes   string
}

// ImportResult summarizes a batch import.
type ImportResult struct {
	Imported int `json:"imported"`
	Skipped  int `json:"skipped"`
}

var importColumns = []string{"name", "email", "phone", "address1", "address2", "city", "state", "zip", "country", "notes"}

// ParseImportRecords parses CSV or XLSX bytes into import records. The format
// is one of "csv" or "xlsx". A header row is expected.
func ParseImportRecords(data []byte, format string) ([]ImportRecord, error) {
	switch format {
	case "xlsx":
		return parseXLSX(data)
	default:
		return parseCSV(data)
	}
}

func parseCSV(data []byte) ([]ImportRecord, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse csv: %w", err)
	}
	return recordsToImport(records)
}

func parseXLSX(data []byte) ([]ImportRecord, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("open xlsx: %w", err)
	}
	defer func() { _ = f.Close() }()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("xlsx has no sheets")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("read xlsx rows: %w", err)
	}
	return recordsToImport(rows)
}

func recordsToImport(rows [][]string) ([]ImportRecord, error) {
	if len(rows) == 0 {
		return nil, nil
	}

	// Map header names to indexes.
	header := make(map[string]int, len(rows[0]))
	for i, col := range rows[0] {
		header[strings.ToLower(strings.TrimSpace(col))] = i
	}

	colIdx := func(name string) int {
		if i, ok := header[name]; ok {
			return i
		}
		return -1
	}

	var records []ImportRecord
	for r := 1; r < len(rows); r++ {
		row := rows[r]
		get := func(name string) string {
			i := colIdx(name)
			if i < 0 || i >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[i])
		}

		rec := ImportRecord{
			Name:    get("name"),
			Email:   get("email"),
			Phone:   get("phone"),
			Line1:   get("address1"),
			Line2:   get("address2"),
			City:    get("city"),
			State:   get("state"),
			Zip:     get("zip"),
			Country: get("country"),
			Notes:   get("notes"),
		}
		if rec.Name == "" && rec.Line1 == "" {
			continue
		}
		records = append(records, rec)
	}
	return records, nil
}

// Import creates contacts from parsed records. It respects the user's contact
// quota and stops once it is exhausted.
func (s *Service) Import(ctx context.Context, userID, collectionID uint, records []ImportRecord) (ImportResult, error) {
	if len(records) == 0 {
		return ImportResult{}, nil
	}

	remaining, err := s.billing.RemainingQuota(ctx, userID, "contacts")
	if err != nil {
		return ImportResult{}, err
	}

	var result ImportResult
	for _, rec := range records {
		if remaining == 0 {
			break
		}

		address := domain.Address{
			Label:   "Home",
			Line1:   rec.Line1,
			Line2:   rec.Line2,
			City:    rec.City,
			State:   rec.State,
			Zip:     rec.Zip,
			Country: rec.Country,
		}

		var emails, phones []string
		if rec.Email != "" {
			emails = []string{rec.Email}
		}
		if rec.Phone != "" {
			phones = []string{rec.Phone}
		}

		if _, err := s.Create(ctx, userID, collectionID, rec.Name, emails, phones, []domain.Address{address}, rec.Notes); err != nil {
			if errors.Is(err, billingDomain.ErrQuotaExceeded) {
				break
			}
			result.Skipped++
			continue
		}
		if remaining > 0 {
			remaining--
		}
		result.Imported++
	}
	return result, nil
}

// ExportColumns returns the header row used for CSV/XLSX exports.
func ExportColumns() []string {
	return importColumns
}

// ExportCSV renders contacts as CSV bytes including a header row.
func ExportCSV(contacts []domain.Contact) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write(importColumns); err != nil {
		return nil, err
	}
	for _, c := range contacts {
		addr := firstAddress(c)
		row := []string{
			c.Name,
			strings.Join(c.Emails, ";"),
			strings.Join(c.Phones, ";"),
			addr.Line1,
			addr.Line2,
			addr.City,
			addr.State,
			addr.Zip,
			addr.Country,
			c.Notes,
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ExportXLSX renders contacts as an XLSX workbook with a single sheet.
func ExportXLSX(contacts []domain.Contact) ([]byte, error) {
	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	sheet := "Contacts"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, err
	}
	f.SetActiveSheet(0)

	for i, col := range importColumns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := f.SetCellValue(sheet, cell, col); err != nil {
			return nil, err
		}
	}

	for r, c := range contacts {
		addr := firstAddress(c)
		values := []interface{}{
			c.Name,
			strings.Join(c.Emails, ";"),
			strings.Join(c.Phones, ";"),
			addr.Line1,
			addr.Line2,
			addr.City,
			addr.State,
			addr.Zip,
			addr.Country,
			c.Notes,
		}
		for i, v := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, r+2)
			if err := f.SetCellValue(sheet, cell, v); err != nil {
				return nil, err
			}
		}
	}

	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func firstAddress(c domain.Contact) domain.Address {
	if len(c.Addresses) == 0 {
		return domain.Address{}
	}
	return c.Addresses[0]
}
