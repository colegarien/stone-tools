package view

import (
	"os"
	"path/filepath"
	"runtime"
	"stone-tools/config"
	"stone-tools/view/archive_picker"
	"stone-tools/view/filters"
	"stone-tools/view/root_picker"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/sys/windows/registry"
)

func Run() error {
	conf, err := config.LoadConfig()
	if err != nil {
		return err
	}

	var model tea.Model
	if conf.DarkstoneDirectory == "" {
		// need to figure out root darkstone directory
		startDirectory, err := determineStartDirectory()
		if err != nil {
			return err
		}

		conf.DarkstoneDirectory = startDirectory
		model = root_picker.New(conf)
	} else {
		// go straight to archive selector
		model = archive_picker.New(conf)
	}

	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithFilter(filters.GlobalFilter))
	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

var possibleWindowRegKeys = []string{
	"SOFTWARE\\WOW6432Node\\DelphineSoft\\Darkstone\\CurrentVersion\\Darkstone",
	"SOFTWARE\\DelphineSoft\\Darkstone\\CurrentVersion\\Darkstone",
	"SOFTWARE\\WOW6432Node\\Delphine Software\\Darkstone\\CurrentVersion\\Darkstone",
	"SOFTWARE\\Delphine Software\\Darkstone\\CurrentVersion\\Darkstone",
	"SOFTWARE\\WOW6432Node\\Delphine Software\\Darkstone",
	"SOFTWARE\\Delphine Software\\Darkstone",
}

var possibleWindowsRegValues = []string{
	"DataPath",
	"InstallPath",
}

func determineStartDirectory() (string, error) {
	userHomeDirectory, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if runtime.GOOS == "windows" {
		keyFound := false
		var key registry.Key
		for _, keyPath := range possibleWindowRegKeys {
			key, err = registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.QUERY_VALUE)
			if err != nil {
				continue
			}

			keyFound = true
			break
		}

		if !keyFound {
			return userHomeDirectory, nil
		}
		defer key.Close()

		// Check if the value exists
		valueNameFound := false
		valueName := ""
		for _, valueName = range possibleWindowsRegValues {
			_, _, err = key.GetStringValue(valueName)
			if err != nil {
				continue
			}

			valueNameFound = true
			break
		}
		if !valueNameFound {
			return userHomeDirectory, nil
		}

		// Get the value's data
		value, _, err := key.GetStringValue(valueName)
		if err != nil {
			return userHomeDirectory, nil
		}

		return filepath.Clean(value), nil
	}

	return userHomeDirectory, nil
}
