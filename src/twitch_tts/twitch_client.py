from __future__ import annotations

import asyncio
import threading
from dataclasses import dataclass

import twitchio
from PyQt6.QtCore import QObject, pyqtSignal


@dataclass
class ChatMessage:
    author: str
    content: str


class TwitchWorker(QObject):
    message_received = pyqtSignal(ChatMessage)
    connected = pyqtSignal()
    disconnected = pyqtSignal()
    error = pyqtSignal(str)

    def __init__(self) -> None:
        super().__init__()
        self._thread: threading.Thread | None = None
        self._loop: asyncio.AbstractEventLoop | None = None
        self._client: _TwitchClient | None = None

    def start(self, token: str, channel: str) -> None:
        self.stop()
        self._thread = threading.Thread(target=self._run, args=(token, channel), daemon=True)
        self._thread.start()

    def _run(self, token: str, channel: str) -> None:
        self._loop = asyncio.new_event_loop()
        asyncio.set_event_loop(self._loop)
        self._client = _TwitchClient(token, channel, self)
        try:
            self._loop.run_until_complete(self._client.start())
        except Exception as e:
            self.error.emit(str(e))
        finally:
            self.disconnected.emit()

    def stop(self) -> None:
        if self._client and self._loop:
            asyncio.run_coroutine_threadsafe(self._client.close(), self._loop)
        if self._thread:
            self._thread.join(timeout=5)
            self._thread = None
        self._client = None
        self._loop = None


class _TwitchClient(twitchio.Client):
    def __init__(self, token: str, channel: str, worker: TwitchWorker) -> None:
        super().__init__(token=token, initial_channels=[channel])
        self._worker = worker

    async def event_ready(self) -> None:
        self._worker.connected.emit()

    async def event_message(self, message: twitchio.Message) -> None:
        if message.echo:
            return
        author = message.author.name if message.author else "???"
        self._worker.message_received.emit(ChatMessage(author=author, content=message.content))
