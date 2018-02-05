package main

import (
    "github.com/faiface/beep"
    "github.com/faiface/beep/mp3"
    //"github.com/faiface/beep/speaker"
    "github.com/hajimehoshi/oto"
    "github.com/ktye/fft"
    "github.com/murlokswarm/app"
    "github.com/pkg/errors"
    "math"
    "math/cmplx"
    "os"
    "sync"
    "time"
)

// Player is the component displaying Player.
type Player struct {
    Time int
    Bar  [5]float64
}

var (
    mu       sync.Mutex
    mixer    beep.Mixer
    samples  [][2]float64
    csamples []complex128
    buf      []byte
    player   *oto.Player
    underrun func()
    done     chan struct{}
    fftc     fft.FFT
)

func (p *Player) Init(sampleRate beep.SampleRate, bufferSize int) error {
    mu.Lock()
    defer mu.Unlock()

    if player != nil {
        done <- struct{}{}
        player.Close()
    }

    mixer = beep.Mixer{}

    numBytes := bufferSize * 4
    fftc, _ = fft.New(bufferSize)
    samples = make([][2]float64, bufferSize)
    csamples = make([]complex128, bufferSize)
    buf = make([]byte, numBytes)

    var err error
    player, err = oto.NewPlayer(int(sampleRate), 2, 2, numBytes)
    if err != nil {
        return errors.Wrap(err, "failed to initialize speaker")
    }

    if underrun != nil {
        player.SetUnderrunCallback(underrun)
    }

    done = make(chan struct{})

    go func() {
        for {
            select {
            default:
                p.update()
            case <-done:
                return
            }
        }
    }()

    return nil
}

func UnderrunCallback(f func()) {
    mu.Lock()
    underrun = f
    if player != nil {
        player.SetUnderrunCallback(underrun)
    }
    mu.Unlock()
}

func Lock() {
    mu.Lock()
}

func Unlock() {
    mu.Unlock()
}

func Play(s ...beep.Streamer) {
    mu.Lock()
    mixer.Play(s...)
    mu.Unlock()
}

func (p *Player) update() {
    mu.Lock()
    mixer.Stream(samples)
    mu.Unlock()

    for i := range samples {
        for c := range samples[i] {
            val := samples[i][c]
            csamples[i] = complex((samples[i][0] + samples[i][1]), 0)
            if val < -1 {
                val = -1
            }
            if val > +1 {
                val = +1
            }
            valInt16 := int16(val * (1<<15 - 1))
            low := byte(valInt16)
            high := byte(valInt16 >> 8)
            buf[i*4+c*2+0] = low
            buf[i*4+c*2+1] = high
        }
    }

    fftc.Transform(csamples)
    for j := 0; j < 5; j++ {
        for i := 0; i < len(csamples)/2/5; i++ {
            p.Bar[j] = 50 * (math.Log(1 + cmplx.Abs(csamples[i+j])))
        }
    }

    player.Write(buf)
}

var ch chan bool

func (p *Player) OnMount() {
    f, _ := os.Open("qotsa.mp3")
    s, format, _ := mp3.Decode(f)
    p.Init(format.SampleRate, 4096)
    Play(beep.Seq(s, beep.Callback(func() {
        //close(done)
    })))

    go func() {
        c := time.Tick(50 * time.Millisecond)
        for _ = range c {
            app.Render(p)
        }
    }()
}

func (p *Player) OnDismount() {
    ch <- true
}

func (p *Player) Render() string {
    return `
<div class="center">
    <div class="graph">
            <div style="height: 80px; background-color: rgba(0,0,0,0)" class="bar"/>
                {{ range $key, $data := .Bar }}
                    <div style="height: {{$data}}px;" class="bar"/>
                {{ end }}
            <div style="height: 80px; background-color: rgba(0,0,0,0)" class="bar"/>
    </div>  
</div>
<h1>Queens of the Stoneage</h1>
<h2>The Way You Used to Do</h2>

`
}

// /!\ Register the component. Required to use the component into a context.
func init() {
    app.RegisterComponent(&Player{})
}
