package main

import (
    "github.com/murlokswarm/app"
    _ "github.com/murlokswarm/mac"
    "log"
    "os"
    "runtime"
)

var (
    win app.Windower
)

func main() {
    log.Println(os.Args)
    runtime.GOMAXPROCS(1)
    if len(os.Args) < 2 {
        return
    }

    Init()
    InitPlaylist(os.Args[1:])

    go func() {
        StartPlaylist()
    }()

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
