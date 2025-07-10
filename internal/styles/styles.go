package styles

import "github.com/charmbracelet/lipgloss"

// Color constants used throughout the application
const (
	ColorPrimary   = "86"  // Cyan - Main titles
	ColorSecondary = "39"  // Blue - Section headers, table headers
	ColorAccent    = "212" // Pink - Field labels
	ColorText      = "252" // Light Gray - Values, table cells
	ColorSuccess   = "120" // Green - Success messages, summaries
	ColorError     = "196" // Red - Error messages
	ColorHelp      = "241" // Dark Gray - Help text
	ColorInfo      = "214" // Orange - Informational messages
	ColorInfoBlue  = "39"  // Blue - Alternative info color
)

// Core UI styles used across multiple commands
var (
	// TitleStyle is used for main page titles and headers
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorPrimary)).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color(ColorPrimary)).
		MarginBottom(1).
		PaddingBottom(1)

	// SectionStyle is used for section headers within content
	SectionStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorSecondary)).
		MarginTop(1).
		MarginBottom(1)

	// LabelStyle is used for field labels like "Time:", "Status:", etc.
	LabelStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorAccent)).
		Width(12)

	// ValueStyle is used for values displayed after labels
	ValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText))

	// HelpStyle is used for help text at bottom of screens
	HelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorHelp)).
		MarginTop(2)
)

// Status styles for different types of messages
var (
	// SuccessStyle is used for success messages and confirmations
	SuccessStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorSuccess))

	// ErrorStyle is used for error messages
	ErrorStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorError))

	// InfoStyle is used for informational messages (orange variant)
	InfoStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorInfo))

	// InfoBlueStyle is used for informational messages (blue variant)
	InfoBlueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorInfoBlue))
)

// Table styles for reports
var (
	// HeaderStyle is used for table headers in reports
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorSecondary)).
		Align(lipgloss.Center)

	// CellStyle is used for table cells in reports
	CellStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorText)).
		Align(lipgloss.Center)

	// SummaryStyle is used for summary sections in reports
	SummaryStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorSuccess)).
		BorderStyle(lipgloss.NormalBorder()).
		BorderTop(true).
		BorderForeground(lipgloss.Color(ColorSuccess)).
		PaddingTop(1).
		MarginTop(2)
)

// Report-specific styles
var (
	// BreakStyle is used for formatting breaks in reports
	BreakStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorInfo)).
		PaddingLeft(2)

	// NoteStyle is used for formatting notes in reports
	NoteStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("159")).
		PaddingLeft(2).
		MarginBottom(1)
)

// Edit command styles
var (
	// EditHeaderStyle is used for headers in edit command
	EditHeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorPrimary)).
		PaddingBottom(1)

	// EditFieldStyle is used for field containers in edit command
	EditFieldStyle = lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingRight(2)

	// EditLabelStyle is used for labels in edit command
	EditLabelStyle = lipgloss.NewStyle().
		Bold(true).
		Width(12)

	// EditHelpStyle is used for help text in edit command
	EditHelpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorHelp)).
		PaddingTop(1)
)