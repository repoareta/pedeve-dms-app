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
  /* background: orange !important; */
  
}

.activity-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  margin-bottom: -10px;
  
}

.activity-title {
  font-size: 14px;
  font-weight: 600;
}

.see-all-btn {
  padding: 0;
}

.activity-list {
  padding: 3px 0;
}

.activity-timeline {
  position: relative;
}

.activity-item {
  display: flex;
  gap: 10px;
  position: relative;
  padding: 10px 0;
}

.activity-avatar {
  width: 25px;
  height: 25px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-weight: 600;
  font-size: 12px;
  margin:0 0 0 5px;
}

.activity-content {
  flex: 1;
}

.activity-user {
  font-weight: 600;
  font-size: 12px;
  text-transform: lowercase;
}

.activity-description {
  color: #555;
  font-size: 12px;
}

.activity-time {
  color: #555555;
  font-size: 12px;
}

.activity-line {
  position: absolute;
  left: 18px;
  top: 48px;
  width: 1px;
  height: calc(100% - 48px);
  background: #ddd;
}

.activity-skeleton,
.activity-empty {
  padding: 16px;
}

.activity-empty {
  text-align: center;
  color: #888;
}
</style>
