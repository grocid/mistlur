package main

import (
    "github.com/faiface/beep"
    "github.com/faiface/beep/mp3"
    id3 "github.com/mikkyang/id3-go"
    "github.com/murlokswarm/app"
    _ "github.com/murlokswarm/mac"
    "log"
    "os"
    "time"
)

var (
    win app.Windower
)

func main() {
    log.Println(os.Args)

    app.OnLaunch = func() {
        win = newMainWindow()
        win.Mount(&Player{})
    }

    app.OnReopen = func() {
        if win != nil {
            return
        }
        win = newMainWindow()
        win.Mount(&Player{})
    }

    go func() {
        Init()

        for _, file := range os.Args[1:] {
            // Read music file.
            f, err := os.Open(file)
            // Skip if error...
            if err != nil {
                continue
            }

            log.Println(file)

            // Decode the data.
            s, format, err := mp3.Decode(f)

            if err != nil {
                continue
            }

            // Make a channel to communicate when done.
            done = make(chan struct{})

            // Read tags.
            mp3File, err := id3.Open(file)
            defer mp3File.Close()
            tag.Artist = mp3File.Artist()
            tag.Title = mp3File.Title()

            // Start playing...
            InitPlayer(format.SampleRate, format.SampleRate.N(time.Second/10))
            Play(beep.Seq(s, beep.Callback(func() {
                close(done)
            })))

            // Wait for done signal.
            <-done

        }

        done <- struct{}{}
        tag.Artist = "Nothing playing"
        tag.Title = "Enjoy silence"
    }()

    app.Run()
}

func newMainWindow() app.Windower {
    return app.NewWindow(app.Window{
        Title:          "mistlur",
        TitlebarHidden: true,
        Width:          500,
        Height:         768,
        //Vibrancy:       app.VibeDark,
        BackgroundColor: "#ffffff",
        OnClose: func() bool {
            win = nil
            return true
        },
    })
}
