package player

import (
	"fmt"

	"github.com/go-gst/go-glib/glib"
	"github.com/go-gst/go-gst/gst"
)

type WhatIsPlaying struct {
	Playing  bool
	MainLoop *glib.MainLoop
	Playbin  *gst.Element
	Err      error
}

func playbin(mainLoop *glib.MainLoop, uri string) WhatIsPlaying {
	gst.Init(nil)

	// Create a new playbin and set the URI on it
	playbin, err := gst.NewElement("playbin")
	if err != nil {
		return WhatIsPlaying{Playing: false, MainLoop: nil, Playbin: nil, Err: err}
	}
	playbin.Set("uri", uri)
	bus := playbin.GetBus()

	playbin.SetState(gst.StatePlaying)

	// artist := ""
	// album := ""
	// title := ""
	// fmt.Println("Tags:")
	bus.AddWatch(func(msg *gst.Message) bool {
		switch msg.Type() {
		case gst.MessageEOS:
			mainLoop.Quit()
			return false
		case gst.MessageError:
			err := msg.ParseError()
			fmt.Println("ERROR:", err.Error())
			if debug := err.DebugString(); debug != "" {
				fmt.Println("DEBUG")
				fmt.Println(debug)
			}
			mainLoop.Quit()
			return false
		// Watch state change events
		case gst.MessageStateChanged:
			if _, newState := msg.ParseStateChanged(); newState == gst.StatePlaying {
				bin := gst.ToGstBin(playbin)
				// Generate a dot graph of the pipeline to GST_DEBUG_DUMP_DOT_DIR if defined
				bin.DebugBinToDotFile(gst.DebugGraphShowAll, "PLAYING")
			}

			// Tag messages contain changes to tags on the stream. This can include metadata about
			// the stream such as codecs, artists, albums, etc.
			// case gst.MessageTag:
			// 	tags := msg.ParseTags()
			// 	if artist_tag, ok := tags.GetString(gst.TagArtist); ok {
			// 		if artist_tag != artist {
			// 			fmt.Println("  Artist:", artist_tag)
			// 		}
			// 		artist = artist_tag
			// 	}
			// 	if album_tag, ok := tags.GetString(gst.TagAlbum); ok {
			// 		if album_tag != album {
			// 			fmt.Println("  album:", album_tag)
			// 		}
			// 		album = album_tag
			// 	}
			// 	if title_tag, ok := tags.GetString(gst.TagTitle); ok {
			// 		if title_tag != title {
			// 			fmt.Println("  title:", title_tag)
			// 		}
			// 		title = title_tag
			// 	}
		}
		return true
	})
	err = mainLoop.RunError()
	return WhatIsPlaying{Playing: true, MainLoop: mainLoop, Playbin: playbin, Err: err}
}

func Play(uri string) WhatIsPlaying {
	mainLoop := glib.NewMainLoop(glib.MainContextDefault(), false)
	return playbin(mainLoop, uri)
}
