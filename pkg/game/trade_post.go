package game

import (
	"bytes"
	"fmt"
	"text/tabwriter"
	"time"

	termbox "github.com/nsf/termbox-go"
	"github.com/thorfour/larn/pkg/game/state/items"
)

// MaxDisplay the max number of inventory items to display at once
const MaxDisplay = 26

func tradingPostWelcome() string {
	return `
  Welcome to the Larn Trading Post.  We buy items that explorers no longer find
  useful.  Since the condition of the items you bring in is not certain,
  and we incur great expense in reconditioning the items, we usually pay
  only 20% of their value were they to be new.  If the items are badly
  damaged, we will pay only 10% of their new value.
	  
  Here are the items we would be willing to buy from you:
`
}

func tradingPost(inv []string) string {
	s := tradingPostWelcome()
	buf := bytes.NewBuffer(make([]byte, 100))
	w := tabwriter.NewWriter(buf, 5, 0, 1, ' ', tabwriter.TabIndent)
	for i, t := range inv {
		if i >= MaxDisplay { // only display up to max
			break
		}
		// TODO don't display items that aren't identified
		if i%2 == 0 {
			fmt.Fprintf(w, "  %s\t\t\t\t", t)
		} else {
			fmt.Fprintf(w, "%s\n", t)
		}
	}
	w.Flush()

	s = s + string(buf.Bytes())

	// Pad out with empty lines
	for i := len(inv); i < MaxDisplay/2; i++ {
		s = s + "\n"
	}

	return s + "  What item do you want to sell us? [Press escape to leave]"
}

// tradingPostHandler input handler for the trading post
func (g *Game) tradingPostHandler() func(termbox.Event) {
	g.renderSplash(tradingPost(g.currentState.C.Inventory()))
	return func(e termbox.Event) {
		switch e.Key {
		case termbox.KeyEsc: // Exit
			g.inputHandler = g.defaultHandler
			g.render(display(g.currentState))
		default:
			switch e.Ch {
			case 'a':
				fallthrough
			case 'b':
				fallthrough
			case 'c':
				fallthrough
			case 'd':
				fallthrough
			case 'e':
				fallthrough
			case 'f':
				fallthrough
			case 'g':
				fallthrough
			case 'h':
				fallthrough
			case 'i':
				fallthrough
			case 'j':
				fallthrough
			case 'k':
				fallthrough
			case 'l':
				fallthrough
			case 'm':
				fallthrough
			case 'n':
				fallthrough
			case 'o':
				fallthrough
			case 'p':
				fallthrough
			case 'q':
				fallthrough
			case 'r':
				fallthrough
			case 's':
				fallthrough
			case 't':
				fallthrough
			case 'u':
				fallthrough
			case 'v':
				fallthrough
			case 'w':
				fallthrough
			case 'x':
				fallthrough
			case 'y':
				fallthrough
			case 'z':
				g.handleSellingInv(e.Ch)
			}
		}
	}
}

func (g *Game) validateItemSale(r rune) error {

	//  Check if they have the item
	i := g.currentState.C.Item(r)
	if i == nil {
		return fmt.Errorf("\n\n  You don't have item %s!", string(r))
	}

	//  check if the item is identified
	known := true
	switch t := i.(type) {
	case *items.Potion:
		known = items.KnownPotion(t.ID)
	case *items.Scroll:
		known = items.KnownScroll(t.ID)
	}

	if !known {
		return fmt.Errorf("\n\n  Sorry, we can't accept unidentified objects")
	}

	return nil
}

func (g *Game) handleSellingInv(r rune) {
	if err := g.validateItemSale(r); err != nil {
		g.renderSplash(tradingPost(g.currentState.C.Inventory()) + err.Error())
		time.Sleep(time.Millisecond * 700)
		g.renderSplash(tradingPost(g.currentState.C.Inventory()))
		return
	}

	// value the item
	val := 0
	i := g.currentState.C.Item(r)
	dndPrice := g.DNDStoreLookup(i)
	switch t := i.(type) {
	// TODO handle eye of larn case
	case *items.Gem:
		val = 20 * t.Value
	case *items.Scroll:
		val = 2 * dndPrice
	case *items.Potion:
		val = 2 * dndPrice
	case items.Attributable:
		val = dndPrice
		// Add value if item has positive modifiers
		if a := t.Attr(); a >= 0 {
			val *= 2
			for ; a > 0 && val < 500000; a-- { // NOTE: This is the goofiest modifier
				val = (14 * (67 + val) / 10)
			}
		}
	default:
		val = dndPrice
	}

	g.renderSplash(tradingPost(g.currentState.C.Inventory()) + fmt.Sprintf("\n\n  Item (%s) is worth %d gold pieces to us. Do you want to sell it?", string(r), val))
	g.inputHandler = g.sellConfirmationHandler(r, val)
}

// sellConfirmationHandler waits for confirmation to sell an item
func (g *Game) sellConfirmationHandler(r rune, val int) func(termbox.Event) {
	return func(e termbox.Event) {
		switch e.Ch {
		case 'y':
			fallthrough
		case 'Y': // Sale
			g.renderSplash(tradingPost(g.currentState.C.Inventory()) + fmt.Sprintf("\n\n  Item (%s) is worth %d gold pieces to us. Do you want to sell it?", string(r), val) + "\n\n  yes")
			g.currentState.C.DropItem(r)             // Remove the item from players inventory
			g.currentState.C.Stats.Gold += uint(val) // Add the value of the sale to the users gold
			time.Sleep(time.Millisecond * 700)
			g.renderSplash(tradingPost(g.currentState.C.Inventory()))
			g.inputHandler = g.tradingPostHandler()
			return
		default: // No sale
			g.renderSplash(tradingPost(g.currentState.C.Inventory()) + fmt.Sprintf("\n\n  Item (%s) is worth %d gold pieces to us. Do you want to sell it?", string(r), val) + "\n\n  no thanks.")
			time.Sleep(time.Millisecond * 700)
			g.renderSplash(tradingPost(g.currentState.C.Inventory()))
			g.inputHandler = g.tradingPostHandler()
			return
		}
	}
}

// DNDStoreLookup reutns the price of an item in the DND store
func (g *Game) DNDStoreLookup(t items.Item) int {
	for i := range store {
		for _, sale := range store[i] {
			if sale.String() == t.String() { // Compare the items based on their display name TODO this is kinda gross and would be nice to be replaced by an actual type comparison
				return sale.price / 10 // (reduce all sales by a factor of 10)
			}
		}
	}

	return 0
}
