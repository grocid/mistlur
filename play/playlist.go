package play

import (
    "github.com/faiface/beep"
    "github.com/faiface/beep/mp3"
    id3 "github.com/mikkyang/id3-go"
    "log"
    "os"
    "time"
)

type Playlist struct{}

type Tag struct {
    Artist string
    Title  string
}

var (
    currentlyPlaying int
    filesToBePlayed  []string
    continuePlayList chan struct{}
    tag              Tag
)

func (p *Playlist) Init(files []string) {
    currentlyPlaying = 0
    filesToBePlayed = files
}

func New() Playlist {
    return Playlist{}
}

func (p *Playlist) Start() {
    go func() {
        for {
            // Set current file.
            file := filesToBePlayed[currentlyPlaying]
            // Read tags
            setTags(file)
            play(file)

            currentlyPlaying = currentlyPlaying + 1

            log.Println(currentlyPlaying)

            if currentlyPlaying >= len(filesToBePlayed) {
                Stop()
                mu.Lock()
                isPlaying = false
                currentlyPlaying = len(filesToBePlayed) - 1

                continuePlayList = make(chan struct{})
                <-continuePlayList
            }
        }
    }()
}

func (p *Playlist) Done() {
    done <- struct{}{}
}

func (p *Playlist) TogglePause() {
    togglePause()
}

func (p *Playlist) Back() {
    currentlyPlaying = currentlyPlaying - 2
    if currentlyPlaying < -1 {
        currentlyPlaying = -1
    }
    p.Done()
}

func (p *Playlist) Next() {
    if currentlyPlaying >= len(filesToBePlayed)-2 {
        currentlyPlaying = len(filesToBePlayed) - 2
    }
    p.Done()
}

func (p *Playlist) IsPlaying() bool {
    return IsPlaying()
}

func (p *Playlist) GetSamples() *[][2]float64 {
    return &samples
}

func (p *Playlist) GetTags() Tag {
    return tag
}

func play(file string) {
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
    isPlaying = false
}

func setTags(file string) {
    // Read tags.
    mp3File, _ := id3.Open(file)
    defer mp3File.Close()

    tag = Tag{
        Artist: mp3File.Artist(),
        Title:  mp3File.Title(),
    }
}
