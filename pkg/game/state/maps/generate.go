package maps

import (
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/thorfour/larn/pkg/game/state/items"
	"github.com/thorfour/larn/pkg/game/state/monster"
	"github.com/thorfour/larn/pkg/game/state/types"
	"github.com/thorfour/larn/pkg/io"
)

const (
	homeLevel  = 0
	maxDungeon = 10
	MaxVolcano = 13 // 3 volcanos. 10 dungeons
)

// newMap is a wrapper of newLevel, it creates the level and places objects in the level.
func newMap(lvl uint) [][]io.Runeable {
	m := newLevel(lvl) // Create the level
	if lvl > 1 {       // Treausre rooms starting on level 2 of the dungeon
		treasureRoom(m) // TODO need to fill treasure rooms
	}

	seed := time.Now().UnixNano()
	log.WithField("seed", seed).Info("new map")

	// Seed the global rand // FIXME maps shouldn't use global rand
	rand.Seed(seed)

	placeMapObjects(lvl, m) // Add objects to the level
	return m
}

// newLevel creates a new map for a given level
// It creates a map full of walls and then carves out the pathways
// If level == 0 it returns an empty map for the home level
func newLevel(lvl uint) [][]io.Runeable {

	base := func() io.Runeable {
		return &Wall{DEBUG} // If DEBUG is set, maze will be visible by default
	}
	if lvl == homeLevel {
		base = func() io.Runeable { return Empty{true} }
	}

	// Generate full grid
	level := make([][]io.Runeable, height)
	for i := range level {
		row := make([]io.Runeable, width)
		for j := range row {
			row[j] = base()
		}
		level[i] = row
	}

	// Carve out the passageways
	if lvl != homeLevel {
		//carve(level)
		eat(types.Coordinate{1, 1}, level) // original larn
	}

	return level
}

// eat is the way orginal larn ate through the map of walls to create a maze
func eat(c types.Coordinate, lvl [][]io.Runeable) {
	dir := rand.Intn(4) + 1  // pick a random direction // TODO THOR use a time seed
	for try := 2; try > 0; { // try all directions twice
		switch dir {
		case 1: // West
			if c.X <= 2 || !isWall(types.Coordinate{c.X - 1, c.Y}, lvl) || !isWall(types.Coordinate{c.X - 2, c.Y}, lvl) { // Only eat if at least the next 2 are walls
				break
			}
			lvl[c.Y][c.X-1] = Empty{}
			lvl[c.Y][c.X-2] = Empty{}
			eat(types.Coordinate{c.X - 2, c.Y}, lvl)
		case 2: // East
			if c.X >= width-3 || !isWall(types.Coordinate{c.X + 1, c.Y}, lvl) || !isWall(types.Coordinate{c.X + 2, c.Y}, lvl) { // Only eat if at least the next 2 are walls
				break
			}

			lvl[c.Y][c.X+1] = Empty{}
			lvl[c.Y][c.X+2] = Empty{}
			eat(types.Coordinate{c.X + 2, c.Y}, lvl)
		case 3: // South
			if c.Y <= 2 || !isWall(types.Coordinate{c.X, c.Y - 1}, lvl) || !isWall(types.Coordinate{c.X, c.Y - 2}, lvl) { // Only eat if at least the next 2 are walls
				break
			}
			lvl[c.Y-1][c.X] = Empty{}
			lvl[c.Y-2][c.X] = Empty{}
			eat(types.Coordinate{c.X, c.Y - 2}, lvl)
		case 4: // North
			if c.Y >= height-3 || !isWall(types.Coordinate{c.X, c.Y + 1}, lvl) || !isWall(types.Coordinate{c.X, c.Y + 2}, lvl) { // Only eat if at least the next 2 are walls
				break
			}
			lvl[c.Y+1][c.X] = Empty{}
			lvl[c.Y+2][c.X] = Empty{}
			eat(types.Coordinate{c.X, c.Y + 2}, lvl)
		default:
			panic("Unknown direction")
		}
		if dir++; dir > 4 { // Pick a different direction
			dir = 1
			try--
		}
	}
}

// carve takes a map that consists of walls and carves out walls to create a maze
// it is an implementation of Randomized Prim's algorithm
func carve(lvl [][]io.Runeable) {

	// Pick a random wall and add it to walls list
	walls := []types.Coordinate{randMapCoord()}

	// Keep carving till the walls list is empty
	for len(walls) != 0 {

		// Get random wall from walls list
		i := rand.Intn(len(walls))
		w := walls[i]

		log.WithFields(log.Fields{
			"count":   len(walls),
			"current": w,
			"walls":   walls,
		}).Debug("carve loop")

		// If that wall is adjacent to one open space (less than for initial case)
		if emptyAdjacent(w, lvl) <= 1 {
			// convert the wall to open;
			lvl[w.Y][w.X] = Empty{}
			// and add all of it's neighboring walls to the list
			walls = append(walls, adjacentWalls(w, lvl)...)
		}

		// remove it from the walls list
		walls = append(walls[0:i], walls[i+1:]...)
	}
}

// adjacentWalls returns all adjacent walls in a map
func adjacentWalls(c types.Coordinate, lvl [][]io.Runeable) []types.Coordinate {
	adj := adjacent(c, true)
	var cords []types.Coordinate
	for _, s := range adj {
		switch lvl[s.Y][s.X].(type) {
		case *Wall:
			cords = append(cords, s)
		}
	}

	return cords
}

// getAdjacent returns orthogonally adjacent coordinates in a map, that are valid coordinates
// mapEdge indicates to include the edge of the map as well
func adjacent(w types.Coordinate, mapEdge bool) []types.Coordinate {
	var adj []types.Coordinate
	min := 0
	maxW := width - 1
	maxH := height - 1
	if mapEdge {
		min = 1
		maxW = width - 2
		maxH = height - 2
	}
	if int(w.X)+1 <= maxW {
		adj = append(adj, types.Coordinate{w.X + 1, w.Y})
	}
	if int(w.X)-1 >= min {
		adj = append(adj, types.Coordinate{w.X - 1, w.Y})
	}
	if int(w.Y)+1 <= maxH {
		adj = append(adj, types.Coordinate{w.X, w.Y + 1})
	}
	if int(w.Y)-1 >= min {
		adj = append(adj, types.Coordinate{w.X, w.Y - 1})
	}

	return adj
}

// diagonal returns the diagonally adjacent coordinates in a map that are valid coordinates
// if mapEdge is true it wont include the map edge
func diagonal(w types.Coordinate, mapEdge bool) []types.Coordinate {
	var adj []types.Coordinate
	min := 0
	maxW := width - 1
	maxH := height - 1
	if mapEdge {
		min = 1
		maxW = width - 2
		maxH = height - 2
	}
	if int(w.X)+1 <= maxW && int(w.Y)+1 <= maxH {
		adj = append(adj, types.Coordinate{w.X + 1, w.Y + 1})
	}
	if int(w.X)-1 >= min && int(w.Y)-1 >= min {
		adj = append(adj, types.Coordinate{w.X - 1, w.Y - 1})
	}
	if int(w.Y)+1 <= maxH && int(w.X)-1 >= min {
		adj = append(adj, types.Coordinate{w.X - 1, w.Y + 1})
	}
	if int(w.Y)-1 >= min && int(w.X)+1 <= maxW {
		adj = append(adj, types.Coordinate{w.X + 1, w.Y - 1})
	}

	return adj
}

// emptyAdjacent returns the number of adjacent spaces that are empty
func emptyAdjacent(c types.Coordinate, lvl [][]io.Runeable) int {
	adj := adjacent(c, true)
	count := 0
	for _, s := range adj {
		switch lvl[s.Y][s.X].(type) {
		case Empty:
			count++
		}
	}

	return count
}

func randMapCoord() types.Coordinate {
	return types.Coordinate{rand.Intn(width-1) + 1, rand.Intn(height-1) + 1}
}

// walkToEmpty takes coordincate c and randomly walks till it finds an empty location
func walkToEmpty(c types.Coordinate, lvl [][]io.Runeable) types.Coordinate {

	// Random walk till and empty room is found
	for {
		switch lvl[c.Y][c.X].(type) {
		case Empty:
			return c
		}
		xadj := rand.Intn(3) - 2 // [-1,1]
		yadj := rand.Intn(3) - 2 // [-1,1]
		if xadj > 0 {
			c.X += xadj
		} else {
			c.X -= xadj
		}

		if yadj > 0 {
			c.Y += yadj
		} else {
			c.Y -= yadj
		}

		if c.X > width-2 {
			c.X = 1
		}
		if c.X < 1 {
			c.X = width - 2
		}
		if c.Y > height-2 {
			c.Y = 1
		}
		if c.Y < 1 {
			c.Y = height - 2
		}
	}

	return c
}

// placeMultipleObjects places [0,N) objects of type o in lvl at random coordinates
func placeMultipleObjects(n int, f func() io.Runeable, lvl [][]io.Runeable) {
	for i := 0; i < n; i++ {
		// generate the object
		o := f()

		// Place the object in the map
		placeObject(randMapCoord(), o, lvl)
	}
}

// placeObject places an object in a maze at arandom open location
func placeObject(c types.Coordinate, o io.Runeable, lvl [][]io.Runeable) (types.Coordinate, io.Runeable) {
	log.WithField("rune", string(o.Rune())).Debug("place object")

	c = walkToEmpty(c, lvl)

	// Set the visibility of the object if possible
	if _, ok := o.(Visible); ok {
		o.(Visible).Visible(DEBUG)
	}

	// Add the object
	d := lvl[c.Y][c.X]
	lvl[c.Y][c.X] = o

	return c, d
}

// placeMapObjects places the required objects for a level
// it calls placeObject many times
func placeMapObjects(lvl uint, m [][]io.Runeable) {

	// Place the stairs
	if lvl == homeLevel {
		placeObject(randMapCoord(), Entrance{dungeonRune, 1, dungeonStr}, m)          // Entrance to the dungeon
		placeObject(randMapCoord(), Entrance{homeRune, HomeLvl, homeStr}, m)          // players home
		placeObject(randMapCoord(), Entrance{collegeRune, CollegeLvl, collegeStr}, m) // college of larn
		placeObject(randMapCoord(), Entrance{bankRune, BankLvl, bankStr}, m)          // 1st national bank of larn
		placeObject(randMapCoord(), Entrance{volRune, maxDungeon + 1, volStr}, m)     // volano shaft
		placeObject(randMapCoord(), Entrance{dndRune, DndLvl, dndStr}, m)             // the DND store
		placeObject(randMapCoord(), Entrance{tradeRune, TradeLvl, tradeStr}, m)       // the trading post
		placeObject(randMapCoord(), Entrance{lrsRune, LrsLvl, lrsStr}, m)             // the larn revenue service
	} else {
		// Place the stairs
		if lvl != 1 { // Dungeon level 1 has an entrance/exit doesn't need stairs up
			placeObject(randMapCoord(), &Stairs{Up, int(lvl - 1), DEBUG}, m)
		}
		if lvl != maxDungeon && lvl != MaxVolcano { // Last dungeon/volcano no stairs down
			placeObject(randMapCoord(), &Stairs{Down, int(lvl + 1), DEBUG}, m)
		}

		// Place random maze objects
		// Up to 2 objects per level
		placeMultipleObjects(rand.Intn(3), func() io.Runeable { return &items.Book{Level: lvl} }, m)
		placeMultipleObjects(rand.Intn(3), func() io.Runeable { return new(items.Altar) }, m)
		placeMultipleObjects(rand.Intn(3), func() io.Runeable { return new(items.Statue) }, m)
		placeMultipleObjects(rand.Intn(3), func() io.Runeable { return new(items.Pit) }, m)
		placeMultipleObjects(rand.Intn(3), func() io.Runeable { return new(items.Fountain) }, m)
		placeMultipleObjects(rand.Intn(3), func() io.Runeable { return &items.Trap{TrapType: items.ArrowTrap} }, m)
		placeMultipleObjects(rand.Intn(3)-1, func() io.Runeable { return &items.Trap{TrapType: items.TeleTrap} }, m)
		placeMultipleObjects(rand.Intn(3)-1, func() io.Runeable { return &items.Trap{TrapType: items.DartTrap} }, m)
		if lvl == 1 {
			placeObject(randMapCoord(), &items.Chest{Level: lvl}, m)
		} else {
			placeMultipleObjects(rand.Intn(2), func() io.Runeable { return &items.Chest{Level: lvl} }, m)
		}

		if lvl != maxDungeon && lvl != MaxVolcano {
			placeMultipleObjects(rand.Intn(2), func() io.Runeable { return &items.Trap{TrapType: items.DoorTrap} }, m)
		}

		if lvl <= 10 {
			placeMultipleObjects(rand.Intn(2), func() io.Runeable { return &items.Gem{Stone: items.Diamond, Value: rand.Intn(10*int(lvl)+1) + 10} }, m)
			placeMultipleObjects(rand.Intn(2), func() io.Runeable { return &items.Gem{Stone: items.Ruby, Value: rand.Intn(6*int(lvl)+1) + 6} }, m)
			placeMultipleObjects(rand.Intn(2), func() io.Runeable { return &items.Gem{Stone: items.Emerald, Value: rand.Intn(4*int(lvl)+1) + 4} }, m)
			placeMultipleObjects(rand.Intn(2), func() io.Runeable { return &items.Gem{Stone: items.Sapphire, Value: rand.Intn(3*int(lvl)+1) + 2} }, m)
		}

		placeMultipleObjects(rand.Intn(4)+4, func() io.Runeable { return items.NewPotion() }, m)
		placeMultipleObjects(rand.Intn(5)+4, func() io.Runeable { return items.NewScroll() }, m)
		placeMultipleObjects(rand.Intn(12)+12, func() io.Runeable {
			return &items.GoldPile{Amount: 12*rand.Intn(int(lvl+1)) + (int(lvl) << 3) + 10}
		}, m)
		// TODO Add level 5 bank branch office

		// Add armor to level
		placeRareObject(2, &items.ArmorClass{Type: items.RingMail}, m)
		placeRareObject(1, &items.ArmorClass{Type: items.StuddedLeather}, m)
		placeRareObject(3, &items.ArmorClass{Type: items.SplintMail}, m)
		s := &items.Shield{}
		s.ResetAttr(rand.Intn(3))
		placeRareObject(5, s, m)

		// Add weaspons to level
		ba := &items.WeaponClass{Type: items.BattleAxe}
		ba.ResetAttr(rand.Intn(3))
		placeRareObject(2, ba, m)
		ls := &items.WeaponClass{Type: items.LongSword}
		ls.ResetAttr(rand.Intn(3))
		placeRareObject(5, ls, m)
		fl := &items.WeaponClass{Type: items.Flail}
		fl.ResetAttr(rand.Intn(3))
		placeRareObject(5, fl, m)
		sp := &items.WeaponClass{Type: items.Spear}
		sp.ResetAttr(rand.Intn(5))
		placeRareObject(7, sp, m)
		placeRareObject(2, &items.WeaponClass{Type: items.SwordOfSlashing}, m)
		if lvl == 1 { // Bessman's hammer can only be created on level 1
			placeRareObject(4, &items.WeaponClass{Type: items.BessmansHammer}, m)
		}

		// TODO don't add these weapons is difficulty >= 3
		if rand.Intn(4) == 3 && lvl > 3 {
			ss := &items.WeaponClass{Type: items.SunSword}
			ss.ResetAttr(3)
			placeRareObject(3, ss, m)
			tws := &items.WeaponClass{Type: items.TwoHandedSword}
			tws.ResetAttr(rand.Intn(3) + 1)
			placeRareObject(5, tws, m)
			b := &items.Belt{}
			b.ResetAttr(4)
			placeRareObject(3, b, m)
			er := &items.Ring{Type: items.Energy}
			er.ResetAttr(3)
			placeRareObject(3, er, m)
			pm := &items.ArmorClass{Type: items.PlateMail}
			pm.ResetAttr(5)
			placeRareObject(4, pm, m)
		}

		// Add rings to level
		rr := &items.Ring{Type: items.Regen}
		rr.ResetAttr(rand.Intn(3))
		placeRareObject(4, rr, m)
		rp := &items.Ring{Type: items.Protection}
		rp.ResetAttr(rand.Intn(3))
		placeRareObject(1, rp, m)
		rs := &items.Ring{Type: items.Strength}
		rs.ResetAttr(4)
		placeRareObject(2, rs, m)

		// place special objects
		placeRareObject(3, &items.Special{Type: items.Orb}, m)
		placeRareObject(4, &items.Special{Type: items.Scarab}, m)
		placeRareObject(4, &items.Special{Type: items.Cube}, m)
		placeRareObject(3, &items.Special{Type: items.Device}, m)
		placeRareObject(3, &items.Special{Type: items.Amulet}, m)
	}
}

// placeRareObject will place the object on the map with a chance of prob/151
func placeRareObject(prob int, o io.Runeable, lvl [][]io.Runeable) {
	if rand.Intn(151) < prob {
		placeObject(randMapCoord(), o, lvl)
	}
}

// isWall returns true if the coordinate c is a wall on map lvl
func isWall(c types.Coordinate, lvl [][]io.Runeable) bool {
	switch lvl[c.Y][c.X].(type) {
	case *Wall:
		return true
	default:
		return false
	}
}

// treasureRoom creates a treasure room on a level
func treasureRoom(m [][]io.Runeable) {
	for x := 2 + rand.Intn(10); x < width-10; x += 10 {
		if rand.Intn(13) == 0 { // not every level gets a room
			tWidth := rand.Intn(6) + 4
			tHeight := rand.Intn(6) + 4
			y := rand.Intn(height-10) + 2 // uper left corner of room
			// TODO special handling for last level of dungeon and volcano?
			makeRoom(tWidth, tHeight, x, y, rand.Intn(9)+1, m)
		}
	}
}

func makeRoom(w, h, x, y, glyph int, m [][]io.Runeable) {
	log.WithFields(log.Fields{
		"coord":  types.Coordinate{X: x, Y: y},
		"width":  w,
		"height": h,
	}).Debug("make room")

	for i := x; i < x+w; i++ { // Create only walls where the room will be
		for j := y; j < y+h; j++ {
			m[j][i] = &Wall{DEBUG}
		}
	}

	// Clear out the interior
	for i := x + 1; i < x+w-1; i++ { // Create only walls where the room will be
		for j := y + 1; j < y+h-1; j++ {
			m[j][i] = Empty{}
		}
	}

	// TODO add objects
	// TODO add monsters

	// Add a door
	doorLoc := rand.Intn((w * 2) + ((h - 2) * 2))
	wallCount := 0
	for i := x; i < x+w; i++ {
		for j := y; j < y+h; j++ {
			// Determine if we're on a wall
			if i == x || i == x+w-1 || j == y || j == y+h-1 {
				if wallCount == doorLoc {
					m[j][i] = &items.Door{
						Open: false,
						DefaultItem: items.DefaultItem{
							Visibility: DEBUG,
						},
					}
					return
				}
				wallCount++
			}
		}
	}
}

// spawnMonsters will add monsters to a given dungeon level. If fresh is set, it will spawn a new set instead of an additive amount
// returns the list of monsters that were added
func spawnMonsters(m [][]io.Runeable, lvl uint, fresh bool) []*monster.Monster {
	num := (int(lvl) >> 1) + 1
	if fresh {
		num += rand.Intn(12) + 2
	}

	var monsterList []*monster.Monster

	// spawn num monsters
	for i := 0; i < num; i++ {
		mon := monster.New(monster.FromLevel(int(lvl)))
		monsterList = append(monsterList, mon)
		placeObject(randMapCoord(), mon, m)
	}

	return monsterList
}
