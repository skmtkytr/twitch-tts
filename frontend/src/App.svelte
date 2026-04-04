<script>
  import {Connect, Disconnect, GetSpeakers, SetSpeaker, SetTTSEnabled, SetReadName, SetNameSuffix, LoadConfig, SaveConfig} from '../wailsjs/go/main/App.js'
  import {EventsOn} from '../wailsjs/runtime/runtime.js'

  let channel = ''
  let token = ''
  let connected = false
  let connecting = false
  let status = 'Disconnected'
  let messages = []
  let speakers = []
  let selectedSpeaker = 1
  let ttsEnabled = true
  let readName = true
  let nameSuffix = 'さん'

  // Load config and speakers on mount
  LoadConfig().then(cfg => {
    channel = cfg.channel || ''
    token = cfg.token || ''
    if (cfg.speaker_id) selectedSpeaker = cfg.speaker_id
    if (cfg.read_name !== undefined) readName = cfg.read_name
    if (cfg.name_suffix !== undefined) nameSuffix = cfg.name_suffix
    SetReadName(readName)
    SetNameSuffix(nameSuffix)
  })

  GetSpeakers()
    .then(s => { speakers = s || [] })
    .catch(() => { status = 'VOICEVOX not running' })

  EventsOn('chat-message', (msg) => {
    messages = [...messages, msg]
    // Auto-scroll: defer to next tick
    setTimeout(() => {
      const el = document.getElementById('chat-log')
      if (el) el.scrollTop = el.scrollHeight
    }, 0)
  })

  EventsOn('connected', () => {
    connected = true
    connecting = false
    status = 'Connected'
    SaveConfig({channel, token, speaker_id: selectedSpeaker, read_name: readName, name_suffix: nameSuffix})
  })

  EventsOn('disconnected', () => {
    connected = false
    connecting = false
    status = 'Disconnected'
  })

  async function toggleConnection() {
    if (connected) {
      await Disconnect()
    } else {
      if (!channel || !token) {
        status = 'Channel and Token are required'
        return
      }
      connecting = true
      status = 'Connecting...'
      try {
        await Connect(channel, token)
      } catch (e) {
        status = 'Error: ' + e
        connecting = false
      }
    }
  }

  function onSpeakerChange() {
    SetSpeaker(selectedSpeaker)
  }

  function onTTSToggle() {
    SetTTSEnabled(ttsEnabled)
  }

  function onReadNameToggle() {
    SetReadName(readName)
  }

  function onNameSuffixChange() {
    SetNameSuffix(nameSuffix)
  }
</script>

<main>
  <div class="connection">
    <label>
      Channel
      <input type="text" bind:value={channel} placeholder="channel" disabled={connected} />
    </label>
    <label>
      Token
      <input type="password" bind:value={token} placeholder="oauth:xxxxx" disabled={connected} />
    </label>
    <button on:click={toggleConnection} disabled={connecting}>
      {connected ? 'Disconnect' : 'Connect'}
    </button>
  </div>

  <div class="controls">
    <label>
      Speaker
      <select bind:value={selectedSpeaker} on:change={onSpeakerChange}>
        {#each speakers as s}
          <option value={s.id}>{s.name}</option>
        {/each}
      </select>
    </label>
    <label class="tts-toggle">
      <input type="checkbox" bind:checked={ttsEnabled} on:change={onTTSToggle} />
      TTS
    </label>
    <label class="tts-toggle">
      <input type="checkbox" bind:checked={readName} on:change={onReadNameToggle} />
      Name
    </label>
    <label>
      Suffix
      <input type="text" class="suffix-input" bind:value={nameSuffix} on:change={onNameSuffixChange} placeholder="さん" />
    </label>
  </div>

  <div class="chat-log" id="chat-log">
    {#each messages as msg}
      <div class="msg"><span class="author">{msg.author}</span>: {msg.content}</div>
    {/each}
  </div>

  <div class="status-bar">{status}</div>
</main>

<style>
  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    background: #1e1e2e;
    color: #cdd6f4;
  }

  main {
    display: flex;
    flex-direction: column;
    height: 100vh;
    padding: 12px;
    box-sizing: border-box;
    gap: 10px;
  }

  .connection {
    display: flex;
    gap: 8px;
    align-items: end;
  }

  .connection label {
    display: flex;
    flex-direction: column;
    font-size: 12px;
    flex: 1;
    gap: 4px;
  }

  .connection input {
    padding: 6px 8px;
    border: 1px solid #45475a;
    border-radius: 6px;
    background: #313244;
    color: #cdd6f4;
    font-size: 14px;
  }

  .connection button {
    padding: 6px 16px;
    border: none;
    border-radius: 6px;
    background: #89b4fa;
    color: #1e1e2e;
    font-weight: 600;
    cursor: pointer;
    font-size: 14px;
    white-space: nowrap;
  }

  .connection button:hover {
    background: #74c7ec;
  }

  .connection button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .controls {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .controls label {
    font-size: 12px;
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .controls select {
    padding: 4px 8px;
    border: 1px solid #45475a;
    border-radius: 6px;
    background: #313244;
    color: #cdd6f4;
    font-size: 13px;
  }

  .tts-toggle {
    font-size: 14px !important;
    font-weight: 600;
  }

  .suffix-input {
    width: 50px;
    padding: 3px 6px;
    border: 1px solid #45475a;
    border-radius: 6px;
    background: #313244;
    color: #cdd6f4;
    font-size: 13px;
  }

  .chat-log {
    flex: 1;
    overflow-y: auto;
    background: #181825;
    border-radius: 8px;
    padding: 8px 12px;
    font-size: 14px;
    line-height: 1.6;
  }

  .msg {
    word-break: break-word;
  }

  .author {
    color: #f38ba8;
    font-weight: 600;
  }

  .status-bar {
    font-size: 12px;
    color: #6c7086;
    text-align: center;
  }
</style>
