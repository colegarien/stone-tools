package archive_extractor

import (
	"context"
	"stone-tools/view/filters"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	padding = 2
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render

type model struct {
	previousModel tea.Model
	rootPath      string
	archivePath   string

	ctx    context.Context
	cancel context.CancelFunc

	sub             chan extractProgressMsg
	progress        progress.Model
	extractProgress extractProgress
}

func New(previousModel tea.Model, rootPath, archivePath string) model {
	ctx, cancel := context.WithCancel(context.Background())
	return model{
		previousModel: previousModel,
		rootPath:      rootPath,
		archivePath:   archivePath,

		ctx:    ctx,
		cancel: cancel,

		sub:      make(chan extractProgressMsg),
		progress: progress.New(progress.WithDefaultGradient(), progress.WithWidth(filters.GlobalWindowSize.Width-(padding*2)-4)),
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.extractProgress.isDone {
			return m.previousModel, nil
		}

		if msg.String() == "c" {
			m.cancel()
		}
		return m, nil
	case extractProgressMsg:
		return m.onExtractProgressMsg(msg)
	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - (padding * 2) - 4
		return m, nil
	default:
		return m, nil
	}
}

func (m model) onExtractProgressMsg(msg extractProgressMsg) (tea.Model, tea.Cmd) {
	m.extractProgress.update(msg)
	if m.extractProgress.isDone {
		// break the wait loop.
		return m, nil
	}

	return m, waitForProgress(m.sub)
}

type extractProgress struct {
	percent      float64
	isDone       bool
	wasCanceled  bool
	messages     []string
	errors       []string
	earliestTime time.Time
	latestTime   time.Time
}

func (p *extractProgress) update(msg extractProgressMsg) {
	p.updatePercent(msg.extractedFiles / msg.totalFiles)
	p.updateTiming(msg.time)

	if msg.message != "" && (len(p.messages) == 0 || p.messages[0] != msg.message) {
		if msg.err != nil {
			p.errors = append([]string{msg.message}, p.errors...)
		} else {
			p.messages = append([]string{msg.message}, p.messages...)
		}
	}

	// specifically checking for true to avoid ever setting these back to false due to async hilarity
	if msg.isDone {
		p.isDone = true
	}
	if msg.wasCanceled {
		p.wasCanceled = true
	}
}

func (p *extractProgress) timeTaken() time.Duration {
	return p.latestTime.Sub(p.earliestTime)
}

func (p *extractProgress) updatePercent(percent float64) {
	if percent >= 1.0 {
		percent = 1.0
	}

	if percent > p.percent {
		p.percent = percent
	}
}

func (p *extractProgress) updateTiming(lastUpdateTime time.Time) {
	if p.earliestTime.IsZero() || p.earliestTime.After(lastUpdateTime) {
		p.earliestTime = lastUpdateTime
	}
	if p.latestTime.IsZero() || p.latestTime.Before(lastUpdateTime) {
		p.latestTime = lastUpdateTime
	}
}
