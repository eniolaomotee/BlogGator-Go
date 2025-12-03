package config

import (
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eniolaomotee/BlogGator-Go/internal/database"
	"github.com/pkg/browser"
	"strings"
)

// Styles
var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).MarginLeft(2)
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Background(lipgloss.Color("##3C3C3C"))
	//normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))
	descStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).MarginLeft(4).Width(80)
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).MarginTop(1).MarginLeft(2)
)

// Postitem reps a post in the list
type PostItem struct {
	title       string
	url         string
	description string
	feedName    string
	PublishedAt string
}

func (i PostItem) FilterValue() string { return i.title }
func (i PostItem) Title() string       { return i.title }
func (i PostItem) Description() string { return fmt.Sprintf("%s . %s", i.feedName, i.PublishedAt) }

// TUI model
type tuiModel struct {
	list  list.Model
	posts []database.GetPostsForUserSortedRow
	//	selected int
	viewing  bool
	quitting bool
}

func (m tuiModel) Init() tea.Cmd {
	return nil
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if !m.viewing {
				m.viewing = true
				return m, nil
			}

		case key.Matches(msg, key.NewBinding(key.WithKeys("o"))):
			// Open in browser
			if len(m.posts) > 0 {
				selectedPost := m.posts[m.list.Index()]
				browser.OpenURL(selectedPost.Url)
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "b"))):
			if m.viewing {
				m.viewing = false
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd

}

func (m tuiModel) View() string {
	if m.quitting {
		return "Thanks for using BlogGator! 👋\n"
	}

	if m.viewing && len(m.posts) > 0 {
		return m.viewPost()
	}

	return m.viewList()
}

func (m tuiModel) viewList() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("📰 BlogGator Posts"))
	s.WriteString("\n\n")
	s.WriteString(m.list.View())
	s.WriteString("\n")
	s.WriteString(helpStyle.Render("↑/↓: navigate • enter: view • o: open in browser • q: quit"))

	return s.String()
}

func (m tuiModel) viewPost() string {
	post := m.posts[m.list.Index()]

	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render(post.Title))
	s.WriteString("\n\n")

	// Metadata
	metadata := fmt.Sprintf("📡 %s  •  📅 %s",
		post.FeedName,
		post.PublishedAt.Time.Format("Mon Jan 2, 2006 3:04 PM"))
	s.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		MarginLeft(2).
		Render(metadata))
	s.WriteString("\n\n")

	// Description
	if post.Description.Valid && post.Description.String != "" {
		wrapped := wordWrap(post.Description.String, 80)
		s.WriteString(descStyle.Render(wrapped))
		s.WriteString("\n\n")
	}

	// URL
	s.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7D56F4")).
		MarginLeft(2).
		Render("🔗 " + post.Url))
	s.WriteString("\n\n")

	// Help
	s.WriteString(helpStyle.Render("o: open in browser • b/esc: back • q: quit"))

	return s.String()
}

// wordWrap wraps text to a specified width
func wordWrap(text string, width int) string {
	var result strings.Builder
	var line strings.Builder
	words := strings.Fields(text)

	for _, word := range words {
		if line.Len()+len(word)+1 > width {
			result.WriteString(line.String())
			result.WriteString("\n")
			line.Reset()
		}
		if line.Len() > 0 {
			line.WriteString(" ")
		}
		line.WriteString(word)
	}

	if line.Len() > 0 {
		result.WriteString(line.String())
	}

	return result.String()
}

// NewTUI creates a new TUI model
func NewTUI(posts []database.GetPostsForUserSortedRow) tuiModel {
	items := make([]list.Item, len(posts))
	for i, post := range posts {
		items[i] = PostItem{
			title:       post.Title,
			url:         post.Url,
			description: post.Description.String,
			feedName:    post.FeedName,
			PublishedAt: post.PublishedAt.Time.Format("Jan 2, 2006"),
		}
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedStyle
	delegate.Styles.SelectedDesc = selectedStyle

	l := list.New(items, delegate, 0, 0)
	l.Title = "Posts"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return tuiModel{
		list:  l,
		posts: posts,
	}
}
