<template>
  <div class="note-wrap">
    <header class="nv-topbar">
      <button class="back" @click="$router.push('/dashboard')" aria-label="Back">← Back</button>
      <div style="flex:1"></div>
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
      loading: true,
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
    summary() {
      if (!this.note || !this.note.transcript) return '—'
      const t = this.note.transcript.trim()
      // naive summary: first 2 sentences or 300 chars
      const parts = t.split(/[.!?]\s+/).slice(0, 2).join('. ')
      return (parts || t).slice(0, 300)
    },
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
      // вычисляем результат по сохранённым ответам
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
.note-wrap { max-width: 960px; margin: 0 auto; padding: 20px 16px 48px; color:#fff; background: #0b0b0b; }
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
.quiz-wrap { padding: 6px 0 0; }
.quiz-head { display:flex; align-items:center; gap:12px; }
.quiz-head .title-sm { font-weight:700; font-size:18px; }
.quiz-head .actions { margin-left:auto; display:flex; gap:8px; }
.ghost { height:32px; padding:0 12px; border-radius:10px; border:1px solid rgba(255,255,255,.14); background:transparent; color:#ddd; cursor:pointer; }
.ghost:hover { border-color: rgba(168,85,247,.55); color:#fff; background: rgba(168,85,247,.12); }
.primary { height:36px; padding:0 14px; border-radius:12px; border:1px solid rgba(168,85,247,.8); background: linear-gradient(180deg, rgba(168,85,247,.25), rgba(168,85,247,.18)); color:#fff; cursor:pointer; box-shadow: 0 6px 18px rgba(168,85,247,.25); }
.primary:disabled { opacity:.5; cursor:not-allowed; }

.progress { height:6px; background: rgba(255,255,255,.06); border-radius:999px; overflow:hidden; margin:12px 0 16px; border:1px solid rgba(255,255,255,.08); }
.progress .bar { height:100%; background: linear-gradient(90deg, #a855f7, #7c3aed); width:0; transition: width .25s ease; }

.q-card { border:1px solid rgba(255,255,255,.12); background: #0f0f10; border-radius:16px; padding:16px; box-shadow: 0 10px 28px rgba(0,0,0,.35); }
.q-meta { color:#b0b0b0; margin-bottom:8px; }
.q-text { font-size:18px; font-weight:700; margin-bottom:10px; }
.opts { list-style:none; margin: 10px 0 12px; padding:0; display:flex; flex-direction:column; gap:8px; }
.opt { display:flex; align-items:center; gap:10px; border:1px solid rgba(255,255,255,.12); background: rgba(255,255,255,.04); border-radius:12px; padding:10px 12px; cursor:pointer; transition: background .15s ease, border-color .15s ease, transform .12s ease; }
.opt:hover { background: rgba(255,255,255,.065); border-color: rgba(255,255,255,.18); transform: translateY(-1px); }
.opt.selected { border-color: rgba(34,197,94,.6); box-shadow: 0 0 0 2px rgba(34,197,94,.25) inset; background: rgba(34,197,94,.08); }
.opt .dot { width:10px; height:10px; border-radius:50%; background: rgba(255,255,255,.35); }
.opt.selected .dot { background: #22c55e; }
.q-actions { display:flex; align-items:center; gap:8px; }
.q-actions .spacer { flex:1; }

.result-card { border:1px solid rgba(255,255,255,.12); background:#0f0f10; border-radius:16px; padding:16px; box-shadow: 0 10px 28px rgba(0,0,0,.35); }
.result-title { font-weight:800; font-size:20px; margin-bottom:10px; }
.result-row { display:flex; align-items:center; gap:12px; margin-bottom:8px; }
.score .percent { font-size:34px; font-weight:900; letter-spacing:-0.5px; }
.muted { color:#9aa0a6; }
.wrong-list { margin-top:12px; }
.wrong-title { font-weight:700; margin-bottom:8px; }
.wrong-item { border:1px dashed rgba(255,255,255,.14); border-radius:12px; padding:10px 12px; margin-bottom:8px; background: rgba(255,255,255,.03); }
.w-q { font-weight:700; margin-bottom:6px; }
.w-rows { display:flex; flex-wrap:wrap; gap:8px; margin-bottom:6px; }
.chip { border-radius:999px; padding:6px 10px; font-size:12px; border:1px solid transparent; }
.chip.red { background: rgba(239,68,68,.12); color:#fecaca; border-color: rgba(239,68,68,.35); }
.chip.green { background: rgba(34,197,94,.12); color:#bbf7d0; border-color: rgba(34,197,94,.35); }
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
