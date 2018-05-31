package monster

import (
	"math/rand"

	termbox "github.com/nsf/termbox-go"
	"github.com/thorfour/larn/pkg/game/state/character"
	"github.com/thorfour/larn/pkg/io"
)

const (
	InvisibleRune = ' '
)

type Monster struct {
	Id         int         // the lookup id for the monster
	Info       MonsterType // The monstertype unique to this monster
	Visibility bool        // if the player can see where this monster is
	Displaced  io.Runeable // the object currently displaced by this monster
}

// Rune implements the io.Runeable interface
func (m *Monster) Rune() rune {
	if m.Visibility {
		return m.Info.MonsterRune
	}
	return InvisibleRune
}

// ID implements the monster interface
func (m *Monster) ID() int { return m.Id }

// Bg implements the io.Runeable interface
func (m *Monster) Bg() termbox.Attribute { return termbox.ColorDefault }

// Fg implements the io.Runeable interface
func (m *Monster) Fg() termbox.Attribute { return termbox.ColorDefault }

// Visible implements the Visibility interface
func (m *Monster) Visible(v bool) { m.Visibility = v }

// Damage implements the Monster interface
func (m *Monster) Damage(c *character.Character) {
	// TODO
}

// Name returns the name of the monster
func (m *Monster) Name() string { return NameFromID(m.Id) }

// New returns a new Monster from a monster id
func New(monster int) *Monster {
	return &Monster{
		Id:        monster,
		Info:      monsterData[monster],
		Displaced: Empty{},
	}
}

// Empty represents an empty map location
type Empty struct {
	visible bool
}

func (e Empty) Visible(v bool) { e.visible = v }

// Displace implementes the Displaceable interface
func (e Empty) Displace() bool { return true }

// Rune implements the io.Runeable interface
func (e Empty) Rune() rune {
	if e.visible {
		return '.'
	} else {
		return ' '
	}
}

// Fg implements the io.Runeable interface
func (e Empty) Fg() termbox.Attribute { return termbox.ColorDefault }

// Bg implements the io.Runeable interface
func (e Empty) Bg() termbox.Attribute { return termbox.ColorDefault }

// BaseDamage returns the base damage of a monster
func (m *Monster) BaseDamage() int {
	switch m.Id {
	case Bat: // bats deal a base of 1 always
		return 1
	default:
		d := m.Info.Dmg
		if d < 1 {
			d += 1
		} else {
			d += rand.Intn(d)
		}
		d += m.Info.Lvl
		return d
	}
}