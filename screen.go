package termboxutil

import (
	"errors"
	termbox "github.com/nsf/termbox-go"
	"sync"
)

type Screen struct {
	sync.Mutex

	xpos, ypos   int
	Fg, Bg       termbox.Attribute
	RowFg, RowBg termbox.Attribute // selected row colors
	Rows         []Row
	selected     int // selected row
	scrollable   bool
	scrollPos    int
}

type Row struct {
	Text   string
	Fg, Bg termbox.Attribute
}

func NewScreen(fg, bg, rowFg, rowBg termbox.Attribute) Screen {
	return Screen{sync.Mutex{}, 0, 0, fg, bg, rowFg, rowBg, nil, 0, false, 0}
}

func (s *Screen) Draw(data []string) error {

	s.Lock()
	s.Rows = make([]Row, len(data))

	for i, str := range data {
		s.Rows[i] = Row{str, s.RowFg, s.RowBg}
	}

	s.Unlock()
	return s.Redraw()
}

func (s *Screen) Redraw() error {

	s.Lock()

	err := termbox.Clear(s.Fg, s.Bg)
	maxx, maxy := termbox.Size()

	if err != nil {
		return err
	}

	for i, row := range s.Rows[s.scrollPos:] {
		for _, c := range row.Text {

			if i == s.selected {
				termbox.SetCell(s.xpos, s.ypos, rune(c), row.Fg, row.Bg)
			} else {
				termbox.SetCell(s.xpos, s.ypos, rune(c), s.Fg, s.Bg)
			}

			if s.xpos += 1; s.xpos > maxx {
				break
			}
		}
		s.xpos = 0

		if s.ypos += 1; s.ypos > maxy {
			break
		}
	}
	s.ypos = 0
	s.xpos = 0 // redundant but lets avoid problems later

	s.Unlock()
	return nil
}

func (s *Screen) CurrentRow() *Row {
	return &s.Rows[s.selected+s.scrollPos]
}

// selects the next row
func (s *Screen) NextRow() {

	_, maxy := termbox.Size()
	if s.scrollable == true {
		if s.selected+s.scrollPos == len(s.Rows)-1 {
			return // bottom of the visible output
		} else if s.selected == maxy-1 {
			s.ScrollDown()
		} else {
			s.selected += 1
		}
	} else {
		s.selected += 1
	}
}

// selects the prev row
func (s *Screen) PrevRow() {

	if s.scrollable == false {
		s.selected -= 1
	} else {
		if s.selected == 0 {
			s.ScrollUp()
		} else {
			s.selected -= 1
		}
	}
}

func (s *Screen) ScrollUp() {
	if s.scrollPos > 0 {
		s.scrollPos -= 1
	}
}

func (s *Screen) ScrollDown() {
	s.scrollPos += 1
}

// toggle if NextRow() and PrevRow will scroll the window
func (s *Screen) Scrollable(togl bool) {
	s.scrollable = togl
}

func (s *Screen) MarkRow(i int, fg, bg termbox.Attribute) error {
	if i < 0 || i > len(s.Rows)-1 {
		return errors.New("termbox: unknown row")
	}

	row := &s.Rows[i]
	row.Fg = fg
	row.Bg = bg

	return nil
}

func (s *Screen) UnmarkRow(i int) error {
	return s.MarkRow(i, s.Fg, s.Bg)
}
