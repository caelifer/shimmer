// +build js,wasm

package shimmer

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"syscall/js"
	"time"
)

type Shimmer struct {
	buf                      bytes.Buffer
	onImgLoadCb, shutdownCb  js.Callback
	brightnessCb, contrastCb js.Callback
	hueCb, satCb             js.Callback
	sourceImg                image.Image

	done chan struct{}
}

// New returns a new instance of shimmer
func New() *Shimmer {
	return &Shimmer{
		done: make(chan struct{}),
	}
}

// Start sets up all the callbacks and waits for the close signal
// to be sent from the browser.
func (s *Shimmer) Start() {
	// Setup callbacks
	s.setupOnImgLoadCb()
	js.Global.Get("document").
		Call("getElementById", "sourceImg").
		Call("addEventListener", "load", s.onImgLoadCb)

	s.setupBrightnessCb()
	js.Global.Get("document").
		Call("getElementById", "brightness").
		Call("addEventListener", "change", s.brightnessCb)

	s.setupContrastCb()
	js.Global.Get("document").
		Call("getElementById", "contrast").
		Call("addEventListener", "change", s.contrastCb)

	s.setupHueCb()
	js.Global.Get("document").
		Call("getElementById", "hue").
		Call("addEventListener", "change", s.hueCb)

	s.setupSatCb()
	js.Global.Get("document").
		Call("getElementById", "sat").
		Call("addEventListener", "change", s.satCb)

	s.setupShutdownCb()
	js.Global.Get("document").
		Call("getElementById", "close").
		Call("addEventListener", "click", s.shutdownCb)

	<-s.done
	s.log("Shutting down app")
	s.onImgLoadCb.Close()
	s.brightnessCb.Close()
	s.contrastCb.Close()
	s.hueCb.Close()
	s.satCb.Close()
	s.shutdownCb.Close()
}

// updateImage writes the image to a byte buffer and then converts it to base64.
// Then it sets the value to the src attribute of the target image.
func (s *Shimmer) updateImage(img *image.RGBA, start time.Time) {
	width := img.Bounds().Max.X - img.Bounds().Min.X
	height := img.Bounds().Max.Y - img.Bounds().Min.Y
	js.Global.Call("UpdateCanvas", img.Pix, width, height)

	fmt.Println("time taken:", time.Now().Sub(start))
}

// utility function to log a msg to the UI from inside a callback
func (s *Shimmer) log(msg string) {
	js.Global.Get("document").
		Call("getElementById", "status").
		Set("innerText", msg)
}

func (s *Shimmer) setupShutdownCb() {
	s.shutdownCb = js.NewEventCallback(js.PreventDefault, func(ev js.Value) {
		s.done <- struct{}{}
	})
}
