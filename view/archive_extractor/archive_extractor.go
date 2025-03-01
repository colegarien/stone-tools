package archive_extractor

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"stone-tools/lib"
	"stone-tools/view/filters"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type extractProgressMsg struct {
	extractedFiles    float64
	totalFiles        float64
	isDone            bool
	lastExtractedFile string
	time              time.Time
}

func extractArchive(sub chan extractProgressMsg, mtfFilePath string) tea.Cmd {
	return func() tea.Msg {
		mtfFileData, err := os.ReadFile(mtfFilePath)
		if err != nil {
			// fmt.Println("Error opening file:", err)
			return nil
		}
		mtfFile := bytes.NewReader(mtfFileData)

		archive, err := lib.ScanMtfFile(mtfFile)
		if err != nil {
			// fmt.Println("Error scanning mtf file:", err)
			return nil
		}

		extractedFiles := 0
		totalFiles := len(archive.VirtualFiles)
		sub <- extractProgressMsg{
			extractedFiles: float64(extractedFiles),
			totalFiles:     float64(totalFiles),
			time:           time.Now().UTC(),
		}

		var wg sync.WaitGroup
		outputDirectory := filepath.Join("out", strings.TrimSuffix(filepath.Base(mtfFilePath), filepath.Ext(mtfFilePath)))
		for _, virtualFile := range archive.VirtualFiles {
			wg.Add(1)
			go func() {
				defer wg.Done()

				section := io.NewSectionReader(mtfFile, 0, math.MaxInt64)
				extractedFile, err := lib.ExtractVirtualFile(section, virtualFile)
				if err != nil {
					// fmt.Printf("Error extracting file `%s`: %+v\r\n", virtualFile.FileName, err)
					return
				}

				writePath := filepath.Join(outputDirectory, virtualFile.FileName)
				// fmt.Printf("Writing `%s` (%d bytes)...\r\n", writePath, len(extractedFile))

				os.MkdirAll(filepath.Dir(writePath), os.ModePerm)
				err = os.WriteFile(writePath, extractedFile, os.ModePerm)
				if err != nil {
					// fmt.Printf("Error writing extracted file `%s`: %+v\r\n", virtualFile.FileName, err)
					return
				}

				extractedFiles++
				sub <- extractProgressMsg{
					extractedFiles:    float64(extractedFiles),
					totalFiles:        float64(totalFiles),
					lastExtractedFile: virtualFile.FileName,
					time:              time.Now().UTC(),
				}
			}()
		}

		wg.Wait()
		return extractProgressMsg{
			extractedFiles: float64(totalFiles),
			totalFiles:     float64(totalFiles),
			isDone:         true,
			time:           time.Now().UTC(),
		}
	}
}

func waitForProgress(sub chan extractProgressMsg) tea.Cmd {
	return func() tea.Msg {
		return extractProgressMsg(<-sub)
	}
}

const (
	padding = 2
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

type model struct {
	previousModel tea.Model
	rootPath      string
	archivePath   string

	sub            chan extractProgressMsg
	percent        float64
	isDone         bool
	progress       progress.Model
	extractedFiles []string
	earliestTime   time.Time
	latestTime     time.Time
}

func New(previousModel tea.Model, rootPath, archivePath string) model {
	return model{
		previousModel: previousModel,
		rootPath:      rootPath,
		archivePath:   archivePath,
		sub:           make(chan extractProgressMsg),
		progress:      progress.New(progress.WithDefaultGradient(), progress.WithWidth(filters.GlobalWindowSize.Width-(padding*2)-4)),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		extractArchive(m.sub, m.archivePath), // asynchronously start extracting the mtf file
		waitForProgress(m.sub),               // wait for activity
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.isDone {
			return m.previousModel, nil
		}
		return m, nil
	case extractProgressMsg:
		percent := msg.extractedFiles / msg.totalFiles
		if percent >= 1.0 {
			percent = 1.0
		}

		if percent > m.percent {
			m.percent = percent
		}

		if m.earliestTime.IsZero() || m.earliestTime.After(msg.time) {
			m.earliestTime = msg.time
		}
		if m.latestTime.IsZero() || m.latestTime.Before(msg.time) {
			m.latestTime = msg.time
		}

		if msg.lastExtractedFile != "" && (len(m.extractedFiles) == 0 || m.extractedFiles[0] != msg.lastExtractedFile) {
			m.extractedFiles = append([]string{msg.lastExtractedFile}, m.extractedFiles...)
		}

		if msg.isDone {
			// break the wait loop.
			m.isDone = true
			return m, nil
		}

		return m, waitForProgress(m.sub)
	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - (padding * 2) - 4
		return m, nil
	default:
		return m, nil
	}
}

func (m model) View() string {
	pad := strings.Repeat(" ", padding)
	listing := ""
	if len(m.extractedFiles) > 0 {
		listingSize := 10
		if listingSize > len(m.extractedFiles) {
			listingSize = len(m.extractedFiles)
		}
		listing = "\n\n" + pad + "- " + strings.Join(m.extractedFiles[:listingSize], "\n"+pad+"- ")
	}

	helpMessage := "Extraction in progress"
	if m.isDone {
		helpMessage = fmt.Sprintf("Extraction took %s, Press any key to return", m.latestTime.Sub(m.earliestTime))
	}

	return "\n" +
		pad + m.progress.ViewAs(m.percent) + "\n\n" +
		pad + helpStyle(helpMessage) +
		listing
}
