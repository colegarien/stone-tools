package archive_extractor

import (
	"fmt"
	"strings"
)

func (m model) View() string {
	pad := strings.Repeat(" ", padding)
	listing := ""
	if len(m.extractProgress.messages) > 0 {
		listingSize := min(10, len(m.extractProgress.messages))
		listing = "\n\n" + pad + "- " + strings.Join(m.extractProgress.messages[:listingSize], "\n"+pad+"- ")
	}

	errors := ""
	if len(m.extractProgress.errors) > 0 {
		errorSize := min(5, len(m.extractProgress.errors))
		errors = "\n\n" + pad + "- " + strings.Join(m.extractProgress.errors[:errorSize], "\n"+pad+"- ")
	}

	helpMessage := "Extraction in progress (press c to cancel)"
	if m.extractProgress.isDone {
		prefix := "Extraction took %s"
		if m.extractProgress.wasCanceled {
			prefix = "Extraction was canceled after %s"
		}

		helpMessage = fmt.Sprintf(prefix+", Press any key to return", m.extractProgress.timeTaken())
	}

	return "\n" +
		pad + m.progress.ViewAs(m.extractProgress.percent) + "\n\n" +
		pad + helpStyle(helpMessage) +
		listing + "\n" +
		errorStyle(errors)
}
