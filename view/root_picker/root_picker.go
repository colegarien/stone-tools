package root_picker

import (
	"strings"

	"stone-tools/view/archive_picker"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	filepicker     filepicker.Model
	selectedFolder string
	quitting       bool
	err            error
}

func New(startDirectory string) model {
	fp := filepicker.New()
	fp.DirAllowed = true
	fp.FileAllowed = false
	fp.CurrentDirectory = startDirectory

	return model{
		filepicker:     fp,
		selectedFolder: startDirectory,
	}
}

func (m model) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "c":
			if m.err == nil && m.selectedFolder != "" {
				return archive_picker.New(m.selectedFolder), nil
			}
		}
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	if didSelect, _ := m.filepicker.DidSelectFile(msg); didSelect {
		m.selectedFolder = m.filepicker.CurrentDirectory
	} else if m.filepicker.CurrentDirectory != "" {
		m.selectedFolder = m.filepicker.CurrentDirectory
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return ""
	}
	var s strings.Builder
	s.WriteString("\n  ")
	if m.err != nil {
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.selectedFolder == "" {
		s.WriteString("Navigate to MTF installation:")
		s.WriteString("\n")
	} else {
		s.WriteString("Current folder: " + m.filepicker.Styles.Selected.Render(m.selectedFolder))
		s.WriteString("\n  - Hit `c` to select as root.")
	}
	s.WriteString("\n\n" + m.filepicker.View() + "\n")
	return s.String()
}
