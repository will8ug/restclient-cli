package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/will8ug/restclient-cli/internal/executor"
	"github.com/will8ug/restclient-cli/internal/parser"
)

type panel int

const (
	panelList panel = iota
	panelDetail
	panelResponse
)

type responseMsg struct {
	resp *executor.Response
}

type errMsg struct {
	err error
}

type Model struct {
	requests    []parser.Request
	fileName    string
	list        list.Model
	detail      viewport.Model
	response    viewport.Model
	spinner     spinner.Model
	activePanel panel
	loading     bool
	currentResp *executor.Response
	currentErr  error
	width       int
	height      int
	ready       bool
	showHelp    bool
}

func NewModel(requests []parser.Request, fileName string) Model {
	items := requestsToItems(requests)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	delegate := requestDelegate{}
	l := list.New(items, delegate, 0, 0)
	l.Title = "Requests"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	return Model{
		requests:    requests,
		fileName:    fileName,
		list:        l,
		spinner:     s,
		activePanel: panelList,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.resizePanels()
		m.ready = true

		if len(m.requests) > 0 {
			m.detail.SetContent(renderRequestDetail(m.requests[0]))
		}
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q":
			if m.activePanel == panelList {
				return m, tea.Quit
			}

		case "tab":
			m.activePanel = (m.activePanel + 1) % 3
			return m, nil

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "enter":
			if m.activePanel == panelList && !m.loading {
				idx := m.list.Index()
				if idx >= 0 && idx < len(m.requests) {
					m.loading = true
					m.currentResp = nil
					m.currentErr = nil
					return m, tea.Batch(
						m.spinner.Tick,
						executeRequest(m.requests[idx]),
					)
				}
			}
		}

		switch m.activePanel {
		case panelList:
			prevIdx := m.list.Index()
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			cmds = append(cmds, cmd)

			if m.list.Index() != prevIdx {
				idx := m.list.Index()
				if idx >= 0 && idx < len(m.requests) {
					m.detail.SetContent(renderRequestDetail(m.requests[idx]))
					m.detail.GotoTop()
				}
			}
			return m, tea.Batch(cmds...)

		case panelDetail:
			var cmd tea.Cmd
			m.detail, cmd = m.detail.Update(msg)
			return m, cmd

		case panelResponse:
			var cmd tea.Cmd
			m.response, cmd = m.response.Update(msg)
			return m, cmd
		}

	case responseMsg:
		m.loading = false
		m.currentResp = msg.resp
		m.currentErr = nil
		m.response.SetContent(renderResponse(msg.resp))
		m.response.GotoTop()
		return m, nil

	case errMsg:
		m.loading = false
		m.currentResp = nil
		m.currentErr = msg.err
		m.response.SetContent(renderError(msg.err))
		m.response.GotoTop()
		return m, nil

	case list.FilterMatchesMsg:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	if m.showHelp {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, renderHelp())
	}

	listWidth := m.width * 30 / 100
	rightWidth := m.width - listWidth - 4
	detailHeight := (m.height - 3) * 40 / 100
	responseHeight := m.height - 3 - detailHeight - 4

	listPanel := panelStyle(m.activePanel == panelList).
		Width(listWidth).
		Height(m.height - 3).
		Render(m.list.View())

	detailTitle := " Request "
	detailBorder := panelStyle(m.activePanel == panelDetail).
		Width(rightWidth).
		Height(detailHeight)
	detailPanel := detailBorder.Render(detailTitle + "\n" + m.detail.View())

	var responseContent string
	if m.loading {
		responseContent = fmt.Sprintf("\n  %s Sending request...", m.spinner.View())
	} else if m.currentResp != nil {
		responseContent = m.response.View()
	} else if m.currentErr != nil {
		responseContent = m.response.View()
	} else {
		responseContent = dimStyle.Render("\n  Press Enter to send the selected request")
	}

	responseTitle := " Response "
	responseBorder := panelStyle(m.activePanel == panelResponse).
		Width(rightWidth).
		Height(responseHeight)
	responsePanel := responseBorder.Render(responseTitle + "\n" + responseContent)

	rightColumn := lipgloss.JoinVertical(lipgloss.Left, detailPanel, responsePanel)
	content := lipgloss.JoinHorizontal(lipgloss.Top, listPanel, rightColumn)

	statusBar := m.renderStatusBar()

	return lipgloss.JoinVertical(lipgloss.Left, content, statusBar)
}

func (m Model) renderStatusBar() string {
	var left string
	if m.loading {
		idx := m.list.Index()
		method := ""
		url := ""
		if idx >= 0 && idx < len(m.requests) {
			method = m.requests[idx].Method
			url = m.requests[idx].URL
		}
		left = fmt.Sprintf(" %s Sending %s %s...", m.spinner.View(), method, url)
	} else {
		left = " ↑↓ navigate  enter send  tab switch  / filter  ? help  q quit"
	}

	right := fmt.Sprintf(" %s ", m.fileName)
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	bar := left + strings.Repeat(" ", gap) + right
	return statusBarStyle.Width(m.width).Render(bar)
}

func (m Model) resizePanels() Model {
	listWidth := m.width * 30 / 100
	rightWidth := m.width - listWidth - 6
	detailHeight := (m.height - 3) * 40 / 100
	responseHeight := m.height - 3 - detailHeight - 8

	m.list.SetSize(listWidth-2, m.height-5)

	m.detail = viewport.New(rightWidth, max(detailHeight-3, 1))
	m.response = viewport.New(rightWidth, max(responseHeight-3, 1))

	return m
}

func executeRequest(req parser.Request) tea.Cmd {
	return func() tea.Msg {
		resp, err := executor.Execute(executor.Request{
			Method:  req.Method,
			URL:     req.URL,
			Headers: req.Headers,
			Body:    req.Body,
		}, 30*time.Second)
		if err != nil {
			return errMsg{err: err}
		}
		return responseMsg{resp: resp}
	}
}

func panelStyle(active bool) lipgloss.Style {
	if active {
		return activePanelStyle
	}
	return inactivePanelStyle
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
