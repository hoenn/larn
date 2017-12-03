package game

import (
	"fmt"

	"github.com/golang/glog"
	termbox "github.com/nsf/termbox-go"
	"github.com/thorfour/larn/pkg/game/state"
	"github.com/thorfour/larn/pkg/io"
)

const (
	logLength = 5 // Number of lines to display for the status log
)

type Simple rune

func (s Simple) Rune() rune            { return rune(s) }
func (s Simple) Fg() termbox.Attribute { return termbox.ColorDefault }
func (s Simple) Bg() termbox.Attribute { return termbox.ColorDefault }

// display returns a 2d slice representation of the game
func display(s *state.State) [][]io.Runeable {
	return cat(s.CurrentMap(), infoBarGrid(s), statusLog(s))
}

// infoBarGrid returns the info bar in display grid format
func infoBarGrid(s *state.State) [][]io.Runeable {
	r := make([][]io.Runeable, 2)

	info := fmt.Sprintf("Spells: %v( %v) AC: %v WC: %v Level %v Exp: %v %s", s.C.Spells, s.C.MaxSpells, s.C.Ac, s.C.Wc, s.C.Level, s.C.Exp, s.C.Title)
	for _, c := range info {
		r[0] = append(r[0], Simple(c))
	}

	info = fmt.Sprintf("HP: %v( %v) STR=%v INT=%v WIS=%v CON=%v DEX=%v CHA=%v LV: %v Gold: %v", s.C.Hp, s.C.MaxHP, s.C.Str, s.C.Intelligence, s.C.Wisdom, s.C.Con, s.C.Dex, s.C.Cha, s.C.Loc, s.C.Gold)
	for _, c := range info {
		r[1] = append(r[1], Simple(c))
	}

	return r
}

// cat concatenats all the maps together
func cat(maps ...[][]io.Runeable) [][]io.Runeable {
	for i, _ := range maps {
		if i == 0 {
			continue
		}
		maps[0] = append(maps[0], maps[i]...)
	}
	return maps[0]
}

// statusLog returns the status log that's displayed on the bottom
func statusLog(s *state.State) [][]io.Runeable {

	r := make([][]io.Runeable, logLength)

	// Convert the status log to runes
	for i, logmsg := range s.StatLog {
		glog.V(6).Info("Adding log message '%s'", logmsg)
		for _, c := range logmsg {
			r[i] = append(r[i], Simple(c))
		}
	}

	return r
}
