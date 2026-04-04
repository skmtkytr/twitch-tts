# Twitch TTS

Twitch チャットを VOICEVOX で読み上げるデスクトップアプリ。
Wails (Go + Svelte) で構築。OBS 配信にそのまま音声を乗せられる。

## Features

- Twitch チャットのリアルタイム表示と読み上げ
- VOICEVOX のキャラクター/スタイル選択
- 発言者名の読み上げ ON/OFF、敬称の設定
- TTS ON/OFF トグル
- OBS 用の仮想オーディオシンク自動作成 + ループバック
- 設定の自動保存・復元

## Requirements

- [VOICEVOX](https://voicevox.hiroshiba.jp/) (起動しておく)
- PipeWire / PulseAudio
- Twitch OAuth Token ([Twitch CLI](https://dev.twitch.tv/docs/cli/) で取得)

## Build

```bash
# Wails CLI が必要
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 依存パッケージ (Arch/CachyOS)
sudo pacman -S webkit2gtk

# ビルド
wails build
./build/bin/twitch-tts
```

## Usage

1. VOICEVOX を起動
2. アプリを起動
3. Channel 名と OAuth Token を入力して Connect
4. OBS で「音声出力キャプチャ」→ `Twitch TTS` を選択

自分のスピーカー/ヘッドホンにも自動でループバックされるので、
読み上げ音声を聞きながら配信できる。

## OAuth Token の取得

```bash
# Twitch CLI をインストール
yay -S twitch-cli

# Twitch Developer Console でアプリ登録後
twitch configure
twitch token -u -s 'chat:read'
```

## Config

設定は `~/.config/twitch-tts/config.json` に保存される。
