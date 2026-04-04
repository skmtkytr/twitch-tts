from __future__ import annotations

import subprocess
import tempfile
from dataclasses import dataclass

import aiohttp


@dataclass
class Speaker:
    name: str
    id: int


class VoicevoxClient:
    def __init__(self, base_url: str = "http://localhost:50021") -> None:
        self.base_url = base_url

    async def speakers(self) -> list[Speaker]:
        async with aiohttp.ClientSession() as session:
            async with session.get(f"{self.base_url}/speakers") as resp:
                resp.raise_for_status()
                data = await resp.json()
        result: list[Speaker] = []
        for speaker in data:
            for style in speaker["styles"]:
                result.append(Speaker(
                    name=f"{speaker['name']} ({style['name']})",
                    id=style["id"],
                ))
        return result

    async def synthesize(self, text: str, speaker_id: int) -> bytes:
        async with aiohttp.ClientSession() as session:
            async with session.post(
                f"{self.base_url}/audio_query",
                params={"text": text, "speaker": speaker_id},
            ) as resp:
                resp.raise_for_status()
                query = await resp.json()

            async with session.post(
                f"{self.base_url}/synthesis",
                params={"speaker": speaker_id},
                json=query,
            ) as resp:
                resp.raise_for_status()
                return await resp.read()

    @staticmethod
    def play_wav(wav_data: bytes) -> None:
        with tempfile.NamedTemporaryFile(suffix=".wav", delete=True) as f:
            f.write(wav_data)
            f.flush()
            subprocess.run(
                ["paplay", f.name],
                check=False,
                capture_output=True,
            )
