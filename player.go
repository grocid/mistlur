package main

import (
    "github.com/ktye/fft"
    "github.com/murlokswarm/app"
    "math"
    "math/cmplx"
    "time"
    "mistlur/cl"
    "log"
)

// Player is the component displaying Player.
type Player struct {
    Bar     [10]float64
    PlayBtn string
}

type Tag struct {
    Artist string
    Title  string
}

const (
    // Let us make it less computationally invasive
    FFTSamples = 4096
    // This is fast enough for the eye, no? Maybe a little choppy
    // but that is a trade-off.
    RefreshEveryMillisec = 40
)

var (
    tag Tag
    // Signal to control UI computations.
    guidone chan struct{}
    play    bool
    fftc    fft.FFT
)

func Init() {
    fftc, _ = fft.New(FFTSamples)
    csamples = make([]complex128, FFTSamples)
}

func (p *Player) OnMount() {

platforms, err := cl.GetPlatforms()
    if err != nil {
        log.Fatalf("Failed to get platforms: %+v", err)
    }
    for i, p := range platforms {
        log.Println("Platform %d:", i)
        log.Println("  Name: %s", p.Name())
        log.Println("  Vendor: %s", p.Vendor())
        log.Println("  Profile: %s", p.Profile())
        log.Println("  Version: %s", p.Version())
        log.Println("  Extensions: %s", p.Extensions())
    }


    play = true
    p.PlayBtn = "pause"
    // Make a channel to control UI.
    guidone = make(chan struct{})

    go func() {
        c := time.Tick(RefreshEveryMillisec * time.Millisecond)
        for _ = range c {
             // Convert channel slice to complex128 (mono).
            for i := 0; i < FFTSamples; i++ {
                csamples[i] = complex((samples[i][0] + samples[i][1]), 0)
            }
            // An FFT walks into...
            fftc.Transform(csamples)
            // ...a bar...
            for j := 0; j < len(p.Bar); j++ {
                // Consider only half of the frequencies.
                for i := 0; i < len(csamples)/2/len(p.Bar); i++ {
                    p.Bar[j] = 20 * (math.Log(1 + cmplx.Abs(csamples[i+j])))
                }
            }
            // ...and the whole scene unfolds with tedious inevitability.
            // #complexjoke
        }
    }()

    go func() {
        c := time.Tick(RefreshEveryMillisec * time.Millisecond)
        for _ = range c {
            select {
            default:
                // Render pl0x.
                app.Render(p)
            case <-guidone:
                return
            }
            
        }
    }()
}

func (p *Player) ClearBars() {
    for i := 0; i < len(p.Bar); i++ {
        p.Bar[i] = 0
    }
}

func (p *Player) NextBtn() {
    // Simply tell the player that it is done with the current song...
    done <- struct{}{}
}

func (p *Player) TogglePlayBtn() {
    play = !play

    // Tell UI to toggle the button.
    if play {
        p.PlayBtn = "pause"
    } else {
        p.PlayBtn = "play"
    }
}

func (p *Player) OnDismount() {
    // Tell UI it is done here.
    guidone <- struct{}{}
}

func (p *Player) Render() string {
    // UI component
    return `
<div class="center">
    <div class="graph">
            <div style="height: 100px; background-color: rgba(0,0,0,0)" class="bar"/>
                {{ range $key, $data := .Bar }}
                    <div style="height: {{$data}}px;" class="bar"/>
                {{ end }}
            <div style="height: 100px; background-color: rgba(0,0,0,0)" class="bar"/>
    </div>
</div>
<h1>` + tag.Artist + `</h1>
<h2>` + tag.Title + `</h2>
<div>
    <button class="button back" onclick="OK"/>
    <button class="button {{.PlayBtn}}" onclick="TogglePlayBtn"/>   
    <button class="button next" onclick="NextBtn"/>                
</div>
`
}

// /!\ Register the component. Required to use the component into a context.
func init() {
    app.RegisterComponent(&Player{})
}
