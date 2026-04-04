from __future__ import annotations

import asyncio
import threading
from functools import partial

from PyQt6.QtCore import Qt, pyqtSlot
from PyQt6.QtWidgets import (
    QApplication,
    QCheckBox,
    QComboBox,
    QHBoxLayout,
    QLabel,
    QLineEdit,
    QMainWindow,
    QPushButton,
    QTextEdit,
    QVBoxLayout,
    QWidget,
)

from twitch_tts.twitch_client import ChatMessage, TwitchWorker
from twitch_tts.voicevox import Speaker, VoicevoxClient


class MainWindow(QMainWindow):
    def __init__(self) -> None:
        super().__init__()
        self.setWindowTitle("Twitch TTS")
        self.setMinimumSize(500, 400)

        self._voicevox = VoicevoxClient()
        self._speakers: list[Speaker] = []
        self._tts_enabled = True
        self._twitch = TwitchWorker()
        self._twitch.message_received.connect(self._on_message)
        self._twitch.connected.connect(self._on_connected)
        self._twitch.disconnected.connect(self._on_disconnected)
        self._twitch.error.connect(self._on_error)

        self._tts_queue: asyncio.Queue[tuple[str, int]] = asyncio.Queue()
        self._tts_thread: threading.Thread | None = None

        self._build_ui()
        self._load_speakers()

    def _build_ui(self) -> None:
        central = QWidget()
        self.setCentralWidget(central)
        layout = QVBoxLayout(central)

        # -- Connection row --
        conn_layout = QHBoxLayout()
        conn_layout.addWidget(QLabel("Channel:"))
        self._channel_input = QLineEdit()
        self._channel_input.setPlaceholderText("channel name")
        conn_layout.addWidget(self._channel_input)

        conn_layout.addWidget(QLabel("Token:"))
        self._token_input = QLineEdit()
        self._token_input.setPlaceholderText("oauth:xxxxx")
        self._token_input.setEchoMode(QLineEdit.EchoMode.Password)
        conn_layout.addWidget(self._token_input)

        self._connect_btn = QPushButton("Connect")
        self._connect_btn.clicked.connect(self._toggle_connection)
        conn_layout.addWidget(self._connect_btn)
        layout.addLayout(conn_layout)

        # -- Speaker row --
        speaker_layout = QHBoxLayout()
        speaker_layout.addWidget(QLabel("Speaker:"))
        self._speaker_combo = QComboBox()
        speaker_layout.addWidget(self._speaker_combo, 1)

        self._tts_check = QCheckBox("TTS ON")
        self._tts_check.setChecked(True)
        self._tts_check.toggled.connect(self._on_tts_toggled)
        speaker_layout.addWidget(self._tts_check)
        layout.addLayout(speaker_layout)

        # -- Chat log --
        self._chat_log = QTextEdit()
        self._chat_log.setReadOnly(True)
        layout.addWidget(self._chat_log)

        # -- Status bar --
        self._status = self.statusBar()
        self._status.showMessage("Disconnected")

    def _load_speakers(self) -> None:
        def _fetch() -> None:
            loop = asyncio.new_event_loop()
            try:
                speakers = loop.run_until_complete(self._voicevox.speakers())
                self._speakers = speakers
                # Update combo box from main thread
                self._speaker_combo.clear()
                for s in speakers:
                    self._speaker_combo.addItem(s.name, s.id)
                self._status.showMessage("VOICEVOX connected / Twitch disconnected")
            except Exception as e:
                self._status.showMessage(f"VOICEVOX error: {e}")
            finally:
                loop.close()

        threading.Thread(target=_fetch, daemon=True).start()

    def _toggle_connection(self) -> None:
        if self._connect_btn.text() == "Connect":
            channel = self._channel_input.text().strip()
            token = self._token_input.text().strip()
            if not channel or not token:
                self._status.showMessage("Channel and Token are required")
                return
            self._connect_btn.setEnabled(False)
            self._status.showMessage("Connecting...")
            self._start_tts_worker()
            self._twitch.start(token, channel)
        else:
            self._twitch.stop()
            self._stop_tts_worker()

    @pyqtSlot()
    def _on_connected(self) -> None:
        self._connect_btn.setText("Disconnect")
        self._connect_btn.setEnabled(True)
        self._channel_input.setEnabled(False)
        self._token_input.setEnabled(False)
        self._status.showMessage("Connected")

    @pyqtSlot()
    def _on_disconnected(self) -> None:
        self._connect_btn.setText("Connect")
        self._connect_btn.setEnabled(True)
        self._channel_input.setEnabled(True)
        self._token_input.setEnabled(True)
        self._status.showMessage("Disconnected")

    @pyqtSlot(str)
    def _on_error(self, msg: str) -> None:
        self._status.showMessage(f"Error: {msg}")

    @pyqtSlot(ChatMessage)
    def _on_message(self, msg: ChatMessage) -> None:
        self._chat_log.append(f"<b>{msg.author}</b>: {msg.content}")

        if self._tts_enabled:
            speaker_id = self._speaker_combo.currentData()
            if speaker_id is not None:
                text = f"{msg.author}。{msg.content}"
                self._tts_queue.put_nowait((text, speaker_id))

    @pyqtSlot(bool)
    def _on_tts_toggled(self, checked: bool) -> None:
        self._tts_enabled = checked

    # -- TTS worker (sequential playback) --

    def _start_tts_worker(self) -> None:
        self._tts_queue = asyncio.Queue()
        self._tts_running = True
        self._tts_thread = threading.Thread(target=self._tts_loop, daemon=True)
        self._tts_thread.start()

    def _stop_tts_worker(self) -> None:
        self._tts_running = False
        # Unblock the queue
        self._tts_queue.put_nowait(None)  # type: ignore[arg-type]
        if self._tts_thread:
            self._tts_thread.join(timeout=5)
            self._tts_thread = None

    def _tts_loop(self) -> None:
        loop = asyncio.new_event_loop()
        asyncio.set_event_loop(loop)
        try:
            loop.run_until_complete(self._tts_consume())
        finally:
            loop.close()

    async def _tts_consume(self) -> None:
        while self._tts_running:
            item = await self._tts_queue.get()
            if item is None:
                break
            text, speaker_id = item
            try:
                wav = await self._voicevox.synthesize(text, speaker_id)
                self._voicevox.play_wav(wav)
            except Exception:
                pass  # skip failed synthesis silently

    def closeEvent(self, event) -> None:  # noqa: N802
        self._twitch.stop()
        self._stop_tts_worker()
        super().closeEvent(event)
