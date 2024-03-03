package main

// A simple program demonstrating the textarea component from the Bubbles
// component library.

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(initialModel())

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
		cmd := exec.Command(
			"obsidian-cli",
			"create",
			note,
			"--append",
			fmt.Sprintf(
				"--content=%s",
				content,
			),
		)
		if err := cmd.Run(); err != nil {
			return errMsg(err)
		}
		return executeDoneMsg{}
	}
}

type model struct {
	textArea   textarea.Model
	textInput  textinput.Model
	focusInput int
	err        error
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Title"
	ti.Focus()
	ti.CharLimit = 50
	ti.Width = 36

	ta := textarea.New()
	ta.Placeholder = "Extra context/links"
	ta.SetWidth(50)

	return model{
		textArea:  ta,
		textInput: ti,
		err:       nil,
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
			contentString := fmt.Sprintf(
				"\n- [ ] %s\n\t- %s",
				m.textInput.Value(),
				m.textArea.Value(),
			)
			return m, appendToObsidianNote("Externals/Entry", contentString)
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
