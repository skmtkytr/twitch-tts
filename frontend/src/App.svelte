<script>
  import {Connect, Disconnect, GetSpeakers, SetSpeaker, SetTTSEnabled, SetReadName, SetNameSuffix, SendChat, TwitchLogin, LoadConfig, SaveConfig} from '../wailsjs/go/main/App.js'
  import {EventsOn} from '../wailsjs/runtime/runtime.js'

  let channel = ''
  let token = ''
  let connected = false
  let connecting = false
  let canWrite = false
  let chatInput = ''
  let sending = false
  let twitchUser = ''
  let loggingIn = false
  let status = ''
  let messages = []
  let speakers = []
  let selectedSpeaker = 1
  let voicevoxOk = false
  let ttsEnabled = true
  let readName = true
  let nameSuffix = 'さん'

  // Stable color per username
  const userColors = {}
  const palette = ['#f38ba8','#fab387','#f9e2af','#a6e3a1','#94e2d5','#89dceb','#74c7ec','#89b4fa','#b4befe','#cba6f7','#f5c2e7','#eba0ac']
  function colorFor(name) {
    if (!userColors[name]) {
      let hash = 0
      for (let i = 0; i < name.length; i++) hash = name.charCodeAt(i) + ((hash << 5) - hash)
      userColors[name] = palette[Math.abs(hash) % palette.length]
    }
    return userColors[name]
  }

  // Load config and speakers on mount
  LoadConfig().then(cfg => {
    channel = cfg.channel || ''
    token = cfg.token || ''
    if (token) twitchUser = '(認証済み)'
    if (cfg.speaker_id) selectedSpeaker = cfg.speaker_id
    if (cfg.read_name !== undefined) readName = cfg.read_name
    if (cfg.name_suffix !== undefined) nameSuffix = cfg.name_suffix
    SetReadName(readName)
    SetNameSuffix(nameSuffix)
  })

  function fetchSpeakers() {
    return GetSpeakers()
      .then(s => { speakers = s || []; voicevoxOk = speakers.length > 0 })
      .catch(() => { speakers = []; voicevoxOk = false })
  }

  fetchSpeakers()

  const voicevoxTimer = setInterval(() => {
    if (!voicevoxOk) fetchSpeakers()
  }, 5000)

  EventsOn('chat-message', (msg) => {
    messages = [...messages, msg]
    setTimeout(() => {
      const el = document.getElementById('chat-log')
      if (el) el.scrollTop = el.scrollHeight
    }, 0)
  })

  EventsOn('connected', (writable) => {
    connected = true
    connecting = false
    canWrite = !!writable
    status = ''
    SaveConfig({channel, token, speaker_id: selectedSpeaker, read_name: readName, name_suffix: nameSuffix})
  })

  EventsOn('disconnected', (err) => {
    connected = false
    connecting = false
    canWrite = false
    status = err ? 'Error: ' + err : ''
  })

  async function toggleConnection() {
    if (connected) {
      await Disconnect()
    } else {
      if (!channel) {
        status = 'Channel is required'
        return
      }
      connecting = true
      status = ''
      try {
        await Connect(channel, token)
      } catch (e) {
        status = 'Error: ' + e
        connecting = false
      }
    }
  }

  function onSpeakerChange() { SetSpeaker(selectedSpeaker) }
  function onTTSToggle() { SetTTSEnabled(ttsEnabled) }
  function onReadNameToggle() { SetReadName(readName) }
  function onNameSuffixChange() { SetNameSuffix(nameSuffix) }

  async function sendMessage() {
    if (!chatInput.trim() || sending) return
    sending = true
    try {
      await SendChat(chatInput)
      chatInput = ''
    } catch (e) {
      status = 'Send error: ' + e
    } finally {
      sending = false
    }
  }

  function onChatKeydown(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      sendMessage()
    }
  }

  async function twitchLogin() {
    loggingIn = true
    status = 'ブラウザで Twitch にログインしてください...'
    try {
      const t = await TwitchLogin()
      token = t
      twitchUser = '(認証済み)'
      status = ''
    } catch (e) {
      status = 'Login error: ' + e
    } finally {
      loggingIn = false
    }
  }

  function twitchLogout() {
    token = ''
    twitchUser = ''
  }
</script>

<main>
  <!-- Status indicators -->
  <div class="status-row">
    <div class="indicator">
      <span class="dot" class:dot-green={connected} class:dot-red={!connected}></span>
      <span>Twitch {connected ? 'Connected' : 'Disconnected'}</span>
    </div>
    <div class="indicator">
      <span class="dot" class:dot-green={voicevoxOk} class:dot-red={!voicevoxOk}></span>
      <span>VOICEVOX {voicevoxOk ? 'OK' : '未検出'}</span>
    </div>
    {#if status}
      <span class="status-msg">{status}</span>
    {/if}
  </div>

  <!-- Connection card -->
  <div class="card">
    <div class="card-header">Twitch 接続</div>
    <div class="card-body connection-body">
      <label class="field">
        <span class="field-label">Channel</span>
        <input type="text" bind:value={channel} placeholder="channel_name" disabled={connected} />
      </label>
      <div class="field">
        <span class="field-label">アカウント</span>
        <div class="twitch-auth">
          {#if twitchUser}
            <span class="twitch-user">{twitchUser}</span>
            <button class="btn btn-ghost btn-xs" on:click={twitchLogout} disabled={connected}>Logout</button>
          {:else}
            <button class="btn btn-twitch btn-xs" on:click={twitchLogin} disabled={connected || loggingIn}>
              {loggingIn ? '認証中...' : 'Twitch でログイン'}
            </button>
            <span class="hint">なくても読み取り専用で接続可</span>
          {/if}
        </div>
      </div>
      <div class="field field-action">
        <button class="btn btn-primary" on:click={toggleConnection} disabled={connecting}>
          {#if connecting}
            Connecting...
          {:else if connected}
            Disconnect
          {:else}
            Connect
          {/if}
        </button>
      </div>
    </div>
  </div>

  <!-- TTS Settings card -->
  <div class="card">
    <div class="card-header">TTS 設定</div>
    <div class="card-body settings-body">
      {#if voicevoxOk}
        <label class="field">
          <span class="field-label">Speaker</span>
          <select bind:value={selectedSpeaker} on:change={onSpeakerChange}>
            {#each speakers as s}
              <option value={s.id}>{s.name}</option>
            {/each}
          </select>
        </label>
      {:else}
        <div class="field">
          <span class="voicevox-warn">VOICEVOX を起動すると自動で接続します</span>
        </div>
      {/if}
      <div class="toggles">
        <label class="toggle-label">
          <input type="checkbox" bind:checked={ttsEnabled} on:change={onTTSToggle} />
          <span>TTS</span>
        </label>
        <label class="toggle-label">
          <input type="checkbox" bind:checked={readName} on:change={onReadNameToggle} />
          <span>名前読み上げ</span>
        </label>
        <label class="suffix-field">
          <span>敬称</span>
          <input type="text" class="suffix-input" bind:value={nameSuffix} on:change={onNameSuffixChange} placeholder="さん" />
        </label>
      </div>
    </div>
  </div>

  <!-- Chat log -->
  <div class="chat-log" id="chat-log">
    {#if messages.length === 0}
      <div class="chat-empty">チャットメッセージがここに表示されます</div>
    {/if}
    {#each messages as msg}
      <div class="msg">
        <span class="author" style="color: {colorFor(msg.author)}">{msg.author}</span>
        <span class="msg-text">{msg.content}</span>
      </div>
    {/each}
  </div>

  <!-- Chat input -->
  {#if connected && canWrite}
    <div class="chat-input">
      <input
        type="text"
        bind:value={chatInput}
        on:keydown={onChatKeydown}
        placeholder="メッセージを入力..."
        disabled={sending}
      />
      <button class="btn btn-send" on:click={sendMessage} disabled={sending || !chatInput.trim()}>
        Send
      </button>
    </div>
  {/if}
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
    gap: 8px;
  }

  /* ---- Status row ---- */
  .status-row {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 6px 12px;
    background: #181825;
    border-radius: 8px;
    font-size: 12px;
  }

  .indicator {
    display: flex;
    align-items: center;
    gap: 6px;
    font-weight: 600;
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    display: inline-block;
    flex-shrink: 0;
  }

  .dot-green { background: #a6e3a1; box-shadow: 0 0 6px #a6e3a144; }
  .dot-red   { background: #f38ba8; box-shadow: 0 0 6px #f38ba844; }

  .status-msg {
    margin-left: auto;
    color: #fab387;
  }

  /* ---- Cards ---- */
  .card {
    background: #181825;
    border-radius: 8px;
    overflow: hidden;
  }

  .card-header {
    font-size: 11px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #6c7086;
    padding: 8px 12px 4px;
  }

  .card-body {
    padding: 8px 12px 12px;
  }

  /* ---- Connection card ---- */
  .connection-body {
    display: flex;
    gap: 12px;
    align-items: end;
  }

  .field {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .field-label {
    font-size: 11px;
    color: #6c7086;
    font-weight: 600;
  }

  .field input[type="text"] {
    padding: 6px 8px;
    border: 1px solid #45475a;
    border-radius: 6px;
    background: #313244;
    color: #cdd6f4;
    font-size: 13px;
    min-width: 0;
  }

  .field input[type="text"]:focus {
    border-color: #89b4fa;
    outline: none;
  }

  .field input[type="text"]:disabled {
    opacity: 0.5;
  }

  .field-action {
    margin-left: auto;
  }

  .twitch-auth {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .twitch-user {
    color: #a6e3a1;
    font-weight: 600;
    font-size: 12px;
  }

  .hint {
    color: #6c7086;
    font-size: 11px;
  }

  /* ---- Settings card ---- */
  .settings-body {
    display: flex;
    gap: 16px;
    align-items: center;
    flex-wrap: wrap;
  }

  .settings-body select {
    padding: 5px 8px;
    border: 1px solid #45475a;
    border-radius: 6px;
    background: #313244;
    color: #cdd6f4;
    font-size: 13px;
  }

  .settings-body select:focus {
    border-color: #89b4fa;
    outline: none;
  }

  .toggles {
    display: flex;
    gap: 14px;
    align-items: center;
  }

  .toggle-label {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
    user-select: none;
  }

  .toggle-label input[type="checkbox"] {
    accent-color: #89b4fa;
    width: 15px;
    height: 15px;
  }

  .suffix-field {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 13px;
    font-weight: 600;
  }

  .suffix-input {
    width: 48px;
    padding: 4px 6px;
    border: 1px solid #45475a;
    border-radius: 6px;
    background: #313244;
    color: #cdd6f4;
    font-size: 13px;
  }

  .suffix-input:focus {
    border-color: #89b4fa;
    outline: none;
  }

  .voicevox-warn {
    color: #fab387;
    font-size: 12px;
    font-weight: 600;
  }

  /* ---- Buttons ---- */
  .btn {
    border: none;
    border-radius: 6px;
    font-weight: 600;
    cursor: pointer;
    white-space: nowrap;
    transition: background 0.15s, opacity 0.15s;
  }

  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .btn-primary {
    padding: 7px 20px;
    background: #89b4fa;
    color: #1e1e2e;
    font-size: 13px;
  }

  .btn-primary:hover:not(:disabled) { background: #74c7ec; }

  .btn-twitch {
    background: #9146ff;
    color: #fff;
  }

  .btn-twitch:hover:not(:disabled) { background: #7c3aed; }

  .btn-ghost {
    background: #45475a;
    color: #cdd6f4;
  }

  .btn-ghost:hover:not(:disabled) { background: #585b70; }

  .btn-xs {
    padding: 4px 10px;
    font-size: 11px;
  }

  .btn-send {
    padding: 7px 18px;
    background: #a6e3a1;
    color: #1e1e2e;
    font-size: 13px;
  }

  .btn-send:hover:not(:disabled) { background: #94e2d5; }

  /* ---- Chat log ---- */
  .chat-log {
    flex: 1;
    overflow-y: auto;
    background: #181825;
    border-radius: 8px;
    padding: 10px 14px;
    font-size: 14px;
    line-height: 1.7;
  }

  .chat-empty {
    color: #45475a;
    text-align: center;
    padding-top: 40px;
    font-size: 13px;
  }

  .msg {
    word-break: break-word;
    padding: 1px 0;
  }

  .author {
    font-weight: 700;
  }

  .author::after {
    content: ': ';
    color: #6c7086;
    font-weight: 400;
  }

  .msg-text {
    color: #cdd6f4;
  }

  /* ---- Chat input ---- */
  .chat-input {
    display: flex;
    gap: 8px;
  }

  .chat-input input {
    flex: 1;
    padding: 7px 10px;
    border: 1px solid #45475a;
    border-radius: 6px;
    background: #313244;
    color: #cdd6f4;
    font-size: 14px;
  }

  .chat-input input:focus {
    border-color: #89b4fa;
    outline: none;
  }

  .chat-input input:disabled {
    opacity: 0.5;
  }
</style>
