<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import * as api from '../services/api'
import { useTrainer } from '../stores/trainer'
const emit = defineEmits<{ library: []; start: [id: string, mode: string] }>()
const store = useTrainer()
const lessons = ref<api.domain.LessonSummary[]>([])
const preview = ref<api.domain.LessonPreview | null>(null)
const view = ref<api.domain.LessonView | null>(null)
const conceptId = ref('')
const showAnswer = ref(false)
const showCase = ref(false)
const heading = ref<HTMLElement | null>(null)
const labels: Record<string, string> = { not_started: '未学习', in_progress: '学习中', understood: '已理解', needs_review: '需复习' }
const orderedIds = computed(() => view.value?.document.sections.flatMap(s => s.conceptIds) ?? [])
const currentIndex = computed(() => orderedIds.value.indexOf(conceptId.value))
const concept = computed(() => view.value?.document.concepts.find(c => c.id === conceptId.value))
const status = (id: string) => view.value?.progress.find(p => p.conceptId === id)?.status ?? 'not_started'
const term = (id: string) => view.value?.document.concepts.find(c => c.id === id)?.term
const fields = computed(() => concept.value ? [
  ['学习目标', concept.value.objective], ['定义', concept.value.definition], ['原理', concept.value.explanation],
  ['类比', concept.value.analogy], ['业务场景', concept.value.scenario]
] : [])
async function refresh() { lessons.value = await api.ListLessons() }
function choose() { store.run(async () => { preview.value = null; const p = await api.ChooseLessonImport(); if (p.token) preview.value = p }) }
function cancel() { store.run(async () => { await api.CancelLessonImport(); preview.value = null }) }
function confirm() { store.run(async () => { if (!preview.value) return; await api.ConfirmLessonImport(preview.value.token); preview.value = null; await refresh(); store.notice = '课程导入成功' }) }
async function openConcept(id: string) {
  if (!view.value) return
  const next = await api.OpenConcept(view.value.document.lesson.id, id)
  showAnswer.value = false; showCase.value = false; view.value = next; conceptId.value = id
  await refresh()
  heading.value?.focus()
}
function open(id: string) { store.run(async () => { view.value = await api.GetLesson(id); await openConcept(view.value.summary.resumeConceptId) }) }
function navigate(id: string) { store.run(async () => openConcept(id)) }
function mark(value: string) { store.run(async () => { if (!view.value) return; view.value = await api.SetConceptStatus(view.value.document.lesson.id, conceptId.value, value); await refresh() }) }
function caseStudy() { showAnswer.value = false; showCase.value = true }
onMounted(() => store.run(refresh))
</script>

<template>
  <section class="knowledge">
    <div class="section-head"><div><h2>知识学习</h2><p>先理解概念，再通过复习题验证。已理解为自主评价，不代表考试掌握。</p></div><button class="primary" :disabled="store.busy" @click="choose">导入 Lesson JSON</button></div>
    <article v-if="preview" class="panel" aria-label="课程导入预览">
      <h3>导入预览：{{ preview.lesson.title }}</h3>
      <p>Week {{ preview.lesson.week }} / Day {{ preview.lesson.day }} / Domain {{ preview.lesson.domain }} · {{ preview.sections.length }} 章节 / {{ preview.conceptCount }} 知识点</p>
      <p>关联题集 ID：{{ preview.lesson.linkedQuestionSetId || '未关联' }}</p>
      <ol><li v-for="s in preview.sections" :key="s.id">{{ s.title }} · {{ s.conceptIds.length }} 知识点</li></ol>
      <p role="status">{{ preview.status === 'ready' ? '校验通过，可以导入。' : preview.status === 'duplicate' ? '重复课程：ID 与内容相同，原有进度保留，无需再次导入。' : '内容冲突：相同课程 ID 已存在，拒绝覆盖，原有内容和进度保留。' }}</p>
      <div class="actions"><button class="primary" :disabled="store.busy || preview.status !== 'ready'" @click="confirm">确认导入课程</button><button :disabled="store.busy" @click="cancel">取消</button></div>
    </article>
    <template v-if="!view">
      <p v-if="!lessons.length" class="empty">暂无课程。点击“导入 Lesson JSON”，选择随包提供的 cissp-week01-day01-lesson.json。</p>
      <div class="grid"><article v-for="s in lessons" :key="s.lesson.id" class="panel">
        <span class="badge">Week {{ s.lesson.week }} · Day {{ s.lesson.day }} · Domain {{ s.lesson.domain }}</span>
        <h3>{{ s.lesson.title }}</h3><p>{{ s.lesson.description }}</p>
        <p>{{ s.sectionCount }} 章节 · {{ s.conceptCount }} 知识点 · 已理解 {{ s.understood }} / {{ s.conceptCount }}{{ s.completed ? ' · 课程已完成' : '' }}</p>
        <button class="primary" :disabled="store.busy" @click="open(s.lesson.id)">继续学习</button>
      </article></div>
    </template>
    <template v-else>
      <div class="actions"><button :disabled="store.busy" @click="view = null; showAnswer = false">返回课程列表</button><span role="status">已理解 {{ view.summary.understood }} / {{ view.summary.conceptCount }}{{ view.summary.completed ? ' · 课程已完成' : '' }}</span></div>
      <h3>{{ view.document.lesson.title }}</h3>
      <div class="lesson-layout">
        <aside class="panel" aria-label="章节目录">
          <div v-for="s in view.document.sections" :key="s.id"><h4>{{ s.title }}</h4><ol>
            <li v-for="id in s.conceptIds" :key="id"><button :disabled="store.busy" :aria-current="!showCase && conceptId === id ? 'step' : undefined" :class="{active: !showCase && conceptId === id}" @click="navigate(id)">{{ term(id)?.en }} · {{ term(id)?.zh }}<br><small>{{ labels[status(id)] }}</small></button></li>
          </ol></div>
          <button v-if="view.document.caseStudy" :disabled="store.busy" :aria-pressed="showCase" @click="caseStudy">综合案例</button>
        </aside>
        <article class="panel lesson-content">
          <template v-if="showCase && view.document.caseStudy">
            <h3>{{ view.document.caseStudy.title }}</h3><p>{{ view.document.caseStudy.background }}</p>
            <ol><li v-for="(step,n) in view.document.caseStudy.steps" :key="n"><strong>{{ step.label }}</strong><p>{{ step.detail }}</p></li></ol>
            <h4>总结</h4><p>{{ view.document.caseStudy.takeaway }}</p>
            <p class="muted">综合案例不计入知识点完成数。继续学习定位最近打开的知识点。</p>
            <button :disabled="store.busy" @click="showCase = false">返回知识点</button>
          </template>
          <template v-else-if="concept">
            <p class="eyebrow">第 {{ currentIndex + 1 }} / {{ orderedIds.length }} 个知识点 · {{ labels[status(conceptId)] }}</p>
            <h3 ref="heading" tabindex="-1">{{ concept.term.en }} · {{ concept.term.zh }}</h3>
            <div v-for="[label, text] in fields" :key="label"><h4>{{ label }}</h4><p>{{ text }}</p></div>
            <template v-if="concept.controls?.length"><h4>相关控制措施</h4><ul><li v-for="control in concept.controls" :key="control">{{ control }}</li></ul></template>
            <h4>常见误区</h4><p>{{ concept.commonMistake }}</p><h4>CISSP 提示</h4><p>{{ concept.examTip }}</p>
            <div class="explanation"><h4>理解自检 · 不计分</h4><p>{{ concept.recall.prompt }}</p><button :aria-expanded="showAnswer" aria-controls="recall-answer" @click="showAnswer = !showAnswer">{{ showAnswer ? '隐藏参考答案' : '查看参考答案' }}</button><p v-if="showAnswer" id="recall-answer">{{ concept.recall.answer }}</p></div>
            <div class="actions" aria-label="理解状态"><button v-for="value in ['understood','needs_review','in_progress']" :key="value" :disabled="store.busy" :aria-pressed="status(conceptId) === value" @click="mark(value)">{{ labels[value] }}</button></div>
            <div class="actions spread"><button :disabled="store.busy || currentIndex <= 0" @click="navigate(orderedIds[currentIndex - 1]!)">上一知识点</button><button :disabled="store.busy || currentIndex >= orderedIds.length - 1" @click="navigate(orderedIds[currentIndex + 1]!)">下一知识点</button></div>
          </template>
        </article>
      </div>
      <article class="panel"><h3>关联复习题</h3><p>题集 ID：{{ view.document.lesson.linkedQuestionSetId || '课程未关联题集' }}</p>
        <div v-if="view.linkedSetStatus === 'available'" class="actions"><button class="primary" :disabled="store.busy" @click="emit('start', view.document.lesson.linkedQuestionSetId!, 'study')">开始复习题（学习）</button><button :disabled="store.busy" @click="emit('start', view.document.lesson.linkedQuestionSetId!, 'exam')">考试模式</button></div>
        <template v-else-if="view.linkedSetStatus !== 'unlinked'"><p>{{ view.linkedSetStatus === 'archived' ? '关联题集已归档，请先在题库中恢复显示。' : '关联题集尚未导入。请导入具有上述 set.id 的 Question JSON 后再开始复习。' }}</p><button :disabled="store.busy" @click="emit('library')">前往题库与导入</button></template>
      </article>
    </template>
  </section>
</template>

<style scoped>
.knowledge {overflow-wrap:anywhere}.lesson-layout{display:grid;grid-template-columns:minmax(220px, 280px) minmax(0,1fr);gap:18px;align-items:start}.lesson-layout .panel{min-width:0}.lesson-layout aside ol{list-style:none;padding:0}.lesson-layout aside button{width:100%;text-align:left;white-space:normal;overflow-wrap:anywhere}.lesson-content p,.lesson-content li{white-space:pre-wrap}.lesson-content h4{margin-bottom:6px}.lesson-content h3:focus{outline:2px solid #c48e32;outline-offset:5px}@media(max-width:850px){.lesson-layout{grid-template-columns:minmax(0,1fr)}}
</style>
