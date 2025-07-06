package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	choices  []string
	cursor   int
	loading  bool
	response string
	errMsg   string
}

func initialModel() model {
	return model{
		choices: []string{
			"📚 List Books",
			"➕ Add Book",
			"❌ Delete Book",
			"✏️ Edit Book",
		},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

type responseMsg string
type errorMsg string

func makeRequest(choice string) tea.Cmd {
	return func() tea.Msg {
		var (
			resp *http.Response
			err  error
		)

		client := &http.Client{}
		url := "http://localhost:8080/books"

		switch choice {
		case "📚 List Books":
			resp, err = http.Get(url + "/list")

		case "➕ Add Book":
			body := []byte(`{
				"Book_name": "The GO Programming Language",
				"Author": "Brian W. Kernighan and Dennis M. Ritchie",
				"ISBN": 0
			}`)
			resp, err = http.Post(url+"/add", "application/json", bytes.NewBuffer(body))

		case "❌ Delete Book":
			body := []byte(`{ "title": "The GO Programming Language" }`)
			req, _ := http.NewRequest("DELETE", url+"/delete", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err = client.Do(req)

		case "✏️  Edit Book":
			body := []byte(`{
				"title": "The C Programming Language",
				"field": "Book_name",
				"value": "The GoLANG Programming Language"
			}`)
			req, _ := http.NewRequest("PATCH", url+"/edit", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err = client.Do(req)

		default:
			return errorMsg("Invalid choice")
		}

		if err != nil {
			return errorMsg(fmt.Sprintf("Request error: %v", err))
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		return responseMsg(bodyBytes)
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

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
			m.loading = true
			m.response = ""
			m.errMsg = ""
			return m, makeRequest(m.choices[m.cursor])
		}

	case responseMsg:
		m.loading = false
		m.response = string(msg)
		m.errMsg = ""

	case errorMsg:
		m.loading = false
		m.errMsg = string(msg)
		m.response = ""
	}

	return m, nil
}

func (m model) View() string {
	s := "📖 Choose an action:\n\n"

	for i, choice := range m.choices {
		cursor := "  "
		if i == m.cursor {
			cursor = "👉"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	if m.loading {
		s += "\n⏳ Sending request..."
	} else if m.response != "" {
		s += fmt.Sprintf("\n✅ Response:\n%s", m.response)
	} else if m.errMsg != "" {
		s += fmt.Sprintf("\n❌ Error:\n%s", m.errMsg)
	}

	s += "\n\nPress 'q' to quit.\n"
	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("💥 Error:", err)
		os.Exit(1)
	}
}

