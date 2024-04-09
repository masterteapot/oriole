// This is a simplified go-reimplementation of the gst-launch-<version> cli tool.
// It has no own parameters and simply parses the cli arguments as launch syntax.
// When the parsing succeeded, the pipeline is run until the stream ends or an error happens.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
)

// Run is used to wrap the given function in a main loop and print any error
func run(f func() error) {
	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

	go func() {
		if err := f(); err != nil {
			fmt.Println("ERROR!", err)
		}
		mainLoop.Quit()
	}()

	mainLoop.Run()
}

// RunLoop is used to wrap the given function in a main loop and print any error.
// The main loop itself is passed to the function for more control over exiting.
func runLoop(f func(*glib.MainLoop) error) {
	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)

	if err := f(mainLoop); err != nil {
		fmt.Println("ERROR!", err)
	}
}

func runPipeline(mainLoop *glib.MainLoop) error {
	if len(os.Args) != 2 {
		return errors.New("We expecte 1 argument which is the location of a flac audio file. You failed to provide.")
	}

	gst.Init(&os.Args)

	// Let GStreamer create a pipeline from the parsed launch syntax on the cli.
	pipeline, err := gst.NewPipelineFromString("filesrc location =" + os.Args[1] + " ! flacparse ! flacdec ! pipewiresink")
	if err != nil {
		return err
	}

	// Add a message handler to the pipeline bus, printing interesting information to the console.
	pipeline.GetPipelineBus().AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS: // When end-of-stream is received stop the main loop
			pipeline.BlockSetState(gst.StateNull)
			mainLoop.Quit()
		case gst.MessageError: // Error messages are always fatal
			err := msg.ParseError()
			fmt.Println("ERROR:", err.Error())
			if debug := err.DebugString(); debug != "" {
				fmt.Println("DEBUG:", debug)
			}
			mainLoop.Quit()
		default:
			tag := msg.ParseTags()
			playingNow := ""
			if tag != nil {
				myTrack, status := tag.GetString(gst.TagTitle)
				if status == true {
					playingNow = playingNow + "'" + myTrack + "'"
				}
				myArtist, status := tag.GetString(gst.TagArtist)
				if status == true {
					playingNow = playingNow + " by: " + myArtist
				}
			}
			fmt.Println(playingNow)
		}
		return true
	})

	// Start the pipeline
	pipeline.SetState(gst.StatePlaying)

	// Block on the main loop
	return mainLoop.RunError()
}

func main() {
	runLoop(func(loop *glib.MainLoop) error {
		return runPipeline(loop)
	})
}
