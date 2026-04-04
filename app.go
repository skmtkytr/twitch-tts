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
	ctx       context.Context
	voicevox  *VoicevoxClient
	twitchCli *twitch.Client
	ttsOn     bool
	speakerID int
	mu        sync.Mutex
	ttsCh     chan ttsRequest
	stopCh    chan struct{}
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
		voicevox:  NewVoicevoxClient(""),
		ttsOn:     true,
		speakerID: 1,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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

func (a *App) Connect(channel, token string) error {
	channel = strings.TrimSpace(channel)
	token = strings.TrimSpace(token)
	if channel == "" || token == "" {
		return fmt.Errorf("channel and token are required")
	}
	if !strings.HasPrefix(token, "oauth:") {
		token = "oauth:" + token
	}

	// Clean up previous connection
	a.Disconnect()

	a.ttsCh = make(chan ttsRequest, 64)
	a.stopCh = make(chan struct{})

	// Start TTS worker
	go a.ttsWorker()

	client := twitch.NewClient(channel, token)
	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		msg := ChatMessage{
			Author:  message.User.DisplayName,
			Content: message.Message,
		}
		runtime.EventsEmit(a.ctx, "chat-message", msg)

		a.mu.Lock()
		ttsOn := a.ttsOn
		speakerID := a.speakerID
		a.mu.Unlock()

		if ttsOn {
			select {
			case a.ttsCh <- ttsRequest{
				Text:      msg.Author + "。" + msg.Content,
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
			if err := PlayWav(wav); err != nil {
				log.Printf("playback error: %v", err)
			}
		}
	}
}
