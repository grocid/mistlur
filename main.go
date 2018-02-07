package main

import (
    "github.com/murlokswarm/app"
    _ "github.com/murlokswarm/mac"
    "log"
    "mistlur/play"
    "os"
    "runtime"
)

var (
    win      app.Windower
    playlist play.Playlist
)

func main() {
    log.Println(os.Args)
    runtime.GOMAXPROCS(1)
    if len(os.Args) < 2 {
        return
    }

    // Initialize FFT.
    Init()

    playlist = play.New()
    playlist.Init(os.Args[1:])

    go func() {
        playlist.Start()
    }()

    app.OnLaunch = func() {
        appMenu := &MenuBar{}

        if menuBar, ok := app.MenuBar(); ok {
            menuBar.Mount(appMenu)
        }

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

    app.Run()
}

func newMainWindow() app.Windower {
    return app.NewWindow(app.Window{
        Title:          "mistlur",
        TitlebarHidden: true,
        Width:          400,
        Height:         70,
        Vibrancy:       app.VibeDark,
        CloseHidden:    true,
        MinimizeHidden: true,
        FixedSize:      true,
        OnClose: func() bool {
            win = nil
            return true
        },
    })
}
