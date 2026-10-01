<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import * as api from './services/api'
import StatisticsTable from './components/StatisticsTable.vue'
import { useTrainer } from './stores/trainer'
const store = useTrainer()
const page = ref('library')
const reviewKind = ref('wrong')
const reviewItems = computed(() => store.review.filter(q => reviewKind.value === 'wrong' ? q.wrong > 0 : q.favorite))
const isFavorite = computed(() => store.review.find(q => q.question.id === item.value?.question.id)?.favorite ?? false)
const preview = ref<api.domain.Preview | null>(null)
const browsing = ref<api.domain.Question[]>([])
const index = ref(0)
const selected = ref('')
const confirmFinish = ref(false)
const item = computed(() => store.session?.items[index.value])
const answered = computed(() => store.session?.items.filter(i => store.session?.mode === 'exam' ? !!i.selected : i.scored).length ?? 0)
const correct = computed(() => store.session?.statistics.total.correct ?? 0)
function navigate(n: number) { index.value = n; selected.value = item.value?.selected ?? ''; confirmFinish.value = false }
async function openSession(s: api.domain.Session) { store.session = s; page.value = 'quiz'; navigate(0) }
function start(id: string, mode = 'study') { store.run(async () => { await openSession(await api.StartQuiz(id, mode)); await store.refresh() }) }
function resume(id: string) { store.run(async () => openSession(await api.GetSession(id))) }
function saveChoice(flagged = item.value?.flagged ?? false) { store.run(async () => { if (!store.session || !item.value) return; try { store.session = await api.SaveChoice(store.session.id, item.value.question.id, selected.value, flagged) } catch(e) { selected.value = item.value.selected; throw e } }) }
function submit() { store.run(async () => { if (!store.session || !item.value) return; store.session = await api.SubmitStudy(store.session.id, item.value.question.id, selected.value); await store.refresh() }) }
function finish() { store.run(async () => { if (!store.session) return; store.session = await api.CompleteQuiz(store.session.id); confirmFinish.value = false; selected.value = item.value?.selected ?? ''; await store.refresh() }) }
function choose() { store.run(async () => { preview.value = null; const p = await api.ChooseImport(); if (p.token) preview.value = p }) }
function importSet() { store.run(async () => { if (!preview.value) return; await api.ConfirmImport(preview.value.token); preview.value = null; store.notice = '题集导入成功'; await store.refresh() }) }
function favorite(id: string, value: boolean) { store.run(async () => { await api.SetFavorite(id, value); await store.refresh() }) }
function reviewStart() { store.run(async () => { await openSession(await api.StartReview(reviewKind.value)); await store.refresh() }) }
function viewSet(id: string) { store.run(async () => { browsing.value = await api.SetQuestions(id) }) }
const date = (value: string) => value ? new Date(value).toLocaleString('zh-CN') : '进行中'
onMounted(() => store.run(store.refresh))
</script>
<template>
  <div class="app-shell">
    <header><div><p class="eyebrow">本地学习 · 每日积累</p><h1>CISSP Quiz Trainer</h1></div><span class="badge">v0.1 · 离线使用</span></header>
    <nav aria-label="主要导航"><button :class="{active: page === 'library'}" @click="page = 'library'">题库与导入</button><button :class="{active: page === 'history'}" @click="page = 'history'">练习记录</button><button :class="{active: page === 'review'}" @click="page = 'review'">错题与收藏</button><button :class="{active: page === 'stats'}" @click="page = 'stats'">统计与备份</button><button v-if="store.session" :class="{active: page === 'quiz'}" @click="page = 'quiz'">当前答题</button></nav>
    <p v-if="store.error" class="message error" role="alert">{{ store.error }}</p><p v-if="store.notice" class="message" role="status">{{ store.notice }}</p><p v-if="store.busy" role="status">正在处理，请稍候…</p>
    <main :aria-busy="store.busy">
      <section v-if="page === 'library'">
        <div class="section-head"><div><h2>我的题库</h2><p>导入标准 JSON 题集，按原题顺序开始学习。</p></div><button class="primary" :disabled="store.busy" @click="choose">导入 JSON 题集</button></div>
        <article v-if="preview" class="panel"><h3>导入预览：{{ preview.set.title }}</h3><p>{{ preview.set.description }} · {{ preview.questions.length }} 题</p><p>ID：{{ preview.set.id }} · {{ preview.status === 'ready' ? '可导入' : preview.status === 'duplicate' ? '重复题集' : '存在冲突' }}</p><ul v-if="preview.messages.length"><li v-for="message in preview.messages" :key="message">{{ message }}</li></ul><details><summary>查看题目内容（不显示答案）</summary><ol><li v-for="q in preview.questions" :key="q.id">{{ q.question }} <span class="muted">Domain {{ q.domain }} / {{ q.type }}</span></li></ol></details><div class="actions"><button class="primary" :disabled="store.busy || preview.status !== 'ready'" @click="importSet">确认导入全部题目</button><button @click="preview = null">取消</button></div></article>
        <p v-if="!store.sets.length" class="empty">还没有题集。点击“导入 JSON 题集”，选择随包提供的 demo-questions.json 开始。</p>
        <div class="grid"><article v-for="s in store.sets" :key="s.id" class="panel"><span class="badge">{{ s.count }} 题</span><h3>{{ s.title }}</h3><p>{{ s.description || '暂无简介' }}</p><div class="actions"><button class="primary" :disabled="store.busy" @click="start(s.id)">开始学习</button><button :disabled="store.busy" @click="start(s.id, 'exam')">开始考试</button><button :disabled="store.busy" @click="viewSet(s.id)">查看题目</button></div></article></div>
        <article v-if="browsing.length" class="panel"><div class="section-head"><h3>题目一览</h3><button @click="browsing = []">收起</button></div><ol><li v-for="q in browsing" :key="q.id"><p>{{ q.question }}</p><p class="muted">{{ q.id }} · Domain {{ q.domain }} · {{ q.type }} · {{ q.difficulty }}</p></li></ol></article>
      </section>
      <section v-if="page === 'history'"><h2>练习记录</h2><p>未完成练习会自动保存，重新启动后可继续。</p><p v-if="!store.sessions.length" class="empty">尚无练习记录。从题库开始一次学习吧。</p><div class="table-wrap" v-else><table><thead><tr><th>题集</th><th>开始时间</th><th>状态</th><th>进度 / 结果</th><th>操作</th></tr></thead><tbody><tr v-for="s in store.sessions" :key="s.id"><td>{{ s.title }}<br><span class="muted">{{ s.mode === 'exam' ? '考试' : '学习' }}</span></td><td>{{ date(s.startedAt) }}</td><td>{{ s.status === 'completed' ? '已完成' : '进行中' }}</td><td>{{ s.status === 'completed' ? '正确 ' + s.correct + ' / ' + s.total : '已提交 ' + s.scored + ' / ' + s.total }}</td><td><button :disabled="store.busy" @click="resume(s.id)">{{ s.status === 'completed' ? '查看结果' : '继续' }}</button></td></tr></tbody></table></div></section>
      <section v-if="page === 'quiz' && store.session && item">
        <div class="section-head"><div><p class="eyebrow">{{ store.session.status === 'completed' ? '练习结果' : store.session.mode === 'exam' ? '考试模式 · 交卷前隐藏答案' : '学习模式' }}</p><h2>{{ store.session.title }}</h2></div><span>第 {{ index + 1 }} / {{ store.session.items.length }} 题</span></div>
        <div v-if="store.session.status === 'completed'" class="result"><strong>正确 {{ correct }} / {{ store.session.items.length }} · {{ store.session.statistics.total.rate?.toFixed(1) }}%</strong><p>漏答 {{ store.session.statistics.total.unanswered }} 题 · 用时 {{ store.session.durationSeconds }} 秒（含暂停时间）。开始：{{ date(store.session.startedAt) }} · 结束：{{ date(store.session.completedAt) }}</p></div>
        <details v-if="store.session.status === 'completed'"><summary>查看本次 Domain / Type 统计</summary><StatisticsTable title="本次 Domain" :rows="store.session.statistics.domains" domains/><StatisticsTable title="本次题型" :rows="store.session.statistics.types"/></details><article class="panel question"><p class="muted">Domain {{ item.question.domain }} · {{ item.question.type }} · {{ item.question.difficulty }}</p><h3>{{ item.question.question }}</h3><button :disabled="store.busy" :aria-pressed="isFavorite" @click="favorite(item.question.id, !isFavorite)">{{ isFavorite ? '取消收藏' : '收藏本题' }}</button><fieldset :disabled="store.busy || item.scored || store.session.status === 'completed'"><legend class="sr-only">选择答案</legend><label v-for="key in ['A','B','C','D']" :key="key" class="option" :class="{chosen: selected === key}"><input v-model="selected" type="radio" name="answer" :value="key" @change="saveChoice()"> <strong>{{ key }}</strong><span>{{ item.question.options[key] }}</span></label></fieldset>
          <button v-if="store.session.status !== 'completed' && store.session.mode === 'study' && !item.scored" class="primary" :disabled="store.busy || !selected" @click="submit">提交本题并查看解析</button>
          <div v-if="store.session.status === 'active' && !item.scored" class="actions"><button :disabled="store.busy" @click="saveChoice(!item.flagged)">{{ item.flagged ? '取消标记' : '标记待复查' }}</button><span class="muted">选择会自动保存{{ item.flagged ? ' · 已标记' : '' }}</span></div><div v-if="item.question.answer" class="explanation"><h4>{{ item.correct ? '回答正确' : item.selected ? '回答错误' : '本题漏答' }} · 正确答案 {{ item.question.answer }}</h4><p>你的答案：{{ item.selected || '未作答' }}</p><p>{{ item.question.explanation }}</p><p v-if="item.question.whyCorrect"><strong>为什么正确：</strong>{{ item.question.whyCorrect }}</p><ul v-if="item.question.whyOthersWrong"><li v-for="(reason,key) in item.question.whyOthersWrong" :key="key">{{ key }}：{{ reason }}</li></ul><p v-if="item.question.tags?.length" class="muted">标签：{{ item.question.tags.join(' · ') }}</p></div>
        </article>
        <div class="actions spread"><button :disabled="store.busy || index === 0" @click="navigate(index-1)">上一题</button><span>已答 {{ answered }} / {{ store.session.items.length }}</span><button :disabled="store.busy || index === store.session.items.length-1" @click="navigate(index+1)">下一题</button></div>
        <div class="question-nav" aria-label="题号"><button v-for="(q,n) in store.session.items" :key="q.question.id" :aria-current="n === index ? 'step' : undefined" :class="{active:n===index}" :disabled="store.busy" @click="navigate(n)">{{ n+1 }}{{ q.scored || q.selected ? ' ✓' : '' }}{{ q.flagged ? ' ★' : '' }}</button></div>
        <div v-if="store.session.status !== 'completed'" class="panel"><button v-if="!confirmFinish" :disabled="store.busy" @click="confirmFinish = true">交卷并查看结果</button><div v-else><p>还有 {{ store.session.items.length - answered }} 题未作答或未提交，交卷后将作为漏答计错。确认交卷？</p><div class="actions"><button class="primary" :disabled="store.busy" @click="finish">确认交卷</button><button @click="confirmFinish = false">继续作答</button></div></div></div>
      </section>
      <section v-if="page === 'review'"><div class="section-head"><div><h2>错题与收藏</h2><p>曾经答错的题目始终保留。连续答对 3 次显示已掌握。</p></div><button class="primary" :disabled="store.busy || !reviewItems.length" @click="reviewStart">开始此范围复习</button></div><div class="actions"><button :class="{active:reviewKind==='wrong'}" @click="reviewKind='wrong'">错题历史</button><button :class="{active:reviewKind==='favorites'}" @click="reviewKind='favorites'">收藏题</button></div><p v-if="!reviewItems.length" class="empty">此范围暂无题目。作答后错题会自动记录，也可以在答题时收藏。</p><article v-for="q in reviewItems" :key="q.question.id" class="panel"><p class="muted">Domain {{ q.question.domain }} · {{ q.question.type }}</p><h3>{{ q.question.question }}</h3><p>作答 {{ q.attempts }} 次 · 答错 {{ q.wrong }} 次 · 连续正确 {{ q.streak }} 次 · {{ q.mastered ? '已掌握（保留历史）' : '待复习' }}</p><button :disabled="store.busy" :aria-pressed="q.favorite" @click="favorite(q.question.id,!q.favorite)">{{ q.favorite ? '取消收藏' : '收藏' }}</button></article></section>
      <section v-if="page === 'stats' && store.statistics"><h2>基础统计</h2><p>累计题次包括学习已提交题和已交卷考试，含漏答；不包含考试草稿。同题多次练习分别计数。</p><div class="result"><strong>累计 {{ store.statistics.total.count }} 题次 · 正确 {{ store.statistics.total.correct }} · 正确率 {{ store.statistics.total.rate == null ? '—' : store.statistics.total.rate.toFixed(1) + '%' }}</strong><p>漏答 {{ store.statistics.total.unanswered }} · 曾错题 {{ store.statistics.wrongQuestions }} · 收藏 {{ store.statistics.favorites }}</p></div><StatisticsTable title="按 Domain" :rows="store.statistics.domains" domains/><StatisticsTable title="按题型" :rows="store.statistics.types"/></section>
    </main><footer>题集仅用于个人学习。此工具不提供官方真题或考试通过预测。</footer>
  </div>
</template>
