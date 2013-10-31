package termboxutil

import (
	"errors"
	termbox "github.com/nsf/termbox-go"
	"sync"
)

var screenMutex sync.Mutex

type Screen []*Window

var curWindow *Window

type Window struct {
	sync.Mutex

	xpos, ypos int
	selected   int // selected row
	scrollable bool
	scrollPos  int
	closing    bool
	autoResize bool
	screen     *Screen

	Fg, Bg       termbox.Attribute
	RowFg, RowBg termbox.Attribute // selected row colors
	Rows         []Row
	CatchEvent   func(termbox.Event)
}

type Row struct {
	Text   string
	Fg, Bg termbox.Attribute
}

func (s *Screen) Focus(w *Window) {

	screenMutex.Lock()
	curWindow = w
	screenMutex.Unlock()
}

func (s *Screen) NewWindow(fg, bg, rowFg, rowBg termbox.Attribute) Window {

	screenMutex.Lock()
	defer screenMutex.Unlock()

	window := Window{sync.Mutex{}, 0, 0, 0, false, 0, false, true, s, fg, bg, rowFg, rowBg, nil, nil}

	*s = append(*s, &window)
	curWindow = &window

	return window
}

func (s *Screen) Loop() {
	for {
		e := termbox.PollEvent()

		// handle error
		if e.Type == termbox.EventError {
			panic(e.Err)
		}

		w := curWindow // TODO rename w to curWindow

		// handle resize
		if w.autoResize && e.Type == termbox.EventResize {
			err := w.Redraw()
			if err != nil {
				panic(err) // TODO dont panic here
			}

			termbox.Flush()
			continue
		}

		if w.CatchEvent != nil {
			w.CatchEvent(e)
		}
	}
}

func (w *Window) Draw(data []string) error {

	w.Lock()
	w.selected = 0
	w.scrollPos = 0
	w.Rows = make([]Row, len(data))

	for i, str := range data {
		w.Rows[i] = Row{str, w.RowFg, w.RowBg}
	}

	w.Unlock()
	return w.Redraw()
}

func (w *Window) Redraw() error {

	w.Lock()

	err := termbox.Clear(w.Fg, w.Bg)
	maxx, maxy := termbox.Size()

	if err != nil {
		return err
	}

	for i, row := range w.Rows[w.scrollPos:] {
		for _, c := range row.Text {

			if i == w.selected {
				termbox.SetCell(w.xpos, w.ypos, rune(c), row.Fg, row.Bg)
			} else {
				termbox.SetCell(w.xpos, w.ypos, rune(c), w.Fg, w.Bg)
			}

			if w.xpos += 1; w.xpos > maxx {
				break
			}
		}
		w.xpos = 0

		if w.ypos += 1; w.ypos > maxy {
			break
		}
	}
	w.ypos = 0
	w.xpos = 0 // redundant but lets avoid problems later

	w.Unlock()
	return nil
}

func (w *Window) CurrentRow() *Row {
	return &w.Rows[w.selected+w.scrollPos]
}

// selects the next row
func (w *Window) NextRow() {

	_, maxy := termbox.Size()
	if w.scrollable == true {
		if w.selected+w.scrollPos == len(w.Rows)-1 {
			return // bottom of the visible output
		} else if w.selected == maxy-1 {
			w.ScrollDown()
		} else {
			w.selected += 1
		}
	} else {
		w.selected += 1
	}
}

// selects the prev row
func (w *Window) PrevRow() {

	if w.scrollable == false {
		w.selected -= 1
	} else {
		if w.selected == 0 {
			w.ScrollUp()
		} else {
			w.selected -= 1
		}
	}
}

func (w *Window) ScrollUp() {
	if w.scrollPos > 0 {
		w.scrollPos -= 1
	}
}

func (w *Window) ScrollDown() {
	w.scrollPos += 1
}

// set the Window as scrollable - pass the offset as int
// if zero scrolling is disabled
func (w *Window) Scrollable(togl bool) {
	w.scrollable = togl
}

func (w *Window) MarkRow(i int, fg, bg termbox.Attribute) error {
	if i < 0 || i > len(w.Rows)-1 {
		return errors.New("termbox: unknown row")
	}

	row := &w.Rows[i]
	row.Fg = fg
	row.Bg = bg

	return nil
}

func (w *Window) UnmarkRow(i int) error {
	return w.MarkRow(i, w.Fg, w.Bg)
}
