package archive_extractor

import (
	"os"
	"path/filepath"
	"stone-tools/lib"
	"stone-tools/view/filters"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type extractProgressMsg struct {
	extractedFiles    float64
	totalFiles        float64
	isDone            bool
	lastExtractedFile string
}

func extractArchive(sub chan extractProgressMsg, mtfFilePath string) tea.Cmd {
	return func() tea.Msg {
		mtfFile, err := os.Open(mtfFilePath)
		if err != nil {
			// fmt.Println("Error opening file:", err)
			return nil
		}
		defer mtfFile.Close()

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
		} // TODO stop doing this synchronously, slows the extract wayyyy down

		for _, virtualFile := range archive.VirtualFiles {
			extractedFile, err := lib.ExtractVirtualFile(mtfFile, virtualFile)
			if err != nil {
				// fmt.Printf("Error extracting file `%s`: %+v\r\n", virtualFile.FileName, err)
				continue
			}

			writePath := filepath.Join("out", virtualFile.FileName)
			// fmt.Printf("Writing `%s` (%d bytes)...\r\n", writePath, len(extractedFile))

			os.MkdirAll(filepath.Dir(writePath), os.ModePerm)
			err = os.WriteFile(writePath, extractedFile, os.ModePerm)
			if err != nil {
				// fmt.Printf("Error writing extracted file `%s`: %+v\r\n", virtualFile.FileName, err)
				continue
			}

			extractedFiles++
			sub <- extractProgressMsg{
				extractedFiles:    float64(extractedFiles),
				totalFiles:        float64(totalFiles),
				lastExtractedFile: virtualFile.FileName,
			}
		}

		return extractProgressMsg{
			extractedFiles: float64(totalFiles),
			totalFiles:     float64(totalFiles),
			isDone:         true,
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
	rootPath    string
	archivePath string

	sub            chan extractProgressMsg
	percent        float64
	progress       progress.Model
	extractedFiles []string
}

func New(rootPath, archivePath string) model {
	return model{
		rootPath:    rootPath,
		archivePath: archivePath,
		sub:         make(chan extractProgressMsg),
		progress:    progress.New(progress.WithDefaultGradient(), progress.WithWidth(filters.GlobalWindowSize.Width-(padding*2)-4)),
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
		return m, tea.Quit
	case extractProgressMsg:
		m.percent = msg.extractedFiles / msg.totalFiles
		if m.percent >= 1.0 {
			m.percent = 1.0
		}

		if msg.lastExtractedFile != "" && (len(m.extractedFiles) == 0 || m.extractedFiles[0] != msg.lastExtractedFile) {
			m.extractedFiles = append([]string{msg.lastExtractedFile}, m.extractedFiles...)
		}

		if msg.isDone {
			return m, nil // break the wait loop // TODO how to go back...
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
	return "\n" +
		pad + m.progress.ViewAs(m.percent) + "\n\n" +
		pad + helpStyle("Press any key to quit") +
		listing
}
