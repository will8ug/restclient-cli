package tui

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/will8ug/restclient-cli/internal/executor"
	"github.com/will8ug/restclient-cli/internal/parser"
)

type requestItem struct {
	index  int
	name   string
	method string
	url    string
}

func (i requestItem) FilterValue() string { return i.name + " " + i.url }
func (i requestItem) Title() string {
	badge := methodStyle(i.method).Render(fmt.Sprintf("%-7s", i.method))
	name := i.name
	if name == "" {
		name = i.url
	}
	return badge + " " + name
}
func (i requestItem) Description() string { return i.url }

type requestDelegate struct{}

func (d requestDelegate) Height() int                             { return 2 }
func (d requestDelegate) Spacing() int                            { return 0 }
func (d requestDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d requestDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(requestItem)
	if !ok {
		return
	}

	badge := methodStyle(item.method).Render(fmt.Sprintf("%-7s", item.method))
	name := item.name
	if name == "" {
		name = item.url
	}

	isSelected := index == m.Index()

	titleLine := fmt.Sprintf("%s %s", badge, name)
	descLine := dimStyle.Render("  " + item.url)

	if isSelected {
		titleLine = lipgloss.NewStyle().Bold(true).Render("> ") + titleLine
		descLine = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Render("  " + item.url)
	} else {
		titleLine = "  " + titleLine
	}

	fmt.Fprintf(w, "%s\n%s", titleLine, descLine)
}

func requestsToItems(requests []parser.Request) []list.Item {
	items := make([]list.Item, len(requests))
	for i, req := range requests {
		items[i] = requestItem{
			index:  i,
			name:   req.Name,
			method: req.Method,
			url:    req.URL,
		}
	}
	return items
}

func renderRequestDetail(req parser.Request) string {
	var b strings.Builder

	b.WriteString(methodStyle(req.Method).Render(req.Method))
	b.WriteString(" ")
	b.WriteString(req.URL)
	b.WriteString("\n")

	if len(req.Headers) > 0 {
		b.WriteString("\n")
		headerKeys := make([]string, 0, len(req.Headers))
		for k := range req.Headers {
			headerKeys = append(headerKeys, k)
		}
		sort.Strings(headerKeys)
		for _, k := range headerKeys {
			b.WriteString(dimStyle.Render(k+": ") + req.Headers[k] + "\n")
		}
	}

	if req.Body != "" {
		b.WriteString("\n")
		b.WriteString(formatBody(req.Body, req.Headers["Content-Type"]))
	}

	return b.String()
}

func renderResponse(resp *executor.Response) string {
	var b strings.Builder

	statusLine := fmt.Sprintf("%s  (%dms)", resp.Status, resp.Duration.Milliseconds())
	b.WriteString(statusStyle(resp.StatusCode).Render(statusLine))
	b.WriteString("\n\n")

	if len(resp.Headers) > 0 {
		headerKeys := make([]string, 0, len(resp.Headers))
		for k := range resp.Headers {
			headerKeys = append(headerKeys, k)
		}
		sort.Strings(headerKeys)
		for _, k := range headerKeys {
			for _, v := range resp.Headers[k] {
				b.WriteString(dimStyle.Render(k+": ") + v + "\n")
			}
		}
		b.WriteString("\n")
	}

	contentType := resp.Headers.Get("Content-Type")
	b.WriteString(formatBody(resp.Body, contentType))

	return b.String()
}

func renderError(err error) string {
	return errorTextStyle.Render("Error: " + err.Error())
}

func formatBody(body string, contentType string) string {
	if strings.Contains(contentType, "json") || looksLikeJSON(body) {
		if formatted, err := prettyJSON(body); err == nil {
			return formatted
		}
	}
	return body
}

func looksLikeJSON(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))
}

func prettyJSON(s string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		return "", err
	}
	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(pretty), nil
}

func renderHelp() string {
	helpContent := `
  Keyboard Shortcuts

  ↑/↓ or j/k     Navigate request list / scroll panels vertically
  ←/→            Scroll panels horizontally
  Shift+←/→      Scroll panels horizontally (fine, 3 columns)
  Home/End       Reset horizontal scroll position
  Enter          Send selected request
  Tab            Switch panel focus
  /              Filter requests
  ?              Toggle this help
  q              Quit
  Ctrl+C         Force quit
`
	style := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 3).
		Align(lipgloss.Center)

	return style.Render(helpContent)
}
