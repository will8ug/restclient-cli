# Horizontal Scrolling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add horizontal scrolling to all 3 TUI panels (request list, request detail, response view) with clamped limits so content that exceeds panel width can be scrolled left/right instead of being truncated.

**Architecture:** The viewport panels (detail, response) already have built-in horizontal scrolling via `bubbles/viewport` (left/right keys, `ScrollLeft`/`ScrollRight`, `SetXOffset`). The main gaps are: (1) reset `xOffset` to 0 when content changes, (2) add shift+left/right for faster scrolling, (3) add Home/End to reset horizontal position, (4) implement horizontal scrolling for the list panel (which uses `bubbles/list` — no horizontal scroll support). For the list, we track a `listXOffset` in the Model and apply ANSI string cutting in the delegate's `Render` method, clamped between 0 and `(longestItemWidth - visibleWidth)`. All horizontal offsets are clamped to prevent over-scrolling.

**Tech Stack:** Go 1.26+, Bubbletea v1.3.10, bubbles v1.0.0 (viewport with built-in horizontal scroll), lipgloss v1.1.0, `charmbracelet/x/ansi` for string cutting in list panel.

---

### Task 1: Reset viewport xOffset on content change

When a new request is selected or a new response arrives, the viewport's horizontal scroll position (`xOffset`) persists from the previous content. This causes the new content to start scrolled sideways, which is confusing. We need to reset `xOffset` to 0 whenever content changes.

**Files:**
- Modify: `internal/tui/model.go:158,179,187`
- Test: `internal/tui/model_test.go`

- [ ] **Step 1: Write the failing test**

Add a test that verifies xOffset resets when switching requests (which changes detail content) and when receiving a response (which changes response content).

```go
func TestUpdateDetailXOffsetResetOnSelection(t *testing.T) {
	m := setupModel(t)

	// Scroll detail viewport right
	m.detail.ScrollRight(10)
	if m.detail.XOffset() == 0 {
		t.Fatal("expected xOffset > 0 after scrolling right")
	}

	// Select a different request — this should reset xOffset
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyDown})
	if m.detail.XOffset() != 0 {
		t.Fatalf("expected xOffset to reset to 0 on selection change, got %d", m.detail.XOffset())
	}
}
```

Wait — the viewport's `xOffset` is a private field. We can't read it directly. Instead, verify indirectly: after scrolling right and switching requests, the detail view's leftmost content should be visible (not cut off). Let me check the API.

The `viewport.Model` has no public `XOffset()` getter. We'll need to verify through the rendered view content. But there's `SetXOffset` which is public. Let me use `SetXOffset` to set it, then check the rendered output for the expected content.

Actually, the simplest approach: set xOffset via `SetXOffset`, then after switching requests, verify the first line of the rendered view starts with the expected prefix (the method name), proving the offset was reset.

```go
func TestUpdateDetailXOffsetResetOnSelection(t *testing.T) {
	m := setupModel(t)

	// Scroll detail viewport right so content is shifted
	m.detail.SetXOffset(20)

	// Select a different request — this should reset xOffset to 0
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyDown})

	// The first line of detail should start with the method name (visible again)
	detailView := m.detail.View()
	if !strings.HasPrefix(strings.TrimSpace(detailView), "POST") {
		t.Fatalf("expected detail view to start with POST after xOffset reset, got: %q", strings.TrimSpace(detailView))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -race -run TestUpdateDetailXOffsetResetOnSelection ./internal/tui/`
Expected: FAIL — `m.detail.SetXOffset(20)` sets the offset, but after switching requests the offset persists because `GotoTop()` only resets YOffset.

- [ ] **Step 3: Write minimal implementation**

In `model.go`, add `m.detail.SetXOffset(0)` after every `m.detail.GotoTop()` and `m.response.SetXOffset(0)` after every `m.response.GotoTop()`:

```go
// In Update(), case panelList section (~line 158):
if m.list.Index() != prevIdx {
    idx := m.list.Index()
    if idx >= 0 && idx < len(m.requests) {
        m.detail.SetContent(renderRequestDetail(m.requests[idx]))
        m.detail.GotoTop()
        m.detail.SetXOffset(0)
    }
}

// In Update(), case responseMsg (~line 179):
m.response.SetContent(renderResponse(msg.resp))
m.response.GotoTop()
m.response.SetXOffset(0)

// In Update(), case errMsg (~line 187):
m.response.SetContent(renderError(msg.err))
m.response.GotoTop()
m.response.SetXOffset(0)
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -race -run TestUpdateDetailXOffsetResetOnSelection ./internal/tui/`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/tui/model.go internal/tui/model_test.go
git commit -m "feat: reset viewport xOffset on content change"
```

---

### Task 2: Add shift+left/right and Home/End key bindings for horizontal scroll

The viewport's built-in left/right keys scroll by 6 columns (default `horizontalStep`). We should add shift+left/right for half-width scrolling (3 columns) and Home/End to reset horizontal position to 0 / max. These keys need to be handled in the model's `Update` before falling through to the viewport, since we want them to apply only to the active viewport panel.

**Files:**
- Modify: `internal/tui/model.go:105-172`
- Modify: `internal/tui/panels.go:167-186` (help text)
- Test: `internal/tui/model_test.go`

- [ ] **Step 1: Write the failing test**

Test that shift+left/right scrolls the active viewport, and Home/End resets horizontal position.

```go
func TestUpdateHorizontalScrollKeys(t *testing.T) {
	// Tab to detail panel
	m := setupModel(t)
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyTab})
	if m.activePanel != panelDetail {
		t.Fatalf("expected panelDetail, got %v", m.activePanel)
	}

	// Set content with a long line so horizontal scrolling is meaningful
	m.detail.SetContent("short\n" + strings.Repeat("x", 200) + "\nshort")
	m.detail.SetXOffset(0)

	// Press right — should scroll detail right
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRight})
	viewAfterRight := m.detail.View()
	// After scrolling right, the long "xxx..." line should be cut at the start
	if strings.Contains(viewAfterRight, "short") && strings.HasPrefix(strings.TrimSpace(viewAfterRight), "short") {
		// "short" on the first line should still be visible (it's shorter than viewport width)
		// But the middle line "xxx..." should start at xOffset=6
	}

	// Press Home — should reset horizontal position to 0
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyHome})
	viewAfterHome := m.detail.View()
	// The long line should start from position 0
	if strings.HasPrefix(strings.TrimSpace(viewAfterHome), "s") {
		// First line "short" should be fully visible from position 0
	}
}
```

Actually, testing horizontal scrolling through the rendered View output is fragile. A better approach: use `SetXOffset` to set a known offset, then send Home key, and verify the view starts from position 0 (the first character of content is visible).

Let me simplify the test to verify the behavior works by checking that content visibility changes:

```go
func TestUpdateHorizontalScrollHomeKey(t *testing.T) {
	m := setupModel(t)
	// Tab to detail panel
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyTab})

	// Set long content and scroll right
	m.detail.SetContent(strings.Repeat("a", 200))
	m.detail.SetXOffset(20)

	beforeHome := m.detail.View()
	// At xOffset=20, the view starts at position 20, so "aaaa..." from pos 20

	// Press Home — should reset xOffset to 0
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyHome})

	afterHome := m.detail.View()
	// At xOffset=0, the view starts from position 0
	// The view should be different after pressing Home
	if beforeHome == afterHome {
		t.Fatal("expected view to change after Home key resets horizontal scroll")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -race -run TestUpdateHorizontalScrollHomeKey ./internal/tui/`
Expected: FAIL — Home key is not handled for horizontal scroll reset in the model's Update method.

- [ ] **Step 3: Write minimal implementation**

In `model.go`, add key handling for Home/End in the `switch msg.String()` block, and add shift+left/right handling. The left/right keys already work via viewport's built-in handler, but shift+left/right need explicit handling since the viewport only handles plain left/right.

```go
// In Update(), inside the switch msg.String() block (after "enter" case, before closing):
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
        maxOffset := m.detail.TotalLineCount() // not quite — need longestLineWidth - width
        // viewport doesn't expose longestLineWidth, but we can scroll right until it clamps
        // Use a large value; SetXOffset clamps internally
        m.detail.SetXOffset(999999)
        return m, nil
    case panelResponse:
        m.response.SetXOffset(999999)
        return m, nil
    }
```

Wait — the viewport clamps `SetXOffset` to `0 .. longestLineWidth - Width`, so passing a very large value will clamp to max. But `999999` is inelegant. Better: add a helper method. Since `longestLineWidth` is private, we can use `SetXOffset` with a large value and rely on the internal clamp. This is fine.

Actually, looking at the viewport source: `func (m *Model) SetXOffset(n int) { m.xOffset = clamp(n, 0, m.longestLineWidth-m.Width) }`. So a large value will clamp correctly. Let's use a clean approach.

For shift+left/right: Bubbletea sends shift+left as a `KeyMsg` with `Type: tea.KeyShiftLeft` and shift+right as `Type: tea.KeyShiftRight`. However, checking the Bubbletea key types:

Actually in Bubbletea, `tea.KeyMsg` has a `Type` field. The shift modifier may be represented differently. Let me check what key types are available.

The `tea.KeyMsg` `String()` method returns things like "shift+left", "shift+right", "home", "end". So we should match on those strings.

```go
// Add these cases to the switch msg.String() block in Update():
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
        m.detail.SetXOffset(1 << 30) // clamped internally to max
        return m, nil
    case panelResponse:
        m.response.SetXOffset(1 << 30)
        return m, nil
    }
```

For the list panel, shift+left/right and Home/End will be handled in Task 3 (list horizontal scrolling).

- [ ] **Step 4: Update help text**

In `panels.go`, update `renderHelp()` to document the new key bindings:

```go
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
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test -race -run TestUpdateHorizontalScrollHomeKey ./internal/tui/`
Expected: PASS

- [ ] **Step 6: Run all tests**

Run: `go test -race -count=1 ./internal/tui/`
Expected: All PASS

- [ ] **Step 7: Commit**

```bash
git add internal/tui/model.go internal/tui/model_test.go internal/tui/panels.go
git commit -m "feat: add shift+left/right and Home/End for horizontal scroll in viewport panels"
```

---

### Task 3: Add horizontal scrolling for the request list panel

The `bubbles/list` component does not support horizontal scrolling. We need to track a `listXOffset` in the Model, apply it when rendering list items in the delegate, and handle left/right key events when the list panel is active. The offset is clamped between 0 and `(longestItemWidth - visibleWidth)`, where `longestItemWidth` is tracked in the Model.

**Files:**
- Modify: `internal/tui/model.go:44-59,147-161,282-295` (Model struct, key handling, resize)
- Modify: `internal/tui/panels.go:17-66` (requestDelegate.Render)
- Test: `internal/tui/model_test.go`

- [ ] **Step 1: Write the failing test**

```go
func TestUpdateListHorizontalScroll(t *testing.T) {
	m := setupModel(t)
	// Ensure we're in list panel
	if m.activePanel != panelList {
		t.Fatalf("expected panelList, got %v", m.activePanel)
	}

	// Initial listXOffset should be 0
	if m.listXOffset != 0 {
		t.Fatalf("expected initial listXOffset=0, got %d", m.listXOffset)
	}

	// Press right in list panel — should scroll list right
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyRight})
	if m.listXOffset == 0 {
		t.Fatal("expected listXOffset > 0 after pressing right in list panel")
	}

	// Press Home — should reset listXOffset to 0
	m, _ = sendKey(t, m, tea.KeyMsg{Type: tea.KeyHome})
	if m.listXOffset != 0 {
		t.Fatalf("expected listXOffset=0 after Home, got %d", m.listXOffset)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -race -run TestUpdateListHorizontalScroll ./internal/tui/`
Expected: FAIL — `m.listXOffset` field doesn't exist yet.

- [ ] **Step 3: Add listXOffset and listMaxXOffset fields to Model**

In `model.go`, add fields to the Model struct:

```go
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
	listXOffset int
	listMaxXOffset int
}
```

- [ ] **Step 4: Handle left/right keys for list panel horizontal scroll**

In the `switch msg.String()` block in `Update()`, add cases for left/right when the list panel is active. The left/right keys currently fall through to the list's Update, which uses them for page up/down. We need to intercept them when we want horizontal scroll instead.

Design decision: We'll intercept left/right ONLY when shift is held (shift+left/shift+right) for list horizontal scroll, similar to the viewport panels. Plain left/right will continue to work as the list's default page up/down bindings. This is consistent with how horizontal scroll works in the viewport panels (plain left/right for 6-column scroll).

Wait, actually for the viewport panels, plain left/right are already wired to horizontal scroll by the viewport's built-in handler. For the list, plain left/right are wired to page navigation. To be consistent, we should remap left/right to horizontal scroll in the list too, but that would break the existing page navigation behavior.

Better approach: Use `h`/`l` for horizontal scrolling in all panels (left/right already used for different purposes in the list). Actually, in the viewport, left/right already scroll horizontally by default. For consistency, let's make left/right scroll horizontally in ALL panels, and use page up/down for vertical page navigation. The list's default keymap already maps left/h/pgup/b/u to prevPage and right/l/pgdown/f/d to nextPage. We should change this.

Actually, let me reconsider. The simplest approach that maintains backward compatibility:
- For viewport panels (detail, response): left/right keys already scroll horizontally via built-in viewport handler. This works.
- For list panel: intercept left/right to scroll horizontally, and let j/k/up/down handle vertical navigation (which they already do).

This means left/right will be horizontal scroll in ALL panels consistently. The list's built-in page up/down key bindings (which use left/right/h/l/pgup/pgdown) will be overridden by our horizontal scroll handling. We can keep pgup/pgdown for vertical page navigation in the list.

In the `switch msg.String()` block, add:

```go
case "left":
    if m.activePanel == panelList {
        m.listXOffset = max(m.listXOffset-6, 0)
        return m, nil
    }
case "right":
    if m.activePanel == panelList {
        m.listXOffset = min(m.listXOffset+6, m.listMaxXOffset)
        return m, nil
    }
case "shift+left":
    if m.activePanel == panelList {
        m.listXOffset = max(m.listXOffset-3, 0)
        return m, nil
    }
case "shift+right":
    if m.activePanel == panelList {
        m.listXOffset = min(m.listXOffset+3, m.listMaxXOffset)
        return m, nil
    }
case "home":
    if m.activePanel == panelList {
        m.listXOffset = 0
        return m, nil
    }
case "end":
    if m.activePanel == panelList {
        m.listXOffset = m.listMaxXOffset
        return m, nil
    }
```

But wait — we already added home/end/shift+left/right for viewport panels in Task 2. Now we need to expand those cases to also handle the list panel. Let me revise: ALL the horizontal scroll key cases should handle ALL 3 panels. This means the cases from Task 2 should be updated to include list panel handling too.

Actually, since the tasks are sequential and each one builds on the previous, let me restructure. In Task 2, we added home/end/shift+left/right for viewport panels only. In Task 3, we'll update those same cases to also handle the list panel. And we'll add plain left/right handling for the list panel.

Here's the combined approach for the key handling section:

```go
// In the switch msg.String() block:
case "left":
    if m.activePanel == panelList {
        m.listXOffset = max(m.listXOffset-6, 0)
        return m, nil
    }
    // For detail/response, left falls through to viewport.Update which handles it
case "right":
    if m.activePanel == panelList {
        m.listXOffset = min(m.listXOffset+6, m.listMaxXOffset)
        return m, nil
    }
    // For detail/response, right falls through to viewport.Update which handles it

case "shift+left":
    switch m.activePanel {
    case panelList:
        m.listXOffset = max(m.listXOffset-3, 0)
        return m, nil
    case panelDetail:
        m.detail.ScrollLeft(3)
        return m, nil
    case panelResponse:
        m.response.ScrollLeft(3)
        return m, nil
    }

case "shift+right":
    switch m.activePanel {
    case panelList:
        m.listXOffset = min(m.listXOffset+3, m.listMaxXOffset)
        return m, nil
    case panelDetail:
        m.detail.ScrollRight(3)
        return m, nil
    case panelResponse:
        m.response.ScrollRight(3)
        return m, nil
    }

case "home":
    switch m.activePanel {
    case panelList:
        m.listXOffset = 0
        return m, nil
    case panelDetail:
        m.detail.SetXOffset(0)
        return m, nil
    case panelResponse:
        m.response.SetXOffset(0)
        return m, nil
    }

case "end":
    switch m.activePanel {
    case panelList:
        m.listXOffset = m.listMaxXOffset
        return m, nil
    case panelDetail:
        m.detail.SetXOffset(1 << 30)
        return m, nil
    case panelResponse:
        m.response.SetXOffset(1 << 30)
        return m, nil
    }
```

Note: For left/right keys, we need to intercept them for the list panel (return early) but let them fall through for viewport panels. The current code structure has the `switch msg.String()` block that returns early for matched cases, then the panel-specific switch handles the rest. So left/right will be caught for the list panel and returned early, but for detail/response they'll fall through to the panel-specific switch which forwards to viewport.Update.

- [ ] **Step 5: Calculate listMaxXOffset in resizePanels**

We need to compute the maximum horizontal scroll offset for the list panel. This requires knowing the longest item width and the visible width. In `resizePanels()`, add:

```go
func (m Model) resizePanels() Model {
    listWidth := m.width * listWidthPercent / 100
    rightWidth := m.width - listWidth - panelHorizInset
    listContentHeight := m.height - statusBarHeight
    detailHeight := listContentHeight * detailHeightPercent / 100
    responseHeight := listContentHeight - detailHeight - panelBorderHeight

    m.list.SetSize(listWidth-panelContentXPad, listContentHeight-panelContentYPad)

    // Calculate listMaxXOffset: max scroll = longestItemWidth - visibleWidth
    visibleWidth := listWidth - panelContentXPad
    longestLineWidth := m.longestListItemWidth()
    m.listMaxXOffset = max(longestLineWidth-visibleWidth, 0)

    // Clamp existing listXOffset to new max
    m.listXOffset = min(m.listXOffset, m.listMaxXOffset)

    m.detail = viewport.New(rightWidth-panelContentXPad, max(detailHeight-panelContentYPad, 1))
    m.response = viewport.New(rightWidth-panelContentXPad, max(responseHeight-panelContentYPad, 1))

    return m
}
```

- [ ] **Step 6: Add longestListItemWidth helper method**

```go
func (m Model) longestListItemWidth() int {
    maxW := 0
    for _, req := range m.requests {
        line := fmt.Sprintf("%s %s %s", req.Method, req.Name, req.URL)
        if req.Name == "" {
            line = fmt.Sprintf("%s %s", req.Method, req.URL)
        }
        w := lipgloss.Width(line)
        if w > maxW {
            maxW = w
        }
    }
    return maxW
}
```

Wait, this needs to account for the styled method badge width too. The badge uses `methodStyle(method).Render(fmt.Sprintf("%-7s", method))` which is wider than the plain method string. Let me match what the delegate renders:

```go
func (m Model) longestListItemWidth() int {
    maxW := 0
    for _, req := range m.requests {
        badge := methodStyle(req.Method).Render(fmt.Sprintf("%-7s", req.Method))
        name := req.Name
        if name == "" {
            name = req.URL
        }
        titleLine := fmt.Sprintf("> %s %s", badge, name)
        descLine := fmt.Sprintf("  %s", req.URL)
        titleW := lipgloss.Width(titleLine)
        descW := lipgloss.Width(descLine)
        w := max(titleW, descW)
        if w > maxW {
            maxW = w
        }
    }
    return maxW
}
```

This requires importing `fmt` in model.go. It's already imported via other files but let me check — looking at the imports: `strings` and the bubbletea packages. We need to add `"fmt"` to the import block in model.go.

Actually, looking at model.go imports more carefully:

```go
import (
    "fmt"         // Wait — is it imported? Let me check.
    "strings"
    "time"

    "github.com/charmbracelet/bubbles/list"
    ...
)
```

Looking at the actual model.go line 1-15: it imports `"fmt"`, `"strings"`, `"time"` plus the bubbletea packages. Yes, `"fmt"` is already imported.

- [ ] **Step 7: Apply listXOffset in requestDelegate.Render**

In `panels.go`, modify `requestDelegate.Render` to apply the horizontal offset. The delegate doesn't have direct access to the Model's `listXOffset`. We need to pass it somehow.

The `list.ItemDelegate` interface's `Render` method signature is:
```go
Render(w io.Writer, m list.Model, index int, listItem list.Item)
```

We can't change this interface. Options:
1. Store `listXOffset` in the `requestDelegate` struct and update it before each View render
2. Use `list.Model.SetWidth` to control visible width (but this doesn't help with offset)
3. Create a custom delegate that wraps `requestDelegate` and carries the offset

Option 1 is simplest: make `requestDelegate` carry the offset.

```go
type requestDelegate struct {
    xOffset int
}
```

Then in `Render`, apply `ansi.Cut` to shift content:

```go
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

    // Apply horizontal scroll offset
    if d.xOffset > 0 {
        titleLine = ansi.Cut(titleLine, d.xOffset, d.xOffset+m.Width())
        descLine = ansi.Cut(descLine, d.xOffset, d.xOffset+m.Width())
    }

    fmt.Fprintf(w, "%s\n%s", titleLine, descLine)
}
```

This requires importing `"github.com/charmbracelet/x/ansi"` in panels.go. Let me check if this is already in go.mod — yes, `github.com/charmbracelet/x/ansi v0.11.6` is listed.

- [ ] **Step 8: Update delegate in NewModel and pass xOffset in View**

In `NewModel()`, the delegate is created as `requestDelegate{}` (no offset). In the `View()` method, before rendering the list, we need to set the delegate's xOffset. But the `list.Model` stores the delegate internally and uses it for rendering. We can update the delegate via `list.SetDelegate()`.

Wait — `list.Model` doesn't have a `SetDelegate` method. Looking at the list API... the delegate is set during `list.New()` and stored as a private field. We can't update it after creation.

Alternative approach: Instead of storing xOffset on the delegate, we can use `list.NewItemDelegate()` to set it. But the API doesn't support updating delegates after creation.

Better approach: Use a pointer-based delegate that can be updated:

```go
type requestDelegate struct {
    xOffset int
}
```

Since Go structs are values, when we pass `requestDelegate{}` to `list.New`, the list stores a copy. We need to use a pointer receiver... but the `list.ItemDelegate` interface methods need pointer receivers for mutation to work, and `Render` takes `d requestDelegate` (not a pointer). Actually, looking at the current code, `requestDelegate` methods use value receivers. But `xOffset` needs to be mutable.

Let me look at this differently. The `Render` method has access to `m list.Model`, which gives us `m.Width()`. If we can make `xOffset` accessible through the list model somehow...

Actually, the simplest solution: make the delegate a pointer type so the xOffset field can be mutated externally.

```go
// In panels.go, change to pointer-based delegate:
type requestDelegate struct {
    xOffset int
}

func (d *requestDelegate) Height() int                             { return 2 }
func (d *requestDelegate) Spacing() int                            { return 0 }
func (d *requestDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d *requestDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
    // ... same rendering but uses d.xOffset for cutting
}
```

And in `NewModel()`:
```go
delegate := &requestDelegate{}
l := list.New(items, delegate, 0, 0)
```

Since `list.New` takes `list.ItemDelegate` (an interface), and `*requestDelegate` implements it (pointer receivers), this works. We store the delegate pointer on the Model so we can update its xOffset:

```go
type Model struct {
    // ... existing fields ...
    listXOffset    int
    listMaxXOffset int
    listDelegate   *requestDelegate
}
```

In `NewModel()`:
```go
delegate := &requestDelegate{}
l := list.New(items, delegate, 0, 0)
// ...
return Model{
    // ... existing fields ...
    listDelegate: delegate,
}
```

In `View()`, before rendering the list:
```go
m.listDelegate.xOffset = m.listXOffset
```

Wait, but `View()` is called on a copy of the Model (since `Model` implements `tea.Model` with value receivers). Changes to `m.listDelegate.xOffset` in `View()` won't persist. But that's fine — we just need the delegate to use the current xOffset when rendering. Since `listDelegate` is a pointer, updating it in `View()` affects the underlying struct even though `View()` operates on a Model copy.

Actually, there's a subtlety: `View()` operates on `m Model` (a copy), but `m.listDelegate` is a `*requestDelegate` pointer, so `m.listDelegate.xOffset = m.listXOffset` modifies the same underlying struct. Then `m.list.View()` will use that delegate with the correct xOffset. This works!

- [ ] **Step 9: Reset listXOffset when content changes (filter, resize)**

When the list is resized, `listMaxXOffset` changes, so we clamp `listXOffset` (already handled in `resizePanels()` step 5). When a filter is applied, the visible items change but the max offset is based on ALL request items (not just filtered), so `listMaxXOffset` stays the same. This is correct — you should be able to scroll horizontally to see the full width of any item, even filtered-out ones don't affect the max scroll.

When a request is executed and we return to the list panel, `listXOffset` should remain as-is (no reset). The user may want to keep their horizontal scroll position while navigating the list.

Actually, on second thought — when the user tabs back to the list panel after viewing response, the horizontal scroll position should persist. This is natural. No reset needed.

- [ ] **Step 10: Run test to verify it passes**

Run: `go test -race -run TestUpdateListHorizontalScroll ./internal/tui/`
Expected: PASS

- [ ] **Step 11: Run all tests**

Run: `go test -race -count=1 ./internal/tui/`
Expected: All PASS

- [ ] **Step 12: Commit**

```bash
git add internal/tui/model.go internal/tui/panels.go internal/tui/model_test.go
git commit -m "feat: add horizontal scrolling for request list panel"
```

---

### Task 4: Update status bar to show horizontal scroll hint

The status bar currently shows key hints like `↑↓ navigate  enter send  tab switch`. We should add a hint for horizontal scrolling so users know it's available.

**Files:**
- Modify: `internal/tui/model.go:257-280`

- [ ] **Step 1: Update status bar text**

In `model.go`, update the `renderStatusBar()` method to include horizontal scroll hints:

```go
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
        left = " ↑↓ navigate  ←→ h-scroll  home/end reset h-scroll  enter send  tab switch  / filter  ? help  q quit"
    }

    right := fmt.Sprintf(" %s ", m.fileName)
    gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
    if gap < 0 {
        gap = 0
    }

    bar := left + strings.Repeat(" ", gap) + right
    return statusBarStyle.Width(m.width).Render(bar)
}
```

- [ ] **Step 2: Run all tests**

Run: `go test -race -count=1 ./internal/tui/`
Expected: All PASS

- [ ] **Step 3: Commit**

```bash
git add internal/tui/model.go
git commit -m "feat: add horizontal scroll hints in status bar"
```

---

### Task 5: Manual integration testing

Build and run the TUI with a test .http file to verify horizontal scrolling works in all 3 panels.

**Files:**
- No code changes — manual testing only

- [ ] **Step 1: Build and run**

```bash
make build
./bin/restclient-cli examples/jsonplaceholder.http
```

- [ ] **Step 2: Verify viewport panels**

1. Press Tab to switch to Request Detail panel
2. Use left/right arrows — content should scroll horizontally
3. Use shift+left/right — content should scroll by 3 columns (fine scroll)
4. Press Home — horizontal scroll should reset to 0
5. Press End — horizontal scroll should go to max
6. Switch to a different request (Tab back to list, select different item) — detail xOffset should reset to 0

- [ ] **Step 3: Verify response panel**

1. Select a request and press Enter to send it
2. Tab to Response panel
3. Use left/right, shift+left/right, Home/End — response body should scroll horizontally
4. Send a different request — response xOffset should reset to 0

- [ ] **Step 4: Verify list panel**

1. Tab to Request List panel
2. Use left/right arrows — list items should scroll horizontally (if they exceed panel width)
3. Use shift+left/right for fine scroll
4. Home/End should reset/go-to-max

- [ ] **Step 5: Verify clamping**

1. Scroll left at position 0 — should stay at 0 (no negative offset)
2. Scroll right past the end — should clamp at max offset
3. Resize terminal to smaller width — existing offset should clamp to new max

---

## Self-Review

**1. Spec coverage:** The spec asks for horizontal scrolling in all 3 panels with clamped limits. Tasks 1-3 cover all 3 panels. Task 1 resets offsets on content change. Task 2 adds shift+left/right and Home/End for viewport panels. Task 3 adds left/right/shift+left/right/Home/End for list panel + clamped offsets. Task 4 updates UI hints. Task 5 is manual testing. All requirements covered.

**2. Placeholder scan:** No TBD, TODO, "implement later", or vague "add appropriate" instructions. All steps have concrete code. No "similar to Task N" shortcuts.

**3. Type consistency:** `listXOffset` and `listMaxXOffset` are `int` fields on `Model`. `requestDelegate.xOffset` is `int`. All arithmetic uses `max()` and `min()` with `int` types. `SetXOffset(int)` and `ScrollLeft(int)/ScrollRight(int)` all use `int`. Consistent throughout.