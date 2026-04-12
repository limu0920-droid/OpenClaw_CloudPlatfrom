<script setup lang="ts">
import { Tab, TabGroup, TabList, TabPanel, TabPanels } from '@headlessui/vue'
import { useRouter } from 'vue-router'

type Plane = {
  name: string
  subtitle: string
  description: string
  stats: string[]
  bullets: string[]
  tone: string
  to: string
  action: string
}

const router = useRouter()

const planes: Plane[] = [
  {
    name: 'Portal',
    subtitle: '租户侧门户',
    description: '保持轻量、可信和低学习成本，把实例、渠道、备份、任务和工单组织成一条业务主链路。',
    stats: ['实例视图', '渠道接入', '备份与恢复'],
    bullets: [
      '实例概览、访问入口、配置发布与购买动作集中在一个视图里。',
      '渠道接入和工单提交是用户最直接的动作面。',
      '适合业务用户和交付同学快速定位运行状态。',
    ],
    tone: 'tone-portal',
    to: '/portal',
    action: '进入 Portal',
  },
  {
    name: 'Admin',
    subtitle: '平台控制面',
    description: '强调状态密度、风险控制和运维节奏，围绕租户、任务、告警、审计和平台资源分层展开。',
    stats: ['租户治理', '任务编排', '审计检索'],
    bullets: [
      '深色控制面更适合长时间操作和状态型任务。',
      '审计、告警和平台级任务集中处理，不和租户入口混杂。',
      '后续接真 IAM、搜索和 runtime 后仍可平滑扩展。',
    ],
    tone: 'tone-admin',
    to: '/admin',
    action: '进入 Admin',
  },
]

function go(to: string) {
  void router.push(to)
}
</script>

<template>
  <TabGroup>
    <div class="tabs-shell">
      <TabList class="tab-list">
        <Tab v-for="plane in planes" :key="plane.name" as="template" v-slot="{ selected }">
          <button :class="['tab-trigger', selected ? 'selected' : '', plane.tone]">
            <span class="tab-sub">{{ plane.subtitle }}</span>
            <strong>{{ plane.name }}</strong>
          </button>
        </Tab>
      </TabList>

      <TabPanels class="tab-panels">
        <TabPanel v-for="plane in planes" :key="plane.name" class="tab-panel">
          <div :class="['panel-hero', plane.tone]">
            <div class="panel-copy">
              <div class="panel-subtitle">{{ plane.subtitle }}</div>
              <h3>{{ plane.name }}</h3>
              <p>{{ plane.description }}</p>
              <button class="panel-cta" type="button" @click="go(plane.to)">{{ plane.action }}</button>
            </div>
            <div class="panel-stats">
              <div v-for="stat in plane.stats" :key="stat" class="stat-chip">{{ stat }}</div>
            </div>
          </div>

          <div class="panel-details">
            <article v-for="bullet in plane.bullets" :key="bullet" class="detail-card">
              <p>{{ bullet }}</p>
            </article>
          </div>
        </TabPanel>
      </TabPanels>
    </div>
  </TabGroup>
</template>

<style scoped>
.tabs-shell {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  gap: 16px;
}

.tab-list {
  display: grid;
  gap: 10px;
}

.tab-trigger {
  display: flex;
  flex-direction: column;
  gap: 6px;
  width: 100%;
  padding: 18px 20px;
  text-align: left;
  border-radius: 22px;
  border: 1px solid rgba(148, 163, 184, 0.18);
  background: rgba(255, 255, 255, 0.74);
  transition:
    transform 0.18s ease,
    border-color 0.18s ease,
    background 0.18s ease;
}

.tab-trigger:hover,
.tab-trigger.selected {
  transform: translateX(2px);
  border-color: rgba(29, 107, 255, 0.22);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(240, 245, 255, 0.96));
}

.tab-sub {
  font-size: 12px;
  color: var(--text-muted);
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.tab-trigger strong {
  font-size: 1.2rem;
  font-variation-settings: "wght" 660;
}

.tab-panels,
.tab-panel {
  min-width: 0;
}

.panel-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 220px;
  gap: 16px;
  padding: 22px;
  border-radius: 28px;
  border: 1px solid rgba(148, 163, 184, 0.16);
}

.panel-copy {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.panel-subtitle {
  font-size: 12px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.panel-copy h3 {
  margin: 0;
  font-size: 2rem;
}

.panel-copy p {
  margin: 0;
  color: var(--text-muted);
  line-height: 1.8;
}

.panel-cta {
  width: fit-content;
  padding: 12px 18px;
  border-radius: 999px;
  border: 1px solid transparent;
  background: linear-gradient(120deg, #0f6bff, #5c7bff 65%, #1ad1ff);
  color: #fff;
  font-variation-settings: "wght" 620;
}

.panel-stats {
  display: grid;
  gap: 10px;
  align-content: start;
}

.stat-chip {
  padding: 14px 16px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.72);
  border: 1px solid rgba(148, 163, 184, 0.14);
  font-variation-settings: "wght" 600;
}

.panel-details {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  margin-top: 14px;
}

.detail-card {
  padding: 18px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid rgba(148, 163, 184, 0.14);
}

.detail-card p {
  margin: 0;
  line-height: 1.75;
  color: var(--text);
}

.tone-portal {
  background:
    radial-gradient(circle at top left, rgba(29, 107, 255, 0.18), transparent 24%),
    linear-gradient(145deg, rgba(255, 255, 255, 0.94), rgba(240, 247, 255, 0.98));
}

.tone-admin {
  color: #edf3ff;
  background:
    radial-gradient(circle at top right, rgba(59, 130, 246, 0.22), transparent 18%),
    linear-gradient(160deg, #10203a, #152a46 58%, #0f223b);
}

.tone-admin .panel-copy p,
.tone-admin .panel-subtitle {
  color: rgba(226, 232, 240, 0.76);
}

.tone-admin .stat-chip,
.tone-admin.detail-card {
  color: #edf3ff;
}

.tone-admin .stat-chip {
  background: rgba(8, 15, 29, 0.52);
  border-color: rgba(148, 163, 184, 0.12);
}

@media (max-width: 980px) {
  .tabs-shell,
  .panel-hero,
  .panel-details {
    grid-template-columns: 1fr;
  }
}
</style>
