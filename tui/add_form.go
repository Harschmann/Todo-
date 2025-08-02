package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// We need a helper type for our list items that satisfies the list.Item interface.
type item string

func (i item) FilterValue() string { return string(i) }

// addFormModel is the Bubbletea model for our "Add Log" form.
type addFormModel struct {
	// A slice of lists for our choice-based fields.
	lists []list.Model

	// A slice of text inputs for our form fields.
	inputs []textinput.Model

	// focused tracks which component is currently in focus.
	// 0-2 for lists (Platform, Topic, Difficulty), 3-4 for inputs (Time, Notes)
	focused int
}

// NewAddForm creates a new model with all the necessary fields initialized.
func NewAddForm() addFormModel {
	m := addFormModel{
		lists:  make([]list.Model, 3),    // For Platform, Topic, Difficulty
		inputs: make([]textinput.Model, 2), // For Time Spent & Notes
	}

	// ---- LISTS INITIALIZATION ----
	platforms := []list.Item{
		item("Codeforces"),
		item("LeetCode"),
		item("AtCoder"),
		item("HackerRank"),
		item("CSES"),
	}

	// common cp topics
	topics := []list.Item{
		item("Ad-Hoc"),
		item("Binary Search"),
		item("Bit Manipulation"),
		item("Combinatorics"),
		item("Data Structures"),
		item("DP (Dynamic Programming)"),
		item("Game Theory"),
		item("Geometry"),
		item("Graph Theory"),
		item("Greedy"),
		item("Implementation"),
		item("Math & Number Theory"),
		item("String Algorithms"),
		item("Two Pointers"),
	}

	difficulties := []list.Item{
		item("Easy"),
		item("Medium"),
		item("Hard"),
		item("Contest-Specific"), // For platform ratings like 1600, Div2C etc.
	}

	m.lists[0] = list.New(platforms, list.NewDefaultDelegate(), 0, 0)
	m.lists[1] = list.New(topics, list.NewDefaultDelegate(), 0, 0)
	m.lists[2] = list.New(difficulties, list.NewDefaultDelegate(), 0, 0)

	m.lists[0].Title = "Platform"
	m.lists[1].Title = "Topic"
	m.lists[2].Title = "Difficulty"

	// ---- TEXT INPUTS INITIALIZATION ----
	var t textinput.Model

	t = textinput.New()
	t.Placeholder = "Time in minutes (e.g., 60)"
	t.Focus() // Set the first field as focused
	m.inputs[0] = t

	t = textinput.New()
	t.Placeholder = "Any notes about the solution..."
	m.inputs[1] = t

	// ---- SET INITIAL FOCUS ----
	m.focused = 0 // Start by focusing the first list (Platform)

	return m
}

// Init is the first command that is run when the program starts.
func (m addFormModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update will be filled in the next steps.
func (m addFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// For now, we only handle quitting.
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

// View will be filled in the next steps.
func (m addFormModel) View() string {
	// This is just a placeholder for now.
	return "Form components initialized. Ready to build the View.\n\n(Press ctrl+c to quit)"
}