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
	viewQuestionID // ADDED
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
	currentView     currentView
	logEntry        model.Log
	mainMenu        list.Model
	platforms       list.Model
	topics          list.Model
	difficulty      list.Model
	questionIDInput textinput.Model // ADDED
	timeInput       textinput.Model
	notesInput      textinput.Model
}

func NewForm() formModel {
	const defaultWidth = 40
	const verticalPadding = 2

	subListDelegate := itemDelegate{}

	// UPDATED: Main Menu now includes Question ID
	mainMenuItems := []list.Item{
		item("Platform"), item("Topic"), item("Difficulty"),
		item("Question ID"), item("Time Spent"), item("Notes"),
		item("Submit"),
	}
	mainMenu := list.New(mainMenuItems, subListDelegate, defaultWidth, len(mainMenuItems)+verticalPadding)
	mainMenu.SetShowTitle(false)

	// Sub-lists...
	platformItems := []list.Item{item("Codeforces"), item("LeetCode"), item("AtCoder"), item("HackerRank"), item("CSES")}
	platformList := list.New(platformItems, subListDelegate, defaultWidth, len(platformItems)+verticalPadding)
	platformList.Title = "Choose a Platform"

	topicItems := []list.Item{
		item("Ad-Hoc"), item("Binary Search"), item("Bit Manipulation"),
		item("Data Structures"), item("DP"), item("Game Theory"),
		item("Graphs"), item("Greedy"), item("Implementation"),
		item("Math"), item("Strings"), item("Two Pointers"),
	}
	topicList := list.New(topicItems, subListDelegate, defaultWidth, len(topicItems)+verticalPadding)
	topicList.Title = "Choose a Topic"

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
	questionIDInput := textinput.New()
	questionIDInput.Placeholder = "e.g., 1337A or two-sum"
	questionIDInput.CharLimit = 40
	questionIDInput.Width = 40

	timeInput := textinput.New()
	timeInput.Placeholder = "e.g., 45"
	timeInput.CharLimit = 3
	timeInput.Width = 20

	notesInput := textinput.New()
	notesInput.Placeholder = "e.g., Used two pointers"
	notesInput.CharLimit = 100
	notesInput.Width = 50

	return formModel{
		currentView:     viewMain,
		mainMenu:        mainMenu,
		platforms:       platformList,
		topics:          topicList,
		difficulty:      difficultyList,
		questionIDInput: questionIDInput,
		timeInput:       timeInput,
		notesInput:      notesInput,
	}
}

func (m formModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := msg.Width - 4
		m.mainMenu.SetWidth(w)
		m.platforms.SetWidth(w)
		m.topics.SetWidth(w)
		m.difficulty.SetWidth(w)
		m.questionIDInput.Width = w
		m.timeInput.Width = w
		m.notesInput.Width = w
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		switch m.currentView {
		case viewMain:
			if msg.String() == "enter" {
				selected := m.mainMenu.SelectedItem().(item)
				switch selected {
				case "Platform":
					m.currentView = viewPlatform
				case "Topic":
					m.currentView = viewTopic
				case "Difficulty":
					m.currentView = viewDifficulty
				case "Question ID": // ADDED
					m.currentView = viewQuestionID
					m.questionIDInput.Focus()
					m.timeInput.Blur()
					m.notesInput.Blur()
				case "Time Spent":
					m.currentView = viewTime
					m.timeInput.Focus()
					m.questionIDInput.Blur()
					m.notesInput.Blur()
				case "Notes":
					m.currentView = viewNotes
					m.notesInput.Focus()
					m.questionIDInput.Blur()
					m.timeInput.Blur()
				case "Submit":
					return m, tea.Quit
				}
				return m, nil
			}

		case viewPlatform, viewTopic, viewDifficulty:
			if msg.String() == "enter" {
				switch m.currentView {
				case viewPlatform:
					m.logEntry.Platform = m.platforms.SelectedItem().(item).FilterValue()
				case viewTopic:
					m.logEntry.Topic = m.topics.SelectedItem().(item).FilterValue()
				case viewDifficulty:
					m.logEntry.Difficulty = m.difficulty.SelectedItem().(item).FilterValue()
				}
				m.mainMenu.CursorDown()
				m.currentView = viewMain
				return m, nil
			}
			if msg.String() == "tab" {
				m.currentView = viewMain
				return m, nil
			}

		case viewQuestionID, viewTime, viewNotes: // UPDATED
			if msg.String() == "enter" || msg.String() == "tab" {
				m.currentView = viewMain
				m.questionIDInput.Blur()
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
	case viewQuestionID: // ADDED
		m.questionIDInput, cmd = m.questionIDInput.Update(msg)
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
	// UPDATED: Summary now includes Question ID
	summary := fmt.Sprintf(
		"Platform: %s\nTopic: %s\nDifficulty: %s\nQuestion ID: %s\nTime: %s\nNotes: %s",
		m.logEntry.Platform, m.logEntry.Topic, m.logEntry.Difficulty,
		m.questionIDInput.Value(), m.timeInput.Value(), m.notesInput.Value(),
	)

	var currentView string
	switch m.currentView {
	case viewPlatform:
		currentView = m.platforms.View()
	case viewTopic:
		currentView = m.topics.View()
	case viewDifficulty:
		currentView = m.difficulty.View()
	case viewQuestionID: // ADDED
		currentView = "Question ID:\n" + focusedStyle.Render(m.questionIDInput.View())
	case viewTime:
		currentView = "Time Spent (minutes):\n" + focusedStyle.Render(m.timeInput.View())
	case viewNotes:
		currentView = "Notes:\n" + focusedStyle.Render(m.notesInput.View())
	default: // viewMain
		currentView = m.mainMenu.View()
	}

	return summaryStyle.Render(summary) + "\n\n" + currentView
}
