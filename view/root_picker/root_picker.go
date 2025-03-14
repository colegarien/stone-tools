package root_picker

import (
	"strings"

	"stone-tools/config"
	"stone-tools/view/archive_picker"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	filepicker filepicker.Model
	conf       config.Config
	quitting   bool
	err        error
}

func New(conf config.Config) model {
	fp := filepicker.New()
	fp.DirAllowed = true
	fp.FileAllowed = false
	fp.CurrentDirectory = conf.DarkstoneDirectory

	return model{
		filepicker: fp,
		conf:       conf,
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
			if m.err == nil && m.conf.DarkstoneDirectory != "" {
				config.SaveConfig(m.conf)
				return archive_picker.New(m.conf), nil
			}
		}
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	if didSelect, _ := m.filepicker.DidSelectFile(msg); didSelect {
		m.conf.DarkstoneDirectory = m.filepicker.CurrentDirectory
	} else if m.filepicker.CurrentDirectory != "" {
		m.conf.DarkstoneDirectory = m.filepicker.CurrentDirectory
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
	} else if m.conf.DarkstoneDirectory == "" {
		s.WriteString("Navigate to MTF installation:")
		s.WriteString("\n")
	} else {
		s.WriteString("Current folder: " + m.filepicker.Styles.Selected.Render(m.conf.DarkstoneDirectory))
		s.WriteString("\n  - Hit `c` to select as root.")
	}
	s.WriteString("\n\n" + m.filepicker.View() + "\n")
	return s.String()
}
