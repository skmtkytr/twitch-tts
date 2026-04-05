package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx        context.Context
	voicevox   *VoicevoxClient
	audio      *AudioRouter
	twitchCli  *twitch.Client
	ttsOn      bool
	speakerID  int
	readName   bool
	nameSuffix string
	mu         sync.Mutex
	ttsCh      chan ttsRequest
	stopCh     chan struct{}
}

type ttsRequest struct {
	Text      string
	SpeakerID int
}

type ChatMessage struct {
	Author  string `json:"author"`
	Content string `json:"content"`
}

func NewApp() *App {
	return &App{
		voicevox:   NewVoicevoxClient(""),
		audio:      NewAudioRouter(),
		ttsOn:      true,
		speakerID:  1,
		readName:   true,
		nameSuffix: "さん",
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.audio.Setup(); err != nil {
		log.Printf("audio setup warning: %v", err)
	}
}

func (a *App) shutdown(ctx context.Context) {
	a.Disconnect()
	a.audio.Teardown()
}

func (a *App) GetSpeakers() ([]Speaker, error) {
	return a.voicevox.GetSpeakers()
}

func (a *App) SetSpeaker(id int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.speakerID = id
}

func (a *App) SetTTSEnabled(on bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ttsOn = on
}

func (a *App) SetReadName(on bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.readName = on
}

func (a *App) SetNameSuffix(suffix string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.nameSuffix = suffix
}

func (a *App) Connect(channel, token string) error {
	channel = strings.TrimSpace(channel)
	token = strings.TrimSpace(token)
	if channel == "" {
		return fmt.Errorf("channel is required")
	}

	// Clean up previous connection
	a.Disconnect()

	a.ttsCh = make(chan ttsRequest, 64)
	a.stopCh = make(chan struct{})

	// Start TTS worker
	go a.ttsWorker()

	var client *twitch.Client
	if token == "" {
		// Anonymous read-only connection (no OAuth token needed)
		client = twitch.NewAnonymousClient()
	} else {
		if !strings.HasPrefix(token, "oauth:") {
			token = "oauth:" + token
		}
		client = twitch.NewClient(channel, token)
	}
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		msg := ChatMessage{
			Author:  message.User.DisplayName,
			Content: message.Message,
		}
		runtime.EventsEmit(a.ctx, "chat-message", msg)

		a.mu.Lock()
		ttsOn := a.ttsOn
		speakerID := a.speakerID
		readName := a.readName
		nameSuffix := a.nameSuffix
		a.mu.Unlock()

		if ttsOn {
			text := msg.Content
			if readName {
				text = msg.Author + nameSuffix + "。" + msg.Content
			}
			select {
			case a.ttsCh <- ttsRequest{
				Text:      text,
				SpeakerID: speakerID,
			}:
			default:
				// queue full, skip
			}
		}
	})

	client.OnConnect(func() {
		runtime.EventsEmit(a.ctx, "connected", nil)
	})

	client.Join(channel)

	a.mu.Lock()
	a.twitchCli = client
	a.mu.Unlock()

	go func() {
		if err := client.Connect(); err != nil {
			log.Printf("twitch connect error: %v", err)
			runtime.EventsEmit(a.ctx, "disconnected", err.Error())
		}
	}()

	return nil
}

func (a *App) Disconnect() {
	a.mu.Lock()
	client := a.twitchCli
	a.twitchCli = nil
	stopCh := a.stopCh
	a.mu.Unlock()

	if stopCh != nil {
		close(stopCh)
	}
	if client != nil {
		client.Disconnect()
	}
	runtime.EventsEmit(a.ctx, "disconnected", nil)
}

func (a *App) ttsWorker() {
	for {
		select {
		case <-a.stopCh:
			return
		case req := <-a.ttsCh:
			wav, err := a.voicevox.Synthesize(req.Text, req.SpeakerID)
			if err != nil {
				log.Printf("synthesis error: %v", err)
				continue
			}
			if err := PlayWav(wav, a.audio.SinkName()); err != nil {
				log.Printf("playback error: %v", err)
			}
		}
	}
}
