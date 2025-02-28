package archive_picker

import (
	"io/fs"
	"path/filepath"
	"stone-tools/view/archive_extractor"
	"stone-tools/view/filters"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type item struct {
	FileName string
	Desc     string
	Path     string
}

func (i item) Title() string       { return i.FileName }
func (i item) Description() string { return i.Desc }
func (i item) FilterValue() string { return i.FileName }

type model struct {
	rootPath string
	list     list.Model
}

func New(rootPath string) model {
	var listItems []list.Item
	filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".mtf") {
			desc := "- an mtf archive"
			listItems = append(listItems, item{FileName: filepath.Base(path), Desc: desc, Path: path})
		}
		return nil
	})

	m := model{
		rootPath: rootPath,
		list:     list.New(listItems, list.NewDefaultDelegate(), 0, 0),
	}
	m.list.Title = "MTF Archives"

	h, v := docStyle.GetFrameSize()
	m.list.SetSize(filters.GlobalWindowSize.Width-h, filters.GlobalWindowSize.Height-v)

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			nextView := archive_extractor.New(m.rootPath, m.list.SelectedItem().(item).Path)
			return nextView, nextView.Init()
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return docStyle.Render(m.list.View())
}
