package main

import (
    "github.com/faiface/beep"
    "github.com/faiface/beep/mp3"
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
        for _, file := range os.Args[1:] {

            f, err := os.Open(file)

            if err != nil {
                continue
            }

            s, format, err := mp3.Decode(f)

            if err != nil {
                continue
            }

            Init(format.SampleRate, format.SampleRate.N(time.Second/10))
            Play(beep.Seq(s, beep.Callback(func() {
                close(done)
            })))
        }
    }()

    app.Run()
}

func newMainWindow() app.Windower {
    return app.NewWindow(app.Window{
        Title:          "mistlur",
        TitlebarHidden: true,
        Width:          500,
        Height:         768,
        Vibrancy:       app.VibeDark,
        OnClose: func() bool {
            win = nil
            return true
        },
    })
}
