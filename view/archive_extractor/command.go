package archive_extractor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"stone-tools/lib"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Init() tea.Cmd {
	return tea.Batch(
		extractArchive(m.ctx, m.sub, m.archivePath), // asynchronously start extracting the mtf file
		waitForProgress(m.sub),                      // wait for results
	)
}

type extractProgressMsg struct {
	// aggregate metrics
	extractedFiles float64
	totalFiles     float64
	isDone         bool
	wasCanceled    bool

	// latest on-the-fly metric
	time       time.Time
	message    string
	err        error
	errorCount int
}

func extractArchive(ctx context.Context, sub chan extractProgressMsg, mtfFilePath string) tea.Cmd {
	return func() tea.Msg {
		mtfFileData, err := os.ReadFile(mtfFilePath)
		if err != nil {
			sub <- extractProgressMsg{
				extractedFiles: 0,
				totalFiles:     0,
				isDone:         true,

				time:       time.Now().UTC(),
				message:    fmt.Sprintf("Error opening file: %v", err),
				errorCount: 1,
				err:        err,
			}

			return nil
		}
		mtfFile := bytes.NewReader(mtfFileData)

		archive, err := lib.ScanMtfFile(mtfFile)
		if err != nil {
			sub <- extractProgressMsg{
				extractedFiles: 0,
				totalFiles:     0,
				isDone:         true,

				time:       time.Now().UTC(),
				message:    fmt.Sprintf("Error scanning mtf file: %v", err),
				errorCount: 1,
				err:        err,
			}
			return nil
		}

		canceled := false
		errorCount := 0
		extractedFiles := 0
		totalFiles := len(archive.VirtualFiles)
		sub <- extractProgressMsg{
			extractedFiles: float64(extractedFiles),
			totalFiles:     float64(totalFiles),

			time:       time.Now().UTC(),
			errorCount: errorCount,
		}

		var wg sync.WaitGroup
		outputDirectory := filepath.Join("out", strings.TrimSuffix(filepath.Base(mtfFilePath), filepath.Ext(mtfFilePath)))
		for _, virtualFile := range archive.VirtualFiles {
			wg.Add(1)
			go func() {
				defer wg.Done()
				isCanceled := checkForCancelation(ctx, sub)
				if isCanceled {
					canceled = true
					return
				}

				section := io.NewSectionReader(mtfFile, 0, math.MaxInt64)
				extractedFile, err := lib.ExtractVirtualFile(section, virtualFile)
				if err != nil {
					errorCount++
					sub <- extractProgressMsg{
						extractedFiles: float64(extractedFiles),
						totalFiles:     float64(totalFiles),

						time:       time.Now().UTC(),
						message:    fmt.Sprintf("Error extracting file `%s`: %+v\r\n", virtualFile.FileName, err),
						errorCount: errorCount,
						err:        err,
					}
					return
				}

				isCanceled = checkForCancelation(ctx, sub)
				if isCanceled {
					canceled = true
					return
				}

				writePath := filepath.Join(outputDirectory, virtualFile.FileName)
				os.MkdirAll(filepath.Dir(writePath), os.ModePerm)
				err = os.WriteFile(writePath, extractedFile, os.ModePerm)
				if err != nil {
					errorCount++
					sub <- extractProgressMsg{
						extractedFiles: float64(extractedFiles),
						totalFiles:     float64(totalFiles),

						time:       time.Now().UTC(),
						message:    fmt.Sprintf("Error writing extracted file `%s`: %+v\r\n", virtualFile.FileName, err),
						errorCount: errorCount,
						err:        err,
					}
					return
				}

				extractedFiles++
				sub <- extractProgressMsg{
					extractedFiles: float64(extractedFiles),
					totalFiles:     float64(totalFiles),

					time:       time.Now().UTC(),
					message:    fmt.Sprintf("%s (%d bytes)", writePath, len(extractedFile)),
					errorCount: errorCount,
				}

				isCanceled = checkForCancelation(ctx, sub)
				if isCanceled {
					canceled = true
					return
				}
			}()
		}

		wg.Wait()
		if canceled {
			return nil
		}

		return extractProgressMsg{
			extractedFiles: float64(totalFiles),
			totalFiles:     float64(totalFiles),
			isDone:         true,

			time:       time.Now().UTC(),
			message:    fmt.Sprintf("Complete. Total files read %d and total errors %d.", totalFiles, errorCount),
			errorCount: errorCount,
		}
	}
}

func checkForCancelation(ctx context.Context, sub chan extractProgressMsg) bool {
	if ctx.Err() != nil {
		sub <- extractProgressMsg{
			isDone:      true,
			wasCanceled: true,

			time:    time.Now().UTC(),
			message: "Extraction canceled.",
		}
		return true
	}
	select {
	case <-ctx.Done():
		sub <- extractProgressMsg{
			isDone:      true,
			wasCanceled: true,

			time:    time.Now().UTC(),
			message: "Extraction canceled.",
		}
		return true
	default:
	}

	return false
}

func waitForProgress(sub chan extractProgressMsg) tea.Cmd {
	return func() tea.Msg {
		return extractProgressMsg(<-sub)
	}
}
