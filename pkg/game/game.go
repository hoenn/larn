package game

import (
	"fmt"

	"github.com/golang/glog"
	termbox "github.com/nsf/termbox-go"
	"github.com/thorfour/larn/pkg/game/data"
	"github.com/thorfour/larn/pkg/game/state"
	"github.com/thorfour/larn/pkg/game/state/character"
	"github.com/thorfour/larn/pkg/io"
)

var (
	Quit = fmt.Errorf("%s", "Quit")
	Save = fmt.Errorf("%s", "Save")
)

const (
	internalKeyBufferSize = 10
	borderRune            = rune('#')
	borderWidth           = 67
	borderHeight          = 17
	invMaxDisplay         = 5 // number of inventory items to display on a page at a time
)

// Game holds all current game information
type Game struct {
	settings     data.Settings
	currentState *state.State

	// input channel from keyboard
	input chan termbox.Event

	// inputHandler is the function that handles input from the keyboard
	inputHandler func(e termbox.Event)

	// Indicates if the game has hit an error
	err error
}

// SaveFilePresent returns true if a save file exists, and the name of the file
func saveFilePresent() (bool, string) {
	// TODO
	return false, ""
}

// New initializes a game state
func New() *Game {
	glog.V(1).Info("Creating new game")
	g := new(Game)
	g.inputHandler = g.defaultHandler
	g.input = make(chan termbox.Event, internalKeyBufferSize)

	if ok, saveFile := saveFilePresent(); ok {
		// TODO handle loading a save game
		fmt.Println(saveFile)
		return g
	}

	// TODO setup game settings
	// g.settings

	// Generate starting game state
	g.currentState = state.New()

	return g
}

// Start is the entrypoint to running a new game, should not return without a request from the user
func (g *Game) Start() error {
	if err := termbox.Init(); err != nil {
		return fmt.Errorf("termbox failed to initialize: %v", err)
	}
	defer termbox.Close()

	// Start a listener for user input
	go io.KeyboardListener(g.input)

	// If the game wasn't from a save file, display the welcome screen
	if !g.settings.FromSaveFile {
		g.renderWelcome()

		// Wait for first key stroke to bypass welcome
		<-g.input
	}

	// Render the game
	g.render(display(g.currentState))

	// Game logic
	return g.run()
}

// run is the main game handler loop
func (g *Game) run() error {
	for {
		// Check for a game error
		if g.err != nil {
			if g.err == Save || g.err == Quit { // Save or Quit aren't errors to return
				return nil
			}
			return g.err
		}

		// Get next input
		e := <-g.input

		// Handle the next input event
		g.inputHandler(e)
	}
}

func (g *Game) defaultHandler(e termbox.Event) {

	switch e.Ch {
	case 'H': // run left
		g.runAction(character.Left)
	case 'J': // run down
		g.runAction(character.Down)
	case 'K': // run up
		g.runAction(character.Up)
	case 'L': // run right
		g.runAction(character.Right)
	case 'Y': // run northwest
		g.runAction(character.UpLeft)
	case 'U': // run northeast
		g.runAction(character.UpRight)
	case 'B': // run southwest
		g.runAction(character.DownLeft)
	case 'N': // run southeast
		g.runAction(character.DownRight)
	case 'h': // move left
		g.currentState.Move(character.Left)
		g.render(display(g.currentState))
	case 'j': // move down
		g.currentState.Move(character.Down)
		g.render(display(g.currentState))
	case 'k': // move up
		g.currentState.Move(character.Up)
		g.render(display(g.currentState))
	case 'l': // move right
		g.currentState.Move(character.Right)
		g.render(display(g.currentState))
	case 'y': // move northwest
		g.currentState.Move(character.UpLeft)
		g.render(display(g.currentState))
	case 'u': // move northeast
		g.currentState.Move(character.UpRight)
		g.render(display(g.currentState))
	case 'b': // move southwest
		g.currentState.Move(character.DownLeft)
		g.render(display(g.currentState))
	case 'n': // move southeast
		g.currentState.Move(character.DownRight)
		g.render(display(g.currentState))
	case ',': // Pick up the item
		g.currentState.PickUp()
		g.render(display(g.currentState))
	case '^': // identify a trap
	case 'd': // drop an item
		g.inputHandler = g.drop()
	case 'v': // print program version
	case '?': // help screen
	case 'g': // give present pack weight
	case 'i': // inventory your pockets
		g.inputHandler = g.inventoryWrapper(g.currentState.Inventory()) // After an inventory request only Esc and space are accepted
	case 'A': // create diagnostic file
	case '.': // stay here
	case 'Z': // teleport yourself
	case 'c': // cast a spell
	case 'r': // read a scroll
	case 'q': // quaff a potion
	case 'W': // wear armor
	case 'T': // take off armor
	case 'w': // wield a weapon
	case 'P': // give tax status
	case 'D': // list all items found
	case 'e': // eat something
	case 'S': // save the game and quit
		g.err = Save
		return
	case 'Q': // quit the game
		g.err = Quit // Set the error to quit
		return
	case 'E': // Enter the building
		g.currentState.Enter()
		g.render(display(g.currentState))
	}
}

//  renderWelcome generates the welcome to larn message
func (g *Game) renderWelcome() {
	if g.err != nil {
		return
	}
	g.err = io.RenderWelcome(welcome)
}

func (g *Game) render(display [][]io.Runeable) {
	if g.err != nil {
		return
	}

	g.err = io.RenderNewGrid(display)
}

func (g *Game) renderCharacter(c character.Coordinate) {
	if g.err != nil {
		return
	}

	g.err = io.RenderCell(c.X, c.Y, '&', termbox.ColorGreen, termbox.ColorGreen)
}

func (g *Game) runAction(d character.Direction) {
	for moved := g.currentState.Move(d); moved; moved = g.currentState.Move(d) {
		g.render(display(g.currentState))
	}
}

// inventoryWrapper returns a truncated input handler, used after a user requests an inventory display
// it will render the first inventory list, and subsequent calls the the function it returns will render the remaining pages
func (g *Game) inventoryWrapper(s []string) func(termbox.Event) {
	offset := 0
	label := 'a' // first inventory item is labled as a)

	generateInv := func() []string {
		var inv []string
		inv = append(inv, "") // empty string at the top
		for i := 0; i < invMaxDisplay && offset < len(s); i++ {
			inv = append(inv, fmt.Sprintf("%s) %v", string(label), s[offset]))
			label++
			offset++
		}
		inv = append(inv, "   --- press space to continue ---") // add the help string at the bottom
		return inv
	}

	g.render(overlay(display(g.currentState), convert(generateInv())))

	return func(e termbox.Event) {
		switch e.Key {
		case termbox.KeyEsc: // Escape key
			g.inputHandler = g.defaultHandler
			g.render(display(g.currentState))
		case termbox.KeySpace: // Space key
			if offset < len(s) { // Render next page
				g.render(overlay(display(g.currentState), convert(generateInv())))
				return
			}
			// No more pages to display, remove the overlay
			g.inputHandler = g.defaultHandler
			g.render(display(g.currentState))
		default:
			glog.V(6).Infof("Receive invalid input: %s", string(e.Ch))
			return
		}
	}
}

// drop func to drop an item
func (g *Game) drop() func(termbox.Event) {
	glog.V(2).Infof("Drop requested")

	g.currentState.Log("What do you want to drop [* for all] ?")
	g.render(display(g.currentState))

	// Capute the input character for the item to drop
	return func(e termbox.Event) {
		g.inputHandler = g.defaultHandler

		switch e.Key {
		case termbox.KeyEsc:
			g.currentState.Log("aborted")
			g.render(display(g.currentState))
		default:
			label := 'a'
			for n, i := range g.currentState.Inventory() { // FIXME the drop function isn't stable (i.e dropping item a results in item be now becoming item a)
				if e.Ch == label {
					if err := g.currentState.Drop(n); err != nil { // drop item n
						g.currentState.Log(err.Error()) // unable to drop
						g.render(display(g.currentState))
						return
					}
					g.currentState.Log("You drop:")
					g.currentState.Log(fmt.Sprintf("%s) %v", string(label), i))
					g.render(display(g.currentState))
					return
				}
				label++
			}
			g.currentState.Log(fmt.Sprintf("You don't have item %s!", string(e.Ch)))
			g.render(display(g.currentState))
		}
	}
}
