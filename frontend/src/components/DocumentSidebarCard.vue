<script setup lang="ts">
import { Icon as IconifyIcon } from '@iconify/vue'

const emit = defineEmits([
  'search',
  'refresh',
  'add-folder',
  'upload-file',
  'nav-dashboard',
  'nav-recent',
  'nav-trash',
])

const props = defineProps<{
  hideSearch?: boolean
}>()
</script>

<template>
  <a-card class="search-card" :bordered="false">
    <div class="search-header">
      <h3 class="search-title">Search</h3>
      <div class="search-actions">
        <IconifyIcon
          v-if="!props.hideSearch"
          icon="mdi:magnify"
          width="20"
          class="action-icon"
          @click="emit('search')"
        />
        <IconifyIcon icon="mdi:refresh" width="20" class="action-icon" @click="emit('refresh')" />
      </div>
    </div>
    <a-dropdown>
      <template #overlay>
        <a-menu>
          <a-menu-item key="add-folder" @click="emit('add-folder')">
            <IconifyIcon icon="mdi:folder-plus-outline" width="18" style="margin-right: 8px;" />
            Add new folder
          </a-menu-item>
          <a-menu-item key="upload-file" @click="emit('upload-file')">
            <IconifyIcon icon="mdi:file-plus-outline" width="18" style="margin-right: 8px;" />
            Add new file
          </a-menu-item>
        </a-menu>
      </template>
      <a-button type="primary" danger block class="new-folder-btn">
        <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
        Baru
      </a-button>
    </a-dropdown>
    <div class="nav-links">
      <a-button type="text" block class="nav-link" @click="emit('nav-dashboard')">
        <IconifyIcon icon="mdi:view-dashboard-outline" width="16" style="margin-right: 8px;" />
        Dashboard
      </a-button>
      <!-- <a-button type="text" block class="nav-link" @click="emit('nav-recent')">
        <IconifyIcon icon="mdi:clock-outline" width="16" style="margin-right: 8px;" />
        Recent
      </a-button>
      <a-button type="text" block class="nav-link" @click="emit('nav-trash')">
        <IconifyIcon icon="mdi:trash-can-outline" width="16" style="margin-right: 8px;" />
        Trash
      </a-button> -->
    </div>
  </a-card>
</template>

<style scoped>
.search-card {
  /* background: #fff;
  border-radius: 12px;
  /* box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05); */
  /* padding: 12px;  */
}

.search-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.search-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #333;
}

.search-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.action-icon {
  color: #666;
  cursor: pointer;
  transition: color 0.2s;
}

.action-icon:hover {
  color: #035cab;
}

.new-folder-btn {
  width: 100%;
  height: 35px;
  margin-bottom: 16px;
  /* background: #db241b;
  border-color: #db241b; */
  font-weight: 500;
  display: flex;
  align-items: center;
  justify-content: center;
}

.new-folder-btn:hover {
  background: #c41d1a;
  border-color: #c41d1a;
}

.nav-links {
  margin-top: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.nav-link {
  text-align: left;
  padding-left: 4px;
  padding-right: 4px;
  color: #444;
}

.nav-link:hover {
  color: #035cab;
}
</style>
