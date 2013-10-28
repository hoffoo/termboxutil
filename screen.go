package termboxutil

import (
	termbox "github.com/nsf/termbox-go"
	"errors"
	"sync"
)

type Screen struct {
	sync.Mutex
	ch chan termbox.Event

	xpos, ypos int
	Fg, Bg     termbox.Attribute
	Rows       []Row
}

type Row struct {
	Text string
	Fg, Bg termbox.Attribute
}

func NewScreen(fg, bg termbox.Attribute) Screen {
	return Screen{sync.Mutex{}, make(chan termbox.Event), 0, 0, fg, bg, nil}
}

func (s *Screen) Draw(data []string) error {

	s.Lock()
	s.Rows = make([]Row,len(data))

	for i, str := range data {
		s.Rows[i] = Row{str, s.Fg, s.Bg}
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

	for _, row := range s.Rows {
		for _, c := range row.Text {
			termbox.SetCell(s.xpos, s.ypos, rune(c), row.Fg, row.Bg)
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
