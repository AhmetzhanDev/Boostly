<template>
  <div class="dash-wrap" :class="{ 'drawer-open': sidebarOpen }">
    <!-- Top bar -->
    <header class="topbar">
      <div class="brand">
        <span class="brand-text">SpeakApper AI</span>
      </div>
      <div class="topbar-actions">
       
        <!-- Profile dropdown -->
        <div class="profile" @keydown.esc="profileMenu=false">
          <button class="avatar" @click="profileMenu=!profileMenu" :aria-expanded="profileMenu ? 'true' : 'false'" aria-haspopup="menu">
            <span>{{ userInitials }}</span>
          </button>
          <div v-if="profileMenu" class="profile-menu" role="menu">
            <div class="pm-head">
              <div class="pm-avatar">{{ userInitials }}</div>
              <div class="pm-meta">
                <div class="pm-name">{{ userDisplayName }}</div>
                <div class="pm-mail">{{ userEmail }}</div>
              </div>
            </div>
            <button class="pm-item" @click="goSettings"><span>⚙️</span> Настройки</button>
            <button class="pm-item danger" @click="logout"><span>🚪</span> Выйти</button>
          </div>
        </div>
        
        <button v-if="!sidebarOpen" class="burger" @click="toggleSidebar" aria-label="Toggle menu">
          <span class="burger-box"><span class="burger-lines"></span></span>
        </button>
      </div>
    </header>

    <!-- Single expanding side pane -->
    <aside class="navpane" :class="{ open: sidebarOpen }" @click.self="profileMenu=false">
      <div v-if="sidebarOpen" class="pane-head">
        <div class="pane-logo">Boostly</div>
        <div style="display:flex; gap:8px; align-items:center;">
          
          <button class="collapse-btn" @click="toggleSidebar" aria-label="Close menu">
            <svg viewBox="0 0 24 24" width="18" height="18"><path d="M15 6 9 12l6 6" fill="none" stroke="#ffffff" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
          </button>
        </div>
      </div>
      <!-- User card in sidebar -->
      <div v-if="sidebarOpen" class="user-card">
        <div class="uc-left">{{ userInitials }}</div>
        <div class="uc-mid">
          <div class="uc-name">{{ userDisplayName }}</div>
          <div class="uc-mail">{{ userEmail }}</div>
        </div>
        
      </div>
      <button v-else class="rail-burger" @click="toggleSidebar" aria-label="Open menu">
        <svg viewBox="0 0 24 24" width="22" height="22" aria-hidden="true">
          <path d="M4 7h16M4 12h16M4 17h16" stroke="#ffffff" stroke-width="2" stroke-linecap="round"/>
        </svg>
      </button>

      <nav class="menu">
        <button class="menu-item active" @click="$router.push('/dashboard')">
          <span class="mi-ico">
            <!-- Home icon (outline for collapsed look) -->
            <svg viewBox="0 0 24 24" aria-hidden="true" class="ico-outline">
              <path d="M3 10l9-7 9 7v8a2 2 0 0 1-2 2h-4v-5H9v5H5a2 2 0 0 1-2-2v-8z" fill="none" stroke="#ffffff" stroke-width="2" stroke-linejoin="round"/>
            </svg>
          </span>
          <span class="mi-text">Dashboard</span>
        </button>

        <button class="menu-item" @click="goSettings">
          <span class="mi-ico">
            <!-- Settings icon (outline for collapsed look) -->
            <svg viewBox="0 0 24 24" aria-hidden="true" class="ico-outline">
              <path d="M12 15.5a3.5 3.5 0 1 0 0-7 3.5 3.5 0 0 0 0 7z" fill="none" stroke="#ffffff" stroke-width="2"/>
              <path d="M19.4 15a8 8 0 0 0 .06-6l2.04-1.59-1.92-3.32-2.39.96a7.8 7.8 0 0 0-2.27-1.3L14.6 1h-5.2L8.98 3.75a7.8 7.8 0 0 0-2.27 1.3l-2.39-.96L2.4 7.41 4.44 9a8 8 0 0 0 0 6l-2.04 1.59 1.92 3.32 2.39-.96c.7.53 1.45.95 2.27 1.3L9.4 23h5.2l.22-1.75c.82-.35 1.57-.77 2.27-1.3l2.39.96 1.92-3.32L19.4 15z" fill="none" stroke="#ffffff" stroke-width="2" stroke-linejoin="round" stroke-linecap="round"/>
            </svg>
          </span>
          <span class="mi-text">Settings</span>
        </button>

        <transition name="upg">
          <button v-show="sidebarOpen" class="upgrade" @click="upgrade">
            <span class="spark">✨</span>
            <span class="upg-text">Upgrade to Premium</span>
          </button>
        </transition>
      </nav>
      <!-- Folders Section -->
      <section v-if="sidebarOpen" class="folders">
        <div class="folders-head">
          <div class="folders-title">Folders</div>
          <button class="folders-add" @click="createFolder" aria-label="Create folder">＋</button>
        </div>
        <div class="folder-list">
          <button
            class="folder-item"
            :class="{ active: activeFolderId===null }"
            @click="activeFolderId=null"
          >
            <span class="fi-ico">🗂️</span>
            <span class="fi-text">All notes</span>
            <span class="fi-count">{{ notes.length }}</span>
          </button>
          <button
            v-for="f in folders"
            :key="f.id"
            class="folder-item"
            :class="{ active: activeFolderId===f.id }"
            @click="activeFolderId=f.id"
          >
            <span class="fi-ico">📁</span>
            <span class="fi-text">{{ f.name }}</span>
            <span class="fi-count">{{ folderCount(f.id) }}</span>
          </button>
        </div>
      </section>

      <!-- Sidebar footer with theme toggle -->
     
    </aside>

    <!-- Main content -->
    <main class="main" @click="profileMenu=false">
      <!-- Header -->
      <header class="page-header">
        <div class="ph-row">
          <div>
            <h1 class="h1">Dashboard</h1>
            <p class="sub">Create new notes</p>
          </div>
          <div class="search-wrap">
            <svg viewBox="0 0 24 24" width="18" height="18" aria-hidden="true">
              <circle cx="11" cy="11" r="7" stroke="#ffffff" stroke-width="2" fill="none"/>
              <path d="M20 20l-3.2-3.2" stroke="#ffffff" stroke-width="2" stroke-linecap="round"/>
            </svg>
            <input v-model="searchQuery" class="search" type="text" placeholder="Search notes..." />
            <kbd class="slash">/</kbd>
          </div>
        </div>
      </header>

      <!-- Quick actions -->
      <section class="quick">
        <div class="qa" v-for="(card, i) in quickActions" :key="i" @click="card.action()">
          <div class="qa-ico" :class="[card.color, card.key]">
            <!-- Blank document -->
            <svg v-if="card.key==='blank'" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8l-6-6zM15 9V3.5L20.5 9H15z"/>
            </svg>
            <!-- Microphone -->
            <svg v-else-if="card.key==='audio'" viewBox="0 0 24 24" aria-hidden="true">
              <path d="M12 15a3 3 0 0 0 3-3V7a3 3 0 0 0-6 0v5a3 3 0 0 0 3 3zm5-4v1a5 5 0 0 1-10 0v-1H5v1a7 7 0 0 0 6 6v3h2v-3a7 7 0 0 0 6-6v-1h-2z"/>
            </svg>
            <!-- Document upload (DOC tag) -->
            <svg v-else-if="card.key==='doc'" viewBox="0 0 24 24" aria-hidden="true" class="doc-svg">
              <rect class="doc-paper" x="4" y="3.5" width="16" height="17" rx="3" ry="3"/>
              <polygon class="doc-fold" points="16,3.5 20,7.5 16,7.5"/>
              <text x="12" y="15" text-anchor="middle" class="doc-text">DOC</text>
            </svg>
            <!-- YouTube -->
            <svg v-else viewBox="0 0 24 24" aria-hidden="true" class="yt-svg">
              <rect x="2" y="5" width="20" height="14" rx="4" ry="4" fill="#FF0000"/>
              <polygon points="10,9 15.5,12 10,15" fill="#FFFFFF"/>
            </svg>
          </div>
          <div class="qa-body">
            <div class="qa-title">{{ card.title }}</div>
            <div class="qa-desc">{{ card.desc }}</div>
          </div>
          <div class="qa-arrow">›</div>
        </div>
      </section>

      <!-- Actions header -->
      <section class="tabs-row">
        <div class="tabs">
          <button class="tab active">My Notes</button>
        </div>
        
      </section>

      <!-- Notes list -->
      <section class="notes-list">
        <article class="note" v-for="note in filteredNotes" :key="note.id" @click="openNote(note)">
          <div class="note-ico" :class="note.type">
            <template v-if="note.type==='audio'">
              <!-- Microphone icon (same as quick action) -->
              <svg viewBox="0 0 24 24" width="20" height="20" aria-hidden="true">
                <path fill="currentColor" d="M12 15a3 3 0 0 0 3-3V7a3 3 0 0 0-6 0v5a3 3 0 0 0 3 3zm5-4v1a5 5 0 0 1-10 0v-1H5v1a7 7 0 0 0 6 6v3h2v-3a7 7 0 0 0 6-6v-1h-2z"/>
              </svg>
            </template>
            <template v-else-if="note.type==='yt'">
              <!-- YouTube icon (same as quick action) -->
              <svg viewBox="0 0 24 24" aria-hidden="true" class="yt-svg">
                <rect x="2" y="5" width="20" height="14" rx="4" ry="4" fill="#FF0000"/>
                <polygon points="10,9 15.5,12 10,15" fill="#FFFFFF"/>
              </svg>
            </template>
            <span v-else>📄</span>
          </div>
          <div class="note-body">
            <div class="note-title">{{ note.title }}</div>
            <div class="note-meta">Last opened {{ note.lastOpened }}</div>
          </div>
          <button class="note-more" @click.stop="moreNote(note)">⋯</button>
        </article>
        <div v-if="filteredNotes.length===0" class="empty">No notes yet</div>
      </section>
    </main>

    <!-- Audio Record Modal -->
    <div v-if="showAudioModal" class="modal-wrap" @keydown.esc="closeAudioModal">
      <div class="modal-backdrop" @click="closeAudioModal"></div>
      <div class="modal" role="dialog" aria-modal="true" aria-label="Record audio">
        <div class="modal-head">
          <div class="modal-title">Record or upload audio</div>
          <button class="modal-close" @click="closeAudioModal" aria-label="Close">
            ✕
          </button>
        </div>
        <div class="modal-body">
          <div class="rec-center">
            <button class="mic-btn" :class="{ recording: isRecording }" @click="toggleRecord" aria-label="Toggle recording">
              <svg v-if="!isRecording" viewBox="0 0 24 24" width="28" height="28" aria-hidden="true">
                <path d="M12 15a3 3 0 0 0 3-3V7a3 3 0 0 0-6 0v5a3 3 0 0 0 3 3zm5-4v1a5 5 0 0 1-10 0v-1H5v1a7 7 0 0 0 6 6v3h2v-3a7 7 0 0 0 6-6v-1h-2z" fill="currentColor"/>
              </svg>
              <svg v-else viewBox="0 0 24 24" width="28" height="28" aria-hidden="true">
                <rect x="6" y="5" width="4" height="14" rx="1" fill="#111"/>
                <rect x="14" y="5" width="4" height="14" rx="1" fill="#111"/>
              </svg>
            </button>
          </div>

          <!-- Voice visualizer -->
          <div class="viz">
            <div v-for="(h, i) in barHeights" :key="i" class="bar" :style="{ height: h + '%'}"></div>
          </div>

          <div class="rec-bar" style="justify-content:center">
            <span class="pill">
              <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true"><rect x="4" y="4" width="16" height="16" rx="3" fill="#fff"/></svg>
              {{ fmtTime(elapsedSec) }}
            </span>
          </div>

          <div class="rec-actions" style="justify-content:center; margin-top:8px;">
            <input ref="fileInput" type="file" accept="audio/*,.m4a,.mp3,.wav,.webm,.ogg,.aac,.flac" @change="onFilePicked" style="display:none" />
            <button class="btn btn-ghost" :class="{disabled: isRecording}" @click="$refs.fileInput && $refs.fileInput.click()" :disabled="isRecording" aria-label="Upload audio">
              <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true"><path fill="currentColor" d="M5 20h14a1 1 0 0 0 1-1v-7h-2v6H6V12H4v7a1 1 0 0 0 1 1zm7-16-5 5h3v4h4v-4h3l-5-5z"/></svg>
              <span style="margin-left:6px">Upload audio</span>
            </button>
            <button class="btn primary" :disabled="!audioUrl || isRecording" @click="generateNote" aria-label="Generate note">
              <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true"><path fill="currentColor" d="M19 3H5a2 2 0 0 0-2 2v14l4-4h12a2 2 0 0 0 2-2V5a2 2 0 0 0-2-2z"/></svg>
              <span style="margin-left:6px">Generate note</span>
            </button>
          </div>

          <div v-if="audioUrl" class="audio-preview">
            <!-- Hidden audio element controlled by custom UI -->
            <audio ref="player" :src="audioUrl" preload="metadata" @timeupdate="onTimeUpdate" @loadedmetadata="onLoadedMeta" @ended="onEnded" style="display:none"></audio>

            <div class="player">
              <button class="pp-btn" @click="togglePlay" :aria-label="isPlaying ? 'Pause' : 'Play'">
                <svg v-if="!isPlaying" viewBox="0 0 24 24" width="18" height="18" aria-hidden="true"><path fill="currentColor" d="M8 5v14l11-7z"/></svg>
                <svg v-else viewBox="0 0 24 24" width="18" height="18" aria-hidden="true"><path fill="currentColor" d="M7 6h4v12H7zM13 6h4v12h-4z"/></svg>
              </button>
              <div class="progress" @click="seek($event)">
                <div class="progress-track">
                  <div class="progress-fill" :style="{ width: progressPercent + '%' }"></div>
                </div>
              </div>
              <div class="time">{{ fmtTime(currentTime) }} / {{ fmtTime(duration) }}</div>
            </div>

            
          </div>
        </div>
      </div>
    </div>

    <!-- Upgrade to Premium Modal -->
    <div v-if="showUpgradeModal" class="modal-wrap" @keydown.esc="closeUpgradeModal">
      <div class="modal-backdrop" @click="closeUpgradeModal"></div>
      <div class="modal" role="dialog" aria-modal="true" aria-label="Upgrade to Premium">
        <div class="modal-head">
          <div class="modal-title">Premium</div>
          <button class="modal-close" @click="closeUpgradeModal" aria-label="Close">✕</button>
        </div>
        <div class="modal-body">
          <div class="premium-intro">Разблокируйте все возможности без ограничений</div>
          <div class="pricing">
            <div class="plan">
              <div class="plan-name">Месяц</div>
              <div class="plan-price"><span class="n">1</span> тг<span class="per">/мес</span></div>
              <button class="btn primary" @click="selectPlan('monthly')">Выбрать</button>
            </div>
            <div class="plan best">
              <div class="badge">Популярно</div>
              <div class="plan-name">Год</div>
              <div class="plan-price"><span class="n">12</span> тг<span class="per">/год</span></div>
              <div class="note">Экономия 0%</div>
              <button class="btn primary" @click="selectPlan('yearly')">Выбрать</button>
            </div>
          </div>
          <div class="fine">Оплата — демо-заглушка. Интеграция с платежами будет добавлена позднее.</div>
        </div>
      </div>
    </div>

    <!-- YouTube URL Modal -->
    <div v-if="showYoutubeModal" class="modal-wrap" @keydown.esc="closeYoutubeModal">
      <div class="modal-backdrop" @click="closeYoutubeModal"></div>
      <div class="modal" role="dialog" aria-modal="true" aria-label="Add YouTube link">
        <div class="modal-head">
          <div class="modal-title">YouTube video</div>
          <button class="modal-close" @click="closeYoutubeModal" aria-label="Close">✕</button>
        </div>
        <div class="modal-body">
          <label class="field-label" for="yt-url">Paste a YouTube link</label>
          <input id="yt-url" v-model.trim="youtubeUrl" class="search" type="url" placeholder="https://www.youtube.com/watch?v=..." style="width:100%" />
          <div class="note" style="margin-top:10px">We will download audio, transcribe it and generate study materials.</div>
          <div class="rec-actions" style="justify-content:flex-end; margin-top:14px; gap:8px">
            <button class="btn outline" @click="closeYoutubeModal">Cancel</button>
            <button class="btn primary" :disabled="!youtubeUrl" @click="startYoutubeFlow">Start</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Processing modal -->
    <div v-if="proc.show" class="modal-wrap" @keydown.esc="proc.show=false">
      <div class="modal-backdrop" @click="proc.show=false"></div>
      <div class="modal" role="dialog" aria-modal="true" aria-label="Processing">
        <div class="modal-head">
          <div class="modal-title">Generating your note…</div>
          <button class="modal-close" @click="proc.show=false" aria-label="Close">
            <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M18.3 5.71L12 12.01 5.7 5.7 4.29 7.11l6.3 6.3-6.3 6.29 1.41 1.41 6.3-6.3 6.29 6.3 1.41-1.41-6.3-6.29 6.3-6.3z"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <div class="note">Don't close this page until the note is ready</div>

          <ul class="steps">
            <li class="step-item" :class="{done: proc.step1.done}">
              <div class="left" :class="{ glow: !proc.step1.done }"><span class="num">1</span></div>
              <div class="mid">
                <div class="stitle">Note is creating</div>
              </div>
              <div class="right">
                <span v-if="!proc.step1.done" class="spinner" aria-hidden="true"></span>
                <span class="badge" :class="proc.step1.done?'ok':''">{{ proc.step1.done ? 'Completed' : 'Pending' }}</span>
              </div>
            </li>
            <li class="step-item" :class="{done: proc.step2.done}">
              <div class="left" :class="{ glow: !proc.step2.done && proc.step1.done }"><span class="num">2</span></div>
              <div class="mid">
                <div class="stitle">Record is uploading</div>
                <div class="progress-line" v-if="!proc.step2.done">
                  <div class="progress-fill" :style="{ width: (proc.step2.progress||0) + '%' }"></div>
                </div>
                <div class="sdesc" v-if="!proc.step2.done">{{ proc.step2.progress }}%</div>
              </div>
              <div class="right">
                <span v-if="!proc.step2.done" class="spinner" aria-hidden="true"></span>
                <span class="badge" :class="proc.step2.done?'ok':''">{{ proc.step2.done ? 'Completed' : proc.step2.progress + '%' }}</span>
              </div>
            </li>
            <li class="step-item" :class="{done: proc.step3.done}">
              <div class="left" :class="{ glow: !proc.step3.done && proc.step2.done }"><span class="num">3</span></div>
              <div class="mid">
                <div class="stitle">Record is transcribing</div>
                <div class="sdesc">Progress {{ fmtTime(proc.step3.elapsed) }}</div>
              </div>
              <div class="right">
                <span v-if="!proc.step3.done" class="spinner" aria-hidden="true"></span>
                <span class="badge" :class="proc.step3.done?'ok':''">{{ proc.step3.done ? 'Completed' : 'Progress' }}</span>
              </div>
            </li>
            <li class="step-item" :class="{done: proc.step4.done}">
              <div class="left" :class="{ glow: !proc.step4.done && proc.step3.done }"><span class="num">4</span></div>
              <div class="mid">
                <div class="stitle">AI is generating note</div>
                <div class="sdesc">{{ proc.step4.done ? 'Completed' : 'Progress ' + fmtTime(proc.step4.elapsed) }}</div>
              </div>
              <div class="right">
                <span v-if="!proc.step4.done" class="spinner" aria-hidden="true"></span>
                <span class="badge" :class="proc.step4.done?'ok':''">{{ proc.step4.done ? 'Completed' : 'Pending' }}</span>
              </div>
            </li>
            <li class="step-item" :class="{done: proc.ready}">
              <div class="left" :class="{ glow: proc.step4.done && !proc.ready }"><span class="num">5</span></div>
              <div class="mid">
                <div class="stitle">Note is ready</div>
              </div>
              <div class="right"><span class="badge" :class="proc.ready?'ok':''">{{ proc.ready ? 'Completed' : 'Pending' }}</span></div>
            </li>
          </ul>

          <div class="footer">
            <button class="btn primary" :disabled="!proc.ready" @click="viewNoteNow">
              <svg viewBox="0 0 24 24" width="16" height="16" aria-hidden="true"><path fill="currentColor" d="M12 6a9.77 9.77 0 0 1 9 6 9.77 9.77 0 0 1-9 6 9.77 9.77 0 0 1-9-6 9.77 9.77 0 0 1 9-6zm0 10a4 4 0 1 0 0-8 4 4 0 0 0 0 8z"/></svg>
              <span style="margin-left:6px">View note now</span>
            </button>
          </div>

          <div v-if="proc.error" class="error">{{ proc.error }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Dashboard',
  data() {
    return {
      sidebarOpen: false,
      activeTab: 'my',
      searchQuery: '',
      activeFolderId: null, // null = All notes
      theme: 'dark',
      // Profile state
      profileMenu: false,
      user: null,
      quickActions: [
        { key: 'blank', title: 'Blank document', desc: 'Start from scratch', color: 'c-purple', action: () => this.createBlank() },
        { key: 'audio', title: 'Record or upload audio', desc: 'Upload an audio file', color: 'c-violet', action: () => this.openAudioModal() },
        { key: 'doc',   title: 'Document upload', desc: 'Any PDF, DOC, PPT, etc', color: 'c-purple', action: () => this.uploadDoc() },
        { key: 'yt',    title: 'YouTube video', desc: 'Paste a YouTube link', color: 'c-red', action: () => this.addYoutube() },
      ],
      folders: [
        { id: 'f1', name: 'Russian' },
        { id: 'f2', name: 'Work' },
      ],
      notes: [],
      // Audio modal state
      showAudioModal: false,
      audioBlob: null,
      // YouTube modal state
      showYoutubeModal: false,
      youtubeUrl: '',
      // Upgrade modal state
      showUpgradeModal: false,
      proc: {
        show: false,
        error: '',
        transcript: '',
        noteId: '',
        ready: false,
        step1: { done: false },
        step2: { done: false, inProgress: false, progress: 0 },
        step3: { done: false, elapsed: 0 },
        step4: { done: false, elapsed: 0 },
      },
      // Settings
      autoRedirect: (localStorage.getItem('autoRedirect') || 'true') === 'true',
      isRecording: false,
      elapsedSec: 0,
      audioUrl: null,
      barHeights: Array(32).fill(8),
      // Player state
      isPlaying: false,
      duration: 0,
      currentTime: 0,
      _mediaRecorder: null,
      _mediaStream: null,
      _timer: null,
      _audioCtx: null,
      _analyser: null,
      _raf: null,
    }
  },
  computed: {
    filteredNotes() {
      const q = this.searchQuery.trim().toLowerCase()
      return this.notes
        .filter(n => n.tab === this.activeTab)
        .filter(n => this.activeFolderId === null ? true : n.folderId === this.activeFolderId)
        .filter(n => q ? n.title.toLowerCase().includes(q) : true)
    },
    progressPercent() {
      if (!this.duration) return 0
      return Math.min(100, Math.max(0, (this.currentTime / this.duration) * 100))
    },
    userEmail(){
      const u = this.user || {}
      return u.email || ''
    },
    userDisplayName(){
      const u = this.user || {}
      const full = u.fullName || u.name || [u.firstName, u.lastName].filter(Boolean).join(' ').trim()
      if (full) return full
      if (u.email) return u.email.split('@')[0]
      return 'User'
    },
    userInitials(){
      const n = this.userDisplayName.trim()
      const parts = n.split(/\s+/).slice(0,2)
      const ini = parts.map(p => p[0]?.toUpperCase() || '').join('') || 'U'
      return ini
    }
  },
  mounted() {
    // Apply persisted theme or system preference
    const stored = localStorage.getItem('theme')
    if (stored === 'light' || stored === 'dark') {
      this.applyTheme(stored)
    } else if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
      this.applyTheme('light')
    } else {
      this.applyTheme('dark')
    }
    // Подхватить пользователя из localStorage
    try { const u = localStorage.getItem('user'); this.user = u ? JSON.parse(u) : null } catch(e) {}
    // Загрузка пользовательских материалов, если есть токен
    try { this.fetchMaterials() } catch(e) { console.warn(e) }
  },
  methods: {
    // --- YouTube note tagging helpers (persist locally) ---
    getYtSet() {
      try {
        const raw = localStorage.getItem('ytNotes')
        const arr = Array.isArray(JSON.parse(raw)) ? JSON.parse(raw) : []
        return new Set(arr.map(String))
      } catch (e) { return new Set() }
    },
    saveYtSet(set) {
      try { localStorage.setItem('ytNotes', JSON.stringify(Array.from(set))) } catch (e) {}
    },
    markNoteAsYt(id) {
      if (!id) return
      const s = this.getYtSet()
      s.add(String(id))
      this.saveYtSet(s)
    },
    isYt(id) {
      if (!id) return false
      return this.getYtSet().has(String(id))
    },

    // Создает короткий осмысленный заголовок из транскрипта
    makeShortTitle(text) {
      if (!text || typeof text !== 'string') return 'Untitled note'
      // Убираем переносы/дублирующиеся пробелы
      let s = text.replace(/\r/g, ' ').replace(/\n/g, ' ').split(/\s+/).join(' ').trim()
      if (!s) return 'Untitled note'
      // Режем по первому разделителю предложения
      const m = s.match(/(.+?[\.\!\?])\s|(.+?$)/)
      s = (m && (m[1] || m[2]) || s).trim()
      // Ограничиваем до 7 слов
      const words = s.split(/\s+/)
      if (words.length > 7) s = words.slice(0, 7).join(' ')
      // Ограничиваем длину до 60 символов, убираем хвостовые знаки
      if (s.length > 60) s = s.slice(0, 60).trim().replace(/[\,\.;:\-\s]+$/, '')
      // Убираем крайние кавычки/скобки
      s = s.replace(/^["'\(\[]+|["'\)\]]+$/g, '').trim()
      return s || 'Untitled note'
    },
    logout(){
      try { localStorage.removeItem('token'); localStorage.removeItem('user') } catch(e){}
      this.user = null
      this.profileMenu = false
      this.notes = []
      this.$router.push('/')
    },
    async fetchMaterials(){
      const token = localStorage.getItem('token')
      if (!token) { this.notes = []; return }
      try {
        const resp = await fetch('http://localhost:8080/api/materials', {
          headers: { 'Authorization': 'Bearer ' + token }
        })
        if (!resp.ok) { this.notes = []; return }
        const data = await resp.json()
        // API может вернуть массив напрямую или объект вида { materials: [...] }
        const list = Array.isArray(data) ? data : (Array.isArray(data?.materials) ? data.materials : [])
        // Map backend materials to dashboard notes list
        const mapped = list.map(m => {
          const transcript = m.transcript || ''
          const title = this.makeShortTitle(m.title || transcript)
          const updated = m.updatedAt || m.updated_at || m.updated || m.createdAt || m.created_at || Date.now()
          const id = m.id || m._id
          const type = this.isYt(id) ? 'yt' : 'audio'
          return {
            id,
            title,
            type,
            lastOpened: new Date(updated).toLocaleString(),
            updatedMs: new Date(updated).getTime(),
            tab: 'my',
            folderId: this.activeFolderId || null,
            transcript
          }
        })
        // Newest first
        this.notes = mapped.sort((a,b) => (b.updatedMs||0) - (a.updatedMs||0))
      } catch (e) {
        console.error('fetchMaterials error', e)
        this.notes = []
      }
    },
    // YouTube helpers moved from data() to methods for proper binding
    transcribeYouTube(url) {
      return new Promise(async (resolve, reject) => {
        try {
          const t0 = Date.now()
          const tick = setInterval(() => { this.proc.step3.elapsed = Math.floor((Date.now()-t0)/1000) }, 1000)
          const resp = await fetch('http://localhost:8080/api/transcribe-youtube', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ url })
          })
          clearInterval(tick)
          const data = await resp.json().catch(() => ({}))
          if (!resp.ok) {
            const msg = data && (data.message || data.details || data.error) || 'YouTube transcribe failed'
            return reject(new Error(msg))
          }
          resolve((data && data.transcription) || '')
        } catch (err) {
          reject(err)
        }
      })
    },
    // Безопасный fetch с таймаутом, чтобы UI не зависал бесконечно
    async fetchJsonWithTimeout(url, options = {}, timeoutMs = 90000) {
      const controller = new AbortController()
      const id = setTimeout(() => controller.abort(), timeoutMs)
      try {
        const resp = await fetch(url, { ...(options||{}), signal: controller.signal })
        return resp
      } finally {
        clearTimeout(id)
      }
    },
    openYoutubeModal() {
      this.youtubeUrl = ''
      this.showYoutubeModal = true
      this.$nextTick(() => {
        try { document.getElementById('yt-url')?.focus() } catch(e){}
      })
    },
    closeYoutubeModal() {
      this.showYoutubeModal = false
    },
    async startYoutubeFlow() {
      const url = this.youtubeUrl
      if (!url) return
      this.closeYoutubeModal()
      // Open processing modal and reset
      this.proc.show = true
      this.proc.error = ''
      this.proc.ready = false
      this.proc.step1.done = true
      // no upload step for YouTube
      this.proc.step2.inProgress = false
      this.proc.step2.progress = 100
      this.proc.step2.done = true
      // Step 3: transcribe from YouTube
      let transcript = ''
      try {
        transcript = await this.transcribeYouTube(url)
        this.proc.step3.done = true
        this.proc.transcript = transcript
      } catch (e) {
        console.error(e)
        this.proc.error = 'Не удалось транскрибировать видео YouTube: ' + (e?.message || e)
        return
      }
      // Step 4: AI generate as usual
      try {
        const gen = await this.generateMaterials(transcript)
        this.proc.step4.done = true
        this.proc.ready = true
        if (gen && (gen.id || gen._id)) {
          const id = gen.id || gen._id
          const payload = {
            id,
            title: gen.title || (transcript.split('\n')[0].slice(0, 40) || 'YouTube note'),
            createdAt: gen.createdAt || new Date().toISOString(),
            audioUrl: null,
            transcript: gen.transcript || transcript,
            flashcards: gen.flashcards || [],
            quiz: gen.quiz || []
          }
          sessionStorage.setItem('note:'+id, JSON.stringify(payload))
          this.proc.noteId = id
          // mark as YouTube-originated
          this.markNoteAsYt(id)
          // Immediately refresh user's notes list if authenticated
          try { if (localStorage.getItem('token')) { await this.fetchMaterials() } } catch(e){}
        } else {
          const id = String(Date.now())
          const payload = {
            id,
            title: transcript.split('\n')[0].slice(0, 40) || 'YouTube note',
            createdAt: new Date().toISOString(),
            audioUrl: null,
            transcript,
            flashcards: gen.flashcards || [],
            quiz: gen.quiz || []
          }
          sessionStorage.setItem('note:'+id, JSON.stringify(payload))
          this.proc.noteId = id
          // mark as YouTube-originated
          this.markNoteAsYt(id)
        }
        if (this.autoRedirect && this.proc.noteId) {
          setTimeout(() => { this.viewNoteNow() }, 350)
        }
      } catch (e) {
        console.error(e)
        this.proc.error = 'Ошибка генерации материалов: ' + (e?.message || e)
        return
      }
    },
    toggleSidebar(){ this.sidebarOpen = !this.sidebarOpen },
    goSettings(){ this.toast('Settings') },
    upgrade(){ this.showUpgradeModal = true },
    closeUpgradeModal(){ this.showUpgradeModal = false },
    selectPlan(kind){
      // Заглушка выбора тарифа. Здесь можно интегрировать платежи.
      this.toast && this.toast('Selected plan: ' + kind)
      this.closeUpgradeModal()
    },

    toggleTheme(){
      this.applyTheme(this.theme === 'dark' ? 'light' : 'dark')
    },
    applyTheme(t){
      this.theme = t
      document.documentElement.setAttribute('data-theme', t)
      localStorage.setItem('theme', t)
    },

    // Quick action stubs
    createBlank() { this.toast('Blank document') },
    openAudioModal() {
      this.showAudioModal = true
      this.audioUrl = null
      this.elapsedSec = 0
      this.isRecording = false
    },
    uploadDoc()   { this.toast('Document upload') },
    addYoutube()  { this.openYoutubeModal() },

    createFolder() {
      const name = prompt('Folder name')
      if (!name) return
      const id = 'f' + Math.random().toString(36).slice(2,7)
      this.folders.push({ id, name: name.trim() })
      this.activeFolderId = id
    },
    folderCount(fid){ return this.notes.filter(n => n.folderId === fid).length },
    openNote(n) {
      try {
        const key = 'note:'+n.id
        const existing = sessionStorage.getItem(key)
        let merged
        if (existing) {
          const cur = JSON.parse(existing)
          merged = {
            id: cur.id || n.id,
            title: cur.title || n.title,
            transcript: typeof cur.transcript === 'string' ? cur.transcript : (n.transcript || ''),
            createdAt: cur.createdAt || Date.now(),
            // если уже есть массивы — не затираем
            flashcards: Array.isArray(cur.flashcards) ? cur.flashcards : [],
            quiz: Array.isArray(cur.quiz) ? cur.quiz : []
          }
        } else {
          // минимальный снапшот без перетирания будущих данных
          merged = {
            id: n.id,
            title: n.title,
            transcript: n.transcript || '',
            createdAt: Date.now()
          }
        }
        sessionStorage.setItem(key, JSON.stringify(merged))
      } catch(_){ /* ignore */ }
      try { this.$router.push('/note/' + n.id) } catch(e) { this.toast('Open: ' + n.title) }
    },
    moreNote() { this.toast('More…') },

    // Audio recording logic
    async toggleRecord() {
      if (this.isRecording) {
        await this.stopRecording()
      } else {
        await this.startRecording()
      }
    },
    async startRecording() {
      try {
        // Request mic
        this._mediaStream = await navigator.mediaDevices.getUserMedia({ audio: true })
        this._mediaRecorder = new MediaRecorder(this._mediaStream)
        const chunks = []
        this._mediaRecorder.ondataavailable = e => { if (e.data && e.data.size) chunks.push(e.data) }
        this._mediaRecorder.onstop = () => {
          const blob = new Blob(chunks, { type: 'audio/webm' })
          this.audioBlob = blob
          if (this.audioUrl) URL.revokeObjectURL(this.audioUrl)
          this.audioUrl = URL.createObjectURL(blob)
          // Reset player state
          this.isPlaying = false
          this.currentTime = 0
          this.duration = 0
          this.cleanupStream()
          // Ensure metadata loads to avoid NaN/Infinity
          this.$nextTick(() => { try { this.$refs.player && this.$refs.player.load() } catch(e){} })
        }
        this._mediaRecorder.start()
        this.isRecording = true
        this.elapsedSec = 0
        this._timer = setInterval(() => { this.elapsedSec++ }, 1000)

        // Setup analyser for visualization
        const AudioCtx = window.AudioContext || window.webkitAudioContext
        this._audioCtx = new AudioCtx()
        const src = this._audioCtx.createMediaStreamSource(this._mediaStream)
        this._analyser = this._audioCtx.createAnalyser()
        this._analyser.fftSize = 64
        src.connect(this._analyser)
        const buf = new Uint8Array(this._analyser.frequencyBinCount)
        const bars = this.barHeights.length
        const draw = () => {
          this._analyser.getByteFrequencyData(buf)
          const step = Math.floor(buf.length / bars)
          for (let i = 0; i < bars; i++) {
            let sum = 0
            for (let j = 0; j < step; j++) sum += buf[i*step + j]
            const avg = sum / step
            // Map 0..255 -> 6..100
            this.$set ? this.$set(this.barHeights, i, Math.max(6, Math.min(100, Math.round(avg / 255 * 100)))) : (this.barHeights[i] = Math.max(6, Math.min(100, Math.round(avg / 255 * 100))))
          }
          this._raf = requestAnimationFrame(draw)
        }
        draw()
      } catch (err) {
        console.error('Mic error', err)
        this.toast('Microphone access denied')
      }
    },
    async stopRecording() {
      try {
        if (this._timer) { clearInterval(this._timer); this._timer = null }
        if (this._mediaRecorder && this._mediaRecorder.state !== 'inactive') {
          this._mediaRecorder.stop()
        }
      } finally {
        this.isRecording = false
      }
    },
    cleanupStream() {
      if (this._mediaStream) {
        this._mediaStream.getTracks().forEach(t => t.stop())
        this._mediaStream = null
      }
      if (this._raf) { cancelAnimationFrame(this._raf); this._raf = null }
      if (this._audioCtx) { try { this._audioCtx.close() } catch(e){} this._audioCtx = null }
      this._analyser = null
      this._mediaRecorder = null
    },
    closeAudioModal() {
      // stop playback if any
      const el = this.$refs.player
      if (el && !el.paused) { try { el.pause() } catch(e){} }
      this.isPlaying = false
      this.currentTime = 0
      if (this.isRecording) { this.stopRecording() }
      this.showAudioModal = false
    },
    fmtTime(s) {
      const val = Number.isFinite(s) && s >= 0 ? s : 0
      const mm = String(Math.floor(val / 60)).padStart(2, '0')
      const ss = String(Math.floor(val % 60)).padStart(2, '0')
      return `${mm}:${ss}`
    },
    async generateNote() {
      if (!this.audioBlob || this.isRecording) {
        this.toast('Сначала запишите аудио')
        return
      }
      // Open processing modal
      this.proc.show = true
      this.proc.error = ''
      // Step 1: creating
      this.proc.step1.done = true

      // Step 2: upload with real progress
      try {
        const transcript = await this.uploadAndTranscribe(this.audioBlob)
        this.proc.step3.done = true
        this.proc.transcript = transcript
      } catch (e) {
        this.proc.error = 'Ошибка загрузки/транскрибации'
        console.error(e)
        return
      }

      // Step 4: AI generate
      try {
        const gen = await this.generateMaterials(this.proc.transcript)
        this.proc.step4.done = true
        this.proc.ready = true
        // persist to session and allow navigate
        // If backend returned saved material with id, use it. Otherwise fallback to local payload
        if (gen && (gen.id || gen._id)) {
          const id = gen.id || gen._id
          const payload = {
            id,
            title: gen.title || (this.proc.transcript.split('\n')[0].slice(0, 40) || 'New note'),
            createdAt: gen.createdAt || new Date().toISOString(),
            audioUrl: this.audioUrl,
            transcript: gen.transcript || this.proc.transcript,
            flashcards: gen.flashcards || [],
            quiz: gen.quiz || []
          }
          sessionStorage.setItem('note:'+id, JSON.stringify(payload))
          this.proc.noteId = id
          // Immediately refresh user's notes list if authenticated
          try { if (localStorage.getItem('token')) { await this.fetchMaterials() } } catch(e){}
        } else {
          const id = String(Date.now())
          const payload = {
            id,
            title: this.proc.transcript.split('\n')[0].slice(0, 40) || 'New note',
            createdAt: new Date().toISOString(),
            audioUrl: this.audioUrl,
            transcript: this.proc.transcript,
            flashcards: gen.flashcards || [],
            quiz: gen.quiz || []
          }
          sessionStorage.setItem('note:'+id, JSON.stringify(payload))
          this.proc.noteId = id
        }
        // Optional auto-redirect
        if (this.autoRedirect && this.proc.noteId) {
          setTimeout(() => { this.viewNoteNow() }, 350)
        }
      } catch (e) {
        this.proc.error = 'Ошибка генерации материалов: ' + (e?.message || e)
        console.error(e)
        return
      }
    },
    uploadAndTranscribe(blob) {
      return new Promise((resolve, reject) => {
        this.proc.step2.progress = 0
        this.proc.step2.inProgress = true
        const form = new FormData()
        const file = new File([blob], 'recording.webm', { type: blob.type || 'audio/webm' })
        form.append('audio', file)
        const xhr = new XMLHttpRequest()
        xhr.open('POST', 'http://localhost:8080/api/transcribe')
        xhr.upload.onprogress = (ev) => {
          if (ev.lengthComputable) {
            this.proc.step2.progress = Math.round((ev.loaded / ev.total) * 100)
          }
        }
        const t0 = Date.now()
        const tick = setInterval(() => { this.proc.step3.elapsed = Math.floor((Date.now()-t0)/1000) }, 1000)
        xhr.onreadystatechange = () => {
          if (xhr.readyState === 4) {
            clearInterval(tick)
            if (xhr.status >= 200 && xhr.status < 300) {
              try {
                const data = JSON.parse(xhr.responseText)
                this.proc.step2.inProgress = false
                this.proc.step2.done = true
                resolve(data.transcription || '')
              } catch (err) {
                reject(err)
              }
            } else {
              reject(new Error('Upload/transcribe failed'))
            }
          }
        }
        xhr.onerror = () => { reject(new Error('Network error')) }
        xhr.send(form)
      })
    },
    async generateMaterials(transcript) {
      const t0 = Date.now()
      const tick = setInterval(() => { this.proc.step4.elapsed = Math.floor((Date.now()-t0)/1000) }, 1000)
      try {
        const token = localStorage.getItem('token')
        // Основной путь: generate-and-save с JWT и таймаутом
        if (token) {
          try {
            const resp = await this.fetchJsonWithTimeout('http://localhost:8080/api/generate-and-save', {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + token
              },
              body: JSON.stringify({ transcript })
            }, 90000)

            if (resp.ok) {
              const data = await resp.json().catch(() => ({}))
              if (data && data.material && (data.material.id || data.material._id)) {
                return {
                  id: data.material.id || data.material._id,
                  transcript: data.material.transcript,
                  flashcards: data.material.flashcards || [],
                  quiz: data.material.quiz || []
                }
              }
              return {
                id: data.id || data._id,
                transcript: data.transcript,
                flashcards: data.flashcards || [],
                quiz: data.quiz || []
              }
            } else if (resp.status !== 401 && resp.status !== 403) {
              const errBody = await resp.text().catch(() => '')
              throw new Error(`Ошибка сервера (${resp.status}): ${errBody || 'generate-and-save failed'}`)
            }
            // Если 401/403 — тихий переход к локальной генерации
          } catch (e) {
            if (e && e.name === 'AbortError') {
              throw new Error('Таймаут генерации (generate-and-save)')
            }
            // Прочие ошибки — пробуем фоллбэк ниже
            if (String(e?.message||'').includes('Таймаут')) throw e
          }
        }

        // Фоллбэк без сохранения: /api/generate с меньшим таймаутом
        const resp2 = await this.fetchJsonWithTimeout('http://localhost:8080/api/generate', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ transcript })
        }, 60000)
        if (!resp2.ok) {
          const errBody2 = await resp2.text().catch(() => '')
          throw new Error(`Ошибка сервера (${resp2.status}): ${errBody2 || 'generate failed'}`)
        }
        const data2 = await resp2.json().catch(() => ({}))
        return { flashcards: data2.flashcards || [], quiz: data2.quiz || [] }
      } finally {
        clearInterval(tick)
      }
    },
    viewNoteNow() {
      if (!this.proc.noteId) return
      const id = this.proc.noteId
      this.proc.show = false
      this.showAudioModal = false
      this.$router.push({ path: `/note/${id}` })
    },
    // Player controls
    togglePlay() {
      const el = this.$refs.player
      if (!el) return
      if (this.isPlaying) {
        el.pause()
        this.isPlaying = false
      } else {
        el.play()
        this.isPlaying = true
      }
    },
    onTimeUpdate(e) {
      const el = e.target
      const ct = el.currentTime
      this.currentTime = Number.isFinite(ct) && ct >= 0 ? ct : 0
    },
    onLoadedMeta(e) {
      const el = e.target
      const dur = el.duration
      const ct = el.currentTime
      this.duration = Number.isFinite(dur) && dur > 0 ? dur : 0
      this.currentTime = Number.isFinite(ct) && ct >= 0 ? ct : 0
    },
    onEnded() {
      this.isPlaying = false
      this.currentTime = this.duration
    },
    seek(ev) {
      const el = this.$refs.player
      if (!el || !this.duration) return
      const rect = ev.currentTarget.getBoundingClientRect()
      const ratio = Math.min(1, Math.max(0, (ev.clientX - rect.left) / rect.width))
      el.currentTime = ratio * this.duration
      this.currentTime = el.currentTime
    },

    logout() {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      this.$router.push('/')
    },
    // Settings helpers (simple toggle can be wired later in UI)
    setAutoRedirect(val) {
      this.autoRedirect = !!val
      localStorage.setItem('autoRedirect', this.autoRedirect ? 'true' : 'false')
    },
    toast(msg) { console.log('[Dashboard]', msg) },
    onFilePicked(e) {
      try {
        const file = e?.target?.files?.[0]
        if (!file) return
        // Визуально сбросить текущий плеер
        this.stopPlayback && this.stopPlayback()
        // Сохранить как общий источник аудио
        this.audioBlob = file
        try { if (this.audioUrl) URL.revokeObjectURL(this.audioUrl) } catch(_){}
        this.audioUrl = URL.createObjectURL(file)
        this.toast && this.toast('Аудиофайл добавлен')
      } catch (err) {
        console.error('onFilePicked error', err)
        this.toast && this.toast('Не удалось прочитать файл')
      } finally {
        // очистим value, чтобы можно было выбрать тот же файл повторно
        try { if (this.$refs.fileInput) this.$refs.fileInput.value = '' } catch(_){}
      }
    }
  }
}
</script>

<style scoped>
:root { --bg:#0f0f12; --panel:#171722; --line:rgba(255,255,255,.12); --muted:#b0b0b0; --text:#fff; --accent:#7c3aed; }
.dash-wrap { display:grid; grid-template-columns: 72px 1fr; min-height:100vh; background:rgba(10,10,10,.9); color:var(--text); transition: grid-template-columns .35s cubic-bezier(.22,.61,.36,1); }
.dash-wrap.drawer-open { grid-template-columns: 300px 1fr; }

/* Topbar */
.topbar { position:sticky; top:0; z-index:30; grid-column: 1 / -1; display:flex; align-items:center; justify-content:space-between; padding:14px 16px; background:rgba(10,10,10,.85); backdrop-filter: blur(10px); border-bottom:1px solid var(--line); }
.brand { display:flex; align-items:center; gap:10px; font-weight:800; font-size:22px; }
.brand-text { letter-spacing:.5px; }
.topbar-actions { display:flex; align-items:center; gap:8px; }
.theme-toggle { height:36px; min-width:44px; padding:0 10px; border-radius:10px; border:1px solid var(--line); background:rgba(255,255,255,.04); color:var(--text); display:flex; align-items:center; justify-content:center; cursor:pointer; }
.theme-toggle.big { width:56px; height:44px; border-radius:12px; border:1px solid var(--line); background:rgba(255,255,255,.04); color:var(--text); display:flex; align-items:center; justify-content:center; gap:10px; margin:4px auto 0; transition: width .35s cubic-bezier(.22,.61,.36,1), background .2s ease; }
/* Profile/avatar */
.profile { position:relative; }
.avatar { height:36px; width:36px; border-radius:50%; border:1px solid var(--line); background: linear-gradient(135deg, #111827, #1f2937); color:#fff; display:grid; place-items:center; font-weight:800; letter-spacing:.5px; box-shadow: 0 4px 16px rgba(0,0,0,.25); }
.avatar:hover { box-shadow: 0 6px 22px rgba(0,0,0,.35); }
.profile-menu { position:absolute; top:44px; right:0; width:280px; border-radius:14px; border:1px solid var(--line); background: rgba(15,15,20,.98); backdrop-filter: blur(10px); box-shadow: 0 14px 44px rgba(0,0,0,.5), 0 0 0 1px rgba(124,58,237,.12) inset; padding:10px; z-index:50; }
.pm-head { display:grid; grid-template-columns:40px 1fr; gap:10px; align-items:center; padding:6px 6px 10px; border-bottom:1px solid rgba(255,255,255,.06); margin-bottom:8px; }
.pm-avatar { width:40px; height:40px; border-radius:50%; display:grid; place-items:center; font-weight:900; background: linear-gradient(135deg, #7C3AED55, #00D4FF33); box-shadow: inset 0 0 0 2px rgba(124,58,237,.45); }
.pm-name { font-weight:800; }
.pm-mail { color: var(--muted); font-size:12px; }
.pm-item { width:100%; display:flex; align-items:center; gap:10px; height:40px; border-radius:10px; border:1px solid var(--line); background: rgba(255,255,255,.04); color: var(--text); cursor:pointer; padding:0 10px; }
.pm-item + .pm-item { margin-top:8px; }
.pm-item:hover { box-shadow: 0 0 0 1px rgba(0,212,255,.18) inset; }
.pm-item.danger { border-color: rgba(239,68,68,.4); background: rgba(239,68,68,.12); color:#fecaca; }
.pm-item.danger:hover { box-shadow: 0 0 0 1px rgba(239,68,68,.35) inset; }

/* Sidebar user card */
.user-card { display:grid; grid-template-columns:40px 1fr auto; gap:10px; align-items:center; padding:10px; margin:8px; border-radius:14px; border:1px solid var(--line); background: rgba(255,255,255,.035); }
.uc-left { width:40px; height:40px; border-radius:50%; display:grid; place-items:center; font-weight:900; background: linear-gradient(135deg, #7C3AED55, #00D4FF33); box-shadow: inset 0 0 0 2px rgba(124,58,237,.45); }
.uc-name { font-weight:800; line-height:1.1; }
.uc-mail { color:var(--muted); font-size:12px; }
.uc-gear { height:36px; width:36px; border-radius:10px; border:1px solid var(--line); background:rgba(255,255,255,.04); color:var(--text); cursor:pointer; }
.theme-toggle.big:hover { background: rgba(124,58,237,.12); border-color: rgba(124,58,237,.35); }
.theme-toggle.big:focus-visible { outline: 2px solid rgba(124,58,237,.6); outline-offset: 2px; }
.burger { height:36px; width:44px; border-radius:10px; border:1px solid var(--line); background:rgba(255,255,255,.04); color:var(--text); display:flex; align-items:center; justify-content:center; }
.burger-box { position:relative; width:18px; height:14px; }
.burger-lines, .burger-lines:before, .burger-lines:after { content:""; position:absolute; left:0; right:0; height:2px; background:#fff; border-radius:2px; }
.burger-lines{ top:6px; }
.burger-lines:before{ top:-6px; }
.burger-lines:after{ top:6px; transform: translateY(6px); }

/* Expanding side pane */
.navpane { position:sticky; top:56px; height:calc(100vh - 56px); padding:12px 8px; border-right:1px solid var(--line); background:#0f0f12; width:72px; transition: width .35s cubic-bezier(.22,.61,.36,1); display:flex; flex-direction:column; gap:18px; align-items:center; }
.navpane.open { width:300px; align-items:stretch; }
.pane-head { display:flex; align-items:center; justify-content:space-between; padding:0 8px; }
.pane-logo { font-weight:800; font-size:22px; letter-spacing:.5px; }
.collapse-btn { height:36px; width:36px; border-radius:10px; border:1px solid var(--line); background:rgba(255,255,255,.04); color:var(--text); display:flex; align-items:center; justify-content:center; }
.collapse-btn svg { width:18px; height:18px; }
.rail-burger { height:54px; width:54px; border-radius:14px; background:rgba(255,255,255,.02); border:2px solid #1f3e83; box-shadow: inset 0 0 0 2px rgba(91,131,255,.25); display:grid; place-items:center; color:#fff; align-self:center; }
.menu { display:flex; flex-direction:column; gap:16px; width:100%; }
.menu-item { display:grid; grid-template-columns:28px 1fr; align-items:center; column-gap:12px; height:56px; padding:0 14px; border-radius:14px; border:1px solid var(--line); background:#1b1b23; color:#fff; text-align:left; width:56px; margin:0 auto; transition: width .35s cubic-bezier(.22,.61,.36,1), background .2s ease; }
.menu-item.active { box-shadow:0 0 0 2px rgba(124,58,237,.25) inset; }
.navpane.open .menu-item { width:100%; }
.mi-ico { width:28px; height:28px; display:grid; place-items:center; }
.mi-ico svg { width:22px; height:22px; fill:#fff; }
.mi-text { white-space:nowrap; overflow:hidden; opacity:0; width:0; will-change: opacity, width; transition: opacity .25s ease .1s, width .35s cubic-bezier(.22,.61,.36,1); font-size:18px; font-weight:700; }
.navpane.open .mi-text { opacity:1; width:auto; }
.upgrade { display:flex; align-items:center; justify-content:center; gap:10px; height:56px; border-radius:16px; border:1px solid rgba(124,58,237,.35); background:#6227e9; color:#fff; font-weight:800; width:56px; margin:8px auto 0; transition: width .35s cubic-bezier(.22,.61,.36,1); }
.navpane.open .upgrade { width:100%; }
.upg-text { white-space:nowrap; overflow:hidden; opacity:0; width:0; transition: opacity .25s ease .1s, width .35s cubic-bezier(.22,.61,.36,1); }
.navpane.open .upg-text { opacity:1; width:auto; }
.spark { font-size:18px; }

/* Folders */
.folders { padding:6px 8px 10px; display:flex; flex-direction:column; gap:10px; }
.folders-head { display:flex; align-items:center; justify-content:space-between; padding:0 8px; }
.folders-title { font-weight:800; letter-spacing:.4px; }
.folders-add { height:28px; width:28px; border-radius:8px; border:1px solid var(--line); background:rgba(255,255,255,.05); color:#fff; }
.folder-list { display:flex; flex-direction:column; gap:8px; }
.folder-item { display:grid; grid-template-columns:22px 1fr auto; align-items:center; gap:10px; height:40px; padding:0 10px; border-radius:10px; border:1px solid var(--line); background:#1b1b23; color:#fff; text-align:left; }
.folder-item.active, .folder-item:hover { box-shadow:0 0 0 2px rgba(124,58,237,.25) inset; }
.fi-ico { width:22px; height:22px; display:grid; place-items:center; }
.fi-text { white-space:nowrap; overflow:hidden; text-overflow:ellipsis; }
.fi-count { color:var(--muted); font-size:12px; }

/* Upgrade appear from bottom */
.upg-enter-from, .upg-leave-to { opacity:0; transform: translateY(12px); }
.upg-enter-active, .upg-leave-active { transition: all .35s cubic-bezier(.22,.61,.36,1); }
.upg-enter-to, .upg-leave-from { opacity:1; transform: translateY(0); }

/* Main */
.main { padding:20px 16px 60px; }
@media (min-width: 980px){
  .topbar{ display:none; }
  .navpane{ top:0; height:100vh; }
}
@media (max-width: 979px){
  .main{ grid-column: 1 / -1; }
}

.page-header { margin:8px 0 18px; }
.h1 { font-size:40px; font-weight:800; margin:0 0 8px; }
.sub { color:var(--muted); margin:0; }

/* Page header row & search */
.ph-row { display:flex; align-items:center; justify-content:space-between; gap:12px; }
.search-wrap { position:relative; display:flex; align-items:center; gap:10px; height:40px; padding:0 12px 0 36px; border-radius:12px; border:1px solid var(--line); background:rgba(255,255,255,.035); min-width:260px; }
.search-wrap svg { position:absolute; left:12px; opacity:.7; }
.search { background:transparent; border:none; outline:none; color:var(--text); width:220px; }
.slash { margin-left:auto; color:var(--muted); background:rgba(255,255,255,.06); border:1px solid var(--line); border-bottom-width:2px; padding:2px 6px; border-radius:6px; font-size:12px; }

/* Quick actions */
.quick { display:grid; grid-template-columns: repeat(4, minmax(200px, 1fr)); gap:16px; margin:18px 0 22px; }
.qa { display:grid; grid-auto-flow:column; grid-template-columns:48px 1fr 16px; align-items:center; gap:14px; padding:16px; border-radius:14px; border:1px solid var(--line); background:rgba(255,255,255,.04); cursor:pointer; transition:transform .15s ease, box-shadow .15s ease; }
.qa:hover { transform: translateY(-1px); box-shadow: 0 6px 18px rgba(0,0,0,.25), 0 0 0 1px rgba(0,212,255,.15) inset; }
.qa-ico { width:48px; height:48px; border-radius:12px; display:flex; align-items:center; justify-content:center; font-size:20px; box-shadow:0 0 0 1px rgba(255,255,255,.08) inset; color:#fff; }
.c-purple { background: linear-gradient(135deg,#7c3aed33,#7c3aed22); }
.c-violet { background: linear-gradient(135deg,#6d28d933,#6d28d922); }
.c-blue   { background: linear-gradient(135deg,#00d4ff33,#00d4ff22); }
.c-red    { background: linear-gradient(135deg,#ef444433,#ef444422); }

/* SVGs inside icons */
.qa-ico svg { width:22px; height:22px; display:block; }
.qa-ico.doc svg { width:28px; height:28px; }
.qa-ico path { fill:#fff; }
.qa-ico rect { fill:#fff; }
.qa-ico.yt rect { fill:#FF0000; }
.qa-ico.yt polygon { fill:#FFFFFF; }
.qa-ico.doc .doc-paper { fill:#5b21b6; }
.qa-ico.doc .doc-fold { fill:#8b5cf6; }
.qa-ico.doc .pill { fill: rgba(255,255,255,.2); }
.qa-ico .doc-text { fill:#fff; font-weight:900; font-size:8.8px; font-family: ui-sans-serif, -apple-system, Segoe UI, Roboto, Helvetica, Arial, "Apple Color Emoji", "Segoe UI Emoji"; letter-spacing:.3px; dominant-baseline: middle; }
.qa-ico.yt .yt-svg { filter: drop-shadow(0 0 0 rgba(0,0,0,0)); }

.qa-title { font-weight:700; }
.qa-desc  { color:var(--muted); font-size:12px; }
.qa-arrow { color:var(--muted); font-size:22px; justify-self:end; }

/* Tabs */
.tabs-row { display:flex; align-items:center; justify-content:space-between; gap:12px; margin:6px 0 14px; }
.tabs { display:flex; gap:8px; }
.tab { height:36px; padding:0 14px; border-radius:10px; border:1px solid var(--line); background:rgba(255,255,255,.04); color:var(--text); cursor:pointer; }
.tab.active, .tab:hover { box-shadow:0 0 0 2px rgba(0,212,255,.15) inset; }
.btn.outline { height:36px; padding:0 12px; border-radius:10px; border:1px solid var(--line); background:rgba(255,255,255,.04); color:var(--text); display:inline-flex; align-items:center; gap:8px; cursor:pointer; }
.btn-ico { font-size:16px; }

/* Notes */
.notes-list { display:flex; flex-direction:column; gap:12px; margin-top:8px; }
.note { display:grid; grid-template-columns:44px 1fr auto; align-items:center; gap:14px; padding:14px; border-radius:12px; border:1px solid var(--line); background:rgba(255,255,255,.035); }
.note:hover { box-shadow:0 0 0 1px rgba(0,212,255,.15) inset; }
.note-ico { width:44px; height:44px; border-radius:12px; display:flex; align-items:center; justify-content:center; font-size:18px; background:rgba(255,255,255,.05); }
.note-ico.audio { background: rgba(255,255,255,.05); }
.note-ico svg { width:20px; height:20px; fill: currentColor; }
.note-title { font-weight:700; }
.note-meta { color:var(--muted); font-size:12px; }
.note-more { height:32px; width:32px; border-radius:8px; border:1px solid var(--line); background:rgba(255,255,255,.04); color:var(--text); cursor:pointer; }
.empty { color:var(--muted); text-align:center; padding:40px 0; }

@media (max-width: 980px){ .quick{ grid-template-columns: 1fr 1fr; } }
@media (max-width: 640px){ .quick{ grid-template-columns: 1fr; } .h1{font-size:32px;} }

/* Light theme overrides */
[data-theme='light'] .dash-wrap { background:#f6f7fb; }
[data-theme='light'] .topbar { background: rgba(255,255,255,.9); }
[data-theme='light'] .navpane { background:#ffffff; }
[data-theme='light'] .menu-item { background:#ffffff; }
[data-theme='light'] .upgrade { background:#4f46e5; }
[data-theme='light'] .search-wrap { background: rgba(0,0,0,.03); }
[data-theme='light'] .qa { background: rgba(0,0,0,.03); }
[data-theme='light'] .tab { background: rgba(0,0,0,.03); }
[data-theme='light'] .btn.outline { background: rgba(0,0,0,.03); }
[data-theme='light'] .folders-add,
[data-theme='light'] .note-more,
[data-theme='light'] .collapse-btn,
[data-theme='light'] .burger { background: rgba(0,0,0,.04); }
[data-theme='light'] .theme-toggle { background: rgba(0,0,0,.04); }
[data-theme='light'] .theme-toggle.big:hover { background: rgba(0,0,0,.06); border-color: rgba(0,0,0,.15); }
[data-theme='light'] .folder-item { background:#ffffff; }
[data-theme='light'] .note { background: rgba(0,0,0,.03); }
[data-theme='light'] .note-ico { background: rgba(0,0,0,.05); }

/* --- Audio/Generic modals --- */
.modal-wrap { position: fixed; inset: 0; display:flex; align-items:center; justify-content:center; z-index: 100; padding: 16px; animation: fadeIn .18s ease; }
.modal-backdrop { position:absolute; inset:0; background: rgba(0,0,0,.6); backdrop-filter: blur(6px); }
.modal { position:relative; width:min(720px, 92vw); background: var(--panel); color: var(--text); border:1px solid var(--line); border-radius:18px; box-shadow: 0 24px 80px rgba(0,0,0,.55); overflow:hidden; transform-origin:center; animation: popIn .22s cubic-bezier(.22,.61,.36,1); }
.modal-head { display:flex; align-items:center; justify-content:space-between; padding:18px 20px; border-bottom:1px solid var(--line); backdrop-filter: saturate(120%); }
.modal-title { font-weight:800; font-size:20px; letter-spacing:.2px; }
.modal-close { height:36px; width:36px; border-radius:10px; border:1px solid var(--line); background:rgba(255,255,255,.06); color:var(--text); display:grid; place-items:center; transition: background .15s ease, box-shadow .15s ease, transform .08s ease; }
.modal-close:hover { background:rgba(255,255,255,.1); box-shadow: 0 0 0 3px rgba(124,58,237,.18) inset; }
.modal-close:focus-visible { outline:none; box-shadow: 0 0 0 3px rgba(0,212,255,.35) inset; }
.modal-body { padding:20px; display:flex; flex-direction:column; gap:16px; max-height: min(70vh, 640px); overflow:auto; }
.modal-body .note { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.rec-center { display:grid; place-items:center; padding:24px 0; }
.mic-btn { width:84px; height:84px; border-radius:50%; display:grid; place-items:center; border:2px solid rgba(255,255,255,.2); background: rgba(255,255,255,.05); color: var(--text); cursor:pointer; box-shadow: 0 0 0 6px rgba(124,58,237,.12); }
.mic-btn.recording { background:#ef4444; box-shadow: 0 0 0 6px rgba(239,68,68,.2); }
.rec-bar { display:flex; align-items:center; gap:12px; background: rgba(255,255,255,.06); border:1px solid var(--line); padding:10px 12px; border-radius:12px; width: max(260px, 60%); margin: 0 auto; }
.pill { display:inline-flex; align-items:center; gap:6px; padding:2px 8px; border-radius:999px; background:#111; color:#fff; font-size:12px; border:1px solid var(--line); }
.rec-actions { display:flex; align-items:center; justify-content:center; gap:12px; }
.btn-ghost { height:36px; padding:0 12px; border-radius:10px; border:1px solid var(--line); background:transparent; color:var(--text); }
.btn.primary { height:36px; padding:0 14px; border-radius:10px; border:1px solid rgba(124,58,237,.5); background: linear-gradient(90deg,#7C3AED,#A78BFA); color:#fff; box-shadow: 0 6px 18px rgba(124,58,237,.35); display:inline-flex; align-items:center; }
.btn.primary:disabled { opacity:.6; cursor:not-allowed; box-shadow:none }
.btn-ghost.disabled { opacity:.5; pointer-events:none }
.audio-preview { margin-top:8px; }

[data-theme='light'] .modal { background: #fff; box-shadow: 0 24px 64px rgba(0,0,0,.18); }

/* Processing modal extras */
.modal-backdrop { position: absolute; inset:0; background: rgba(0,0,0,.6); backdrop-filter: blur(6px); z-index:0; }
.steps { display:flex; flex-direction:column; gap:10px; margin-top:8px; }
.step-item { display:grid; grid-template-columns:28px 1fr auto; align-items:center; gap:12px; padding:10px 12px; border-radius:12px; border:1px solid var(--line); background: rgba(255,255,255,.035); }
.step-item .left { display:grid; place-items:center; }
.step-item .left.glow .num { box-shadow: 0 0 0 6px rgba(0,212,255,.12); border-color: rgba(0,212,255,.45); }

/* Premium modal */
.premium-intro { text-align:center; color:var(--muted); margin-top:2px; }
.pricing { display:grid; grid-template-columns: 1fr 1fr; gap:12px; margin-top:12px; }
.plan { position:relative; display:flex; flex-direction:column; gap:10px; padding:14px; border-radius:14px; border:1px solid var(--line); background: rgba(255,255,255,.035); align-items:center; text-align:center; }
.plan.best { background: linear-gradient(180deg, rgba(124,58,237,.12), rgba(124,58,237,.06)); border-color: rgba(124,58,237,.4); box-shadow: 0 10px 26px rgba(124,58,237,.2); }
.plan .badge { position:absolute; top:10px; right:10px; font-size:10px; padding:3px 8px; border-radius:999px; background:#6227e9; color:#fff; border:1px solid rgba(255,255,255,.25); }
.plan-name { font-weight:800; letter-spacing:.3px; }
.plan-price { font-weight:900; font-size:20px; }
.plan-price .n { font-size:34px; line-height:1; margin-right:6px; }
.plan-price .per { color: var(--muted); font-size:12px; margin-left:6px; }
.plan .note { color: var(--muted); font-size:12px; margin-top:-4px; }
.fine { color: var(--muted); font-size:12px; margin-top:12px; text-align:center; }
.num { width:22px; height:22px; border-radius:999px; display:grid; place-items:center; background: rgba(255,255,255,.06); border:1px solid rgba(255,255,255,.18); font-size:12px; font-weight:800; }
.stitle { font-weight:700; }
.sdesc { color: var(--muted); font-size:12px; margin-top:4px; }
.right { display:flex; align-items:center; gap:8px; }
.badge { height:24px; padding:0 8px; border-radius:999px; border:1px solid var(--line); background: rgba(255,255,255,.04); color: var(--text); font-size:12px; display:inline-flex; align-items:center; }
.badge.ok { background: rgba(34,197,94,.15); border-color: rgba(34,197,94,.35); color:#10B981; }
.spinner { width:16px; height:16px; border-radius:50%; border:2px solid rgba(255,255,255,.2); border-top-color:#7C3AED; animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg) } }
@keyframes fadeIn { from { opacity: 0 } to { opacity: 1 } }
@keyframes popIn { from { opacity:.6; transform: scale(.98) } to { opacity:1; transform: scale(1) } }
/* Ensure modal sits above backdrop */
.modal { position: relative; z-index: 1; }
.progress-line { height:8px; background: rgba(255,255,255,.08); border-radius:999px; border:1px solid var(--line); overflow:hidden; margin-top:6px; }
.progress-fill { height:100%; background: linear-gradient(90deg, #00D4FF, #7C3AED); box-shadow: 0 0 12px rgba(124,58,237,.35); transition: width .2s ease; }

[data-theme='light'] .player { background: rgba(0,0,0,.03); }
[data-theme='light'] .pp-btn { background: rgba(0,0,0,.04); }
[data-theme='light'] .btn.primary { border-color: rgba(0,0,0,.15); box-shadow: 0 6px 18px rgba(124,58,237,.25); }
/* visualizer */
.viz { display:flex; align-items:flex-end; justify-content:center; gap:6px; height:76px; margin: 0 auto 8px; }
.bar { width:6px; background: linear-gradient(180deg, #A78BFA, #7C3AED); border-radius:6px; height:10%; transition: height .08s ease; box-shadow: 0 2px 10px rgba(124,58,237,.25); }
.mic-btn.recording + .viz .bar { background: linear-gradient(180deg, #FCA5A5, #EF4444); box-shadow: 0 2px 10px rgba(239,68,68,.25); }
[data-theme='light'] .bar { box-shadow:none }

/* custom player */
.player { display:flex; align-items:center; gap:12px; background: rgba(255,255,255,.06); border:1px solid var(--line); padding:10px 12px; border-radius:12px; }
.pp-btn { height:36px; width:36px; display:grid; place-items:center; border-radius:10px; border:1px solid var(--line); background:rgba(255,255,255,.06); color:var(--text); }
.progress { flex:1; cursor:pointer; }
.progress-track { height:8px; background: rgba(255,255,255,.08); border-radius:999px; position:relative; overflow:hidden; border:1px solid var(--line); }
.progress-fill { position:absolute; inset:0 auto 0 0; width:0%; background: linear-gradient(90deg, #A78BFA, #7C3AED); border-radius:999px; box-shadow: 0 0 12px rgba(124,58,237,.35); }
.time { min-width:98px; text-align:right; color: var(--muted); font-variant-numeric: tabular-nums; }

[data-theme='light'] .player { background: rgba(0,0,0,.03); }
[data-theme='light'] .pp-btn { background: rgba(0,0,0,.04); }
[data-theme='light'] .btn.primary { border-color: rgba(0,0,0,.15); box-shadow: 0 6px 18px rgba(124,58,237,.25); }
</style>