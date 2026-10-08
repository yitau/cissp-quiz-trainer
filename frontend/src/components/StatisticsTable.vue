<script setup lang="ts">
import { domain } from '../services/api'
defineProps<{ title: string; rows: domain.Metric[]; domains?: boolean }>()
const names = ['','安全与风险管理','资产安全','安全架构与工程','通信与网络安全','身份与访问管理','安全评估与测试','安全运营','软件开发安全']
</script>
<template><div><h3>{{ title }}</h3><div class="table-wrap"><table><thead><tr><th>类别</th><th>题次（样本量）</th><th>正确</th><th>漏答</th><th>正确率</th></tr></thead><tbody><tr v-for="m in rows" :key="m.label"><td>{{ domains ? m.label + ' · ' + names[Number(m.label)] : m.label }}</td><td>{{ m.count }}</td><td>{{ m.correct }}</td><td>{{ m.unanswered }}</td><td>{{ m.rate == null ? '—' : m.rate.toFixed(1) + '%' }}</td></tr></tbody></table></div></div></template>
