package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles using Lipgloss for rich TUI aesthetics
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EE6FF8")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C1C1C1"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	scoreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8A8A8A"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)
)

type model struct {
	choices  []DirectoryEntry
	cursor   int
	selected string
	quitting bool
	height   int
	width    int
}

func initialModel(entries []DirectoryEntry) model {
	return model{
		choices: entries,
		cursor:  0,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if len(m.choices) == 0 {
				return m, nil
			}
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.choices) - 1 // wrap around
			}

		case "down", "j":
			if len(m.choices) == 0 {
				return m, nil
			}
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			} else {
				m.cursor = 0 // wrap around
			}

		case "enter":
			if len(m.choices) > 0 {
				m.selected = m.choices[m.cursor].Path
			}
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	s.WriteString(titleStyle.Render(" QCD - Quick Directory Changer ") + "\n")

	if len(m.choices) == 0 {
		s.WriteString(errorStyle.Render("No directories registered yet.") + "\n\n")
		s.WriteString(normalStyle.Render("Try registering the current directory:") + "\n")
		s.WriteString(selectedStyle.Render("  qcd add .") + "\n\n")
		s.WriteString(helpStyle.Render("Press 'q' or 'esc' to exit.") + "\n")
		return s.String()
	}

	// Simple scroll windowing based on available heights
	maxVisible := 12
	if m.height > 6 {
		maxVisible = m.height - 6 // reserve lines for title and help
	}
	if maxVisible < 1 {
		maxVisible = 1
	}

	start := 0
	end := len(m.choices)

	if len(m.choices) > maxVisible {
		half := maxVisible / 2
		start = m.cursor - half
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(m.choices) {
			end = len(m.choices)
			start = end - maxVisible
		}
	}

	for i := start; i < end; i++ {
		entry := m.choices[i]
		cursorStr := "  "
		var line string

		// Calculate fuzzy relative time for last accessed
		timeDiff := time.Since(entry.LastAccessed).Round(time.Second)
		var timeStr string
		if timeDiff < time.Minute {
			timeStr = "just now"
		} else if timeDiff < time.Hour {
			timeStr = fmt.Sprintf("%dm ago", int(timeDiff.Minutes()))
		} else if timeDiff < 24*time.Hour {
			timeStr = fmt.Sprintf("%dh ago", int(timeDiff.Hours()))
		} else {
			timeStr = fmt.Sprintf("%dd ago", int(timeDiff.Hours()/24))
		}

		meta := fmt.Sprintf(" (visits: %d, %s)", entry.Score, timeStr)

		if m.cursor == i {
			cursorStr = "❯ "
			line = selectedStyle.Render(entry.Path) + scoreStyle.Render(meta)
		} else {
			line = normalStyle.Render(entry.Path) + scoreStyle.Render(meta)
		}
		s.WriteString(fmt.Sprintf("%s%s\n", cursorStr, line))
	}

	s.WriteString("\n" + helpStyle.Render("↑/↓ or k/j: Navigate • Enter: Select & cd • Esc/q: Cancel") + "\n")
	return s.String()
}

// runTUI starts the bubbletea program and returns the selected path.
// It uses Stderr for rendering TUI, keeping Stdout clean for the output path.
func runTUI() (string, error) {
	entries, err := getSortedDirectories()
	if err != nil {
		return "", err
	}

	p := tea.NewProgram(
		initialModel(entries),
		tea.WithOutput(os.Stderr),
	)

	m, err := p.Run()
	if err != nil {
		return "", err
	}

	if finalModel, ok := m.(model); ok {
		if finalModel.selected != "" {
			// Increment score for next time
			_ = incrementScore(finalModel.selected)
			return finalModel.selected, nil
		}
	}

	return "", nil
}
