package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
)

type VoicevoxClient struct {
	BaseURL string
}

type Speaker struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type rawSpeaker struct {
	Name   string     `json:"name"`
	Styles []rawStyle `json:"styles"`
}

type rawStyle struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

func NewVoicevoxClient(baseURL string) *VoicevoxClient {
	if baseURL == "" {
		baseURL = "http://localhost:50021"
	}
	return &VoicevoxClient{BaseURL: baseURL}
}

func (v *VoicevoxClient) GetSpeakers() ([]Speaker, error) {
	resp, err := http.Get(v.BaseURL + "/speakers")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to VOICEVOX: %w", err)
	}
	defer resp.Body.Close()

	var raw []rawSpeaker
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	var speakers []Speaker
	for _, s := range raw {
		for _, st := range s.Styles {
			speakers = append(speakers, Speaker{
				Name: fmt.Sprintf("%s (%s)", s.Name, st.Name),
				ID:   st.ID,
			})
		}
	}
	return speakers, nil
}

func (v *VoicevoxClient) Synthesize(text string, speakerID int) ([]byte, error) {
	// Create audio query
	queryURL := fmt.Sprintf("%s/audio_query?text=%s&speaker=%d", v.BaseURL, url.QueryEscape(text), speakerID)
	queryResp, err := http.Post(queryURL, "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("audio_query failed: %w", err)
	}
	defer queryResp.Body.Close()

	if queryResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(queryResp.Body)
		return nil, fmt.Errorf("audio_query returned %d: %s", queryResp.StatusCode, body)
	}

	queryBody, err := io.ReadAll(queryResp.Body)
	if err != nil {
		return nil, err
	}

	// Synthesize
	synthURL := fmt.Sprintf("%s/synthesis?speaker=%d", v.BaseURL, speakerID)
	synthResp, err := http.Post(synthURL, "application/json", bytes.NewReader(queryBody))
	if err != nil {
		return nil, fmt.Errorf("synthesis failed: %w", err)
	}
	defer synthResp.Body.Close()

	if synthResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(synthResp.Body)
		return nil, fmt.Errorf("synthesis returned %d: %s", synthResp.StatusCode, body)
	}

	return io.ReadAll(synthResp.Body)
}

func PlayWav(wavData []byte) error {
	f, err := os.CreateTemp("", "twitch-tts-*.wav")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	if _, err := f.Write(wavData); err != nil {
		f.Close()
		return err
	}
	f.Close()

	cmd := exec.Command("pw-play", f.Name())
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
