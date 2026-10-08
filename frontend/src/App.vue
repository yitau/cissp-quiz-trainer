<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import * as api from './services/api'
import KnowledgeLearning from './components/KnowledgeLearning.vue'
import StatisticsTable from './components/StatisticsTable.vue'
import DeleteSetDialog from './components/DeleteSetDialog.vue'
import { useTrainer } from './stores/trainer'
const store = useTrainer()
const dataDir = ref('')
const page = ref('library')
const showArchived = ref(false)
const visibleSets = computed(() => store.sets.filter(s => s.archived === showArchived.value))
const deleteTarget = ref<api.domain.SetSummary | null>(null)
const reviewKind = ref('wrong')
const reviewItems = computed(() => store.review.filter(q => reviewKind.value === 'wrong' ? q.wrong > 0 : q.favorite))
const isFavorite = computed(() => store.review.find(q => q.question.id === item.value?.question.id)?.favorite ?? false)
const preview = ref<api.domain.Preview | null>(null)
const browsing = ref<api.domain.Question[]>([])
const index = ref(0)
const selected = ref('')
const confirmFinish = ref(false)
const cardFilter = ref('all')
const progressLabels: Record<string, string> = { unanswered: '未答', draft: '待提交', submitted: '已提交', selected: '已选', correct: '正确', wrong: '错误', omitted: '漏答' }
const numberedItems = computed(() => store.session?.items.map((q, n) => ({ q, n })) ?? [])
const pendingItems = computed(() => numberedItems.value.filter(({ q }) => ['unanswered', 'draft'].includes(q.progress)))
const cardItems = computed(() => numberedItems.value.filter(({ q }) => cardFilter.value === 'all' || (cardFilter.value === 'flagged' ? q.flagged : cardFilter.value === 'wrong' ? ['wrong', 'omitted'].includes(q.progress) : q.progress === cardFilter.value)))
const cardFilters = computed(() => store.session?.status === 'completed'
  ? [{ key: 'all', label: '全部结果' }, { key: 'wrong', label: '本次错题（含漏答）' }, { key: 'omitted', label: '本次漏答' }]
  : [{ key: 'all', label: '全部题目' }, { key: 'unanswered', label: '未答' }, ...(store.session?.mode === 'study' ? [{ key: 'draft', label: '待提交' }] : []), { key: 'flagged', label: '已标记' }])
const item = computed(() => store.session?.items[index.value])
const answered = computed(() => store.session?.items.filter(i => store.session?.mode === 'exam' ? !!i.selected : i.scored).length ?? 0)
const currentVisible = computed(() => store.session?.status !== 'completed' || cardItems.value.some(entry => entry.n === index.value))
const navigationItems = computed(() => store.session?.status === 'completed' ? cardItems.value : numberedItems.value)
const previousIndex = computed(() => [...navigationItems.value].reverse().find(entry => entry.n < index.value)?.n)
const nextIndex = computed(() => navigationItems.value.find(entry => entry.n > index.value)?.n)
const correct = computed(() => store.session?.statistics.total.correct ?? 0)
function navigate(n: number) { index.value = n; selected.value = item.value?.selected ?? ''; confirmFinish.value = false }
async function openSession(s: api.domain.Session) { store.session = s; page.value = 'quiz'; cardFilter.value = 'all'; navigate(s.resumeIndex) }
function filterCard(value: string) { cardFilter.value = value; const first = cardItems.value[0]; if (first) navigate(first.n) }
function retryMistakes() { store.run(async () => { if (!store.session) return; await openSession(await api.StartSessionReview(store.session.id)); await store.refresh() }) }
function start(id: string, mode = 'study') { store.run(async () => { await openSession(await api.StartQuiz(id, mode)); await store.refresh() }) }
function resume(id: string) { store.run(async () => openSession(await api.GetSession(id))) }
function saveChoice(flagged = item.value?.flagged ?? false) { store.run(async () => { if (!store.session || !item.value) return; try { store.session = await api.SaveChoice(store.session.id, item.value.question.id, selected.value, flagged); store.sessions = await api.ListSessions() } catch(e) { selected.value = item.value.selected; throw e } }) }
function submit() { store.run(async () => { if (!store.session || !item.value) return; store.session = await api.SubmitStudy(store.session.id, item.value.question.id, selected.value); await store.refresh() }) }
function finish() { store.run(async () => { if (!store.session) return; store.session = await api.CompleteQuiz(store.session.id); cardFilter.value = 'all'; confirmFinish.value = false; selected.value = item.value?.selected ?? ''; await store.refresh() }) }
function choose() { store.run(async () => { preview.value = null; const p = await api.ChooseImport(); if (p.token) preview.value = p }) }
function importSet() { store.run(async () => { if (!preview.value) return; await api.ConfirmImport(preview.value.token); preview.value = null; store.notice = '题集导入成功'; await store.refresh() }) }
function favorite(id: string, value: boolean) { store.run(async () => { await api.SetFavorite(id, value); await store.refresh() }) }
function reviewStart() { store.run(async () => { await openSession(await api.StartReview(reviewKind.value)); await store.refresh() }) }
function askDelete(set: api.domain.SetSummary) { store.error = ''; deleteTarget.value = set }
function deleteSet(id: string) {
  store.run(async () => {
    await api.DeleteUnusedSet(id)
    deleteTarget.value = null; browsing.value = []; preview.value = null
    await store.refresh()
    store.notice = '题集已删除'
  })
}
function archiveSet(id: string, archived: boolean) {
  store.run(async () => {
    await api.SetArchived(id, archived)
    browsing.value = []
    await store.refresh()
    store.notice = archived ? '题集已归档，历史、错题、收藏和统计继续保留' : '题集已恢复显示'
  })
}
function viewSet(id: string) { store.run(async () => { browsing.value = await api.SetQuestions(id) }) }
const date = (value: string) => value ? new Date(value).toLocaleString('zh-CN') : '进行中'
function backup() { store.run(async () => { const path = await api.CreateBackup(); if (path) store.notice = '备份已保存：' + path }) }
function restore() { store.run(async () => { const safety = await api.RestoreBackup(); if (!safety) return; store.session = null; preview.value = null; browsing.value = []; await store.refresh(); store.notice = '恢复成功。恢复前的安全备份：' + safety }) }
onMounted(() => store.run(async () => { dataDir.value = await api.DataDirectory(); await store.refresh() }))
</script>
<template>
  <div class="app-shell">
    <DeleteSetDialog :target="deleteTarget" :busy="store.busy" :error="store.error" @confirm="deleteSet" @cancel="deleteTarget = null"/>
    <header><div><p class="eyebrow">本地学习 · 每日积累</p><h1>CISSP Quiz Trainer</h1></div><span class="badge">v0.2.0 · 离线使用</span></header>
    <nav aria-label="主要导航"><button :disabled="store.busy" :class="{active: page === 'knowledge'}" @click="page = 'knowledge'">知识学习</button><button :disabled="store.busy" :class="{active: page === 'library'}" @click="page = 'library'">题库与导入</button><button :disabled="store.busy" :class="{active: page === 'history'}" @click="page = 'history'">练习记录</button><button :disabled="store.busy" :class="{active: page === 'review'}" @click="page = 'review'">错题与收藏</button><button :disabled="store.busy" :class="{active: page === 'stats'}" @click="page = 'stats'">统计与备份</button><button :disabled="store.busy" v-if="store.session" :class="{active: page === 'quiz'}" @click="page = 'quiz'">当前答题</button></nav>
    <p v-if="store.error" class="message error" role="alert">{{ store.error }}</p><p v-if="store.notice" class="message" role="status">{{ store.notice }}</p><p v-if="store.busy" role="status">正在处理，请稍候…</p>
    <main :aria-busy="store.busy"><KnowledgeLearning v-if="page === 'knowledge'" @library="page = 'library'" @start="start"/>
      <section v-if="page === 'library'">
        <div class="section-head"><div><h2>我的题库</h2><p>导入标准 JSON 题集，按原题顺序开始学习。</p></div><button class="primary" :disabled="store.busy" @click="choose">导入 JSON 题集</button></div>
        <article v-if="preview" class="panel"><h3>导入预览：{{ preview.set.title }}</h3><p>{{ preview.set.description }} · {{ preview.questions.length }} 题</p><p>ID：{{ preview.set.id }} · {{ preview.status === 'ready' ? '可导入' : preview.status === 'duplicate' ? '重复题集' : '存在冲突' }}</p><ul v-if="preview.messages.length"><li v-for="message in preview.messages" :key="message">{{ message }}</li></ul><details><summary>查看题目内容（不显示答案）</summary><ol><li v-for="q in preview.questions" :key="q.id">{{ q.question }} <span class="muted">Domain {{ q.domain }} / {{ q.type }}</span></li></ol></details><div class="actions"><button class="primary" :disabled="store.busy || preview.status !== 'ready'" @click="importSet">确认导入全部题目</button><button @click="preview = null">取消</button></div></article>
        <div class="actions" aria-label="题库范围">
          <button :class="{active: !showArchived}" :aria-pressed="!showArchived" @click="showArchived = false; browsing = []">日常题库</button>
          <button :class="{active: showArchived}" :aria-pressed="showArchived" @click="showArchived = true; browsing = []">已归档</button>
        </div>
        <p v-if="showArchived" class="muted">归档仅隐藏日常题库入口，已有会话可继续，错题、收藏和统计保持不变。恢复显示后可开始整套练习。</p>
        <p v-if="!visibleSets.length" class="empty">{{ showArchived ? '暂无已归档题集。' : '暂无日常题集。可导入 JSON，或从“已归档”恢复显示。' }}</p>
        <div class="grid">
          <article v-for="s in visibleSets" :key="s.id" class="panel">
            <span class="badge">{{ s.count }} 题{{ s.archived ? ' · 已归档' : '' }}</span>
            <h3>{{ s.title }}</h3><p>{{ s.description || '暂无简介' }}</p>
            <div class="actions">
              <button v-if="!s.archived" class="primary" :disabled="store.busy" @click="start(s.id)">开始学习</button>
              <button v-if="!s.archived" :disabled="store.busy" @click="start(s.id, 'exam')">开始考试</button>
              <button :disabled="store.busy" @click="viewSet(s.id)">查看题目</button>
              <button v-if="s.archived" :disabled="store.busy" @click="archiveSet(s.id, false)">恢复显示</button>
              <button v-else-if="s.hasHistory" :disabled="store.busy" @click="archiveSet(s.id, true)">归档题集</button>
              <button v-if="!s.hasHistory" :disabled="store.busy" @click="askDelete(s)">删除题集…</button>
            </div>
            <p class="muted">{{ s.hasHistory ? '已有练习记录，保留历史，仅可归档。' : '尚无练习记录，可确认删除。' }}</p>
          </article>
        </div>
        <article v-if="browsing.length" class="panel"><div class="section-head"><h3>题目一览</h3><button @click="browsing = []">收起</button></div><ol><li v-for="q in browsing" :key="q.id"><p>{{ q.question }}</p><p class="muted">{{ q.id }} · Domain {{ q.domain }} · {{ q.type }} · {{ q.difficulty }}</p></li></ol></article>
      </section>
      <section v-if="page === 'history'"><h2>练习记录</h2><p>未完成练习会自动保存。继续时定位下一道未完成题；全部完成待交卷时优先定位标记题。</p><p v-if="!store.sessions.length" class="empty">尚无练习记录。从题库开始一次学习吧。</p><div class="table-wrap" v-else><table><thead><tr><th>题集</th><th>开始时间</th><th>状态</th><th>进度 / 结果</th><th>操作</th></tr></thead><tbody><tr v-for="s in store.sessions" :key="s.id"><td>{{ s.title }}<br><span class="muted">{{ s.mode === 'exam' ? '考试' : '学习' }}</span></td><td>{{ date(s.startedAt) }}</td><td>{{ s.status === 'completed' ? '已完成' : '进行中' }}</td><td>{{ s.status === 'completed' ? '正确 ' + s.correct + ' / ' + s.total : (s.mode === 'exam' ? '已选 ' + s.selected : '已提交 ' + s.scored) + ' / ' + s.total }}</td><td><button :disabled="store.busy" @click="resume(s.id)">{{ s.status === 'completed' ? '查看结果' : '继续' }}</button></td></tr></tbody></table></div></section>
      <section v-if="page === 'quiz' && store.session && item">
        <div class="section-head"><div><p class="eyebrow">{{ store.session.status === 'completed' ? '练习结果' : store.session.mode === 'exam' ? '考试模式 · 交卷前隐藏答案' : '学习模式' }}</p><h2>{{ store.session.title }}</h2></div><span>第 {{ index + 1 }} / {{ store.session.items.length }} 题</span></div>
        <div v-if="store.session.status === 'completed'" class="result"><strong>正确 {{ correct }} / {{ store.session.items.length }} · 错误 {{ store.session.statistics.total.wrong }} · {{ store.session.statistics.total.rate?.toFixed(1) }}%</strong><p>漏答 {{ store.session.statistics.total.unanswered }} 题 · 用时 {{ store.session.durationSeconds }} 秒（含暂停时间）。开始：{{ date(store.session.startedAt) }} · 结束：{{ date(store.session.completedAt) }}</p></div>
        <details v-if="store.session.status === 'completed'"><summary>查看本次 Domain / Type 统计</summary><StatisticsTable title="本次 Domain" :rows="store.session.statistics.domains" domains/><StatisticsTable title="本次题型" :rows="store.session.statistics.types"/></details>
        <details class="panel" :open="store.session.status === 'completed'" aria-label="答题检查与结果筛选">
          <summary>{{ store.session.status === 'completed' ? '本次结果检查' : '答题检查 · 未完成 ' + pendingItems.length + ' 题（展开筛选）' }}</summary>
          <div class="actions" aria-label="题号筛选"><button v-for="filter in cardFilters" :key="filter.key" :aria-pressed="cardFilter === filter.key" :class="{active: cardFilter === filter.key}" :disabled="store.busy" @click="filterCard(filter.key)">{{ filter.label }}</button></div>
          <p class="muted">{{ store.session.status === 'completed' ? '选择题号查看解析；错题范围包含漏答。' : '筛选题号可跳转检查；学习已选答案仍需提交本题。' }}</p>
          <p v-if="!cardItems.length" role="status">当前筛选范围没有题目。</p>
          <div class="question-nav" aria-label="题号"><button v-for="{q,n} in cardItems" :key="q.question.id" :aria-current="n === index ? 'step' : undefined" :class="{active:n===index}" :disabled="store.busy" @click="navigate(n)">第 {{ n+1 }} 题 · {{ progressLabels[q.progress] }}{{ q.flagged ? ' · 已标记' : '' }}</button></div>
          <button v-if="store.session.status === 'completed'" :disabled="store.busy || !store.session.statistics.total.wrong" @click="retryMistakes">重练本次错题（含漏答）</button>
        </details>
        <article v-if="currentVisible" class="panel question"><p class="muted">Domain {{ item.question.domain }} · {{ item.question.type }} · {{ item.question.difficulty }}</p><h3>{{ item.question.question }}</h3><button :disabled="store.busy" :aria-pressed="isFavorite" @click="favorite(item.question.id, !isFavorite)">{{ isFavorite ? '取消收藏' : '收藏本题' }}</button><fieldset :disabled="store.busy || item.scored || store.session.status === 'completed'"><legend class="sr-only">选择答案</legend><label v-for="key in ['A','B','C','D']" :key="key" class="option" :class="{chosen: selected === key}"><input v-model="selected" type="radio" name="answer" :value="key" @change="saveChoice()"> <strong>{{ key }}</strong><span>{{ item.question.options[key] }}</span></label></fieldset>
          <button v-if="store.session.status !== 'completed' && store.session.mode === 'study' && !item.scored" class="primary" :disabled="store.busy || !selected" @click="submit">提交本题并查看解析</button>
          <div v-if="store.session.status === 'active' && !item.scored" class="actions"><button :disabled="store.busy" @click="saveChoice(!item.flagged)">{{ item.flagged ? '取消标记' : '标记待复查' }}</button><span class="muted">选择会自动保存{{ item.flagged ? ' · 已标记' : '' }}</span></div><div v-if="item.question.answer" class="explanation"><h4>{{ item.correct ? '回答正确' : item.selected ? '回答错误' : '本题漏答' }} · 正确答案 {{ item.question.answer }}</h4><p>你的答案：{{ item.selected || '未作答' }}</p><p>{{ item.question.explanation }}</p><p v-if="item.question.whyCorrect"><strong>为什么正确：</strong>{{ item.question.whyCorrect }}</p><ul v-if="item.question.whyOthersWrong"><li v-for="(reason,key) in item.question.whyOthersWrong" :key="key">{{ key }}：{{ reason }}</li></ul><p v-if="item.question.tags?.length" class="muted">标签：{{ item.question.tags.join(' · ') }}</p></div>
        </article>
        <div v-if="currentVisible" class="actions spread"><button :disabled="store.busy || previousIndex === undefined" @click="previousIndex !== undefined && navigate(previousIndex)">上一题</button><span>{{ store.session.status === 'completed' ? '已完成' : store.session.mode === 'exam' ? '已选' : '已提交' }} {{ store.session.status === 'completed' ? store.session.items.length : answered }} / {{ store.session.items.length }}</span><button :disabled="store.busy || nextIndex === undefined" @click="nextIndex !== undefined && navigate(nextIndex)">下一题</button></div>
        <div v-if="store.session.status !== 'completed'" class="panel"><button v-if="!confirmFinish" :disabled="store.busy" @click="confirmFinish = true">交卷并查看结果</button><div v-else><p>还有 {{ pendingItems.length }} 题未作答或未提交，交卷后将作为漏答计错。确认交卷？</p><div v-if="pendingItems.length" class="question-nav" aria-label="交卷前未完成题目"><button v-for="{q,n} in pendingItems" :key="q.question.id" :disabled="store.busy" @click="cardFilter = 'all'; navigate(n)">返回第 {{ n+1 }} 题 · {{ progressLabels[q.progress] }}</button></div><p v-else>所有题目均已{{ store.session.mode === 'exam' ? '选择答案' : '提交' }}，可确认交卷。</p><div class="actions"><button class="primary" :disabled="store.busy" @click="finish">确认交卷</button><button @click="confirmFinish = false">继续作答</button></div></div></div>
      </section>
      <section v-if="page === 'review'"><div class="section-head"><div><h2>错题与收藏</h2><p>曾经答错的题目始终保留。连续答对 3 次显示已掌握。</p></div><button class="primary" :disabled="store.busy || !reviewItems.length" @click="reviewStart">开始此范围复习</button></div><div class="actions"><button :class="{active:reviewKind==='wrong'}" @click="reviewKind='wrong'">错题历史</button><button :class="{active:reviewKind==='favorites'}" @click="reviewKind='favorites'">收藏题</button></div><p v-if="!reviewItems.length" class="empty">此范围暂无题目。作答后错题会自动记录，也可以在答题时收藏。</p><article v-for="q in reviewItems" :key="q.question.id" class="panel"><p class="muted">Domain {{ q.question.domain }} · {{ q.question.type }}</p><h3>{{ q.question.question }}</h3><p>作答 {{ q.attempts }} 次 · 答错 {{ q.wrong }} 次 · 连续正确 {{ q.streak }} 次 · {{ q.mastered ? '已掌握（保留历史）' : '待复习' }}</p><button :disabled="store.busy" :aria-pressed="q.favorite" @click="favorite(q.question.id,!q.favorite)">{{ q.favorite ? '取消收藏' : '收藏' }}</button></article></section>
      <section v-if="page === 'stats' && store.statistics"><h2>基础统计</h2><p>累计题次包括学习已提交题和已交卷考试，含漏答；不包含考试草稿。同题多次练习分别计数。</p><div class="result"><strong>累计 {{ store.statistics.total.count }} 题次 · 正确 {{ store.statistics.total.correct }} · 正确率 {{ store.statistics.total.rate == null ? '—' : store.statistics.total.rate.toFixed(1) + '%' }}</strong><p>漏答 {{ store.statistics.total.unanswered }} · 曾错题 {{ store.statistics.wrongQuestions }} · 收藏 {{ store.statistics.favorites }}</p></div><StatisticsTable title="按 Domain" :rows="store.statistics.domains" domains/><StatisticsTable title="按题型" :rows="store.statistics.types"/><article class="panel"><h3>完整备份与恢复</h3><p>备份包含课程、知识学习进度、题库、全部练习记录、收藏和应用配置。恢复会替换当前数据，校验成功后先自动保存恢复前安全备份。</p><p class="muted">数据目录：{{ dataDir }}</p><div class="actions"><button class="primary" :disabled="store.busy" @click="backup">保存完整备份</button><button :disabled="store.busy" @click="restore">从备份恢复…</button></div><p class="muted">请使用新文件名；不会覆盖已有备份。文件操作期间请勿关闭应用。</p></article></section>
    </main><footer>题集仅用于个人学习。此工具不提供官方真题或考试通过预测。</footer>
  </div>
</template>
