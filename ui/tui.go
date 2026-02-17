package ui

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/asciimoo/hister/config"
	"github.com/asciimoo/hister/server/indexer"
	"github.com/asciimoo/hister/server/model"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/jroimartin/gocui"
	"github.com/pkg/browser"
)

const Banner = `
░█░█░▀█▀░█▀▀░▀█▀░█▀▀░█▀▄
░█▀█░░█░░▀▀█░░█░░█▀▀░█▀▄
░▀░▀░▀▀▀░▀▀▀░░▀░░▀▀▀░▀░▀
`

var (
	boldWhite = color.New(color.FgWhite, color.Bold).SprintFunc()
	blue      = color.New(color.FgBlue).SprintFunc()
	gray      = color.New(color.FgHiBlack).SprintFunc()
	red       = color.New(color.FgRed, color.Bold).SprintFunc()
)

type tui struct {
	SearchInput *gocui.View
	ResultsView *gocui.View
	conn        *websocket.Conn
	g           *gocui.Gui
	results     *indexer.Results
	selectedIdx int
	lineOffsets []int
	cfg         *config.Config
}

type singleLineEditor struct {
	editor   gocui.Editor
	callback func()
}

type query struct {
	Text      string `json:"text"`
	Highlight string `json:"highlight"`
}

func newTUI(cfg *config.Config) (*tui, error) {
	t := &tui{cfg: cfg}
	return t, t.init(cfg.WebSocketURL())
}

func (e singleLineEditor) Edit(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case (ch != 0 || key == gocui.KeySpace) && mod == 0:
		e.editor.Edit(v, key, ch, mod)
		if e.callback != nil {
			e.callback()
		}
		return
	case key == gocui.KeyCtrlW:
		e.deleteWord(v)
		if e.callback != nil {
			e.callback()
		}
		return
	case key == gocui.KeyArrowRight:
		ox, _ := v.Cursor()
		if ox >= len(v.Buffer())-1 {
			return
		}
	case key == gocui.KeyHome || key == gocui.KeyArrowUp:
		v.SetCursor(0, 0)
		v.SetOrigin(0, 0)
		return
	case key == gocui.KeyEnd || key == gocui.KeyArrowDown:
		width, _ := v.Size()
		lineWidth := len(v.Buffer()) - 1
		if lineWidth > width {
			v.SetOrigin(lineWidth-width, 0)
			lineWidth = width - 1
		}
		v.SetCursor(lineWidth, 0)
		return
	}
	e.editor.Edit(v, key, ch, mod)
	if e.callback != nil {
		e.callback()
	}
}

func (e singleLineEditor) deleteWord(v *gocui.View) {
	cx, _ := v.Cursor()
	if cx == 0 {
		return
	}

	line := v.Buffer()
	if len(line) == 0 {
		return
	}

	ox, _ := v.Origin()
	pos := min(ox+cx, len(line))

	start := pos - 1
	for start > 0 && line[start] == ' ' {
		start--
	}
	for start > 0 && line[start-1] != ' ' {
		start--
	}

	newLine := line[:start] + line[pos:]
	v.Clear()
	fmt.Fprint(v, newLine)

	newCursorPos := start - ox
	if newCursorPos < 0 {
		v.SetOrigin(start, 0)
		v.SetCursor(0, 0)
	} else {
		v.SetCursor(newCursorPos, 0)
	}
}

type keybinding struct {
	view    string
	key     interface{}
	mod     gocui.Modifier
	handler func(g *gocui.Gui, v *gocui.View) error
}

func SearchTUI(cfg *config.Config) error {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer g.Close()

	g.Cursor = true

	t, err := newTUI(cfg)
	if err != nil {
		return err
	}
	defer t.close()

	t.g = g

	g.SetManagerFunc(t.layout)

	quit := func(g *gocui.Gui, v *gocui.View) error {
		return gocui.ErrQuit
	}

	nop := func(g *gocui.Gui, v *gocui.View) error { return nil }

	keybindings := []keybinding{
		{"", gocui.KeyCtrlC, gocui.ModNone, quit},

		{"search-input", gocui.KeyTab, gocui.ModNone, t.enterResultsMode},
		{"search-input", gocui.KeyEnter, gocui.ModNone, nop},

		{"results", 'j', gocui.ModNone, t.moveCursor(1)},
		{"results", 'k', gocui.ModNone, t.moveCursor(-1)},
		{"results", gocui.KeyArrowDown, gocui.ModNone, t.moveCursor(1)},
		{"results", gocui.KeyArrowUp, gocui.ModNone, t.moveCursor(-1)},

		{"results", gocui.KeyEnter, gocui.ModNone, t.openSelected},
		{"results", 'd', gocui.ModNone, t.deleteSelected},
		{"results", gocui.KeyTab, gocui.ModNone, t.exitResultsMode},
	}

	for _, kb := range keybindings {
		if err := g.SetKeybinding(kb.view, kb.key, kb.mod, kb.handler); err != nil {
			return err
		}
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

func (t *tui) init(wsurl string) error {
	var err error
	t.conn, _, err = websocket.DefaultDialer.Dial(wsurl, nil)
	if err != nil {
		return err
	}
	go func() {
		for {
			_, msg, err := t.conn.ReadMessage()
			if err != nil {
				// TODO
				return
			}
			t.handleResults(msg)
		}
	}()
	return nil
}

func (t *tui) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	bannerW := 24
	bannerH := 5
	if v, err := g.SetView("banner", maxX/2-bannerW/2-1, 0, maxX/2+bannerW/2, bannerH); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.FgColor = gocui.ColorBlue
		fmt.Fprintf(v, "%s", Banner)
	}
	if v, err := g.SetView("search-input", 5, 5, maxX-6, 7); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Search"
		v.Editable = true
		v.Editor = singleLineEditor{
			editor:   gocui.DefaultEditor,
			callback: t.search,
		}
		t.SearchInput = v
		g.SetCurrentView(v.Name())
	}
	if v, err := g.SetView("results", 0, 8, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Frame = false
		v.Wrap = false
		v.Overwrite = true
		v.Autoscroll = false
		v.Highlight = true
		v.SelBgColor = gocui.ColorBlue
		v.SelFgColor = gocui.ColorWhite
		t.ResultsView = v
	}
	return nil
}

func (t *tui) search() {
	if t.Query() == "" {
		t.ResultsView.Clear()
		return
	}
	q := query{
		Text:      t.Query(),
		Highlight: "text",
	}
	b, err := json.Marshal(q)
	// TODO error handling
	if err != nil {
		return
	}
	t.conn.WriteMessage(websocket.TextMessage, b)
}

func (t *tui) handleResults(msg []byte) {
	var res *indexer.Results
	if err := json.Unmarshal(msg, &res); err != nil {
		// TODO
		return
	}
	t.results = res
	if t.selectedIdx >= t.getTotalResults() {
		t.selectedIdx = t.getTotalResults() - 1
	}
	if t.selectedIdx < 0 && t.getTotalResults() > 0 {
		t.selectedIdx = 0
	}
	t.renderResults()
}

func (t *tui) renderResults() {
	t.g.Update(func(_ *gocui.Gui) error {
		t.ResultsView.Clear()
		t.lineOffsets = make([]int, 0, t.getTotalResults())
		if t.results == nil || (len(t.results.Documents) == 0 && len(t.results.History) == 0) {
			if t.Query() != "" {
				fmt.Fprintf(t.ResultsView, "%s", gray("No results found"))
			}
			return nil
		}

		currentLine := 0
		for _, r := range t.results.History {
			t.lineOffsets = append(t.lineOffsets, currentLine)
			currentLine += t.renderHistoryItem(r)
		}
		for _, r := range t.results.Documents {
			t.lineOffsets = append(t.lineOffsets, currentLine)
			currentLine += t.renderResult(r)
		}
		return nil
	})
}

func (t *tui) enterResultsMode(g *gocui.Gui, v *gocui.View) error {
	g.Cursor = false
	_, err := g.SetCurrentView("results")
	if err != nil {
		return err
	}
	if t.selectedIdx < 0 && t.getTotalResults() > 0 {
		t.selectedIdx = 0
	}
	t.setCursorToSelected()
	return nil
}

func (t *tui) exitResultsMode(g *gocui.Gui, v *gocui.View) error {
	g.Cursor = true
	_, err := g.SetCurrentView("search-input")
	return err
}

func (t *tui) moveCursor(delta int) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		newIdx := t.selectedIdx + delta
		if t.results != nil && newIdx >= 0 && newIdx < t.getTotalResults() {
			t.selectedIdx = newIdx
			t.refreshResultsWithCursor()
		}
		return nil
	}
}

func (t *tui) refreshResultsWithCursor() {
	t.renderResults()
	t.setCursorToSelected()
}

func (t *tui) setCursorToSelected() {
	if t.results == nil || t.selectedIdx < 0 || t.selectedIdx >= len(t.lineOffsets) {
		return
	}

	targetLine := t.lineOffsets[t.selectedIdx]

	_, viewHeight := t.ResultsView.Size()
	_, oy := t.ResultsView.Origin()

	newOrigin := max(0, min(targetLine, targetLine-viewHeight/2))
	if targetLine < oy || targetLine >= oy+viewHeight {
		t.ResultsView.SetOrigin(0, newOrigin)
		oy = newOrigin
	}
	t.ResultsView.SetCursor(0, targetLine-oy)
}

func (t *tui) openSelected(g *gocui.Gui, v *gocui.View) error {
	if u := t.getSelectedURL(); u != "" {
		return browser.OpenURL(u)
	}
	return nil
}

func (t *tui) deleteSelected(g *gocui.Gui, v *gocui.View) error {
	if u := t.getSelectedURL(); u != "" {
		t.showConfirmationDialog(g, "Delete this result? (y/n)", func() {
			http.PostForm(t.cfg.BaseURL("/delete"), url.Values{"url": []string{u}})
			t.search()
		})
	}
	return nil
}

func (t *tui) showConfirmationDialog(g *gocui.Gui, message string, onConfirm func()) error {
	maxX, maxY := g.Size()
	width := len(message) + 4
	height := 5
	x0 := (maxX - width) / 2
	y0 := (maxY - height) / 2

	v, err := g.SetView("confirm-dialog", x0, y0, x0+width, y0+height)
	if err != nil && err != gocui.ErrUnknownView {
		return err
	}

	v.Frame = true
	v.Clear()
	fmt.Fprintln(v, "")
	fmt.Fprintf(v, "  %s\n", red(message))
	fmt.Fprint(v, "  [y/n]")

	g.SetCurrentView("confirm-dialog")

	cleanup := func(g *gocui.Gui, v *gocui.View) error {
		g.DeleteKeybindings("confirm-dialog")
		g.DeleteView("confirm-dialog")
		g.SetCurrentView("results")
		return nil
	}

	g.SetKeybinding("confirm-dialog", 'y', gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		onConfirm()
		return cleanup(g, v)
	})

	g.SetKeybinding("confirm-dialog", 'n', gocui.ModNone, cleanup)
	g.SetKeybinding("confirm-dialog", gocui.KeyTab, gocui.ModNone, cleanup)

	return nil
}

func (t *tui) getTotalResults() int {
	if t.results == nil {
		return 0
	}
	return len(t.results.History) + len(t.results.Documents)
}

func (t *tui) getSelectedURL() string {
	if t.results == nil || t.selectedIdx < 0 {
		return ""
	}
	if t.selectedIdx < len(t.results.History) {
		return t.results.History[t.selectedIdx].URL
	}
	docIdx := t.selectedIdx - len(t.results.History)
	if docIdx < len(t.results.Documents) {
		return t.results.Documents[docIdx].URL
	}
	return ""
}

func formatText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func (t *tui) renderHistoryItem(d *model.URLCount) int {
	fmt.Fprintf(t.ResultsView, "%s[History] %s\n", boldWhite(""), formatText(d.URL))
	fmt.Fprintf(t.ResultsView, "%s\n", blue(d.URL))
	fmt.Fprintln(t.ResultsView, "")

	return 3
}

func (t *tui) renderResult(d *indexer.Document) int {
	linesPrinted := 3
	fmt.Fprintf(t.ResultsView, "%s\n", boldWhite(formatText(d.Title)))
	fmt.Fprintf(t.ResultsView, "%s\n", blue(d.URL))
	if d.Text != "" {
		fmt.Fprintln(t.ResultsView, formatText(d.Text))
		linesPrinted++
	}
	fmt.Fprintln(t.ResultsView, "")

	return linesPrinted
}

func (t *tui) close() {
	t.conn.Close()
}

func (t *tui) Query() string {
	return strings.TrimSpace(t.SearchInput.Buffer())
}
