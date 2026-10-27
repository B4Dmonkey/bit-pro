package tui

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/glamour/v2"
	"charm.land/glamour/v2/styles"
	"charm.land/lipgloss/v2"
	"github.com/B4Dmonkey/bit-pro/task"
)

type item struct {
	t        *task.Task
	hasBars  bool
	expanded bool
}

type reloadedMsg struct {
	tasks []*task.Task
	err   error
}

type tickMsg struct{}

type costMsg struct {
	track string
	md    string
	err   error
}

func (i item) FilterValue() string { return i.t.Title }
func (i item) Title() string       { return i.t.Title }
func (i item) Description() string { return i.t.ID + " · " + i.t.Status }

type keyMap struct {
	focus  key.Binding
	move   key.Binding
	expand key.Binding
	help   key.Binding
	quit   key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		focus:  key.NewBinding(key.WithKeys("left", "right"), key.WithHelp("←/→", "focus")),
		move:   key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑/↓", "move·scroll")),
		expand: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "expand")),
		help:   key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		quit:   key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.focus, k.move, k.expand, k.help, k.quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.focus, k.move, k.expand}, {k.help, k.quit}}
}

type model struct {
	list.Model
	viewport      viewport.Model
	modalViewport viewport.Model
	help          help.Model
	keys          keyMap
	detailWidth   int
	listWidth     int
	winWidth      int
	winHeight     int
	height        int
	style         string
	renderer      *glamour.TermRenderer
	reload        func() ([]*task.Task, error)
	approve       func(id string, approved bool) error
	cost          func(track string) (string, error)
	costs         map[string]string
	costInFlight  map[string]bool
	loaded        []*task.Task
	detailFocused bool
	modalOpen     bool
	expanded      map[string]bool
}

func visibleItems(tasks []*task.Task, expanded map[string]bool) []list.Item {
	items := make([]list.Item, 0, len(tasks))
	open := false

	for i, t := range tasks {
		if !isBar(t.ID) {
			open = expanded[t.ID]
			hasBars := i+1 < len(tasks) && isBar(tasks[i+1].ID)
			items = append(items, item{t: t, hasBars: hasBars, expanded: open})

			continue
		}

		if open {
			items = append(items, item{t: t})
		}
	}

	return items
}

func New(tasks []*task.Task) model {
	expanded := map[string]bool{}
	l := list.New(visibleItems(tasks, expanded), delegate{}, 0, 0)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)

	style := styles.LightStyle
	if lipgloss.HasDarkBackground(os.Stdin, os.Stdout) {
		style = styles.DarkStyle
	}

	vp := viewport.New()
	mvp := viewport.New()

	return model{
		Model:         l,
		viewport:      vp,
		modalViewport: mvp,
		help:          help.New(),
		keys:          newKeyMap(),
		style:         style,
		loaded:        tasks,
		expanded:      expanded,
		costs:         map[string]string{},
		costInFlight:  map[string]bool{},
	}
}

func sameTasks(a, b []*task.Task) bool {
	return slices.EqualFunc(a, b, func(x, y *task.Task) bool {
		return x.ID == y.ID &&
			x.Status == y.Status &&
			x.Title == y.Title &&
			x.Body == y.Body &&
			x.Phase == y.Phase &&
			x.PhaseLabel == y.PhaseLabel &&
			x.Approved == y.Approved
	})
}

const noUsage = "_no usage recorded_"

const pollInterval = time.Second

func tick() tea.Cmd {
	return tea.Tick(pollInterval, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m model) WithReload(r func() ([]*task.Task, error)) model {
	m.reload = r
	return m
}

func (m model) WithApprove(f func(id string, approved bool) error) model {
	m.approve = f
	return m
}

func (m model) WithCost(f func(track string) (string, error)) model {
	m.cost = f
	return m
}

func (m model) Init() tea.Cmd {
	var poll tea.Cmd
	if m.reload != nil {
		poll = tick()
	}

	return tea.Batch(poll, m.costCmd())
}

func (m model) costCmd() tea.Cmd {
	t := m.selected()
	if m.cost == nil || t == nil || isBar(t.ID) || m.costInFlight[t.ID] {
		return nil
	}

	cost, id := m.cost, t.ID
	m.costInFlight[id] = true

	return func() tea.Msg {
		md, err := cost(id)

		return costMsg{track: id, md: md, err: err}
	}
}

func (m model) handleCost(msg costMsg) (tea.Model, tea.Cmd) {
	delete(m.costInFlight, msg.track)

	if msg.err != nil {
		return m, nil
	}

	if prev, ok := m.costs[msg.track]; ok && prev == msg.md {
		return m, nil
	}

	m.costs[msg.track] = msg.md

	if t := m.selected(); t != nil && t.ID == msg.track {
		m.refreshDetail()
	}

	return m, nil
}

func (m model) reloadCmd() tea.Cmd {
	if m.reload == nil {
		return nil
	}

	reload := m.reload

	return func() tea.Msg {
		tasks, err := reload()

		return reloadedMsg{tasks: tasks, err: err}
	}
}

func (m *model) setTasks(tasks []*task.Task) {
	var prevID string

	prevIndex := m.Index()
	if t := m.selected(); t != nil {
		prevID = t.ID
	}

	items := visibleItems(tasks, m.expanded)

	m.SetItems(items)

	if prevID != "" && len(items) > 0 {
		target := min(prevIndex, len(items)-1)

		for i, it := range items {
			if it.(item).t.ID == prevID {
				target = i
				break
			}
		}

		m.Select(target)
	}

	if m.modalOpen {
		m.refreshModal()
	}

	m.loaded = tasks
	m.layout()
	m.refreshDetail()
}

func (m model) handleReloaded(msg reloadedMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		return m, tick()
	}

	if sameTasks(m.loaded, msg.tasks) {
		return m, tea.Batch(tick(), m.costCmd())
	}

	m.setTasks(msg.tasks)

	return m, tea.Batch(tick(), m.costCmd())
}

func (m model) handleKeyPress(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.modalOpen {
		return m.updateModal(msg)
	}

	if key.Matches(msg, m.keys.help) {
		m.help.ShowAll = !m.help.ShowAll
		m.layout()

		return m, nil
	}

	return m.handleListKey(msg)
}

func (m model) handleApprove() (tea.Model, tea.Cmd) {
	if m.approve != nil {
		if t := m.selected(); t != nil {
			_ = m.approve(t.ID, !t.Approved)

			return m, m.reloadCmd()
		}
	}

	return m, nil
}

func (m model) handleDetailFocusedKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}

func (m model) handleListKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", keyEsc, keyCtrlC:
		return m, tea.Quit
	case keyRight, "l":
		m.detailFocused = true

		return m, nil
	case keyLeft, "h":
		m.detailFocused = false

		return m, nil
	}

	if msg.Code == tea.KeyEnter {
		if m.selected() != nil {
			m.modalOpen = true
			m.refreshModal()
		}

		return m, nil
	}

	if msg.Code == ' ' {
		return m.handleApprove()
	}

	if m.detailFocused {
		return m.handleDetailFocusedKey(msg)
	}

	if msg.Code == tea.KeyTab {
		if t := m.selected(); t != nil {
			m.expanded[t.ID] = !m.expanded[t.ID]
			m.SetItems(visibleItems(m.loaded, m.expanded))
			m.refreshDetail()
		}

		return m, nil
	}

	prev := m.Index()

	var cmd tea.Cmd

	m.Model, cmd = m.Model.Update(msg)

	if m.Index() != prev {
		m.refreshDetail()

		return m, tea.Batch(cmd, m.costCmd())
	}

	return m, cmd
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		return m, m.reloadCmd()
	case reloadedMsg:
		return m.handleReloaded(msg)
	case costMsg:
		return m.handleCost(msg)
	case tea.WindowSizeMsg:
		m.winWidth, m.winHeight = msg.Width, msg.Height
		m.relayout()

		return m, nil
	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	var cmd tea.Cmd

	if m.detailFocused {
		m.viewport, cmd = m.viewport.Update(msg)

		return m, cmd
	}

	prev := m.Index()
	m.Model, cmd = m.Model.Update(msg)

	if m.Index() != prev {
		m.refreshDetail()

		return m, tea.Batch(cmd, m.costCmd())
	}

	return m, cmd
}

func newRenderer(style string, width int) *glamour.TermRenderer {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil
	}

	return r
}

func (m *model) refreshDetail() {
	t := m.selected()
	if t == nil {
		m.viewport.SetContent("")
		return
	}

	md := t.Body
	if cost, ok := m.costs[t.ID]; ok && !isBar(t.ID) {
		if cost == "" {
			cost = noUsage
		}

		md += "\n\n## Cost\n\n" + cost
	}

	body := md
	if m.renderer != nil {
		if out, err := m.renderer.Render(md); err == nil {
			body = out
		}
	}

	m.viewport.SetContent(body)
	m.viewport.GotoTop()
}

func (m *model) refreshModal() {
	t := m.selected()
	if t == nil {
		m.modalViewport.SetContent("")
		return
	}

	innerW, innerH := modalInner(m.winWidth, m.winHeight)
	m.modalViewport.SetWidth(innerW)
	m.modalViewport.SetHeight(innerH)

	body := t.Body
	if r := newRenderer(m.style, max(innerW, 1)); r != nil {
		if out, err := r.Render(t.Body); err == nil {
			body = out
		}
	}

	m.modalViewport.SetContent(body)
	m.modalViewport.GotoTop()
}

func (m model) helpKeys() help.KeyMap {
	if m.modalOpen {
		k := m.keys
		k.focus.SetHelp("←/→", "page")
		k.expand.SetEnabled(false)

		return k
	}

	return m.keys
}

func (m *model) relayout() {
	m.layout()
	m.renderer = newRenderer(m.style, max(m.detailWidth-2, 1))
	m.refreshDetail()
}

func (m *model) layout() {
	listW, detailW := splitWidth(m.winWidth)

	m.help.SetWidth(m.winWidth)
	helpHeight := lipgloss.Height(m.help.View(m.helpKeys()))
	paneHeight := max(m.winHeight-helpHeight, 0)
	m.listWidth, m.detailWidth, m.height = listW, detailW, paneHeight
	m.SetSize(max(listW-2, 0), max(paneHeight-2, 0))
	m.viewport.SetWidth(max(detailW-2, 0))
	m.viewport.SetHeight(max(paneHeight-2, 0))
}

func (m model) View() tea.View {
	v := tea.NewView(m.content())
	v.AltScreen = true

	return v
}

func (m model) content() string {
	listTitle := fmt.Sprintf("Tasks (%d/%d)", doneCount(m.loaded), len(m.loaded))
	listPane := titledBorder(m.Model.View(), listTitle, max(m.listWidth-2, 0), max(m.height-2, 0), !m.detailFocused)
	detailPane := titledBorder(m.viewport.View(), "Details", max(m.detailWidth-2, 0), max(m.height-2, 0), m.detailFocused)
	canvas := lipgloss.JoinHorizontal(lipgloss.Top, listPane, detailPane)

	if m.modalOpen {
		canvas = modalView(m, canvas)
	}

	return lipgloss.JoinVertical(lipgloss.Left, canvas, m.help.View(m.helpKeys()))
}

func doneCount(tasks []*task.Task) int {
	n := 0

	for _, t := range tasks {
		if t.Status == task.StatusDone {
			n++
		}
	}

	return n
}

func titledBorder(content, title string, width, height int, active bool) string {
	border := lipgloss.RoundedBorder()

	boxStyle := lipgloss.NewStyle().
		Border(border).
		BorderTop(false).
		Width(width + 2).
		Height(height + 1)
	topStyle := lipgloss.NewStyle()
	titleStyle := lipgloss.NewStyle()

	if active {
		accent := lipgloss.Color("2")
		boxStyle = boxStyle.BorderForeground(accent)
		topStyle = topStyle.Foreground(accent)
		titleStyle = titleStyle.Foreground(accent).Reverse(true)
	}

	titleSeg := " " + title + " "
	if !active {
		titleSeg = "| " + title + " |"
	}

	fill := max(width-lipgloss.Width(titleSeg)-1, 0)
	left := border.TopLeft + border.Top
	right := strings.Repeat(border.Top, fill) + border.TopRight
	top := topStyle.Render(left) + titleStyle.Render(titleSeg) + topStyle.Render(right)

	return lipgloss.JoinVertical(lipgloss.Left, top, boxStyle.Render(content))
}

func splitWidth(total int) (listW, detailW int) {
	listW = total * 40 / 100
	detailW = max(total-listW-1, 0)

	return listW, detailW
}

func isBar(id string) bool {
	return strings.Contains(id, ".")
}

func verse(t *task.Task) string {
	if !isBar(t.ID) || t.Phase == 0 {
		return ""
	}

	return fmt.Sprintf("phase %d — %s", t.Phase, t.PhaseLabel)
}

func (m model) selected() *task.Task {
	it, ok := m.SelectedItem().(item)
	if !ok {
		return nil
	}

	return it.t
}
