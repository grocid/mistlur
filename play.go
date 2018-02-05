// Based on the internals of https://github.com/faiface/beep

package main

import (
    "github.com/faiface/beep"
    "github.com/hajimehoshi/oto"

    "github.com/pkg/errors"
    //"os"
    "sync"
)

var (
    mu       sync.Mutex
    mixer    beep.Mixer
    samples  [][2]float64
    csamples []complex128
    buf      []byte
    player   *oto.Player
    underrun func()
    done     chan struct{}
)

func InitPlayer(sampleRate beep.SampleRate, bufferSize int) error {
    mu.Lock()
    defer mu.Unlock()

    if player != nil {
        done <- struct{}{}
        player.Close()
    }

    mixer = beep.Mixer{}
    numBytes := bufferSize * 4
    samples = make([][2]float64, bufferSize)

    buf = make([]byte, numBytes)

    var err error
    player, err = oto.NewPlayer(int(sampleRate), 2, 2, numBytes)

    if err != nil {
        return errors.Wrap(err, "failed to initialize speaker")
    }

    if underrun != nil {
        player.SetUnderrunCallback(underrun)
    }

    go func() {
        for {
            select {
            default:
                update()
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

func update() {
    mu.Lock()

    mixer.Stream(samples)
    mu.Unlock()
    for i := range samples {
        for c := range samples[i] {
            val := samples[i][c]
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

    player.Write(buf)

}
