package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/Harschmann/Todo-/model"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- VIEWS ---
type currentView int

const (
	viewMain currentView = iota
	viewPlatform
	viewTopic
	viewDifficulty
	viewTime
	viewNotes
)

// --- STYLES ---
var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	summaryStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	focusedStyle      = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 1)
)

// --- LIST ITEM & DELEGATE ---
type item string

func (i item) FilterValue() string { return string(i) }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	str := fmt.Sprintf("%d. %s", index+1, i)
	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string { return selectedItemStyle.Render("> " + strings.Join(s, " ")) }
	}
	fmt.Fprint(w, fn(str))
}

// --- MODEL ---
type formModel struct {
	currentView currentView
	logEntry    model.Log
	mainMenu    list.Model
	platforms   list.Model
	topics      list.Model
	difficulty  list.Model
	timeInput   textinput.Model
	notesInput  textinput.Model
}

func NewForm() formModel {
	const defaultWidth = 40
	const verticalPadding = 2

	subListDelegate := itemDelegate{}

	// Main Menu
	mainMenuItems := []list.Item{
		item("Platform"), item("Topic"), item("Difficulty"),
		item("Time Spent"), item("Notes"), item("Submit"),
	}
	mainMenu := list.New(mainMenuItems, subListDelegate, defaultWidth, len(mainMenuItems)+verticalPadding)
	mainMenu.SetShowTitle(false)

	// Platform List
	// CORRECTED: Restored the full platform list.
	platformItems := []list.Item{item("Codeforces"), item("LeetCode"), item("AtCoder"), item("HackerRank"), item("CSES")}
	platformList := list.New(platformItems, subListDelegate, defaultWidth, len(platformItems)+verticalPadding)
	platformList.Title = "Choose a Platform"

	// Topic List
	topicItems := []list.Item{
		item("Ad-Hoc"), item("Binary Search"), item("Bit Manipulation"),
		item("Data Structures"), item("DP"), item("Game Theory"),
		item("Graphs"), item("Greedy"), item("Implementation"),
		item("Math"), item("Strings"), item("Two Pointers"),
	}
	topicList := list.New(topicItems, subListDelegate, defaultWidth, len(topicItems)+verticalPadding)
	topicList.Title = "Choose a Topic"

	// Difficulty List
	difficultyItems := []list.Item{item("Easy"), item("Medium"), item("Hard")}
	difficultyList := list.New(difficultyItems, subListDelegate, defaultWidth, len(difficultyItems)+verticalPadding)
	difficultyList.Title = "Choose a Difficulty"

	lists := []*list.Model{&mainMenu, &platformList, &topicList, &difficultyList}
	for _, l := range lists {
		l.SetShowStatusBar(false)
		l.SetShowFilter(false)
		l.SetShowPagination(false)
	}

	// --- Text Inputs ---
	timeInput := textinput.New()
	timeInput.Placeholder = "e.g., 45"
	timeInput.CharLimit = 3
	timeInput.Width = 20

	notesInput := textinput.New()
	notesInput.Placeholder = "e.g., Used two pointers"
	notesInput.CharLimit = 100
	notesInput.Width = 50

	return formModel{
		currentView: viewMain,
		mainMenu:    mainMenu,
		platforms:   platformList,
		topics:      topicList,
		difficulty:  difficultyList,
		timeInput:   timeInput,
		notesInput:  notesInput,
	}
}

func (m formModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.mainMenu.SetWidth(msg.Width)
		m.platforms.SetWidth(msg.Width)
		m.topics.SetWidth(msg.Width)
		m.difficulty.SetWidth(msg.Width)
		m.timeInput.Width = msg.Width - 4
		m.notesInput.Width = msg.Width - 4
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		// Handle state-specific key presses
		switch m.currentView {
		case viewMain:
			if msg.String() == "enter" {
				selectedItem := m.mainMenu.SelectedItem().(item)
				switch selectedItem {
				case "Platform":
					m.currentView = viewPlatform
				case "Topic":
					m.currentView = viewTopic
				case "Difficulty":
					m.currentView = viewDifficulty
				case "Time Spent":
					// CORRECTED: Focus the input when switching to its view.
					m.currentView = viewTime
					m.timeInput.Focus()
					m.notesInput.Blur()
				case "Notes":
					// CORRECTED: Focus the input when switching to its view.
					m.currentView = viewNotes
					m.notesInput.Focus()
					m.timeInput.Blur()
				case "Submit":
					return m, tea.Quit
				}
				return m, nil
			}

		case viewPlatform, viewTopic, viewDifficulty:
			if msg.String() == "tab" {
				m.currentView = viewMain
				return m, nil
			}

		case viewTime, viewNotes:
			if msg.String() == "enter" || msg.String() == "tab" {
				m.currentView = viewMain
				// Blur both inputs when returning to the main menu
				m.timeInput.Blur()
				m.notesInput.Blur()
				return m, nil
			}
		}
	}

	// Delegate messages to the active component
	switch m.currentView {
	case viewPlatform:
		m.platforms, cmd = m.platforms.Update(msg)
	case viewTopic:
		m.topics, cmd = m.topics.Update(msg)
	case viewDifficulty:
		m.difficulty, cmd = m.difficulty.Update(msg)
	case viewTime:
		m.timeInput, cmd = m.timeInput.Update(msg)
	case viewNotes:
		m.notesInput, cmd = m.notesInput.Update(msg)
	default: // viewMain
		m.mainMenu, cmd = m.mainMenu.Update(msg)
	}

	return m, cmd
}

func (m formModel) View() string {
	summary := fmt.Sprintf(
		"Platform: %s\nTopic: %s\nDifficulty: %s\nTime: %s\nNotes: %s",
		m.logEntry.Platform, m.logEntry.Topic, m.logEntry.Difficulty,
		m.timeInput.Value(), m.notesInput.Value(),
	)

	var currentView string
	switch m.currentView {
	case viewPlatform:
		currentView = m.platforms.View()
	case viewTopic:
		currentView = m.topics.View()
	case viewDifficulty:
		currentView = m.difficulty.View()
	case viewTime:
		currentView = "Time Spent (minutes):\n" + focusedStyle.Render(m.timeInput.View())
	case viewNotes:
		currentView = "Notes:\n" + focusedStyle.Render(m.notesInput.View())
	default: // viewMain
		currentView = m.mainMenu.View()
	}

	return summaryStyle.Render(summary) + "\n\n" + currentView
}
