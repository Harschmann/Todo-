package tui

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Harschmann/Todo-/db"
	"github.com/Harschmann/Todo-/model"
	"github.com/charmbracelet/bubbles/key"
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
	viewQuestionID
	viewTime
	viewNotes
	viewLogs
	viewLogDetails
	viewConfirmDelete
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
	descriptionStyle = lipgloss.NewStyle().Faint(true)
	detailsStyle     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	errorStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

// --- LIST ITEMS & DELEGATES ---
type menuItem string

func (i menuItem) FilterValue() string { return string(i) }

type menuItemDelegate struct{}

func (d menuItemDelegate) Height() int                             { return 1 }
func (d menuItemDelegate) Spacing() int                            { return 0 }
func (d menuItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d menuItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(menuItem)
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

type logListItem model.Log

func (l logListItem) FilterValue() string { return l.QuestionID }
func (l logListItem) Title() string       { return l.QuestionID }
func (l logListItem) Description() string {
	notes := l.Notes
	if len(notes) > 30 {
		notes = notes[:27] + "..."
	}
	return fmt.Sprintf("%s | %s | %s | %s | Notes: %s",
		l.Platform, l.Topic, l.Difficulty, l.Date.Format("2006-01-02"), notes)
}

// --- MODEL ---
type formModel struct {
	currentView     currentView
	logEntry        model.Log
	selectedLog     model.Log
	mainMenu        list.Model
	platforms       list.Model
	topics          list.Model
	difficulty      list.Model
	logsList        list.Model
	questionIDInput textinput.Model
	timeInput       textinput.Model
	notesInput      textinput.Model
	errorMsg        string
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func NewForm() formModel {
	const defaultWidth = 40
	const listPadding = 2

	mainMenuItems := []list.Item{
		menuItem("Platform"), menuItem("Topic"), menuItem("Difficulty"),
		menuItem("Question ID"), menuItem("Time Spent"), menuItem("Notes"),
		menuItem("Submit & Add Another"),
		menuItem("View Logs"),
	}
	mainMenu := list.New(mainMenuItems, menuItemDelegate{}, defaultWidth, len(mainMenuItems)+listPadding)
	mainMenu.SetShowTitle(false)

	subListDelegate := menuItemDelegate{}
	platformItems := []list.Item{menuItem("Codeforces"), menuItem("LeetCode"), menuItem("AtCoder"), menuItem("HackerRank"), menuItem("CSES")}
	platformList := list.New(platformItems, subListDelegate, defaultWidth, len(platformItems)+listPadding)
	platformList.Title = "Choose a Platform"
	topicItems := []list.Item{
		menuItem("Ad-Hoc"), menuItem("Binary Search"), menuItem("Bit Manipulation"),
		menuItem("Data Structures"), menuItem("DP"), menuItem("Game Theory"),
		menuItem("Graphs"), menuItem("Greedy"), menuItem("Implementation"),
		menuItem("Math"), menuItem("Strings"), menuItem("Two Pointers"),
	}
	topicList := list.New(topicItems, subListDelegate, defaultWidth, len(topicItems)+listPadding)
	topicList.Title = "Choose a Topic"
	difficultyItems := []list.Item{menuItem("Easy"), menuItem("Medium"), menuItem("Hard")}
	difficultyList := list.New(difficultyItems, subListDelegate, defaultWidth, len(difficultyItems)+listPadding)
	difficultyList.Title = "Choose a Difficulty"

	allLogs, err := db.GetAllLogs()
	if err != nil {
		allLogs = []model.Log{}
	}
	logItems := make([]list.Item, len(allLogs))
	for i, lg := range allLogs {
		logItems[i] = logListItem(lg)
	}
	logDelegate := list.NewDefaultDelegate()
	logDelegate.Styles.SelectedTitle = selectedItemStyle
	logDelegate.Styles.SelectedDesc = descriptionStyle
	logDelegate.Styles.NormalDesc = descriptionStyle
	logDelegate.Styles.DimmedDesc = descriptionStyle
	logsList := list.New(logItems, logDelegate, defaultWidth, 14)
	logsList.Title = "Saved Logs"

	// CORRECTED: Simplified and corrected way to add help text
	logsList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "delete")),
		}
	}

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

	m := formModel{
		currentView:     viewMain,
		mainMenu:        mainMenu,
		platforms:       platformList,
		topics:          topicList,
		difficulty:      difficultyList,
		logsList:        logsList,
		questionIDInput: questionIDInput,
		timeInput:       timeInput,
		notesInput:      notesInput,
	}

	lists := []*list.Model{&m.mainMenu, &m.platforms, &m.topics, &m.difficulty, &m.logsList}
	for _, l := range lists {
		l.SetShowStatusBar(false)
		l.SetShowFilter(false)
		l.SetShowPagination(false)
	}

	return m
}

func (m formModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case clearErrorMsg:
		m.errorMsg = ""
		return m, nil
	case tea.WindowSizeMsg:
		w := msg.Width - 4
		h := msg.Height - 8
		m.mainMenu.SetWidth(w)
		m.platforms.SetWidth(w)
		m.topics.SetWidth(w)
		m.difficulty.SetWidth(w)
		m.logsList.SetSize(w, h)
		m.questionIDInput.Width = w
		m.timeInput.Width = w
		m.notesInput.Width = w
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "ctrl+q" {
			return m, tea.Quit
		}

		switch m.currentView {
		case viewMain:
			if msg.String() == "enter" {
				selected := m.mainMenu.SelectedItem().(menuItem)
				switch selected {
				case "Platform":
					m.currentView = viewPlatform
				case "Topic":
					m.currentView = viewTopic
				case "Difficulty":
					m.currentView = viewDifficulty
				case "Question ID":
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
				case "Submit & Add Another":
					m.logEntry.QuestionID = m.questionIDInput.Value()
					t, _ := strconv.Atoi(m.timeInput.Value())
					m.logEntry.TimeSpent = t
					m.logEntry.Notes = m.notesInput.Value()
					if m.logEntry.Platform == "" || m.logEntry.Topic == "" || m.logEntry.Difficulty == "" || m.logEntry.QuestionID == "" {
						m.errorMsg = "Error: Please fill out all fields before submitting."
						return m, clearErrorAfter(2 * time.Second)
					}
					if err := db.SaveLog(&m.logEntry); err != nil {
						log.Fatal(err)
					}
					return NewForm(), nil
				case "View Logs":
					m.currentView = viewLogs
				}
				return m, nil
			}

		case viewPlatform, viewTopic, viewDifficulty:
			if msg.String() == "enter" {
				switch m.currentView {
				case viewPlatform:
					m.logEntry.Platform = m.platforms.SelectedItem().(menuItem).FilterValue()
				case viewTopic:
					m.logEntry.Topic = m.topics.SelectedItem().(menuItem).FilterValue()
				case viewDifficulty:
					m.logEntry.Difficulty = m.difficulty.SelectedItem().(menuItem).FilterValue()
				}
				m.mainMenu.CursorDown()
				m.currentView = viewMain
				return m, nil
			}
			if msg.String() == "tab" {
				m.currentView = viewMain
				return m, nil
			}

		case viewLogs:
			switch msg.String() {
			case "enter":
				selected := m.logsList.SelectedItem().(logListItem)
				m.selectedLog = model.Log(selected)
				m.currentView = viewLogDetails
				return m, nil
			case "ctrl+d":
				selected := m.logsList.SelectedItem().(logListItem)
				m.selectedLog = model.Log(selected)
				m.currentView = viewConfirmDelete
				return m, nil
			case "tab":
				m.currentView = viewMain
				return m, nil
			}

		case viewLogDetails:
			if msg.String() != "" {
				m.currentView = viewLogs
				return m, nil
			}

		case viewConfirmDelete:
			switch msg.String() {
			case "y", "Y":
				if err := db.DeleteLog(m.selectedLog.Date); err != nil {
					log.Fatal(err)
				}
				freshModel := NewForm()
				freshModel.currentView = viewLogs
				return freshModel, tea.ClearScreen
			case "n", "N", "esc":
				m.currentView = viewLogs
				return m, nil
			}

		case viewQuestionID, viewTime, viewNotes:
			if msg.String() == "enter" || msg.String() == "tab" {
				switch m.currentView {
				case viewQuestionID:
					m.logEntry.QuestionID = m.questionIDInput.Value()
				case viewTime:
					t, _ := strconv.Atoi(m.timeInput.Value())
					m.logEntry.TimeSpent = t
				case viewNotes:
					m.logEntry.Notes = m.notesInput.Value()
				}
				m.mainMenu.CursorDown()
				m.currentView = viewMain
				m.questionIDInput.Blur()
				m.timeInput.Blur()
				m.notesInput.Blur()
				return m, nil
			}
		}
	}

	// Delegate messages
	switch m.currentView {
	case viewPlatform:
		m.platforms, cmd = m.platforms.Update(msg)
	case viewTopic:
		m.topics, cmd = m.topics.Update(msg)
	case viewDifficulty:
		m.difficulty, cmd = m.difficulty.Update(msg)
	case viewQuestionID:
		m.questionIDInput, cmd = m.questionIDInput.Update(msg)
	case viewTime:
		m.timeInput, cmd = m.timeInput.Update(msg)
	case viewNotes:
		m.notesInput, cmd = m.notesInput.Update(msg)
	case viewLogs:
		m.logsList, cmd = m.logsList.Update(msg)
	default: // viewMain
		m.mainMenu, cmd = m.mainMenu.Update(msg)
	}

	return m, cmd
}

func (m formModel) View() string {
	var b strings.Builder

	switch m.currentView {
	case viewLogs:
		b.WriteString(m.logsList.View())
	case viewLogDetails:
		details := fmt.Sprintf(
			"Question ID: %s\nPlatform:    %s\nTopic:       %s\nDifficulty:  %s\nDate:        %s\nTime Spent:  %d mins\n\nNotes:\n%s",
			m.selectedLog.QuestionID, m.selectedLog.Platform, m.selectedLog.Topic, m.selectedLog.Difficulty,
			m.selectedLog.Date.Format("2006-01-02"), m.selectedLog.TimeSpent, m.selectedLog.Notes,
		)
		b.WriteString(detailsStyle.Render(details) + "\n\n(Press any key to return to list)")

	case viewConfirmDelete:
		question := fmt.Sprintf("Are you sure you want to delete this log?\n\n%s\n%s",
			m.selectedLog.QuestionID,
			m.selectedLog.Platform,
		)
		b.WriteString(detailsStyle.Render(question) + "\n\n(y/n)")
	default:
		summary := fmt.Sprintf(
			"Platform: %s\nTopic: %s\nDifficulty: %s\nQuestion ID: %s\nTime: %d\nNotes: %s",
			m.logEntry.Platform, m.logEntry.Topic, m.logEntry.Difficulty,
			m.logEntry.QuestionID, m.logEntry.TimeSpent, m.logEntry.Notes,
		)
		var currentInputView string
		switch m.currentView {
		case viewPlatform:
			currentInputView = m.platforms.View()
		case viewTopic:
			currentInputView = m.topics.View()
		case viewDifficulty:
			currentInputView = m.difficulty.View()
		case viewQuestionID:
			currentInputView = "Question ID:\n" + focusedStyle.Render(m.questionIDInput.View())
		case viewTime:
			currentInputView = "Time Spent (minutes):\n" + focusedStyle.Render(m.timeInput.View())
		case viewNotes:
			// CORRECTED: The typo is fixed here.
			currentInputView = "Notes:\n" + focusedStyle.Render(m.notesInput.View())
		default: // viewMain
			currentInputView = m.mainMenu.View()
		}
		b.WriteString(summaryStyle.Render(summary) + "\n\n" + currentInputView)
	}

	if m.errorMsg != "" {
		b.WriteString("\n\n" + errorStyle.Render(m.errorMsg))
	}

	return b.String()
}
