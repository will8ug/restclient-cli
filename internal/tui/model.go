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

const (
	listWidthPercent    = 30 // percentage of terminal width for request list
	detailHeightPercent = 30 // percentage of available height for detail panel
	statusBarHeight     = 3  // height reserved for bottom status bar (1 line + padding)
	panelHorizInset     = 4  // left+right borders and spacing between panels
	panelBorderHeight   = 2  // top+bottom border lines per panel (lipgloss RoundedBorder)
	numPanels           = 3  // total number of panels
	panelContentXPad    = 2  // horizontal padding for views inside bordered panels
	panelContentYPad    = 3  // vertical padding for views inside bordered panels (title + borders)
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

		idx := m.list.Index()
		if idx >= 0 && idx < len(m.requests) {
			m.detail.SetContent(renderRequestDetail(m.requests[idx]))
		}
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			if msg.String() == "enter" && strings.TrimSpace(m.list.FilterInput.Value()) == "" {
				m.list.ResetFilter()
			}
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
			m.activePanel = (m.activePanel + 1) % numPanels
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

		case "shift+left":
			switch m.activePanel {
			case panelDetail:
				m.detail.ScrollLeft(3)
				return m, nil
			case panelResponse:
				m.response.ScrollLeft(3)
				return m, nil
			}

		case "shift+right":
			switch m.activePanel {
			case panelDetail:
				m.detail.ScrollRight(3)
				return m, nil
			case panelResponse:
				m.response.ScrollRight(3)
				return m, nil
			}

		case "home":
			switch m.activePanel {
			case panelDetail:
				m.detail.SetXOffset(0)
				return m, nil
			case panelResponse:
				m.response.SetXOffset(0)
				return m, nil
			}

		case "end":
			switch m.activePanel {
			case panelDetail:
				m.detail.SetXOffset(1 << 30) // clamped internally by viewport to max
				return m, nil
			case panelResponse:
				m.response.SetXOffset(1 << 30)
				return m, nil
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
					m.detail.SetXOffset(0)
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
		m.response.SetXOffset(0)
		return m, nil

	case errMsg:
		m.loading = false
		m.currentResp = nil
		m.currentErr = msg.err
		m.response.SetContent(renderError(msg.err))
		m.response.GotoTop()
		m.response.SetXOffset(0)
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

	listWidth := m.width * listWidthPercent / 100
	rightWidth := m.width - listWidth - panelHorizInset
	listContentHeight := m.height - statusBarHeight
	detailHeight := listContentHeight * detailHeightPercent / 100
	responseHeight := listContentHeight - detailHeight - panelBorderHeight

	listPanel := panelStyle(m.activePanel == panelList).
		Width(listWidth).
		Height(listContentHeight).
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
	listWidth := m.width * listWidthPercent / 100
	rightWidth := m.width - listWidth - panelHorizInset
	listContentHeight := m.height - statusBarHeight
	detailHeight := listContentHeight * detailHeightPercent / 100
	responseHeight := listContentHeight - detailHeight - panelBorderHeight

	m.list.SetSize(listWidth-panelContentXPad, listContentHeight-panelContentYPad)

	m.detail = viewport.New(rightWidth-panelContentXPad, max(detailHeight-panelContentYPad, 1))
	m.response = viewport.New(rightWidth-panelContentXPad, max(responseHeight-panelContentYPad, 1))

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
