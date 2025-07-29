package styles

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorConstants(t *testing.T) {
	tests := []struct {
		name     string
		color    string
		expected string
	}{
		{"ColorPrimary", ColorPrimary, "86"},
		{"ColorSecondary", ColorSecondary, "39"},
		{"ColorAccent", ColorAccent, "212"},
		{"ColorText", ColorText, "252"},
		{"ColorSuccess", ColorSuccess, "120"},
		{"ColorError", ColorError, "196"},
		{"ColorHelp", ColorHelp, "241"},
		{"ColorInfo", ColorInfo, "214"},
		{"ColorInfoBlue", ColorInfoBlue, "39"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.color != tt.expected {
				t.Errorf("Expected %s to be %s, got %s", tt.name, tt.expected, tt.color)
			}
		})
	}
}

func TestTitleStyle(t *testing.T) {
	// Test that it's bold
	if !TitleStyle.GetBold() {
		t.Error("TitleStyle should be bold")
	}

	// Test color
	if TitleStyle.GetForeground() != lipgloss.Color(ColorPrimary) {
		t.Error("TitleStyle should have primary color")
	}

	// Test border
	if !TitleStyle.GetBorderBottom() {
		t.Error("TitleStyle should have bottom border")
	}

	// Test margins
	if TitleStyle.GetMarginBottom() != 1 {
		t.Error("TitleStyle should have bottom margin of 1")
	}
}

func TestSectionStyle(t *testing.T) {
	// Test that it's bold
	if !SectionStyle.GetBold() {
		t.Error("SectionStyle should be bold")
	}

	// Test color
	if SectionStyle.GetForeground() != lipgloss.Color(ColorSecondary) {
		t.Error("SectionStyle should have secondary color")
	}

	// Test margins
	if SectionStyle.GetMarginTop() != 1 {
		t.Error("SectionStyle should have top margin of 1")
	}

	if SectionStyle.GetMarginBottom() != 1 {
		t.Error("SectionStyle should have bottom margin of 1")
	}
}

func TestLabelStyle(t *testing.T) {
	// Test that it's bold
	if !LabelStyle.GetBold() {
		t.Error("LabelStyle should be bold")
	}

	// Test color
	if LabelStyle.GetForeground() != lipgloss.Color(ColorAccent) {
		t.Error("LabelStyle should have accent color")
	}

	// Test width
	if LabelStyle.GetWidth() != 12 {
		t.Error("LabelStyle should have width of 12")
	}
}

func TestValueStyle(t *testing.T) {
	// Test color
	if ValueStyle.GetForeground() != lipgloss.Color(ColorText) {
		t.Error("ValueStyle should have text color")
	}
}

func TestHelpStyle(t *testing.T) {
	// Test color
	if HelpStyle.GetForeground() != lipgloss.Color(ColorHelp) {
		t.Error("HelpStyle should have help color")
	}

	// Test margin
	if HelpStyle.GetMarginTop() != 2 {
		t.Error("HelpStyle should have top margin of 2")
	}
}

func TestSuccessStyle(t *testing.T) {
	// Test that it's bold
	if !SuccessStyle.GetBold() {
		t.Error("SuccessStyle should be bold")
	}

	// Test color
	if SuccessStyle.GetForeground() != lipgloss.Color(ColorSuccess) {
		t.Error("SuccessStyle should have success color")
	}
}

func TestErrorStyle(t *testing.T) {
	// Test that it's bold
	if !ErrorStyle.GetBold() {
		t.Error("ErrorStyle should be bold")
	}

	// Test color
	if ErrorStyle.GetForeground() != lipgloss.Color(ColorError) {
		t.Error("ErrorStyle should have error color")
	}
}

func TestInfoStyle(t *testing.T) {
	// Test color
	if InfoStyle.GetForeground() != lipgloss.Color(ColorInfo) {
		t.Error("InfoStyle should have info color")
	}
}

func TestInfoBlueStyle(t *testing.T) {
	// Test color
	if InfoBlueStyle.GetForeground() != lipgloss.Color(ColorInfoBlue) {
		t.Error("InfoBlueStyle should have info blue color")
	}
}

func TestHeaderStyle(t *testing.T) {
	// Test that it's bold
	if !HeaderStyle.GetBold() {
		t.Error("HeaderStyle should be bold")
	}

	// Test color
	if HeaderStyle.GetForeground() != lipgloss.Color(ColorSecondary) {
		t.Error("HeaderStyle should have secondary color")
	}

	// Test alignment
	if HeaderStyle.GetAlign() != lipgloss.Center {
		t.Error("HeaderStyle should be center aligned")
	}
}

func TestCellStyle(t *testing.T) {
	// Test color
	if CellStyle.GetForeground() != lipgloss.Color(ColorText) {
		t.Error("CellStyle should have text color")
	}

	// Test alignment
	if CellStyle.GetAlign() != lipgloss.Center {
		t.Error("CellStyle should be center aligned")
	}
}

func TestSummaryStyle(t *testing.T) {
	// Test that it's bold
	if !SummaryStyle.GetBold() {
		t.Error("SummaryStyle should be bold")
	}

	// Test color
	if SummaryStyle.GetForeground() != lipgloss.Color(ColorSuccess) {
		t.Error("SummaryStyle should have success color")
	}

	// Test border
	if !SummaryStyle.GetBorderTop() {
		t.Error("SummaryStyle should have top border")
	}
}

func TestBreakStyle(t *testing.T) {
	// Test color
	if BreakStyle.GetForeground() != lipgloss.Color(ColorInfo) {
		t.Error("BreakStyle should have info color")
	}

	// Test padding
	if BreakStyle.GetPaddingLeft() != 2 {
		t.Error("BreakStyle should have left padding of 2")
	}
}

func TestNoteStyle(t *testing.T) {
	// Test color
	if NoteStyle.GetForeground() != lipgloss.Color("159") {
		t.Error("NoteStyle should have color 159")
	}

	// Test padding
	if NoteStyle.GetPaddingLeft() != 2 {
		t.Error("NoteStyle should have left padding of 2")
	}

	// Test margin
	if NoteStyle.GetMarginBottom() != 1 {
		t.Error("NoteStyle should have bottom margin of 1")
	}
}

func TestEditHeaderStyle(t *testing.T) {
	// Test that it's bold
	if !EditHeaderStyle.GetBold() {
		t.Error("EditHeaderStyle should be bold")
	}

	// Test color
	if EditHeaderStyle.GetForeground() != lipgloss.Color(ColorPrimary) {
		t.Error("EditHeaderStyle should have primary color")
	}
}

func TestEditFieldStyle(t *testing.T) {
	// Test padding
	if EditFieldStyle.GetPaddingLeft() != 2 {
		t.Error("EditFieldStyle should have left padding of 2")
	}

	if EditFieldStyle.GetPaddingRight() != 2 {
		t.Error("EditFieldStyle should have right padding of 2")
	}
}

func TestEditLabelStyle(t *testing.T) {
	// Test that it's bold
	if !EditLabelStyle.GetBold() {
		t.Error("EditLabelStyle should be bold")
	}

	// Test width
	if EditLabelStyle.GetWidth() != 12 {
		t.Error("EditLabelStyle should have width of 12")
	}
}

func TestEditHelpStyle(t *testing.T) {
	// Test color
	if EditHelpStyle.GetForeground() != lipgloss.Color(ColorHelp) {
		t.Error("EditHelpStyle should have help color")
	}

	// Test padding
	if EditHelpStyle.GetPaddingTop() != 1 {
		t.Error("EditHelpStyle should have top padding of 1")
	}
}

func TestStylesCanRender(t *testing.T) {
	// Test that styles can actually render text without panicking
	testText := "Test Text"
	
	styles := []lipgloss.Style{
		TitleStyle,
		SectionStyle,
		LabelStyle,
		ValueStyle,
		HelpStyle,
		SuccessStyle,
		ErrorStyle,
		InfoStyle,
		InfoBlueStyle,
		HeaderStyle,
		CellStyle,
		SummaryStyle,
		BreakStyle,
		NoteStyle,
		EditHeaderStyle,
		EditFieldStyle,
		EditLabelStyle,
		EditHelpStyle,
	}

	for i, style := range styles {
		t.Run("style_"+string(rune(i)), func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Style panicked when rendering: %v", r)
				}
			}()
			
			rendered := style.Render(testText)
			if rendered == "" {
				t.Error("Style rendered empty string")
			}
		})
	}
}