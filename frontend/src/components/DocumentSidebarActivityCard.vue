<script setup lang="ts">
import { Icon as IconifyIcon } from '@iconify/vue'
import DocumentSidebarCard from './DocumentSidebarCard.vue'
import type { UserActivityLog } from '../api/audit'

defineProps<{
  activities: UserActivityLog[]
  activityLoading?: boolean
  pageLoading?: boolean
  hideSearch?: boolean
  showSeeAll?: boolean
  getDisplayName: (username: string) => string
  getActivityDescription: (activity: UserActivityLog) => string
  formatTime: (timestamp: string) => string
  getUserAvatarColor: (username: string) => string
}>()

const emit = defineEmits([
  'search',
  'refresh',
  'add-folder',
  'upload-file',
  'nav-dashboard',
  'nav-recent',
  'nav-trash',
  'see-all',
])
</script>

<template>
  <div class="sidebar-stack">
    <DocumentSidebarCard
      :hide-search="hideSearch"
      @search="emit('search')"
      @refresh="emit('refresh')"
      @add-folder="emit('add-folder')"
      @upload-file="emit('upload-file')"
      @nav-dashboard="emit('nav-dashboard')"
      @nav-recent="emit('nav-recent')"
      @nav-trash="emit('nav-trash')"
    />

    <a-card class="activity-card" :bordered="false">
      <div class="activity-header">
        <h3 class="activity-title">Activity</h3>
        <a-button v-if="showSeeAll !== false" type="link" class="see-all-btn" @click="emit('see-all')">
          See All
          <IconifyIcon icon="mdi:chevron-right" width="16" style="margin-left: 4px;" />
        </a-button>
      </div>
      <div class="activity-list" v-if="!(activityLoading || pageLoading) && activities.length > 0">
        <div class="activity-timeline">
          <div
            v-for="(activity, index) in activities"
            :key="activity.id"
            class="activity-item"
          >
            <div class="activity-avatar" :style="{ backgroundColor: getUserAvatarColor(activity.username) }">
              {{ getDisplayName(activity.username).charAt(0).toUpperCase() }}
            </div>
            <div class="activity-content">
              <div class="activity-time">{{ formatTime(activity.created_at) }}</div>
              <div class="activity-description"><strong>{{ getDisplayName(activity.username) }}</strong> {{ getActivityDescription(activity) }}</div>
              
            </div>
            <div v-if="index < activities.length - 1" class="activity-line"></div>
          </div>
        </div>
      </div>
      <div v-else-if="activityLoading || pageLoading" class="activity-skeleton">
        <a-skeleton active :paragraph="{ rows: 5 }" />
      </div>
      <div v-else class="activity-empty">
        <p>No activities yet</p>
      </div>
    </a-card>
  </div>
</template>

<style scoped>
.activity-card {
  margin-top: 16px;
}

.activity-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
  padding-bottom: -0px;
  /* border-bottom: 1px solid #ff6b35; */
}

.activity-title {
  font-size: 14px;
  font-weight: 600;
  margin: 0;
  color: #333;
}

.see-all-btn {
  padding: 0;
  height: auto;
  color: #1890ff;
}

.see-all-btn:hover {
  color: #40a9ff;
}

.activity-list {
  padding: 8px 0;
}

.activity-timeline {
  position: relative;
  padding-left: 8px;
}

.activity-item {
  display: flex;
  gap: 12px;
  position: relative;
  padding: 12px 0;
  padding-left: 4px;
}

.activity-item:not(:last-child) {
  /* border-bottom: 1px solid #f0f0f0; */
}

.activity-avatar {
  width: 32px !important;
  height: 32px !important;
  min-width: 32px;
  min-height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 600;
  font-size: 13px;
  flex-shrink: 0;
  position: relative;
  z-index: 1;
}

.activity-content {
  flex: 1;
  min-width: 0;
  padding-right: 8px;
}

.activity-time {
  color: #8c8c8c;
  font-size: 11px;
  margin-bottom: 4px;
  line-height: 1.4;
}

.activity-description {
  color: #333;
  font-size: 12px;
  line-height: 1.5;
  word-wrap: break-word;
  overflow-wrap: break-word;
}

.activity-description strong {
  color: #1890ff;
  font-weight: 600;
}

.activity-line {
  position: absolute;
  left: 16px;
  top: 44px;
  bottom: 0;
  width: 2px;
  background: #e8e8e8;
  z-index: 0;
}

.activity-item:last-child .activity-line {
  display: none;
}

.activity-skeleton,
.activity-empty {
  padding: 16px;
}

.activity-empty {
  text-align: center;
  color: #8c8c8c;
  font-size: 12px;
}
</style>
