<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'

import SectionHeader from '../../components/SectionHeader.vue'
import { api } from '../../lib/api'
import type { ApprovalDetail, ApprovalSummary } from '../../lib/types'

const filters = reactive({
  status: '',
  type: '',
})

const items = ref<ApprovalSummary[]>([])
const selectedId = ref('')
const detail = ref<ApprovalDetail | null>(null)
const loading = ref(true)
const detailLoading = ref(false)
const acting = ref('')
const error = ref('')
const actionError = ref('')

const selectedApproval = computed(() => detail.value?.approval ?? items.value.find((item) => item.id === selectedId.value) ?? null)

async function loadApprovals() {
  loading.value = true
  error.value = ''
  try {
    items.value = await api.getApprovals('admin', {
      status: filters.status || undefined,
      type: filters.type || undefined,
    })
    if (!selectedId.value && items.value.length) {
      selectedId.value = items.value[0].id
    }
    if (selectedId.value) {
      await loadDetail(selectedId.value)
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : '加载审批列表失败'
  } finally {
    loading.value = false
  }
}

async function loadDetail(id: string) {
  detailLoading.value = true
  actionError.value = ''
  try {
    detail.value = await api.getAdminApprovalDetail(id)
    selectedId.value = id
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '加载审批详情失败'
  } finally {
    detailLoading.value = false
  }
}

async function approve() {
  if (!selectedApproval.value) return
  acting.value = 'approve'
  actionError.value = ''
  try {
    await api.approveAdminApproval(selectedApproval.value.id, '审批通过，允许进入执行阶段。')
    await loadApprovals()
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '审批通过失败'
  } finally {
    acting.value = ''
  }
}

async function reject() {
  if (!selectedApproval.value) return
  const reason = window.prompt('填写驳回理由', selectedApproval.value.rejectReason || '需要补充风险说明与执行窗口。')
  if (!reason) {
    return
  }
  acting.value = 'reject'
  actionError.value = ''
  try {
    await api.rejectAdminApproval(selectedApproval.value.id, reason, reason)
    await loadApprovals()
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '驳回失败'
  } finally {
    acting.value = ''
  }
}

async function execute() {
  if (!selectedApproval.value) return
  acting.value = 'execute'
  actionError.value = ''
  try {
    await api.executeAdminApproval(selectedApproval.value.id)
    await loadApprovals()
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '执行审批失败'
  } finally {
    acting.value = ''
  }
}

function isSelected(item: ApprovalSummary) {
  return item.id === selectedId.value
}

onMounted(async () => {
  await loadApprovals()
})
</script>

<template>
  <div class="approvals-shell">
    <section class="card panel">
      <SectionHeader title="审批中心" subtitle="统一处理配置发布、运行时控制、实例删除与诊断授权" />
      <div class="filters">
        <el-select v-model="filters.status" placeholder="全部状态" clearable>
          <el-option label="pending" value="pending" />
          <el-option label="approved" value="approved" />
          <el-option label="rejected" value="rejected" />
          <el-option label="executed" value="executed" />
          <el-option label="expired" value="expired" />
        </el-select>
        <el-select v-model="filters.type" placeholder="全部类型" clearable>
          <el-option label="config_publish" value="config_publish" />
          <el-option label="runtime_restart" value="runtime_restart" />
          <el-option label="runtime_stop" value="runtime_stop" />
          <el-option label="runtime_scale" value="runtime_scale" />
          <el-option label="delete_instance" value="delete_instance" />
          <el-option label="diagnostic_access" value="diagnostic_access" />
        </el-select>
        <el-button type="primary" @click="loadApprovals">刷新</el-button>
      </div>
      <div v-if="loading" class="state-card">正在同步审批单…</div>
      <el-alert v-else-if="error" :closable="false" show-icon type="error" :title="error" />
      <div v-else class="approval-list">
        <button
          v-for="item in items"
          :key="item.id"
          type="button"
          :class="['approval-item', { 'approval-item--active': isSelected(item) }]"
          @click="loadDetail(item.id)"
        >
          <div class="approval-item__head">
            <strong>{{ item.approvalNo }}</strong>
            <span :class="['pill', `pill--${item.riskLevel}`]">{{ item.riskLevel }}</span>
          </div>
          <div class="muted">{{ item.approvalType }} · {{ item.status }}</div>
          <div>{{ item.reason || '未填写原因' }}</div>
          <small class="muted">
            {{ item.applicantName || '未知申请人' }} · {{ item.createdAt }}
          </small>
        </button>
      </div>
    </section>

    <section class="card panel detail-panel">
      <SectionHeader title="审批详情" subtitle="状态机、执行上下文与审批历史" />
      <div v-if="detailLoading" class="state-card">正在读取审批详情…</div>
      <div v-else-if="selectedApproval" class="detail-stack">
        <div class="detail-meta">
          <div>
            <div class="eyebrow">{{ selectedApproval.approvalType }}</div>
            <h3>{{ selectedApproval.approvalNo }}</h3>
            <p class="muted">
              {{ selectedApproval.applicantName || '未知申请人' }} · {{ selectedApproval.status }} ·
              {{ selectedApproval.createdAt }}
            </p>
          </div>
          <div class="detail-actions">
            <el-button
              v-if="selectedApproval.status === 'pending'"
              type="success"
              :loading="acting === 'approve'"
              @click="approve"
            >
              审批通过
            </el-button>
            <el-button
              v-if="selectedApproval.status === 'pending' || selectedApproval.status === 'approved'"
              plain
              :loading="acting === 'reject'"
              @click="reject"
            >
              驳回
            </el-button>
            <el-button
              v-if="selectedApproval.status === 'approved'"
              type="primary"
              :loading="acting === 'execute'"
              @click="execute"
            >
              执行审批
            </el-button>
          </div>
        </div>

        <el-alert v-if="actionError" :closable="false" show-icon type="error" :title="actionError" />

        <div class="meta-grid">
          <div class="meta-card">
            <span class="muted">风险等级</span>
            <strong>{{ selectedApproval.riskLevel }}</strong>
          </div>
          <div class="meta-card">
            <span class="muted">实例</span>
            <strong>{{ detail?.instance?.name || `#${selectedApproval.instanceId || '—'}` }}</strong>
          </div>
          <div class="meta-card">
            <span class="muted">审批人</span>
            <strong>{{ selectedApproval.approverName || '—' }}</strong>
          </div>
          <div class="meta-card">
            <span class="muted">执行人</span>
            <strong>{{ selectedApproval.executorName || '—' }}</strong>
          </div>
        </div>

        <div class="reason-card">
          <strong>申请说明</strong>
          <p>{{ selectedApproval.reason || '未填写申请说明。' }}</p>
          <p v-if="selectedApproval.approvalComment" class="muted">审批意见：{{ selectedApproval.approvalComment }}</p>
          <p v-if="selectedApproval.rejectReason" class="muted">驳回理由：{{ selectedApproval.rejectReason }}</p>
        </div>

        <div v-if="detail?.instance" class="link-row">
          <RouterLink :to="`/admin/instances/${detail.instance.id}?tab=monitoring`">查看实例监控</RouterLink>
          <RouterLink :to="`/admin/instances/${detail.instance.id}?tab=diagnostics`">进入实例诊断</RouterLink>
          <RouterLink :to="`/admin/instances/${detail.instance.id}/workspace`">回到工作台</RouterLink>
        </div>

        <div class="history">
          <strong>审批历史</strong>
          <div v-if="detail?.actions.length" class="history-list">
            <div v-for="item in detail.actions" :key="item.id" class="history-item">
              <strong>{{ item.actorName }} · {{ item.action }}</strong>
              <span class="muted">{{ item.createdAt }}</span>
              <small v-if="item.comment" class="muted">{{ item.comment }}</small>
            </div>
          </div>
          <div v-else class="muted">当前还没有审批历史记录。</div>
        </div>
      </div>
      <div v-else class="state-card">请选择一条审批单查看详情。</div>
    </section>
  </div>
</template>

<style scoped>
.approvals-shell {
  display: grid;
  grid-template-columns: 0.96fr 1.04fr;
  gap: 14px;
}

.panel {
  padding: 16px;
}

.filters {
  display: grid;
  grid-template-columns: 1fr 1fr auto;
  gap: 10px;
  margin: 14px 0;
}

.state-card {
  padding: 24px;
  text-align: center;
  border: 1px dashed var(--stroke);
  border-radius: var(--radius-lg);
  background: var(--panel-muted);
}

.approval-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.approval-item {
  padding: 14px;
  border: 1px solid var(--stroke);
  border-radius: var(--radius-lg);
  background: var(--panel-muted);
  text-align: left;
  color: inherit;
}

.approval-item--active {
  border-color: rgba(59, 130, 246, 0.55);
  box-shadow: 0 0 0 1px rgba(59, 130, 246, 0.2);
}

.approval-item__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 6px;
}

.pill--medium {
  background: rgba(245, 158, 11, 0.16);
}

.pill--high {
  background: rgba(59, 130, 246, 0.16);
}

.pill--critical {
  background: rgba(239, 68, 68, 0.16);
}

.detail-stack {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.detail-meta {
  display: flex;
  justify-content: space-between;
  gap: 16px;
}

.detail-meta h3 {
  margin: 6px 0;
}

.detail-actions {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.meta-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.meta-card,
.reason-card,
.history-item {
  padding: 12px;
  border: 1px solid var(--stroke);
  border-radius: var(--radius-md);
  background: var(--panel-muted);
}

.reason-card p {
  margin: 8px 0 0;
}

.link-row {
  display: flex;
  gap: 14px;
  flex-wrap: wrap;
}

.history {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.history-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

@media (max-width: 1180px) {
  .approvals-shell {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 860px) {
  .filters,
  .meta-grid {
    grid-template-columns: 1fr;
  }

  .detail-meta {
    flex-direction: column;
  }

  .detail-actions {
    flex-wrap: wrap;
  }
}
</style>
