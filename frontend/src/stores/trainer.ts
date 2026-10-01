import { defineStore } from 'pinia'
import { ref } from 'vue'
import * as api from '../services/api'

export const useTrainer = defineStore('trainer', () => {
  const sets = ref<api.domain.SetSummary[]>([])
  const sessions = ref<api.domain.SessionSummary[]>([])
  const session = ref<api.domain.Session | null>(null)
  const busy = ref(false)
  const error = ref('')
  const notice = ref('')
  async function run(action: () => Promise<void>) {
    if (busy.value) return
    busy.value = true; error.value = ''; notice.value = ''
    try { await action() } catch (e) { error.value = String(e) }
    finally { busy.value = false }
  }
  async function refresh() {
    sets.value = await api.ListSets()
    sessions.value = await api.ListSessions()
  }
  return { sets, sessions, session, busy, error, notice, run, refresh }
})
