package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Table renders data as a formatted table.
type Table struct {
	headers []string
	rows    [][]string
	writer  io.Writer
}

// NewTable creates a new table with the given headers.
func NewTable(headers ...string) *Table {
	return &Table{
		headers: headers,
		writer:  os.Stdout,
	}
}

// AddRow adds a row to the table.
func (t *Table) AddRow(cols ...string) {
	t.rows = append(t.rows, cols)
}

// Render writes the table to stdout.
func (t *Table) Render() {
	w := tabwriter.NewWriter(t.writer, 0, 0, 2, ' ', 0)

	// Header
	fmt.Fprintln(w, strings.Join(t.headers, "\t"))

	// Separator
	sep := make([]string, len(t.headers))
	for i, h := range t.headers {
		sep[i] = strings.Repeat("-", len(h))
	}
	fmt.Fprintln(w, strings.Join(sep, "\t"))

	// Rows
	for _, row := range t.rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}

	w.Flush()
}

// printOutput prints data in the requested format.
func printOutput(data interface{}) error {
	format := getOutputFormat()
	switch format {
	case "json":
		return printJSON(data)
	case "yaml":
		return printYAML(data)
	default:
		// For non-table formats, the caller should use Table directly.
		// This fallback prints JSON if someone calls printOutput with table format.
		return printJSON(data)
	}
}

func printJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func printYAML(data interface{}) error {
	enc := yaml.NewEncoder(os.Stdout)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(data)
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatSeverity returns a severity string with visual indicator.
func formatSeverity(severity string) string {
	switch strings.ToLower(severity) {
	case "critical":
		return "[!] CRITICAL"
	case "high":
		return "[H] HIGH"
	case "medium":
		return "[M] MEDIUM"
	case "low":
		return "[L] LOW"
	default:
		return severity
	}
}

// formatStatus returns a status string with visual indicator.
func formatStatus(status string) string {
	switch strings.ToLower(status) {
	case "connected", "active", "resolved", "applied", "healthy", "completed":
		return "[+] " + status
	case "disconnected", "error", "failed":
		return "[-] " + status
	case "open", "detected", "pending", "investigating":
		return "[*] " + status
	case "acknowledged":
		return "[~] " + status
	default:
		return status
	}
}
