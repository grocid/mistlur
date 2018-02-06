package main

import (
    "github.com/faiface/beep"
    "github.com/faiface/beep/mp3"
    id3 "github.com/mikkyang/id3-go"
    "log"
    "os"
    "time"
)

type Tag struct {
    Artist string
    Title  string
}

var (
    currentlyPlaying int
    filesToBePlayed  []string
    continuePlayList chan struct{}
)

func InitPlaylist(files []string) {
    currentlyPlaying = 0
    filesToBePlayed = files
}

func StartPlaylist() {
    for {

        file := filesToBePlayed[currentlyPlaying]

        GetTags(file)
        PlayIt(file)

        currentlyPlaying = currentlyPlaying + 1

        log.Println(currentlyPlaying)

        if currentlyPlaying >= len(filesToBePlayed) {
            Stop()
            mu.Lock()
            currentlyPlaying = len(filesToBePlayed) - 1

            continuePlayList = make(chan struct{})
            <-continuePlayList
        }
    }
}

func DecrementPosition() {
    currentlyPlaying = currentlyPlaying - 2
    if currentlyPlaying < -1 {
        currentlyPlaying = -1
    }
}

func PlayIt(file string) {
    // Read music file.
    f, err := os.Open(file)
    // Skip if error...
    if err != nil {
        return
    }

    // Decode the data.
    s, format, err := mp3.Decode(f)

    if err != nil {
        return
    }

    // Make a channel to communicate when done.
    done = make(chan struct{})

    // Start playing...
    InitPlayer(
        format.SampleRate,
        format.SampleRate.N(time.Second/10),
    )
    Play(beep.Seq(s, beep.Callback(
        func() {
            close(done)
        })))

    // Wait for done signal, so that the player
    // has finished crunching the file.
    <-done
}

func GetTags(file string) {
    // Read tags.
    mp3File, _ := id3.Open(file)
    defer mp3File.Close()

    // Set artist and title.
    tag.Artist = mp3File.Artist()
    tag.Title = mp3File.Title()
}
