package bot

import (
	"image"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KalebHawkins/gamebot"
)

const (
	// procName is the name of the targeted process.
	procName = "NewWorld.exe"
	// actionImage is the name of the image for the action icon.
	actionImage = "img/actionIcon.png"
	// treeTexture is a sample of the texture of some trees in monarch's bluff.
	treeTexture = "img/monarchsTreeTexture01.png"

	turnPixelRangeMin = 100
	turnPixelRangeMax = 200

	// actionImageConfidence is the confidence level required to determine if an actual action is present.
	actionImageConfidence = 0.85
	// treeTextureConfidence is the confidence level required to determine if an actual action is present.
	treeTextureConfidence = 0.50
	// actionKey is the key that is tapped when an action is to be performed.
	actionKey = "e"
	// moveForwardKey the key to press to move forward.
	moveForwardKey = "w"
)

type direction int

const (
	Left direction = iota
	Right
)

type nwBot struct {
	// rwMut   sync.RWMutex
	gamebot *gamebot.Bot

	actionImage           *image.Image
	actionImageConfidence float32
	treeTexture           *image.Image
	treeTextureConfidence float32

	turnPixelRangeMin int
	turnPixelRangeMax int

	actionKey      string
	moveForwardKey string

	// turnPixelRangeMin int
	// turnPixelRangeMax int
}

// NewNwBot() create a new bot instance.
func NewNwBot() (*nwBot, error) {
	log.Printf("initializing new world bot using process %s\n", procName)
	gb, err := gamebot.NewBot(procName)
	if err != nil {
		log.Fatal(err)
	}
	nwb := &nwBot{}

	log.Printf("loading action img %s\n", actionImage)
	actionImg, err := nwb.gamebot.OpenImage(actionImage)
	if err != nil {
		return nil, err
	}
	log.Printf("loading action img %s\n", treeTexture)
	treeImg, err := nwb.gamebot.OpenImage(treeTexture)
	if err != nil {
		return nil, err
	}

	nwb.gamebot = gb
	nwb.actionImage = actionImg
	nwb.treeTexture = treeImg
	nwb.actionImageConfidence = actionImageConfidence
	nwb.treeTextureConfidence = treeTextureConfidence
	nwb.turnPixelRangeMin = turnPixelRangeMin
	nwb.turnPixelRangeMax = turnPixelRangeMax
	nwb.actionKey = actionKey
	nwb.moveForwardKey = moveForwardKey

	log.Printf("window set to %s\n", nwb.gamebot.Window())
	log.Printf("actionImage set to %s\n", actionImage)
	log.Printf("treeTexture set to %s\n", treeTexture)
	log.Printf("actionKey set to %s\n", nwb.actionKey)
	log.Printf("actionImageConfidence set to %.2f\n", nwb.actionImageConfidence)
	log.Printf("treeTextureConfidence set to %.2f\n", nwb.treeTextureConfidence)
	log.Printf("moveForwardKey set to %s\n", nwb.moveForwardKey)
	return nwb, nil
}

// (nwb *nwBot) moveForward will move the character forward pressing the w key.
func (nwb *nwBot) moveForward() {
	if !nwb.gamebot.IsKeyDown("e") {
		log.Printf("pressed %s key to move forward\n", nwb.moveForwardKey)
		nwb.gamebot.PressKey(nwb.moveForwardKey)
	}
}

// (nwb *nwBot) isActionAvailable check to see if there is an action currently available.
func (nwb *nwBot) scanForAction(errChan chan error) {
	for {
		_, mxv, _, mxl, err := nwb.gamebot.DetectImage(nwb.gamebot.CaptureWindow(), nwb.actionImage)
		if err != nil {
			errChan <- err
		}

		// fmt.Println(mxv)
		if mxv >= nwb.actionImageConfidence {
			wX, wY := nwb.gamebot.Window().Position()
			log.Printf("action was detected with %.2f confidence at (%d, %d)\n", mxv, mxl.X+wX, mxl.Y+wY)
			for _, v := range nwb.gamebot.KeysDown() {
				log.Printf("releasing %s key to perform detected action\n", v)
				nwb.gamebot.ReleaseKey(v)
			}

			log.Printf("pressing the %s key to perform detected action\n", nwb.actionKey)
			nwb.gamebot.PressKey(nwb.actionKey)
			nwb.gamebot.Sleep(4)
			nwb.gamebot.ReleaseKey(nwb.actionKey)
		}
	}
}

func (nwb *nwBot) turnRandomDirection() {
	dir := nwb.gamebot.RandomInt(int(Left), int(Right))

	wpx, wpy := nwb.gamebot.Window().Position()
	wsx, wsy := nwb.gamebot.Window().Size()
	dx, dy := (wpx+wsx)/2, (wpy+wsy)/2
	nwb.gamebot.SetCursor(dx, dy)

	turnPixels := nwb.gamebot.RandomInt(nwb.turnPixelRangeMin, nwb.turnPixelRangeMax)

	cmpx, cmpy := nwb.gamebot.MousePosition()
	log.Printf("Mouse position before turning is (%d, %d)", cmpx, cmpy)
	switch direction(dir) {
	case Left:
		log.Printf("turning %d pixels to the left\n", turnPixels)
		nwb.gamebot.MoveCursorRelative(-turnPixels, 0)
	case Right:
		log.Printf("turning %d pixels to the right\n", turnPixels)
		nwb.gamebot.MoveCursorRelative(turnPixels, 0)
	}

	cmpx, cmpy = nwb.gamebot.MousePosition()
	log.Printf("Mouse position after turning is (%d, %d)", cmpx, cmpy)
}

func (nwb *nwBot) funcManager(errChan chan error) {
	moveForwardTicker := time.NewTicker(time.Second * 7)
	turnRandomTicker := time.NewTicker(time.Second * 10)

	for {
		select {
		case err := <-errChan:
			log.Fatal(err)
		case <-moveForwardTicker.C:
			nwb.moveForward()
		case <-turnRandomTicker.C:
			nwb.turnRandomDirection()
		}
	}
}

func (nwb *nwBot) Run() error {
	rand.Seed(time.Now().Unix())
	nwb.gamebot.Window().SetActive()
	errChan := make(chan error, 1)

	go nwb.funcManager(errChan)
	go nwb.scanForAction(errChan)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
	return nil
}
