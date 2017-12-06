package items

import (
	"strconv"

	"github.com/thorfour/larn/pkg/game/state/stats"
)

type ArmorType int

const armorRune = '['

const (
	Leather ArmorType = iota
	StuddedLeather
	RingMail
	ChainMail
	SplintMail
	PlateMail
	PlateArmor
	StainlessPlateArmor
)

// Map of all the armor base values
var armorBase = map[ArmorType]int{
	Leather:             2,
	StuddedLeather:      3,
	RingMail:            5,
	ChainMail:           6,
	SplintMail:          7,
	PlateMail:           9,
	PlateArmor:          10,
	StainlessPlateArmor: 12,
}

// Map of all the displayable armor names
var armorName = map[ArmorType]string{
	Leather:             "leather",
	StuddedLeather:      "studded leather",
	RingMail:            "ring mail",
	ChainMail:           "chain mail",
	SplintMail:          "splint mail",
	PlateMail:           "plate mail",
	PlateArmor:          "plate armor",
	StainlessPlateArmor: "stainless plate armor",
}

// ArmorClass satisfies the item interface as well as the Armor Interface
type ArmorClass struct {
	Type      ArmorType // the type of armor
	Attribute int       // the attributes of the armor that add/subtract from the class
	DefaultItem
	NoStats
}

// Rune implements the io.Runeable interface
func (a *ArmorClass) Rune() rune {
	return armorRune
}

// Log implements the Loggable interface
func (a *ArmorClass) Log() string {
	return "You have found a " + a.String()
}

// String implements the Item interface
func (a *ArmorClass) String() string {
	if a.Attribute < 0 {
		return armorName[a.Type] + " " + strconv.Itoa(a.Attribute)
	} else if a.Attribute > 0 {
		return armorName[a.Type] + " +" + strconv.Itoa(a.Attribute)
	}
	return armorName[a.Type]
}

// Wear implements the Armor interface
func (a *ArmorClass) Wear(c *stats.Stats) {
	c.Ac += (armorBase[a.Type] + a.Attribute)

}

// TakeOff implements the Armor interface
func (a *ArmorClass) TakeOff(c *stats.Stats) {
	c.Ac -= (armorBase[a.Type] + a.Attribute)
}
