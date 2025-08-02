package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- VIEWS ---
// We use this to track which view is active.
type currentView int

const (
	viewMain currentView = iota
	viewPlatform
	viewTopic
	viewDifficulty
)

// --- STYLES ---
var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

// --- LIST ITEM & DELEGATE ---
type item string

func (i item) FilterValue() string { return string(i) }

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
	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string { return selectedItemStyle.Render("> " + strings.Join(s, " ")) }
	}
	fmt.Fprint(w, fn(str))
}

// --- MODEL ---
// The model now holds all lists and the current view state.
type formModel struct {
	currentView  currentView
	mainMenu     list.Model
	platformList list.Model
	topicList    list.Model
	difficultyList list.Model
}

func NewForm() formModel {
	const defaultWidth = 40
	const verticalPadding = 2 // For help text

	// --- Create all the lists ---
	delegate := itemDelegate{}

	// Main Menu
	mainMenuItems := []list.Item{item("Platform"), item("Topic"), item("Difficulty")}
	mainMenu := list.New(mainMenuItems, delegate, defaultWidth, len(mainMenuItems)+verticalPadding)

	// Platform List
	platformItems := []list.Item{item("Codeforces"), item("LeetCode"), item("AtCoder"), item("HackerRank"), item("CSES")}
	platformList := list.New(platformItems, delegate, defaultWidth, len(platformItems)+verticalPadding)
	platformList.Title = "Choose a Platform"

	// Topic List
	topicItems := []list.Item{
		item("Ad-Hoc"), item("Binary Search"), item("Bit Manipulation"),
		item("Data Structures"), item("DP"), item("Game Theory"),
		item("Graphs"), item("Greedy"), item("Implementation"),
		item("Math"), item("Strings"), item("Two Pointers"),
	}
	topicList := list.New(topicItems, delegate, defaultWidth, len(topicItems)+verticalPadding)
	topicList.Title = "Choose a Topic"

	// Difficulty List
	difficultyItems := []list.Item{item("Easy"), item("Medium"), item("Hard")}
	difficultyList := list.New(difficultyItems, delegate, defaultWidth, len(difficultyItems)+verticalPadding)
	difficultyList.Title = "Choose a Difficulty"

	// --- Configure all lists ---
	lists := []*list.Model{&mainMenu, &platformList, &topicList, &difficultyList}
	for _, l := range lists {
		l.SetShowStatusBar(false)
		l.SetShowFilter(false)
		l.SetShowPagination(false)
	}
	mainMenu.SetShowTitle(false) // Only main menu has no title

	return formModel{
		currentView:  viewMain,
		mainMenu:     mainMenu,
		platformList: platformList,
		topicList:    topicList,
		difficultyList: difficultyList,
	}
}

func (m formModel) Init() tea.Cmd {
	return nil
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Top-level messages (quit and resize)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.mainMenu.SetWidth(msg.Width)
		m.platformList.SetWidth(msg.Width)
		m.topicList.SetWidth(msg.Width)
		m.difficultyList.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		// --- Navigation Logic ---
		case "enter":
			// If on main menu, switch to a sub-list view.
			if m.currentView == viewMain {
				selectedItem := m.mainMenu.SelectedItem().(item)
				switch selectedItem {
				case "Platform":
					m.currentView = viewPlatform
				case "Topic":
					m.currentView = viewTopic
				case "Difficulty":
					m.currentView = viewDifficulty
				}
				return m, nil
			}
		case "tab":
			// If in a sub-list, return to the main menu.
			if m.currentView != viewMain {
				m.currentView = viewMain
				return m, nil
			}
		}
	}

	// Delegate messages to the list of the currently active view
	var cmd tea.Cmd
	switch m.currentView {
	case viewPlatform:
		m.platformList, cmd = m.platformList.Update(msg)
	case viewTopic:
		m.topicList, cmd = m.topicList.Update(msg)
	case viewDifficulty:
		m.difficultyList, cmd = m.difficultyList.Update(msg)
	default: // viewMain
		m.mainMenu, cmd = m.mainMenu.Update(msg)
	}
	return m, cmd
}

func (m formModel) View() string {
	// --- View Router ---
	// Render the view of the currently active list.
	switch m.currentView {
	case viewPlatform:
		return "\n" + m.platformList.View()
	case viewTopic:
		return "\n" + m.topicList.View()
	case viewDifficulty:
		return "\n" + m.difficultyList.View()
	default: // viewMain
		return "\n" + m.mainMenu.View()
	}
}