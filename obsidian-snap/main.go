package main

// A simple program demonstrating the textarea component from the Bubbles
// component library.

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	var notePath string
	flag.StringVar(
		&notePath,
		"note-path",
		"",
		"path to note where one should append",
	)
	flag.Parse()

	p := tea.NewProgram(initialModel(notePath))

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type errMsg error

type executeDoneMsg struct{}

func appendToObsidianNote(
	note string,
	content string,
) tea.Cmd {
	return func() tea.Msg {
		b, err := os.ReadFile(note)
		if err != nil {
			return err
		}
		if b[len(b)-1] != '\n' {
			content = "\n" + content
		}
		content := fmt.Sprintf("%s%s", string(b), content)
		if err := os.WriteFile(
			note,
			[]byte(content),
			fs.FileMode(os.O_TRUNC),
		); err != nil {
			return errMsg(err)
		}

		return executeDoneMsg{}
	}
}

type model struct {
	notePath  string
	textArea  textarea.Model
	textInput textinput.Model
	err       error
}

func initialModel(notePath string) model {
	ti := textinput.New()
	ti.Placeholder = "Title"
	ti.Focus()
	ti.Width = 80

	ta := textarea.New()
	ta.Placeholder = "Extra context/links"
	ta.SetWidth(80)

	return model{
		textArea:  ta,
		textInput: ti,
		err:       nil,
		notePath:  notePath,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case executeDoneMsg:
		return m, tea.Quit
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlQ:
			val := m.textArea.Value()
			val = strings.Trim(val, "\n ")
			var lines []string
			if val != "" {
				lines = strings.Split(val, "\n")
				for index := range lines {
					lines[index] = fmt.Sprintf("\t- %s", lines[index])
				}
			}
			contentString := fmt.Sprintf(
				"- [ ] %s\n%s",
				m.textInput.Value(),
				strings.Join(lines, "\n"),
			)
			return m, appendToObsidianNote(m.notePath, contentString)
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyTab:
			if m.textInput.Focused() {
				m.textInput.Blur()
				cmds = append(cmds, m.textArea.Focus())
			} else {
				m.textArea.Blur()
				cmds = append(cmds, m.textInput.Focus())
			}
		}
	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}
	m.textArea, cmd = m.textArea.Update(msg)
	cmds = append(cmds, cmd)
	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.textInput.View(),
		m.textArea.View(),
		"(ctrl+c to quit, ctrl+q to submit)",
	) + "\n\n"
}
