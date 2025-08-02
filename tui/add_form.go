// tui/form.go
package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// styling for list items
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("170"))
)

// item is our list element
type item string

func (i item) FilterValue() string { return string(i) }

// itemDelegate renders each item
type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd    { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	str := fmt.Sprintf("%d. %s", index+1, i)

	renderFn := itemStyle.Render
	if index == m.Index() {
		// highlight selected
		renderFn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}
	fmt.Fprint(w, renderFn(str))
}

// formModel holds our Bubble Tea model
type formModel struct {
	list list.Model
}

// NewForm constructs it, with pagination disabled
func NewForm() formModel {
	items := []list.Item{
		item("Platform"),
		item("Topic"),
		item("Difficulty"),
	}

	const defaultWidth = 40
	const verticalPadding = 2 // 1 line of help + 1 margin

	// compute listHeight at runtime
	listHeight := len(items) + verticalPadding

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowFilter(false)
	l.SetShowHelp(true)
	l.SetShowPagination(false) // ← disable dots/pages

	return formModel{list: l}
}

func (m formModel) Init() tea.Cmd {
	return nil
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// only adjust width; keep the fixed height from NewForm
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m formModel) View() string {
	return "\n" + m.list.View()
}