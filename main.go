package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/masterteapot/oriole/player"
)

type model struct {
	filepicker    filepicker.Model
	selectedFile  string
	quitting      bool
	thisIsPlaying player.WhatIsPlaying
	err           error
}

type clearErrorMsg struct{}

func playSong(uri string, m *model) tea.Cmd {
	return func() tea.Msg {
		if m.thisIsPlaying.Playing {
			tea.Println("We are trying to stop\n")		
			m.thisIsPlaying.MainLoop.Quit()
		}
		thisIsPlaying := player.Play(uri)
		thisIsPlaying.Playing = true
		m.thisIsPlaying = thisIsPlaying
		return nil
	}
}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
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
		}
	case clearErrorMsg:
		m.err = nil
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		m.selectedFile = path
		return m, playSong("file://"+path, &m)
	}

	if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
		// Let's clear the selectedFile and display an error.
		m.err = errors.New(path + " is not valid.")
		m.selectedFile = ""
		return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
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
	} else if m.selectedFile == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + m.filepicker.Styles.Selected.Render(m.selectedFile))
	}
	s.WriteString("\n\n" + m.filepicker.View() + "\n")
	return s.String()
}

func main() {
	fp := filepicker.New()
	fp.AllowedTypes = []string{".mp3", ".flac", ".ogg", ".wav"}
	fp.CurrentDirectory = "/home/jared/Music"

	m := model{
		filepicker:    fp,
		thisIsPlaying: player.WhatIsPlaying{Playing: false},
	}
	tm, _ := tea.NewProgram(&m).Run()
	mm := tm.(model)
	fmt.Println("\n  You are playing: " + m.filepicker.Styles.Selected.Render(mm.selectedFile) + "\n")
}
