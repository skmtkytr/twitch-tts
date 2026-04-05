package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/pkg/browser"
)

const (
	twitchClientID = "xkqtiiurh5r7nyldw8relggss9doyx"
	oauthRedirect  = "http://localhost:21821"
	oauthScopes    = "chat:read chat:edit"
)

// StartOAuthFlow opens the browser for Twitch login and returns the access token.
func StartOAuthFlow() (string, error) {
	tokenCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()

	// The redirect page: Twitch sends the token as a URL fragment (#access_token=...),
	// so we need JavaScript to extract it and POST it back to us.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>Twitch TTS - Login</title>
<style>
body { background: #1e1e2e; color: #cdd6f4; font-family: sans-serif;
       display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; }
.msg { text-align: center; font-size: 1.2em; }
</style></head><body>
<div class="msg" id="msg">認証中...</div>
<script>
const hash = window.location.hash.substring(1);
const params = new URLSearchParams(hash);
const token = params.get('access_token');
if (token) {
  fetch('/callback', { method: 'POST', body: token })
    .then(() => { document.getElementById('msg').textContent = '認証完了！このタブを閉じてください。'; })
    .catch(() => { document.getElementById('msg').textContent = 'エラーが発生しました。'; });
} else {
  document.getElementById('msg').textContent = '認証がキャンセルされました。';
}
</script></body></html>`)
	})

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		buf := make([]byte, 512)
		n, _ := r.Body.Read(buf)
		token := string(buf[:n])
		if token == "" {
			http.Error(w, "no token", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		tokenCh <- token
	})

	listener, err := net.Listen("tcp", "127.0.0.1:21821")
	if err != nil {
		return "", fmt.Errorf("failed to start OAuth listener: %w", err)
	}

	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	authURL := fmt.Sprintf(
		"https://id.twitch.tv/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=token&scope=%s",
		twitchClientID, oauthRedirect, oauthScopes,
	)
	log.Printf("Opening browser for Twitch login: %s", authURL)
	if err := browser.OpenURL(authURL); err != nil {
		server.Shutdown(context.Background())
		return "", fmt.Errorf("failed to open browser: %w", err)
	}

	select {
	case token := <-tokenCh:
		server.Shutdown(context.Background())
		return token, nil
	case err := <-errCh:
		server.Shutdown(context.Background())
		return "", err
	}
}
