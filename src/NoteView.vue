<template>
  <div class="note-wrap">
    <header class="nv-topbar">
      <button class="back" @click="$router.push('/dashboard')" aria-label="Back">← Back</button>
      <h1 class="nv-title" v-if="note">{{ note.title }}</h1>
      <div class="spacer"></div>
      <nav class="tabs">
        <button class="tab" :class="{active: active==='note'}" @click="active='note'">
          <svg class="ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <rect x="4" y="4" width="16" height="16" rx="2"/>
            <path d="M9 3h6v4H9z"/>
            <path d="M8 12h8M8 16h8"/>
          </svg>
          <span>Note</span>
        </button>
        <button class="tab" :class="{active: active==='quiz'}" @click="active='quiz'">
          <svg class="ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <rect x="3" y="3" width="7" height="7" rx="1.5"/>
            <circle cx="17" cy="7" r="3.5"/>
            <path d="M4 17l5-3 5 3 5-3v7H4z"/>
          </svg>
          <span>Quiz</span>
        </button>
        <button class="tab" :class="{active: active==='flash'}" @click="active='flash'">
          <svg class="ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <rect x="5" y="6" width="12" height="14" rx="2"/>
            <path d="M9 4h10v12"/>
          </svg>
          <span>Flashcards</span>
        </button>
        <button class="tab" :class="{active: active==='transcript'}" @click="active='transcript'">
          <svg class="ico" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <rect x="3" y="5" width="18" height="14" rx="2"/>
            <rect x="6" y="9" width="4" height="4" rx="1"/>
            <path d="M12.5 9.5h5M12.5 13.5h5"/>
          </svg>
          <span>Transcript</span>
        </button>
      </nav>
    </header>

    <section class="meta" v-if="note">
      <div class="created">Created: {{ new Date(note.createdAt).toLocaleString() }}</div>
    </section>

    <!-- Loading skeletons -->
    <section v-if="loading" class="content">
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
        <div v-if="active==='quiz'" key="quiz" class="quiz-wrap">
          <div class="quiz-head">
            <div class="title-sm">Quiz</div>
            <div class="actions">
              <button class="ghost" @click="resetQuiz">Reset</button>
              <button class="ghost" @click="shuffleQuiz">Shuffle questions</button>
            </div>
          </div>

          <div v-if="(quizList || []).length===0" class="empty">No quiz items</div>

          <div v-else>
            <div class="progress">
              <div class="bar" :style="{width: progressPercent + '%'}"></div>
            </div>

            <div v-if="!quizUi.completed" class="q-card">
              <div class="q-meta">Question {{ quizUi.index+1 }} of {{ quizList.length }}</div>
              <div class="q-text">{{ currentQuestion.question || currentQuestion.q || ('Question #' + (quizUi.index+1)) }}</div>

              <ul class="opts" v-if="currentQuestion.options && currentQuestion.options.length">
                <li v-for="(opt, j) in currentQuestion.options"
                    :key="j"
                    class="opt"
                    :class="{
                      selected: isMSQ ? (quizUi.selectedMulti.has(j)) : (quizUi.selected===j)
                    }"
                    @click="isMSQ ? toggleOption(j) : selectOption(j)">
                  <span class="dot"></span>
                  <span class="label">{{ opt }}</span>
                </li>
              </ul>

              <div class="q-actions">
                <button class="ghost" :disabled="quizUi.index===0" @click="prev">Prev</button>
                <div class="spacer"></div>
                <button class="primary" v-if="quizUi.index < quizList.length-1" :disabled="!canSubmitCurrent" @click="commitAnswerThenNext">Next →</button>
                <button class="primary" v-else :disabled="!canSubmitCurrent" @click="submitQuiz">Finish</button>
              </div>
            </div>

            <div v-else class="result-card">
              <div class="result-title">Quiz result</div>
              <div class="result-row">
                <div class="score">
                  <div class="percent">{{ scorePercent }}%</div>
                  <div class="muted">{{ results.correct }} / {{ quizList.length }} correct</div>
                </div>
                <div class="spacer"></div>
                <button class="ghost" @click="resetQuiz">Try again</button>
              </div>

              <div v-if="results.wrongItems.length" class="wrong-list">
                <div class="wrong-title">Incorrect questions</div>
                <ol>
                  <li v-for="(w, wi) in results.wrongItems" :key="wi" class="wrong-item">
                    <div class="w-q">{{ w.q.question || w.q.q || ('Question #' + (w.index+1)) }}</div>
                    <div class="w-rows">
                      <div class="chip red">Your answer: {{ w.your }}</div>
                      <div class="chip green">Correct: {{ w.correct }}</div>
                    </div>
                    <div v-if="w.q.rationale" class="muted">{{ w.q.rationale }}</div>
                  </li>
                </ol>
              </div>
            </div>
          </div>
        </div>
      </transition>

      <!-- FLASHCARDS -->
      <transition name="fade-slide" mode="out-in">
        <div v-if="active==='flash'" key="flash">
        <h3>Flashcards</h3>
        <div v-if="(note.flashcards || []).length===0" class="empty">No flashcards</div>
        <div v-else class="study">
          <div class="study-controls">
            <button class="btn ghost" @click="shuffleCards" :disabled="count<=1">Shuffle</button>
            <div class="progress">Card {{ study.index + 1 }} of {{ count }}</div>
            <button class="btn ghost" @click="resetOrder" :disabled="count<=1">Reset</button>
          </div>

          <div
            class="card study-card"
            :class="{ flipped: flipped.has(currentIdx) }"
            @click="toggleFlip(currentIdx)"
            role="button"
            :aria-pressed="flipped.has(currentIdx) ? 'true' : 'false'"
            tabindex="0"
            @keydown.enter.prevent="toggleFlip(currentIdx)"
            @keydown.space.prevent="toggleFlip(currentIdx)"
          >
            <div class="card-inner">
              <div class="card-face front">
                <div class="fc-title">{{ currentCard.front || currentCard.term || ('Card #' + (study.index+1)) }}</div>
                <div v-if="currentCard.example" class="fc-example">{{ currentCard.example }}</div>
              </div>
              <div class="card-face back">
                <div class="fc-back-title">Definition</div>
                <div class="fc-text">{{ currentCard.back || currentCard.definition || currentCard.meaning }}</div>
              </div>
            </div>
          </div>

          <div class="study-nav">
            <button class="btn" @click="prevCard" :disabled="count===0">Previous</button>
            <span class="hint">Press to flip</span>
            <button class="btn" @click="nextCard" :disabled="count===0">Next</button>
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
      loading: true,
      // Индексы перевёрнутых карточек
      flipped: new Set(),
      study: { order: [], index: 0 },
      quizUi: {
        index: 0,
        // для MCQ: selected = number | null; для MSQ используем selectedMulti: Set
        selected: null,
        selectedMulti: new Set(),
        answers: [], // [{index, value}] — value: number | number[] depending on type
        completed: false,
      },
      quizList: []
    }
  },
  computed: {
    // Flashcards study view
    count() {
      const list = (this.note && this.note.flashcards) ? this.note.flashcards : []
      return Array.isArray(list) ? list.length : 0
    },
    currentIdx() {
      if (!this.study.order.length) return 0
      const idx = this.study.order[this.study.index] ?? 0
      return Math.min(Math.max(0, idx), Math.max(0, this.count - 1))
    },
    currentCard() {
      const list = (this.note && this.note.flashcards) ? this.note.flashcards : []
      return list[this.currentIdx] || {}
    },
    // Note summary
    summary() {
      if (!this.note || !this.note.transcript) return '—'
      const t = this.note.transcript.trim()
      const parts = t.split(/[.!?]\s+/).slice(0, 2).join('. ')
      return (parts || t).slice(0, 300)
    },
    // Quiz computed
    currentQuestion() {
      return (this.quizList[this.quizUi.index] || {})
    },
    isMSQ() {
      const q = this.currentQuestion
      return Array.isArray(q?.correct) && q.correct.length > 1
    },
    canSubmitCurrent() {
      return this.isMSQ ? (this.quizUi.selectedMulti.size > 0) : (this.quizUi.selected !== null)
    },
    progressPercent() {
      if (!this.quizList.length) return 0
      const answered = this.quizUi.answers.length
      return Math.round((answered / this.quizList.length) * 100)
    },
    results() {
      let correct = 0
      const wrongItems = []
      this.quizUi.answers.forEach((entry) => {
        const q = this.quizList[entry.index]
        if (!q) return
        const truth = Array.isArray(q.correct) && q.correct.length
          ? q.correct
          : (typeof q.answer === 'number' ? [q.answer] : [q.options ? q.options.indexOf(q.answer) : -1])
        const compareAs = (val) => Array.isArray(val) ? [...val].sort().join(',') : String(val)
        const normalize = (arr) => arr.filter(x => x !== -1 && x !== undefined && x !== null)
        let your = entry.value
        let isCorrect = false
        if (Array.isArray(your)) {
          const yn = normalize(your)
          const tn = normalize(truth.map((t) => (typeof t === 'number' ? t : q.options ? q.options.indexOf(t) : -1)))
          isCorrect = compareAs(yn) === compareAs(tn)
        } else {
          const tIdx = (Array.isArray(truth) ? truth[0] : truth)
          const tIndex = (typeof tIdx === 'number') ? tIdx : (q.options ? q.options.indexOf(tIdx) : -1)
          isCorrect = Number(your) === Number(tIndex)
        }
        if (isCorrect) correct++
        else {
          const yourText = Array.isArray(entry.value)
            ? entry.value.map(i => q.options?.[i]).filter(Boolean).join(' | ')
            : q.options?.[entry.value]
          const corrText = (Array.isArray(truth) ? truth : [truth])
            .map(v => (typeof v === 'number') ? q.options?.[v] : v)
            .filter(Boolean).join(' | ')
          wrongItems.push({ index: entry.index, q, your: yourText || '—', correct: corrText || '—' })
        }
      })
      return { correct, wrongItems }
    },
    scorePercent() {
      if (!this.quizList.length) return 0
      return Math.round((this.results.correct / this.quizList.length) * 100)
    }
  },
  watch: {
    active(val) {
      if (val === 'flash') this.initStudyOrder()
    },
    note: {
      handler() { this.initStudyOrder() },
      deep: true
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
      // Инициализируем квиз из локальной заметки, если она есть
      if (this.note && Array.isArray(this.note.quiz)) {
        this.quizList = this.normalizeQuiz(this.note.quiz || [])
        this.resetQuiz()
      }
      this.loading = false
      return
    }
    // Показываем снапшот сразу, но обновляем с бэкенда при валидном id
    this.loading = !this.note
    // Если уже есть снапшот – подхватим квиз немедленно (лучший UX)
    if (this.note && Array.isArray(this.note.quiz)) {
      this.quizList = this.normalizeQuiz(this.note.quiz || [])
      this.resetQuiz()
    }
    this.fetchMaterialById(id).catch(err => {
      console.warn('Fetch material failed', err)
      // Если данных нет вообще — возвращаемся
      if (!this.note) this.$router.replace('/dashboard')
      else this.loading = false
    })
  },
  methods: {
    // Study view controls
    toggleFlip(i) {
      const s = new Set(this.flipped)
      if (s.has(i)) s.delete(i)
      else s.add(i)
      this.flipped = s
    },
    initStudyOrder() {
      const n = this.count
      this.study.order = Array.from({ length: n }, (_, i) => i)
      this.study.index = 0
      this.flipped = new Set()
    },
    shuffleCards() {
      const arr = Array.from({ length: this.count }, (_, i) => i)
      for (let i = arr.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1))
        ;[arr[i], arr[j]] = [arr[j], arr[i]]
      }
      this.study.order = arr
      this.study.index = 0
      this.flipped = new Set()
    },
    resetOrder() {
      this.initStudyOrder()
    },
    nextCard() {
      if (this.count === 0) return
      this.study.index = (this.study.index + 1) % this.count
      this.flipped = new Set()
    },
    prevCard() {
      if (this.count === 0) return
      this.study.index = (this.study.index - 1 + this.count) % this.count
      this.flipped = new Set()
    },
    normalizeQuiz(list) {
      try {
        const out = []
        ;(list || []).forEach((q) => {
          if (!q || typeof q !== 'object') return
          const type = (q.type || q.Type || '').toUpperCase()
          let options = Array.isArray(q.options) ? q.options.slice() : []
          let answer = q.answer
          let correct = Array.isArray(q.correct) ? q.correct.slice() : undefined

          // Нормализуем True/False
          if (type === 'TF' && options.length === 0) {
            options = ['True', 'False']
          }

          // Оставляем только поддерживаемые типы с вариантами ответов
          const hasVariants = Array.isArray(options) && options.length > 0
          const hasAnswer = (answer !== undefined && answer !== null && answer !== '') || (Array.isArray(correct) && correct.length > 0)
          if (!hasVariants || !hasAnswer) return

          // Возвращаем унифицированный объект
          out.push({
            question: q.question || q.q || '',
            options,
            answer,
            correct,
            rationale: q.rationale || ''
          })
        })
        return out
      } catch(_){ return Array.isArray(list) ? list : [] }
    },
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
      // init quiz
      this.quizList = this.normalizeQuiz(this.note.quiz || [])
      this.resetQuiz()
    },
    resetQuiz() {
      this.quizUi.index = 0
      this.quizUi.selected = null
      this.quizUi.selectedMulti = new Set()
      this.quizUi.answers = []
      this.quizUi.completed = false
      this.loadSelectionForIndex()
    },
    shuffleQuiz() {
      const a = this.quizList.slice()
      for (let i = a.length - 1; i > 0; i--) {
        const j = Math.floor(Math.random() * (i + 1))
        ;[a[i], a[j]] = [a[j], a[i]]
      }
      this.quizList = a
      this.resetQuiz()
    },
    loadSelectionForIndex() {
      const found = this.quizUi.answers.find(x => x.index === this.quizUi.index)
      if (!found) {
        this.quizUi.selected = null
        this.quizUi.selectedMulti = new Set()
        return
      }
      if (Array.isArray(found.value)) {
        this.quizUi.selected = null
        this.quizUi.selectedMulti = new Set(found.value)
      } else {
        this.quizUi.selected = found.value
        this.quizUi.selectedMulti = new Set()
      }
    },
    selectOption(j) {
      this.quizUi.selected = j
    },
    toggleOption(j) {
      const set = new Set(this.quizUi.selectedMulti)
      if (set.has(j)) set.delete(j)
      else set.add(j)
      this.quizUi.selectedMulti = set
    },
    prev() {
      if (this.quizUi.index === 0) return
      this.quizUi.index--
      this.loadSelectionForIndex()
    },
    commitAnswerThenNext() {
      const idx = this.quizUi.index
      const isMSQ = Array.isArray(this.currentQuestion?.correct) && this.currentQuestion.correct.length > 1
      const value = isMSQ ? Array.from(this.quizUi.selectedMulti) : this.quizUi.selected
      const existingIdx = this.quizUi.answers.findIndex(x => x.index === idx)
      if (existingIdx >= 0) this.quizUi.answers.splice(existingIdx, 1, { index: idx, value })
      else this.quizUi.answers.push({ index: idx, value })
      if (this.quizUi.index < this.quizList.length - 1) {
        this.quizUi.index++
        this.loadSelectionForIndex()
      }
    },
    submitQuiz() {
      // фиксируем ответ последнего вопроса
      this.commitAnswerThenNext()
      this.quizUi.completed = true
    }
  }
}
</script>

<style scoped>
.note-wrap { position: relative; min-height: 100vh; width: 100vw; max-width: none; margin: 0; padding: 20px 16px 48px; color: var(--text); background: #000; border: none; border-radius: 0; box-shadow: none; }
.nv-topbar { position: sticky; top: 0; z-index: 20; display:flex; align-items:center; gap:12px; padding:10px 12px; background:#000; border-bottom:1px solid rgba(255,255,255,.08); }
.nv-topbar .spacer { flex: 1; }
.back { border:1px solid rgba(255,255,255,.12); background:#101012; color:var(--text); height:40px; padding:0 14px; border-radius:12px; cursor:pointer; font-weight:700; transition: background .2s ease, border-color .2s ease, box-shadow .2s ease, transform .12s ease; box-shadow: 0 8px 20px rgba(124,58,237,.18); }
.back:hover { background: rgba(124,58,237,.12); border-color: rgba(124,58,237,.35); box-shadow: 0 8px 22px rgba(124,58,237,.25); }
.title { font-size:20px; font-weight:700; }
.titlebar { padding: 6px 0 2px; }
.nv-title { font-size: 28px; font-weight: 900; letter-spacing: -.2px; line-height: 1.2; margin-bottom: 2px; background: linear-gradient(90deg, #fff, #d9d4ff 50%, #c4b5fd); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; text-shadow: 0 2px 30px rgba(124,58,237,.18); }
.meta { color: var(--muted); margin: 8px 0 14px; }
.tabs { display:flex; gap:6px; margin:0; justify-content:flex-end; align-items:center; background: rgba(255,255,255,.92); border:1px solid rgba(0,0,0,.08); border-radius:999px; padding:6px; }
.tab { height:40px; padding:0 16px; border-radius:12px; border:none; background:transparent; color:#111; cursor:pointer; font-weight:800; letter-spacing:.2px; display:flex; align-items:center; gap:8px; transition: background .18s ease, box-shadow .18s ease, transform .12s ease, color .18s ease; }
.tab .ico { width:18px; height:18px; opacity:.9; }
.tab:hover { background: rgba(0,0,0,.06); transform: translateY(-1px); }
.tab.active { background:#111; color:#fff; box-shadow: 0 6px 16px rgba(0,0,0,.35); }
.tab.active .ico { opacity:1; }
.content { background: #0b0b0c; border:1px solid rgba(255,255,255,.08); border-radius:12px; padding:16px; box-shadow: 0 14px 34px rgba(0,0,0,.45), 0 0 0 1px rgba(255,255,255,.03) inset; }
.section { margin-bottom:18px; }
.section h3 { font-size: 16px; font-weight: 800; text-transform: uppercase; letter-spacing: .6px; color: #c4b5fd; margin-bottom: 8px; }
.section p { color: #e6e6f0; opacity: .95; line-height: 1.6; }
.empty { color: var(--muted); padding:8px 0; }
.empty.big { text-align:center; padding:40px 0; }
.quiz-wrap { padding: 6px 0 0; }
.quiz-head { display:flex; align-items:center; gap:12px; }
.quiz-head .title-sm { font-weight:700; font-size:18px; }
.quiz-head .actions { margin-left:auto; display:flex; gap:8px; }
.ghost { height:40px; padding:0 16px; border-radius:12px; border:1px solid rgba(255,255,255,.12); background:#101012; color:var(--text); cursor:pointer; font-weight:700; transition: background .18s ease, border-color .18s ease, box-shadow .18s ease, transform .12s ease; }
.ghost:hover { background: #121214; border-color: rgba(124,58,237,.35); box-shadow: 0 6px 18px rgba(124,58,237,.2); }
.primary { height:46px; padding:0 18px; border-radius:14px; border:1px solid rgba(124,58,237,.55); background: linear-gradient(90deg,#7C3AED,#A78BFA); color:#fff; box-shadow: 0 10px 26px rgba(124,58,237,.40); cursor:pointer; font-weight:800; letter-spacing:.2px; transition: transform .12s ease; }
.primary:hover { transform: translateY(-1px); }
.primary:disabled { opacity:.6; cursor:not-allowed; box-shadow:none }

.progress { height:6px; background: rgba(255,255,255,.06); border-radius:999px; overflow:hidden; margin:12px 0 16px; border:1px solid rgba(255,255,255,.08); }
.progress .bar { height:100%; background: linear-gradient(90deg, #a855f7, #7c3aed); width:0; transition: width .25s ease; }

.q-card { border:1px solid rgba(255,255,255,.08); background: #0b0b0c; border-radius:16px; padding:16px; box-shadow: 0 12px 30px rgba(0,0,0,.45), 0 0 0 1px rgba(255,255,255,.03) inset; }
.q-meta { color: var(--muted); margin-bottom:8px; }
.q-text { font-size:18px; font-weight:700; margin-bottom:10px; }
.opts { list-style:none; margin: 10px 0 12px; padding:0; display:flex; flex-direction:column; gap:8px; }
.opt { display:flex; align-items:center; gap:10px; border:1px solid rgba(255,255,255,.08); background: #0f0f10; border-radius:12px; padding:10px 12px; cursor:pointer; transition: background .15s ease, border-color .15s ease, transform .12s ease, box-shadow .15s ease; }
.opt:hover { background: #121214; border-color: rgba(255,255,255,.18); transform: translateY(-1px); }
.opt.selected { border-color: rgba(34,197,94,.6); box-shadow: 0 0 0 2px rgba(34,197,94,.25) inset; background: rgba(34,197,94,.08); }
.opt .dot { width:10px; height:10px; border-radius:50%; background: rgba(255,255,255,.35); }
.opt.selected .dot { background: #22c55e; }
.q-actions { display:flex; align-items:center; gap:8px; }
.q-actions .spacer { flex:1; }

.result-card { border:1px solid rgba(255,255,255,.08); background: #0b0b0c; border-radius:16px; padding:16px; box-shadow: 0 12px 30px rgba(0,0,0,.45), 0 0 0 1px rgba(255,255,255,.03) inset; }
.result-title { font-weight:800; font-size:20px; margin-bottom:10px; }
.result-row { display:flex; align-items:center; gap:12px; margin-bottom:8px; }
.score .percent { font-size:34px; font-weight:900; letter-spacing:-0.5px; }
.muted { color: var(--muted); }
.wrong-list { margin-top:12px; }
.wrong-title { font-weight:700; margin-bottom:8px; }
.wrong-item { border:1px dashed rgba(255,255,255,.14); border-radius:12px; padding:10px 12px; margin-bottom:8px; background: rgba(255,255,255,.03); }
.w-q { font-weight:700; margin-bottom:6px; }
.w-rows { display:flex; flex-wrap:wrap; gap:8px; margin-bottom:6px; }
.chip { border-radius:999px; padding:6px 10px; font-size:12px; border:1px solid transparent; }
.chip.red { background: rgba(239,68,68,.12); color:#fecaca; border-color: rgba(239,68,68,.35); }
.chip.green { background: rgba(34,197,94,.12); color:#bbf7d0; border-color: rgba(34,197,94,.35); }
.cards { display:grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap:12px; }
.card { border:1px solid var(--line); border-radius:12px; padding:12px; background:rgba(255,255,255,.04); transition: transform .18s ease, box-shadow .18s ease, border-color .18s ease; }
.card:hover { transform: translateY(-2px); box-shadow: 0 10px 24px rgba(124,58,237,.25); border-color: rgba(124,58,237,.35); }
.front { font-weight:700; margin-bottom:6px; }
.transcript { white-space: pre-wrap; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace; background:rgba(0,0,0,.25); border-radius:12px; padding:14px; border:1px solid var(--line); box-shadow: 0 6px 18px rgba(0,0,0,.25) inset; }
audio { width: 100%; height: 40px; border-radius: 10px; background: rgba(255,255,255,.06); border: 1px solid var(--line); }

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

<style scoped>
/* Study View layout */
.study { display:flex; flex-direction: column; align-items:center; gap:18px; margin-top: 12px; }
.study-controls { display:flex; align-items:center; gap:12px; opacity:.95; }
.study-controls .progress { font-weight: 700; letter-spacing:.2px; opacity:.9; }
.study-nav { display:flex; align-items:center; gap:16px; }
.hint { opacity:.7; font-size: 12px; }
.btn { padding:8px 12px; border-radius:10px; border:1px solid rgba(255,255,255,.12); background: rgba(255,255,255,.06); cursor:pointer; }
.btn.ghost { background: transparent; }
.btn:disabled { opacity:.5; cursor:not-allowed; }

.study-card { width: min(680px, 96%); height: 300px; }
@media (max-width: 600px) { .study-card { height: 240px; } }
</style>

<style scoped>
/* Fancy flashcards */
.cards.fancy { perspective: 1200px; gap: 16px; }
.card { position: relative; height: 180px; border: none; background: transparent; padding: 0; }
.card .card-inner {
  position: relative;
  width: 100%;
  height: 100%;
  transform-style: preserve-3d;
  transition: transform .6s cubic-bezier(.2,.8,.2,1), filter .3s ease, box-shadow .3s ease;
}
.card.flipped .card-inner { transform: rotateY(180deg); }

.card-face {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 14px;
  border-radius: 16px;
  backface-visibility: hidden;
  background: radial-gradient(1200px 400px at 10% -20%, rgba(168,85,247,.28), rgba(124,58,237,.14) 40%, rgba(255,255,255,.05) 70%), rgba(255,255,255,.04);
  border: 1px solid rgba(168,85,247,.28);
  box-shadow:
    0 10px 30px rgba(0,0,0,.35),
    0 0 0 1px rgba(255,255,255,.05) inset,
    0 0 80px rgba(168,85,247,.12) inset;
  overflow: hidden;
}
.card-face::after {
  /* Shine */
  content: "";
  position: absolute; inset: -40%;
  background: conic-gradient(from 180deg, rgba(168,85,247,.0), rgba(168,85,247,.25), rgba(124,58,237,.0));
  filter: blur(22px);
  transform: translate3d(0,0,0);
  opacity: .35;
  pointer-events: none;
}
.card .front { transform: rotateY(0deg); }
.card .back { transform: rotateY(180deg); }

.fc-title { font-weight: 800; font-size: 18px; letter-spacing: .2px; }
.fc-example { margin-top: 6px; color: #c6b8f8; opacity: .9; font-size: 13px; }
.fc-back-title { font-weight: 800; text-transform: uppercase; font-size: 12px; letter-spacing: .6px; color:#c4b5fd; margin-bottom: 6px; }
.fc-text { font-size: 15px; line-height: 1.4; }

/* Float animation */
.card { animation: float 6s ease-in-out infinite; }
.card:nth-child(2n) { animation-delay: .6s }
.card:nth-child(3n) { animation-delay: 1.2s }
@keyframes float { 0%,100%{ transform: translateY(0) } 50%{ transform: translateY(-4px) } }

/* Reduced motion */
@media (prefers-reduced-motion: reduce) {
  .card, .card .card-inner { transition: none !important; animation: none !important; }
}
</style>
