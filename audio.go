package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

const (
	sinkName        = "twitch_tts"
	sinkDescription = "Twitch TTS"
)

type AudioRouter struct {
	sinkModuleID     string
	loopbackModuleID string
}

func NewAudioRouter() *AudioRouter {
	return &AudioRouter{}
}

// Setup creates a virtual sink for TTS output and a loopback
// so the user can also hear the TTS on their default output.
// OBS captures "Twitch TTS" via Audio Output Capture.
func (ar *AudioRouter) Setup() error {
	// Check if sink already exists
	out, err := exec.Command("pactl", "list", "short", "sinks").Output()
	if err != nil {
		return fmt.Errorf("pactl list sinks failed: %w", err)
	}
	if strings.Contains(string(out), sinkName) {
		log.Println("audio: virtual sink already exists")
		return nil
	}

	// Create null sink (virtual output device)
	out, err = exec.Command("pactl", "load-module", "module-null-sink",
		fmt.Sprintf("sink_name=%s", sinkName),
		fmt.Sprintf("sink_properties=device.description=\"%s\"", sinkDescription),
	).Output()
	if err != nil {
		return fmt.Errorf("failed to create virtual sink: %w", err)
	}
	ar.sinkModuleID = strings.TrimSpace(string(out))
	log.Printf("audio: created virtual sink (module %s)", ar.sinkModuleID)

	// Create loopback: copy TTS sink → default output so user can hear it too
	out, err = exec.Command("pactl", "load-module", "module-loopback",
		fmt.Sprintf("source=%s.monitor", sinkName),
		"latency_msec=50",
	).Output()
	if err != nil {
		log.Printf("audio: loopback creation failed (non-fatal): %v", err)
	} else {
		ar.loopbackModuleID = strings.TrimSpace(string(out))
		log.Printf("audio: created loopback to default output (module %s)", ar.loopbackModuleID)
	}

	return nil
}

// Teardown removes the virtual sink and loopback.
func (ar *AudioRouter) Teardown() {
	if ar.loopbackModuleID != "" {
		if err := exec.Command("pactl", "unload-module", ar.loopbackModuleID).Run(); err != nil {
			log.Printf("audio: failed to remove loopback: %v", err)
		}
		ar.loopbackModuleID = ""
	}
	if ar.sinkModuleID != "" {
		if err := exec.Command("pactl", "unload-module", ar.sinkModuleID).Run(); err != nil {
			log.Printf("audio: failed to remove virtual sink: %v", err)
		}
		ar.sinkModuleID = ""
	}
	log.Println("audio: cleaned up")
}

// SinkName returns the PipeWire sink name for pw-play --target.
func (ar *AudioRouter) SinkName() string {
	return sinkName
}
