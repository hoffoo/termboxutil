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
	Data		 []Row
	Rows         []Row
	selected     int               // selected row
	RowFg, RowBg termbox.Attribute // selected row colors
	scrollable   bool
	scrollOffset int
}

type Row struct {
	Text   string
	Fg, Bg termbox.Attribute
}

func NewScreen(fg, bg, rowFg, rowBg termbox.Attribute) Screen {
	return Screen{sync.Mutex{}, 0, 0, fg, bg, nil, 0, rowFg, rowBg, false, 0}
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

	for i, row := range s.Rows[s.scrollOffset:] {
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

// selects the next row
func (s *Screen) NextRow() {
	_, maxy := termbox.Size()
	if s.selected == maxy-1 || s.selected > len(s.Rows) {
		return
	}
	s.selected += 1
}

// selects the prev row
func (s *Screen) PrevRow() {
	if s.selected == 0 {
		return
	}
	s.selected -= 1
}

func (s *Screen) Scrollable(f bool) {
	s.scrollable = f
}

func (s *Screen) ScrollUp() {

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
