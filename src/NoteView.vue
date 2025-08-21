<template>
  <div class="note-wrap">
    <header class="nv-topbar">
      <button class="back" @click="$router.push('/dashboard')" aria-label="Back">← Back</button>
      <div class="title">{{ note?.title || 'Note' }}</div>
      <div style="width:72px"></div>
    </header>

    <section class="meta" v-if="note">
      <div class="created">Created: {{ new Date(note.createdAt).toLocaleString() }}</div>
    </section>

    <section class="tabs">
      <button class="tab" :class="{active: active==='note'}" @click="active='note'">Note</button>
      <button class="tab" :class="{active: active==='quiz'}" @click="active='quiz'">Quiz</button>
      <button class="tab" :class="{active: active==='flash'}" @click="active='flash'">Flashcards</button>
      <button class="tab" :class="{active: active==='transcript'}" @click="active='transcript'">Transcript</button>
    </section>

    <!-- Loading skeletons -->
    <section v-if="loading" class="content">
      <div class="skeleton title"></div>
      <div class="skeleton line"></div>
      <div class="skeleton line w60"></div>
      <div class="skeleton block"></div>
    </section>

    <section class="content" v-else-if="note">
      <!-- NOTE SUMMARY (simple: first lines of transcript) -->
      <transition name="fade-slide" mode="out-in">
        <div v-if="active==='note'" key="note">
        <div class="section">
          <h3>Summary</h3>
          <p>{{ summary }}</p>
        </div>

        <div class="section" v-if="note.audioUrl">
          <h3>Audio</h3>
          <audio :src="note.audioUrl" controls preload="metadata"></audio>
        </div>
        </div>
      </transition>

      <!-- QUIZ -->
      <transition name="fade-slide" mode="out-in">
        <div v-if="active==='quiz'" key="quiz">
        <h3>Quiz</h3>
        <div v-if="(note.quiz || []).length===0" class="empty">No quiz items</div>
        <ol class="quiz">
          <li v-for="(q, i) in note.quiz" :key="i" class="quiz-item">
            <div class="q">{{ q.question || q.q || ('Question #' + (i+1)) }}</div>
            <ul class="opts" v-if="q.options && q.options.length">
              <li v-for="(opt, j) in q.options" :key="j">{{ opt }}</li>
            </ul>
            <details class="ans" v-if="q.answer || q.correct">
              <summary>Answer</summary>
              <div>{{ q.answer || q.correct }}</div>
            </details>
          </li>
        </ol>
        </div>
      </transition>

      <!-- FLASHCARDS -->
      <transition name="fade-slide" mode="out-in">
        <div v-if="active==='flash'" key="flash">
        <h3>Flashcards</h3>
        <div v-if="(note.flashcards || []).length===0" class="empty">No flashcards</div>
        <div class="cards">
          <div v-for="(c, i) in note.flashcards" :key="i" class="card">
            <div class="front">{{ c.front || c.term || ('Card #' + (i+1)) }}</div>
            <div class="back">{{ c.back || c.definition || c.meaning }}</div>
          </div>
        </div>
        </div>
      </transition>

      <!-- TRANSCRIPT -->
      <transition name="fade-slide" mode="out-in">
        <div v-if="active==='transcript'" key="transcript">
        <h3>Transcript</h3>
        <pre class="transcript">{{ note.transcript }}</pre>
        </div>
      </transition>
    </section>

    <div v-else class="empty big">Note data not found</div>
  </div>
</template>

<script>
export default {
  name: 'NoteView',
  data() {
    return {
      active: 'note',
      note: null,
      loading: true
    }
  },
  computed: {
    summary() {
      if (!this.note || !this.note.transcript) return '—'
      const t = this.note.transcript.trim()
      // naive summary: first 2 sentences or 300 chars
      const parts = t.split(/[.!?]\s+/).slice(0, 2).join('. ')
      return (parts || t).slice(0, 300)
    }
  },
  created() {
    const id = this.$route.params.id
    // 1) Try sessionStorage (local fallback)
    try {
      const raw = sessionStorage.getItem('note:'+id)
      if (raw) this.note = JSON.parse(raw)
    } catch (e) {
      console.error('Failed to parse local note', e)
    }
    // Если id не валиден как Mongo ObjectID (24 hex), не дергаем API
    const isHex24 = /^[a-fA-F0-9]{24}$/.test(id)
    if (!isHex24) {
      this.loading = false
      return
    }
    // Показываем снапшот сразу, но обновляем с бэкенда при валидном id
    this.loading = !this.note
    this.fetchMaterialById(id).catch(err => {
      console.warn('Fetch material failed', err)
      // Если данных нет вообще — возвращаемся
      if (!this.note) this.$router.replace('/dashboard')
      else this.loading = false
    })
  },
  methods: {
    async fetchMaterialById(id) {
      const token = localStorage.getItem('token')
      if (!token) throw new Error('No JWT')
      const resp = await fetch(`http://localhost:8080/api/materials/${id}`, {
        headers: { 'Authorization': 'Bearer ' + token }
      })
      if (!resp.ok) throw new Error('Failed to load material')
      const data = await resp.json()
      const m = data && data.material ? data.material : data
      // Normalize to local note shape
      const transcript = m?.transcript || ''
      const firstLine = (transcript.split('\n')[0] || '').trim()
      const title = m?.title || firstLine || 'Note'
      const created = m?.createdAt || m?.created_at || Date.now()
      this.note = {
        id: m?.id || m?._id || id,
        title,
        createdAt: created,
        audioUrl: m?.audioUrl, // may be undefined
        transcript,
        flashcards: m?.flashcards || [],
        quiz: m?.quiz || []
      }
      // Cache to session for faster reopen
      sessionStorage.setItem('note:'+this.note.id, JSON.stringify(this.note))
      this.loading = false
    }
  }
}
</script>

<style scoped>
.note-wrap { max-width: 960px; margin: 0 auto; padding: 20px 16px 48px; color:#fff; }
.nv-topbar { display:flex; align-items:center; justify-content:space-between; gap:10px; padding:8px 0 16px; }
.back { border:1px solid rgba(255,255,255,.15); background:rgba(255,255,255,.06); color:#fff; height:36px; padding:0 12px; border-radius:10px; cursor:pointer; transition: background .2s ease, border-color .2s ease; }
.back:hover { background: rgba(124,58,237,.12); border-color: rgba(124,58,237,.35); }
.title { font-size:20px; font-weight:700; }
.meta { color:#b0b0b0; margin: 8px 0 14px; }
.tabs { display:flex; gap:10px; border-bottom:1px solid rgba(255,255,255,.12); margin-bottom:16px; position:sticky; top:0; z-index:2; backdrop-filter: blur(6px); }
.tab { height:36px; padding:0 14px; border-radius:10px 10px 0 0; background:transparent; color:#b0b0b0; border:1px solid transparent; cursor:pointer; transition: color .2s ease, background .2s ease, box-shadow .2s ease; }
.tab:hover { color:#fff; }
.tab.active { color:#fff; background: rgba(255,255,255,.06); border-color: rgba(255,255,255,.12); border-bottom-color: transparent; box-shadow: 0 6px 18px rgba(124,58,237,.25) inset; }
.content { background: rgba(255,255,255,.03); border:1px solid rgba(255,255,255,.12); border-radius:12px; padding:16px; box-shadow: 0 12px 30px rgba(0,0,0,.25); backdrop-filter: blur(4px); }
.section { margin-bottom:18px; }
.empty { color:#b0b0b0; padding:8px 0; }
.empty.big { text-align:center; padding:40px 0; }
.quiz { padding-left: 18px; }
.quiz-item { margin-bottom: 12px; }
.opts { margin: 6px 0 0 16px; }
.cards { display:grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap:12px; }
.card { border:1px solid rgba(255,255,255,.12); border-radius:12px; padding:12px; background:rgba(255,255,255,.04); transition: transform .18s ease, box-shadow .18s ease, border-color .18s ease; }
.card:hover { transform: translateY(-2px); box-shadow: 0 10px 24px rgba(124,58,237,.25); border-color: rgba(124,58,237,.35); }
.front { font-weight:700; margin-bottom:6px; }
.transcript { white-space: pre-wrap; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; background:rgba(0,0,0,.25); border-radius:10px; padding:12px; border:1px solid rgba(255,255,255,.08); }

/* Fade-slide transitions */
.fade-slide-enter-active, .fade-slide-leave-active { transition: opacity .18s ease, transform .18s ease; }
.fade-slide-enter-from, .fade-slide-leave-to { opacity: 0; transform: translateY(6px); }

/* Skeleton placeholders */
.skeleton { position: relative; overflow: hidden; background: rgba(255,255,255,.06); border-radius: 8px; margin: 10px 0; border:1px solid rgba(255,255,255,.08); }
.skeleton:after { content:""; position:absolute; inset:0; background: linear-gradient(90deg, rgba(255,255,255,0) 0%, rgba(255,255,255,.15) 50%, rgba(255,255,255,0) 100%); transform: translateX(-100%); animation: shimmer 1.2s infinite; }
.skeleton.title { height: 22px; width: 40%; }
.skeleton.line { height: 12px; width: 90%; }
.skeleton.line.w60 { width: 60%; }
.skeleton.block { height: 120px; width: 100%; border-radius: 12px; }
@keyframes shimmer { to { transform: translateX(100%);} }
</style>
