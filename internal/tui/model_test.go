package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/will8ug/restclient-cli/internal/parser"
)

// setupModel creates a Model with 3 test requests and a WindowSizeMsg applied.
func setupModel(t *testing.T) Model {
	t.Helper()
	requests := []parser.Request{
		{Name: "Get users", Method: "GET", URL: "https://api.example.com/users"},
		{Name: "Create user", Method: "POST", URL: "https://api.example.com/users"},
		{Name: "Delete user", Method: "DELETE", URL: "https://api.example.com/users/1"},
	}
	m := NewModel(requests, "test.http")
	newM, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	return newM.(Model)
}

func sendKey(t *testing.T, m Model, msg tea.KeyMsg) (Model, tea.Cmd) {
	t.Helper()
	newM, cmd := m.Update(msg)
	m = newM.(Model)
	m = routeCmd(t, m, cmd)
	return m, cmd
}

func routeCmd(t *testing.T, m Model, cmd tea.Cmd) Model {
	t.Helper()
	if cmd == nil {
		return m
	}
	return routeMsg(t, m, cmd())
}

func routeMsg(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	switch msg := msg.(type) {
	case nil:
		return m
	case list.FilterMatchesMsg:
		newM, _ := m.Update(msg)
		return newM.(Model)
	case tea.BatchMsg:
		for _, cmd := range msg {
			m = routeCmd(t, m, cmd)
		}
		return m
	default:
		return m
	}
}

func TestUpdateInitialState(t *testing.T) {
	m := setupModel(t)

	if !m.ready {
		t.Fatal("expected model to be ready after window size update")
	}
	if m.activePanel != panelList {
		t.Fatalf("expected activePanel=%v, got %v", panelList, m.activePanel)
	}
	if m.showHelp {
		t.Fatal("expected help to be hidden")
	}
	if m.loading {
		t.Fatal("expected loading to be false")
	}
}

func TestUpdateWindowSizeMsg(t *testing.T) {
	m := setupModel(t)

	for i := 0; i < 2; i++ {
		var cmd tea.Cmd
		m, cmd = sendKey(t, m, tea.KeyMsg{Type: tea.KeyDown})
		if cmd != nil {
			t.Logf("down key returned cmd: %T", cmd)
		}
	}

	if got := m.list.Index(); got != 2 {
		t.Fatalf("expected selected index 2, got %d", got)
	}

	newM, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 32})
	m = newM.(Model)

	if got := m.requests[m.list.Index()].Name; got != "Delete user" {
		t.Fatalf("expected selected request name %q, got %q", "Delete user", got)
	}
	if got := m.detail.View(); !strings.Contains(got, "/users/1") {
		t.Fatalf("expected detail view to contain selected request URL %q, got %q", "/users/1", got)
	}
}

func TestUpdateTabSwitching(t *testing.T) {
	m := setupModel(t)

	if m.activePanel != panelList {
		t.Fatalf("expected initial panelList, got %v", m.activePanel)
	}

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyTab})
	if m.activePanel != panelDetail {
		t.Fatalf("expected panelDetail after first tab, got %v", m.activePanel)
	}

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyTab})
	if m.activePanel != panelResponse {
		t.Fatalf("expected panelResponse after second tab, got %v", m.activePanel)
	}

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyTab})
	if m.activePanel != panelList {
		t.Fatalf("expected panelList after wrapping tab, got %v", m.activePanel)
	}
}

func TestUpdateQuitInList(t *testing.T) {
	m := setupModel(t)

	_, cmd := sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit command from list panel")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Fatalf("expected tea.QuitMsg, got %T", cmd())
	}

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyTab})
	_, cmd = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Fatalf("expected no quit command from non-list panel, got %T", cmd)
	}
}

func TestUpdateHelpToggle(t *testing.T) {
	m := setupModel(t)

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !m.showHelp {
		t.Fatal("expected help to be shown after first toggle")
	}

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if m.showHelp {
		t.Fatal("expected help to be hidden after second toggle")
	}
}

func TestUpdateFilterApplied(t *testing.T) {
	m := setupModel(t)
	total := len(m.list.VisibleItems())

	if m.list.FilterState() == list.Filtering {
		t.Fatal("expected filter to be inactive initially")
	}

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if m.list.FilterState() != list.Filtering {
		t.Fatalf("expected filtering after '/', got %v", m.list.FilterState())
	}

	for _, r := range []rune{'G', 'E', 'T'} {
		var cmd tea.Cmd
		m, cmd = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		if cmd != nil {
			t.Logf("filter key %q returned cmd: %T", r, cmd)
		}
	}

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	if state := m.list.FilterState(); state != list.FilterApplied && state != list.Filtering {
		t.Fatalf("expected filter to be applied or still filtering, got %v", state)
	}
	if got := len(m.list.VisibleItems()); got >= total {
		t.Fatalf("expected visible items to narrow after filtering, got %d of %d", got, total)
	}
}

func TestUpdateFilterShortcut(t *testing.T) {
	m := setupModel(t)
	all := len(m.list.VisibleItems())

	if m.list.FilterState() == list.Filtering {
		t.Fatal("expected filter to be inactive initially")
	}

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if m.list.FilterState() != list.Filtering {
		t.Fatalf("expected filtering after '/', got %v", m.list.FilterState())
	}

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	if got := len(m.list.VisibleItems()); got != all {
		t.Fatalf("expected all items visible after empty filter apply, got %d of %d", got, all)
	}
	if m.list.FilterState() == list.Filtering {
		t.Fatal("expected filtering to stop after enter")
	}
}

func TestUpdateDetailUpdatesOnSelection(t *testing.T) {
	m := setupModel(t)

	if got := m.requests[m.list.Index()].Name; got != "Get users" {
		t.Fatalf("expected initial selected request %q, got %q", "Get users", got)
	}
	if got := m.detail.View(); !strings.Contains(got, "/users") {
		t.Fatalf("expected initial detail to contain request URL %q, got %q", "/users", got)
	}

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyDown})
	if got := m.list.Index(); got != 1 {
		t.Fatalf("expected selected index 1, got %d", got)
	}
	if got := m.requests[m.list.Index()].Name; got != "Create user" {
		t.Fatalf("expected selected request %q, got %q", "Create user", got)
	}
	if got := m.detail.View(); !strings.Contains(got, "/users") {
		t.Fatalf("expected detail to contain request URL %q, got %q", "/users", got)
	}

	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyDown})
	if got := m.requests[m.list.Index()].Name; got != "Delete user" {
		t.Fatalf("expected selected request %q, got %q", "Delete user", got)
	}
	if got := m.detail.View(); !strings.Contains(got, "/users/1") {
		t.Fatalf("expected detail to contain request URL %q, got %q", "/users/1", got)
	}
}

func TestUpdateDetailXOffsetResetOnSelection(t *testing.T) {
	// Use a narrow window so content lines are wider than the viewport,
	// making horizontal scrolling possible (SetXOffset has actual effect).
	requests := []parser.Request{
		{Name: "Get users", Method: "GET", URL: "https://api.example.com/users"},
		{Name: "Create user", Method: "POST", URL: "https://api.example.com/users"},
	}
	m := NewModel(requests, "test.http")
	newM, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 30})
	m = newM.(Model)

	// Navigate to the POST request first so its detail content is loaded
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyDown})

	// Scroll detail viewport right so content is shifted.
	m.detail.SetXOffset(20)

	// Return to the GET request — this triggers SetContent + GotoTop,
	// but xOffset persists because GotoTop only resets YOffset.
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyUp})

	// The first line of detail should start with the method name (visible again
	// only if xOffset was reset to 0; otherwise it's scrolled off-screen).
	detailView := m.detail.View()
	firstLine := strings.Split(detailView, "\n")[0]
	if !strings.Contains(firstLine, "GET") {
		t.Fatalf("expected detail view to start with GET after xOffset reset, got: %q", firstLine)
	}
}

// uniqueLine creates a line where each character position is identifiable:
// the character at offset i is 'A' + (i % 26), so different offsets show different text.
func uniqueLine(length int) string {
	buf := make([]byte, length)
	for i := range buf {
		buf[i] = byte('A' + (i % 26))
	}
	return string(buf)
}

func TestUpdateHorizontalScrollHomeKey(t *testing.T) {
	m := setupModel(t)
	// Tab to detail panel
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyTab})

	// Set long content with distinct chars per position and scroll right
	m.detail.SetContent(uniqueLine(200))
	m.detail.SetXOffset(20)

	beforeHome := m.detail.View()

	// Press Home -- should reset xOffset to 0
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyHome})

	afterHome := m.detail.View()
	if beforeHome == afterHome {
		t.Fatal("expected view to change after Home key resets horizontal scroll")
	}
}

func TestUpdateHorizontalScrollEndKey(t *testing.T) {
	m := setupModel(t)
	// Tab to detail panel
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyTab})

	// Set long content with distinct chars per position at offset 0
	m.detail.SetContent(uniqueLine(200))

	beforeEnd := m.detail.View()

	// Press End -- should scroll to max horizontal position
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyEnd})

	afterEnd := m.detail.View()
	if beforeEnd == afterEnd {
		t.Fatal("expected view to change after End key scrolls to max horizontal position")
	}
}

func TestUpdateHorizontalScrollShiftLeftRight(t *testing.T) {
	m := setupModel(t)
	// Tab to detail panel
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyTab})

	// Set long content with distinct chars per position at offset 0
	m.detail.SetContent(uniqueLine(200))

	beforeShiftRight := m.detail.View()

	// Press Shift+Right -- should scroll right by 3
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyShiftRight})

	afterShiftRight := m.detail.View()
	if beforeShiftRight == afterShiftRight {
		t.Fatal("expected view to change after Shift+Right scrolls horizontally")
	}

	beforeShiftLeft := m.detail.View()

	// Press Shift+Left -- should scroll left by 3
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyShiftLeft})

	afterShiftLeft := m.detail.View()
	if beforeShiftLeft == afterShiftLeft {
		t.Fatal("expected view to change after Shift+Left scrolls horizontally")
	}
}
