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
- OAuth Token なしでも接続可能（匿名読み取り専用モード）

## Requirements

- [VOICEVOX](https://voicevox.hiroshiba.jp/) (起動しておく)
- PipeWire / PulseAudio

## Build

```bash
# Wails CLI が必要
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# 依存パッケージ (Arch/CachyOS)
sudo pacman -S webkit2gtk-4.1

# ビルド
wails build
./build/bin/twitch-tts
```

## Install

```bash
# ビルド + ~/.local/bin にインストール + .desktop ファイル作成
make install
```

## Usage

1. VOICEVOX を起動
2. アプリを起動
3. Channel 名を入力して Connect（OAuth Token はオプション）
4. OBS で「音声出力キャプチャ」→ `Twitch TTS` を選択

自分のスピーカー/ヘッドホンにも自動でループバックされるので、
読み上げ音声を聞きながら配信できる。

## OAuth Token（オプション）

Token がなくてもチャットの読み取りは可能（匿名接続）。
チャットへの書き込みが必要な場合のみ Token を設定する。

```bash
# Twitch CLI をインストール
yay -S twitch-cli

# Twitch Developer Console でアプリ登録後
twitch configure
twitch token -u -s 'chat:read'
```

## Config

設定は `~/.config/twitch-tts/config.toml` に自動保存される。
