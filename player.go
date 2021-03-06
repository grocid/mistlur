package main

import (
    "github.com/ktye/fft"
    "github.com/murlokswarm/app"
    //"log"
    "math"
    "math/cmplx"
    "mistlur/play"
    "time"
)

// Player is the component disPlaying Player.
type Player struct {
    Bar     [10]float64
    PlayBtn string
    Tag     play.Tag
}

const (
    // Let us make it less computationally invasive
    FFTSamples = 1024
    // This is fast enough for the eye, no? Maybe a little choppy
    // but that is a trade-off.
    RefreshEveryMillisec = 10

    BtnPlay  = "play"
    BtnPause = "pause"
)

var (
    guiIsDone chan struct{}
    fftc      fft.FFT
    csamples  []complex128
)

func Init() {
    fftc, _ = fft.New(FFTSamples)
    csamples = make([]complex128, FFTSamples)
}

func (p *Player) OnMount() {
    p.PlayBtn = BtnPause

    // Make a channel to control UI.
    guiIsDone = make(chan struct{})

    // Rendering loop.
    go func() {
        c := time.Tick(RefreshEveryMillisec * time.Millisecond)
        for _ = range c {
            select {
            default:
                // Render pl0x.
                app.Render(p)
            case <-guiIsDone:
                return
            }
        }
    }()

    // FFT loop.
    go func() {
        c := time.Tick(RefreshEveryMillisec * time.Millisecond)
        for _ = range c {
            select {
            default:
                if !playlist.IsPlaying() {
                    p.ClearBars()
                    continue
                }

                s := playlist.GetSamples()
                samples := *s
                // Convert channel slices to complex128 (mono).
                for i := 0; i < FFTSamples; i++ {
                    csamples[i] = complex((samples[i][0] + samples[i][1]), 0)
                }
                // An FFT walks into...
                fftc.Transform(csamples)
                // ...a bar...
                for j := 0; j < len(p.Bar); j++ {
                    // Consider only half of the frequencies.
                    for i := 0; i < FFTSamples/len(p.Bar)/2; i++ {
                        p.Bar[j] = 20 * (math.Log(1 + cmplx.Abs(csamples[i+j])))
                    }
                }
                // ...and the whole scene unfolds with tedious inevitability.
                // #complexjoke
            case <-guiIsDone:
                return
            }
        }
    }()
    //done <- struct{}{}
}

func (p *Player) BackBtn() {
    playlist.Back()
}

func (p *Player) NextBtn() {
    playlist.Next()
}

func (p *Player) TogglePlayBtn() {
    playlist.TogglePause()
}

func (p *Player) ClearBars() {
    for j := 0; j < len(p.Bar); j++ {
        p.Bar[j] = 0
    }
}

func (p *Player) OnDismount() {
    // Tell UI it is done here.
    guiIsDone <- struct{}{}
    playlist.Done()
}

func (p *Player) Render() string {

    // Tell UI to toggle the button.
    if playlist.IsPlaying() {
        p.PlayBtn = BtnPause
    } else {
        p.PlayBtn = BtnPlay
    }

    p.Tag = playlist.GetTags()

    // UI component
    return `
<div class="center">
    <div class="graph">
            <div style="height: 120px; background-color: rgba(0,0,0,0)" class="bar"/>
                {{ range $key, $data := .Bar }}
                    <div style="height: {{$data}}px;" class="bar"/>
                {{ end }}
            <div style="height: 120px; background-color: rgba(0,0,0,0)" class="bar"/>
    </div>
</div>
<h1>{{ .Tag.Artist }} </h1>
<h2>{{ .Tag.Title }} </h2>
<div>
    <button class="button back" onclick="BackBtn"/>
    <button class="button {{.PlayBtn}}" onclick="TogglePlayBtn"/>   
    <button class="button next" onclick="NextBtn"/>                
</div>
`
}

// /!\ Register the component. Required to use the component into a context.
func init() {
    app.RegisterComponent(&Player{})
}
