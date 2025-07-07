package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	responseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Border(lipgloss.RoundedBorder()).
			Padding(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Border(lipgloss.RoundedBorder()).
			Padding(1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1)
)

// Book represents a book structure
type Book struct {
	BookName string `json:"Book_name"`
	Author   string `json:"Author"`
	ISBN     int    `json:"ISBN"`
}

// EditRequest represents an edit request
type EditRequest struct {
	Title string `json:"title"`
	Field string `json:"field"`
	Value string `json:"value"`
}

// DeleteRequest represents a delete request
type DeleteRequest struct {
	Title string `json:"title"`
}

// State represents the current state of the application
type State int

const (
	StateMenu State = iota
	StateAddBook
	StateDeleteBook
	StateEditBook
	StateLoading
	StateShowResponse
)

// Model represents the state of the application
type model struct {
	state    State
	choices  []string
	cursor   int
	response string
	errMsg   string

	// Input fields for adding books
	bookNameInput textinput.Model
	authorInput   textinput.Model
	isbnInput     textinput.Model

	// Input fields for delete/edit
	titleInput textinput.Model
	fieldInput textinput.Model
	valueInput textinput.Model

	// Current input focus
	currentInput int
	maxInputs    int
}

// initialModel initializes the model with default values
func initialModel() model {
	// Initialize text inputs
	bookNameInput := textinput.New()
	bookNameInput.Placeholder = "Enter book name"
	bookNameInput.Focus()
	bookNameInput.CharLimit = 100
	bookNameInput.Width = 50

	authorInput := textinput.New()
	authorInput.Placeholder = "Enter author name"
	authorInput.CharLimit = 100
	authorInput.Width = 50

	isbnInput := textinput.New()
	isbnInput.Placeholder = "Enter ISBN (numbers only)"
	isbnInput.CharLimit = 20
	isbnInput.Width = 50

	titleInput := textinput.New()
	titleInput.Placeholder = "Enter book title"
	titleInput.CharLimit = 100
	titleInput.Width = 50

	fieldInput := textinput.New()
	fieldInput.Placeholder = "Enter field to edit (Book_name, Author, ISBN)"
	fieldInput.CharLimit = 50
	fieldInput.Width = 50

	valueInput := textinput.New()
	valueInput.Placeholder = "Enter new value to update"
	valueInput.CharLimit = 100
	valueInput.Width = 50

	return model{
		state: StateMenu,
		choices: []string{
			"List Books",
			"Add Book",
			"Delete Book",
			"Edit Book",
		},
		bookNameInput: bookNameInput,
		authorInput:   authorInput,
		isbnInput:     isbnInput,
		titleInput:    titleInput,
		fieldInput:    fieldInput,
		valueInput:    valueInput,
	}
}

// initializes tea model
func (m model) Init() tea.Cmd {
	return textinput.Blink
}

// Messages
type responseMsg string
type errorMsg string

// contains the logic for making a list request
func makeListRequest() tea.Cmd {
	return func() tea.Msg {
		resp, err := http.Get("http://localhost:8080/books/list")
		if err != nil {
			return errorMsg(fmt.Sprintf("Request error: %v", err))
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errorMsg(fmt.Sprintf("Read error: %v", err))
		}

		return responseMsg(string(bodyBytes))
	}
}

// contains the logic for making an add request
func makeAddRequest(book Book) tea.Cmd {
	return func() tea.Msg {
		jsonData, err := json.Marshal(book)
		if err != nil {
			return errorMsg(fmt.Sprintf("JSON marshal error: %v", err))
		}

		resp, err := http.Post("http://localhost:8080/books/add", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return errorMsg(fmt.Sprintf("Request error: %v", err))
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errorMsg(fmt.Sprintf("Read error: %v", err))
		}

		return responseMsg(string(bodyBytes))
	}
}

// contains the logic for making a delete request
func makeDeleteRequest(title string) tea.Cmd {
	return func() tea.Msg {
		deleteReq := DeleteRequest{Title: title}
		jsonData, err := json.Marshal(deleteReq)
		if err != nil {
			return errorMsg(fmt.Sprintf("JSON marshal error: %v", err))
		}

		client := &http.Client{}
		req, err := http.NewRequest("DELETE", "http://localhost:8080/books/delete", bytes.NewBuffer(jsonData))
		if err != nil {
			return errorMsg(fmt.Sprintf("Request creation error: %v", err))
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return errorMsg(fmt.Sprintf("Request error: %v", err))
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errorMsg(fmt.Sprintf("Read error: %v", err))
		}

		return responseMsg(string(bodyBytes))
	}
}

// contains the logic for making an edit request
func makeEditRequest(title, field, value string) tea.Cmd {
	return func() tea.Msg {
		editReq := EditRequest{
			Title: title,
			Field: field,
			Value: value,
		}
		jsonData, err := json.Marshal(editReq)
		if err != nil {
			return errorMsg(fmt.Sprintf("JSON marshal error: %v", err))
		}

		client := &http.Client{}
		req, err := http.NewRequest("PATCH", "http://localhost:8080/books/edit", bytes.NewBuffer(jsonData))
		if err != nil {
			return errorMsg(fmt.Sprintf("Request creation error: %v", err))
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return errorMsg(fmt.Sprintf("Request error: %v", err))
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errorMsg(fmt.Sprintf("Read error: %v", err))
		}

		return responseMsg(string(bodyBytes))
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Always update text inputs first so they can receive keystrokes
	m.bookNameInput, cmd = m.bookNameInput.Update(msg)
	cmds = append(cmds, cmd)

	m.authorInput, cmd = m.authorInput.Update(msg)
	cmds = append(cmds, cmd)

	m.isbnInput, cmd = m.isbnInput.Update(msg)
	cmds = append(cmds, cmd)

	m.titleInput, cmd = m.titleInput.Update(msg)
	cmds = append(cmds, cmd)

	m.fieldInput, cmd = m.fieldInput.Update(msg)
	cmds = append(cmds, cmd)

	m.valueInput, cmd = m.valueInput.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case StateMenu:
			newModel, newCmd := m.updateMenu(msg)
			cmds = append(cmds, newCmd)
			return newModel, tea.Batch(cmds...)
		case StateAddBook:
			newModel, newCmd := m.updateAddBook(msg)
			cmds = append(cmds, newCmd)
			return newModel, tea.Batch(cmds...)
		case StateDeleteBook:
			newModel, newCmd := m.updateDeleteBook(msg)
			cmds = append(cmds, newCmd)
			return newModel, tea.Batch(cmds...)
		case StateEditBook:
			newModel, newCmd := m.updateEditBook(msg)
			cmds = append(cmds, newCmd)
			return newModel, tea.Batch(cmds...)
		case StateLoading:
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, tea.Batch(cmds...)
		case StateShowResponse:
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			if msg.String() == "enter" || msg.String() == "esc" {
				m.state = StateMenu
				m.response = ""
				m.errMsg = ""
				return m, tea.Batch(cmds...)
			}
			return m, tea.Batch(cmds...)
		}

	case responseMsg:
		m.state = StateShowResponse
		m.response = string(msg)
		m.errMsg = ""
		return m, tea.Batch(cmds...)

	case errorMsg:
		m.state = StateShowResponse
		m.errMsg = string(msg)
		m.response = ""
		return m, tea.Batch(cmds...)
	}

	return m, tea.Batch(cmds...)
}

func (m model) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.choices)-1 {
			m.cursor++
		}
	case "enter":
		switch m.choices[m.cursor] {
		case "List Books":
			m.state = StateLoading
			return m, makeListRequest()
		case "Add Book":
			m.state = StateAddBook
			m.currentInput = 0
			m.maxInputs = 3
			m.bookNameInput.Focus()
			m.authorInput.Blur()
			m.isbnInput.Blur()
			m.bookNameInput.SetValue("")
			m.authorInput.SetValue("")
			m.isbnInput.SetValue("")
			return m, textinput.Blink
		case "Delete Book":
			m.state = StateDeleteBook
			m.titleInput.Focus()
			m.titleInput.SetValue("")
			return m, textinput.Blink
		case "Edit Book":
			m.state = StateEditBook
			m.currentInput = 0
			m.maxInputs = 3
			m.titleInput.Focus()
			m.fieldInput.Blur()
			m.valueInput.Blur()
			m.titleInput.SetValue("")
			m.fieldInput.SetValue("")
			m.valueInput.SetValue("")
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m model) updateAddBook(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = StateMenu
		return m, nil
	case "tab":
		m.currentInput++
		if m.currentInput >= m.maxInputs {
			m.currentInput = 0
		}
		return m.updateInputFocus(), nil
	case "shift+tab":
		m.currentInput--
		if m.currentInput < 0 {
			m.currentInput = m.maxInputs - 1
		}
		return m.updateInputFocus(), nil
	case "ctrl+n":
		// Alternative way to move to next input
		m.currentInput++
		if m.currentInput >= m.maxInputs {
			m.currentInput = 0
		}
		return m.updateInputFocus(), nil
	case "ctrl+p":
		// Alternative way to move to previous input
		m.currentInput--
		if m.currentInput < 0 {
			m.currentInput = m.maxInputs - 1
		}
		return m.updateInputFocus(), nil
	case "ctrl+s":
		// Submit the form with Ctrl+S
		isbn, err := strconv.Atoi(m.isbnInput.Value())
		if err != nil && m.isbnInput.Value() != "" {
			return m, func() tea.Msg { return errorMsg("ISBN must be a valid number") }
		}

		book := Book{
			BookName: m.bookNameInput.Value(),
			Author:   m.authorInput.Value(),
			ISBN:     isbn,
		}

		if book.BookName == "" || book.Author == "" {
			return m, func() tea.Msg { return errorMsg("Book name and author are required") }
		}

		m.state = StateLoading
		return m, makeAddRequest(book)
	}
	return m, nil
}

func (m model) updateDeleteBook(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = StateMenu
		return m, nil
	case "ctrl+s":
		// Submit with Ctrl+S
		if m.titleInput.Value() == "" {
			return m, func() tea.Msg { return errorMsg("Title is required") }
		}
		m.state = StateLoading
		return m, makeDeleteRequest(m.titleInput.Value())
	}
	return m, nil
}

func (m model) updateEditBook(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = StateMenu
		return m, nil
	case "tab":
		m.currentInput++
		if m.currentInput >= m.maxInputs {
			m.currentInput = 0
		}
		return m.updateEditInputFocus(), nil
	case "shift+tab":
		m.currentInput--
		if m.currentInput < 0 {
			m.currentInput = m.maxInputs - 1
		}
		return m.updateEditInputFocus(), nil
	case "ctrl+n":
		// Alternative way to move to next input
		m.currentInput++
		if m.currentInput >= m.maxInputs {
			m.currentInput = 0
		}
		return m.updateEditInputFocus(), nil
	case "ctrl+p":
		// Alternative way to move to previous input
		m.currentInput--
		if m.currentInput < 0 {
			m.currentInput = m.maxInputs - 1
		}
		return m.updateEditInputFocus(), nil
	case "ctrl+s":
		// Submit the form with Ctrl+S
		if m.titleInput.Value() == "" || m.fieldInput.Value() == "" || m.valueInput.Value() == "" {
			return m, func() tea.Msg { return errorMsg("All fields are required") }
		}

		validFields := []string{"Book_name", "Author", "ISBN"}
		fieldValid := false
		for _, field := range validFields {
			if m.fieldInput.Value() == field {
				fieldValid = true
				break
			}
		}

		if !fieldValid {
			return m, func() tea.Msg { return errorMsg("Field must be one of: Book_name, Author, ISBN") }
		}

		m.state = StateLoading
		return m, makeEditRequest(m.titleInput.Value(), m.fieldInput.Value(), m.valueInput.Value())
	}
	return m, nil
}

func (m model) updateInputFocus() model {
	m.bookNameInput.Blur()
	m.authorInput.Blur()
	m.isbnInput.Blur()

	switch m.currentInput {
	case 0:
		m.bookNameInput.Focus()
	case 1:
		m.authorInput.Focus()
	case 2:
		m.isbnInput.Focus()
	}

	return m
}

func (m model) updateEditInputFocus() model {
	m.titleInput.Blur()
	m.fieldInput.Blur()
	m.valueInput.Blur()

	switch m.currentInput {
	case 0:
		m.titleInput.Focus()
	case 1:
		m.fieldInput.Focus()
	case 2:
		m.valueInput.Focus()
	}

	return m
}

func (m model) View() string {
	switch m.state {
	case StateMenu:
		return m.viewMenu()
	case StateAddBook:
		return m.viewAddBook()
	case StateDeleteBook:
		return m.viewDeleteBook()
	case StateEditBook:
		return m.viewEditBook()
	case StateLoading:
		return m.viewLoading()
	case StateShowResponse:
		return m.viewResponse()
	}
	return ""
}

func (m model) viewMenu() string {
	s := titleStyle.Render("📚 Book Management System") + "\n\n"
	s += "Choose an action:\n\n"

	for i, choice := range m.choices {
		cursor := "  "
		if i == m.cursor {
			cursor = "▶ "
			choice = selectedStyle.Render(choice)
		}
		s += fmt.Sprintf("%s%s\n", cursor, choice)
	}

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("↑/↓: navigate • enter: select • q: quit")
	return s
}

func (m model) viewAddBook() string {
	s := titleStyle.Render("Add New Book") + "\n\n"

	s += inputStyle.Render("Book Name:\n"+m.bookNameInput.View()) + "\n\n"
	s += inputStyle.Render("Author:\n"+m.authorInput.View()) + "\n\n"
	s += inputStyle.Render("ISBN:\n"+m.isbnInput.View()) + "\n\n"

	s += lipgloss.NewStyle().Faint(true).Render("tab: next field • shift+tab: prev field • ctrl+s: submit • esc: back • ctrl+c: quit")
	return s
}

func (m model) viewDeleteBook() string {
	s := titleStyle.Render("Delete Book") + "\n\n"

	s += inputStyle.Render("Book Title:\n"+m.titleInput.View()) + "\n\n"

	s += lipgloss.NewStyle().Faint(true).Render("ctrl+s: delete • esc: back • ctrl+c: quit")
	return s
}

func (m model) viewEditBook() string {
	s := titleStyle.Render("Edit Book") + "\n\n"

	s += inputStyle.Render("Book Title:\n"+m.titleInput.View()) + "\n\n"
	s += inputStyle.Render("Field to Edit (Book_name/Author/ISBN):\n"+m.fieldInput.View()) + "\n\n"
	s += inputStyle.Render("New Value:\n"+m.valueInput.View()) + "\n\n"

	s += lipgloss.NewStyle().Faint(true).Render("tab: next field • shift+tab: prev field • ctrl+s: submit • esc: back • ctrl+c: quit")
	return s
}

func (m model) viewLoading() string {
	s := titleStyle.Render("📚 Book Management System") + "\n\n"
	s += "⏳ Processing request...\n\n"
	s += lipgloss.NewStyle().Faint(true).Render("q: quit")
	return s
}

func (m model) viewResponse() string {
	s := titleStyle.Render("📚 Book Management System") + "\n\n"

	if m.response != "" {
		s += "✅ " + lipgloss.NewStyle().Bold(true).Render("Response:") + "\n\n"
		s += responseStyle.Render(m.formatResponse(m.response)) + "\n\n"
	}

	if m.errMsg != "" {
		s += "❌ " + lipgloss.NewStyle().Bold(true).Render("Error:") + "\n\n"
		s += errorStyle.Render(m.errMsg) + "\n\n"
	}

	s += lipgloss.NewStyle().Faint(true).Render("enter: back to menu • q: quit")
	return s
}

func (m model) formatResponse(response string) string {
	// Try to format JSON response nicely
	var jsonData interface{}
	if err := json.Unmarshal([]byte(response), &jsonData); err == nil {
		if formatted, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
			return string(formatted)
		}
	}

	// If not JSON, return as is but with some basic formatting
	return strings.TrimSpace(response)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("💥 Error: %v\n", err)
		os.Exit(1)
	}
}
