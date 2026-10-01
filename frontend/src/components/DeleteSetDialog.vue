<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import { domain } from '../services/api'

const props = defineProps<{ target: domain.SetSummary | null; busy: boolean; error: string }>()
const emit = defineEmits<{ confirm: [id: string]; cancel: [] }>()
const dialog = ref<HTMLDialogElement | null>(null)
watch(() => props.target, async target => {
  await nextTick()
  if (target && !dialog.value?.open) dialog.value?.showModal()
  if (!target && dialog.value?.open) dialog.value.close()
})
function cancel(event: Event) {
  event.preventDefault()
  if (!props.busy) emit('cancel')
}
</script>

<template>
  <dialog ref="dialog" aria-labelledby="delete-set-title" aria-describedby="delete-set-description" @cancel="cancel">
    <form v-if="target" @submit.prevent="emit('confirm', target.id)">
      <h2 id="delete-set-title">确认删除题集</h2>
      <p><strong>{{ target.title }}</strong> · {{ target.count }} 题</p>
      <p id="delete-set-description">将永久删除这个题集及其题目、相关收藏，无法撤销。已有练习记录（含未完成会话）的题集会被禁止删除。</p>
      <p v-if="error" class="message error" role="alert">{{ error }}</p>
      <div class="actions">
        <button type="button" autofocus :disabled="busy" @click="emit('cancel')">取消</button>
        <button class="danger" type="submit" :disabled="busy">{{ busy ? '正在删除…' : '确认永久删除' }}</button>
      </div>
    </form>
  </dialog>
</template>

<style scoped>
dialog { width: min(560px, calc(100vw - 40px)); border: 1px solid #bacbc5; border-radius: 10px; color: inherit; padding: 28px; }
dialog::backdrop { background: rgb(15 30 25 / 45%); }
.danger { color: #fff; background: #9b3425; border-color: #9b3425; }
.danger:hover:not(:disabled) { background: #78291e; }
</style>
