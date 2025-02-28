package filters

import (
	tea "github.com/charmbracelet/bubbletea"
)

type WindowSize struct {
	Width  int
	Height int
}

var GlobalWindowSize WindowSize

func GlobalFilter(m tea.Model, msg tea.Msg) tea.Msg {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		GlobalWindowSize.Width = msg.Width
		GlobalWindowSize.Height = msg.Height
	}

	return msg
}
