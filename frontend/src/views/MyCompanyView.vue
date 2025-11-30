<template>
  <div class="my-company-layout">
    <DashboardHeader @logout="handleLogout" />

    <div class="detail-content">
      <!-- Loading State -->
      <div v-if="loading || loadingCompanies" class="loading-container">
        <a-spin size="large" />
      </div>

      <!-- Company Selector (if multiple companies) -->
      <div v-else-if="showCompanySelector && allUserCompanies.length > 1" class="company-selector-container">
        <div class="selector-header">
          <h1 class="selector-title">My Company</h1>
          <p class="selector-description">Anda di-assign ke beberapa perusahaan. Pilih perusahaan yang ingin dilihat:</p>
        </div>
        <div class="company-cards-grid">
          <div
            v-for="comp in sortedUserCompanies"
            :key="comp.company.id"
            class="company-selector-card"
            @click="selectCompany(comp.company.id)"
          >
            <!-- Card Header -->
            <div class="selector-card-header">
              <div class="selector-company-icon">
                <img v-if="getCompanyLogo(comp.company)" :src="getCompanyLogo(comp.company)" :alt="comp.company.name" class="selector-logo" />
                <div v-else class="selector-icon-placeholder" :style="{ backgroundColor: getIconColor(comp.company.name) }">
                  {{ getCompanyInitial(comp.company.name) }}
                </div>
              </div>
              <div class="selector-company-info">
                <h3 class="selector-company-name">{{ comp.company.name }}</h3>
                <p class="selector-company-reg">No Reg {{ comp.company.nib || 'N/A' }}</p>
                <div class="selector-company-meta">
                  <a-tag :color="getLevelColor(comp.company.level)" size="small">
                    {{ getLevelLabel(comp.company.level) }}
                  </a-tag>
                  <a-tag v-if="comp.role" :color="getRoleColor(comp.role)" size="small">
                    {{ comp.role }}
                  </a-tag>
                  <a-tag :color="comp.company.is_active ? 'green' : 'red'" size="small">
                    {{ comp.company.is_active ? 'Aktif' : 'Tidak Aktif' }}
                  </a-tag>
                </div>
              </div>
            </div>

            <!-- Card Divider -->
            <div class="selector-card-divider"></div>

            <!-- Card Content -->
            <div class="selector-card-content">
              <div class="latest-month-header">
                <IconifyIcon icon="mdi:information-outline" width="16" style="margin-right: 4px;" />
                <span>Latest Month</span>
              </div>

              <div class="metrics-row">
                <!-- RKAP vs Realization -->
                <div class="metric-item">
                  <div class="metric-value">{{ formatCurrency(getCompanyRKAP(comp.company.id)) }}</div>
                  <div class="metric-meta">
                    <span class="metric-year">{{ getCompanyRKAPYear(comp.company.id) }}</span>
                    <span class="metric-change positive">+{{ getCompanyRKAPChange(comp.company.id) }}%</span>
                  </div>
                  <div class="metric-label">RKAP vs Realization</div>
                </div>

                <!-- Opex Trend -->
                <div class="metric-item">
                  <div class="metric-value">{{ formatCurrency(getCompanyOpex(comp.company.id)) }}</div>
                  <div class="metric-meta">
                    <span class="metric-quarter">{{ getCompanyOpexQuarter(comp.company.id) }}</span>
                    <span class="metric-change negative">-{{ getCompanyOpexChange(comp.company.id) }}%</span>
                  </div>
                  <div class="metric-label">Opex Trend</div>
                </div>
              </div>
            </div>

            <!-- Card Footer -->
            <div class="selector-card-footer">
              <a-button type="link" class="learn-more-link" @click.stop="selectCompany(comp.company.id)">
                Learn more
                <IconifyIcon icon="mdi:arrow-right" width="16" style="margin-left: 4px;" />
              </a-button>
            </div>
          </div>
        </div>
      </div>

      <!-- Company Detail -->
      <div v-else-if="company" class="detail-card">
        <div class="page-header-container">
          <!-- Header Section -->
          <div class="detail-header">
            <div class="header-left">
              <div class="company-icon-large">
                <img v-if="getCompanyLogo(company)" :src="getCompanyLogo(company)" :alt="company.name"
                  class="logo-image-large" />
                <div v-else class="icon-placeholder-large" :style="{ backgroundColor: getIconColor(company.name) }">
                  {{ getCompanyInitial(company.name) }}
                </div>
              </div>
              <div class="header-info">
                <h1 class="company-title">{{ company.name }}</h1>
                <p class="company-subtitle">{{ company.short_name || company.name }}</p>
                <div class="company-meta">
                  <a-tag :color="company.is_active ? 'green' : 'red'">
                    {{ company.is_active ? 'Aktif' : 'Tidak Aktif' }}
                  </a-tag>
                  <a-tag :color="getLevelColor(company.level)">
                    {{ getLevelLabel(company.level) }}
                  </a-tag>
                  <span v-if="company.code" class="meta-item">Kode: {{ company.code }}</span>
                  <span v-if="company.nib" class="meta-item">No Reg {{ company.nib }}</span>
                </div>
              </div>
            </div>
            <div class="header-right">
              <div class="header-actions-top">
                <a-input-search placeholder="Cari..." style="width: 200px;" />
                <a-button type="default">
                  <IconifyIcon icon="mdi:view-grid" width="16" />
                </a-button>
                <a-button type="default">
                  <IconifyIcon icon="mdi:format-list-bulleted" width="16" />
                </a-button>
              </div>
              <div class="header-actions-bottom">
                <a-button v-if="currentUserRole" type="link" size="small" class="role-button">
                  <IconifyIcon icon="mdi:account-tie" width="16" style="margin-right: 4px;" />
                  Role: {{ currentUserRole }}
                </a-button>
                <a-dropdown v-if="allUserCompanies.length > 1" trigger="click">
                  <template #overlay>
                    <a-menu @click="(e: { key: string }) => selectCompany(e.key)">
                      <a-menu-item v-for="comp in sortedUserCompanies" :key="comp.company.id" :value="comp.company.id">
                        {{ comp.company.name }}
                        <a-tag :color="getLevelColor(comp.company.level)" style="margin-left: 8px;">
                          {{ getLevelLabel(comp.company.level) }}
                        </a-tag>
                        <a-tag v-if="comp.role" :color="getRoleColor(comp.role)" style="margin-left: 4px;">
                          {{ comp.role }}
                        </a-tag>
                      </a-menu-item>
                    </a-menu>
                  </template>
                  <a-button type="link" size="small">
                    <IconifyIcon icon="mdi:swap-horizontal" width="16" style="margin-right: 4px;" />
                    Ganti Company
                    <IconifyIcon icon="mdi:chevron-down" width="16" style="margin-left: 4px;" />
                  </a-button>
                </a-dropdown>
                <a-button v-if="allUserCompanies.length > 1" type="link" size="small" @click="showSelector">
                  <IconifyIcon icon="mdi:view-grid" width="16" style="margin-right: 4px;" />
                  Lihat Semua
                </a-button>
              </div>
            </div>
          </div>
          
          <!-- Action Buttons Row -->
          <div class="action-buttons-row">
            <a-space>
              <a-button @click="handleExportPDF">
                <IconifyIcon icon="mdi:file-pdf-box" width="16" style="margin-right: 8px;" />
                PDF
              </a-button>
              <a-button @click="handleExportExcel">
                <IconifyIcon icon="mdi:file-excel" width="16" style="margin-right: 8px;" />
                Excel
              </a-button>
              <a-date-picker v-model:value="selectedPeriod" picker="month" placeholder="Select Periode"
                style="width: 150px;" />
              <a-dropdown v-if="hasAnyMenuOption">
                <template #overlay>
                  <a-menu @click="handleMenuClick">
                    <a-menu-item v-if="canAddSubsidiary" key="add-subsidiary">
                      <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
                      Add Subsidiary
                    </a-menu-item>
                    <a-menu-item v-if="canEdit" key="edit">
                      <IconifyIcon icon="mdi:pencil" width="16" style="margin-right: 8px;" />
                      Edit
                    </a-menu-item>
                    <a-menu-item v-if="canAssignRole" key="assign-role">
                      <IconifyIcon icon="mdi:account-plus" width="16" style="margin-right: 8px;" />
                      Assign Role
                    </a-menu-item>
                  </a-menu>
                </template>
                <a-button>
                  <IconifyIcon icon="mdi:dots-vertical" width="16" style="margin-right: 8px;" />
                  Options
                </a-button>
              </a-dropdown>
            </a-space>
          </div>
        </div>

        <!-- Tabs -->
        <div class="tabs-container">
          <a-tabs v-model:activeKey="activeTab" type="card" size="large">
            <a-tab-pane key="performance" tab="Performance">
              <!-- Performance Tab Content -->
              <div class="performance-content">
                <!-- Financial Trend Cards -->
                <div class="trend-cards-row">
                  <!-- RKAP vs Realization Card -->
                  <a-card class="trend-card" :bordered="false">
                    <template #title>
                      <span class="card-title">RKAP vs Realization</span>
                    </template>
                    <div class="trend-card-content">
                      <div class="trend-metric">
                        <span class="metric-value">{{ formatCurrency(rkapData.value) }}</span>
                        <div class="trend-meta">
                          <span class="trend-period">{{ rkapData.year }}</span>
                          <span class="trend-change positive">+{{ rkapData.change }}%</span>
                        </div>
                      </div>
                      <div class="mini-chart-container">
                        <svg width="100%" height="60" viewBox="0 0 200 60" class="mini-chart">
                          <defs>
                            <linearGradient id="rkapGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                              <stop offset="0%" style="stop-color:#ff9800;stop-opacity:0.3" />
                              <stop offset="100%" style="stop-color:#ff9800;stop-opacity:0.05" />
                            </linearGradient>
                          </defs>
                          <path :d="rkapChartFillPath" fill="url(#rkapGradient)" class="chart-fill" />
                          <path :d="rkapChartPath" stroke="#ff9800" stroke-width="2" fill="none" class="chart-line" />
                        </svg>
                        <div class="chart-labels">
                          <span>Jan</span>
                          <span>Des</span>
                        </div>
                      </div>
                    </div>
                  </a-card>

                  <!-- Opex Trend Card -->
                  <a-card class="trend-card" :bordered="false">
                    <template #title>
                      <span class="card-title">Opex Trend</span>
                    </template>
                    <div class="trend-card-content">
                      <div class="trend-metric">
                        <span class="metric-value">{{ formatCurrency(opexData.value) }}</span>
                        <div class="trend-meta">
                          <span class="trend-period">{{ opexData.quarter }}</span>
                          <span class="trend-change negative">-{{ opexData.change }}%</span>
                        </div>
                      </div>
                      <div class="mini-chart-container">
                        <svg width="100%" height="60" viewBox="0 0 200 60" class="mini-chart">
                          <defs>
                            <linearGradient id="opexGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                              <stop offset="0%" style="stop-color:#666;stop-opacity:0.3" />
                              <stop offset="100%" style="stop-color:#666;stop-opacity:0.05" />
                            </linearGradient>
                          </defs>
                          <path :d="opexChartFillPath" fill="url(#opexGradient)" class="chart-fill" />
                          <path :d="opexChartPath" stroke="#666" stroke-width="2" fill="none" class="chart-line" />
                        </svg>
                        <div class="chart-labels">
                          <span>Jan</span>
                          <span>Des</span>
                        </div>
                      </div>
                    </div>
                  </a-card>

                  <!-- NPAT Trend Card -->
                  <a-card class="trend-card" :bordered="false">
                    <template #title>
                      <span class="card-title">NPAT Trend</span>
                    </template>
                    <div class="trend-card-content">
                      <div class="trend-metric">
                        <span class="metric-value">{{ formatCurrency(npatData.value) }}</span>
                        <div class="trend-meta">
                          <span class="trend-period">{{ npatData.quarter }}</span>
                          <span class="trend-change positive">+{{ npatData.change }}%</span>
                        </div>
                      </div>
                      <div class="mini-chart-container">
                        <svg width="100%" height="60" viewBox="0 0 200 60" class="mini-chart">
                          <defs>
                            <linearGradient id="npatGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                              <stop offset="0%" style="stop-color:#666;stop-opacity:0.3" />
                              <stop offset="100%" style="stop-color:#666;stop-opacity:0.05" />
                            </linearGradient>
                          </defs>
                          <path :d="npatChartFillPath" fill="url(#npatGradient)" class="chart-fill" />
                          <path :d="npatChartPath" stroke="#666" stroke-width="2" fill="none" class="chart-line" />
                        </svg>
                        <div class="chart-labels">
                          <span>Jan</span>
                          <span>Des</span>
                        </div>
                      </div>
                    </div>
                  </a-card>
                </div>

                <!-- Recent Files and Reports -->
                <div class="recent-section">
                  <!-- Recent Files -->
                  <a-card class="recent-card" :bordered="false">
                    <template #title>
                      <div class="card-header-title">
                        <IconifyIcon icon="mdi:clock-outline" width="20" style="margin-right: 8px;" />
                        <span>Recently Files</span>
                      </div>
                    </template>
                    <template #extra>
                      <a-button type="link" @click="handleManageFiles">
                        Manage file upload
                        <IconifyIcon icon="mdi:arrow-right" width="16" style="margin-left: 4px;" />
                      </a-button>
                    </template>
                    <a-table :columns="fileColumns" :data-source="recentFiles" :pagination="false" :show-header="true"
                      size="small">
                      <template #bodyCell="{ column, record }">
                        <template v-if="column.key === 'type'">
                          <a-tag :color="record.type === 'Pdf' ? 'red' : 'green'">{{ record.type }}</a-tag>
                        </template>
                        <template v-if="column.key === 'status'">
                          <a-tag v-if="record.status === 'complete'" color="green">Meta Data ✓</a-tag>
                          <a-button v-else type="link" size="small">
                            Lengkapi Meta Data
                            <IconifyIcon icon="mdi:arrow-right" width="14" style="margin-left: 4px;" />
                          </a-button>
                        </template>
                        <template v-if="column.key === 'action'">
                          <IconifyIcon icon="mdi:chevron-right" width="20" style="color: #999; cursor: pointer;" />
                        </template>
                      </template>
                    </a-table>
                  </a-card>

                  <!-- Recent Reports -->
                  <a-card class="recent-card" :bordered="false">
                    <template #title>
                      <div class="card-header-title">
                        <IconifyIcon icon="mdi:clock-outline" width="20" style="margin-right: 8px;" />
                        <span>Recently Reports</span>
                      </div>
                    </template>
                    <template #extra>
                      <a-button type="link" @click="handleManageReports">
                        Manage Reports
                        <IconifyIcon icon="mdi:arrow-right" width="16" style="margin-left: 4px;" />
                      </a-button>
                    </template>
                    <a-table :columns="reportColumns" :data-source="recentReports" :pagination="false" :show-header="true"
                      size="small">
                      <template #bodyCell="{ column, record }">
                        <template v-if="column.key === 'rkap_percent'">
                          {{ record.rkap_percent }}%
                        </template>
                        <template v-if="column.key === 'action'">
                          <IconifyIcon icon="mdi:chevron-right" width="20" style="color: #999; cursor: pointer;" />
                        </template>
                      </template>
                    </a-table>
                  </a-card>
                </div>
              </div>
            </a-tab-pane>

            <a-tab-pane key="profile" tab="Profile">
              <!-- Profile Tab Content -->
              <div v-if="company" class="profile-content">
                <!-- Informasi Dasar -->
                <div class="detail-section">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:information" width="20" style="margin-right: 8px;" />
                    Informasi Dasar
                  </h2>
                  <a-descriptions :column="2" bordered>
                    <a-descriptions-item label="Nama Lengkap">{{ company.name }}</a-descriptions-item>
                    <a-descriptions-item label="Nama Singkat">{{ company.short_name || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="Kode Perusahaan">{{ company.code }}</a-descriptions-item>
                    <a-descriptions-item label="Status">{{ company.status || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="NPWP">{{ company.npwp || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="NIB">{{ company.nib || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="Deskripsi" :span="2">
                      {{ company.description || '-' }}
                    </a-descriptions-item>
                  </a-descriptions>
                </div>

                <!-- Informasi Kontak -->
                <div class="detail-section">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:phone" width="20" style="margin-right: 8px;" />
                    Informasi Kontak
                  </h2>
                  <a-descriptions :column="2" bordered>
                    <a-descriptions-item label="Telepon">{{ company.phone || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="Fax">{{ company.fax || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="Email">{{ company.email || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="Website">{{ company.website || '-' }}</a-descriptions-item>
                    <a-descriptions-item label="Alamat Perusahaan" :span="2">
                      {{ company.address || '-' }}
                    </a-descriptions-item>
                    <a-descriptions-item label="Alamat Operasional" :span="2">
                      {{ company.operational_address || '-' }}
                    </a-descriptions-item>
                  </a-descriptions>
                </div>

                <!-- Struktur Kepemilikan -->
                <div v-if="company.shareholders && company.shareholders.length > 0" class="detail-section">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:account-group" width="20" style="margin-right: 8px;" />
                    Struktur Kepemilikan
                  </h2>
                  <a-table :columns="shareholderColumns" :data-source="company.shareholders" :pagination="false"
                    row-key="id">
                    <template #bodyCell="{ column, record }">
                      <template v-if="column.key === 'ownership_percent'">
                        {{ record.ownership_percent }}%
                      </template>
                      <template v-if="column.key === 'share_count'">
                        {{ record.share_count?.toLocaleString() || '-' }}
                      </template>
                      <template v-if="column.key === 'is_main_parent'">
                        <a-tag v-if="record.is_main_parent" color="blue">Ya</a-tag>
                        <span v-else>-</span>
                      </template>
                    </template>
                  </a-table>
                </div>

                <!-- Bidang Usaha -->
                <div v-if="company.main_business || (company.business_fields && company.business_fields.length > 0)"
                  class="detail-section">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:briefcase" width="20" style="margin-right: 8px;" />
                    Bidang Usaha
                  </h2>
                  <a-descriptions :column="1" bordered>
                    <a-descriptions-item label="Sektor Industri">
                      {{ getMainBusiness(company)?.industry_sector || '-' }}
                    </a-descriptions-item>
                    <a-descriptions-item label="KBLI">
                      {{ getMainBusiness(company)?.kbli || '-' }}
                    </a-descriptions-item>
                    <a-descriptions-item label="Uraian Kegiatan Usaha Utama">
                      {{ getMainBusiness(company)?.main_business_activity || '-' }}
                    </a-descriptions-item>
                    <a-descriptions-item label="Kegiatan Usaha Tambahan">
                      {{ getMainBusiness(company)?.additional_activities || '-' }}
                    </a-descriptions-item>
                    <a-descriptions-item label="Tanggal Mulai Beroperasi">
                      {{ formatDate(getMainBusiness(company)?.start_operation_date) }}
                    </a-descriptions-item>
                  </a-descriptions>
                </div>

                <!-- Pengurus/Dewan Direksi -->
                <div v-if="company.directors && company.directors.length > 0" class="detail-section">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:account-tie" width="20" style="margin-right: 8px;" />
                    Pengurus/Dewan Direksi
                  </h2>
                  <a-table :columns="directorColumns" :data-source="company.directors" :pagination="false" row-key="id">
                    <template #bodyCell="{ column, record }">
                      <template v-if="column.key === 'start_date'">
                        {{ record.start_date ? formatDate(record.start_date) : '-' }}
                      </template>
                    </template>
                  </a-table>
                </div>
              </div>
            </a-tab-pane>
          </a-tabs>
        </div>
      </div>

      <!-- Not Found / No Company -->
      <div v-else class="not-found">
        <IconifyIcon icon="mdi:alert-circle-outline" width="64" style="color: #ccc; margin-bottom: 16px;" />
        <p>Perusahaan tidak ditemukan atau Anda belum di-assign ke perusahaan</p>
        <a-button type="primary" @click="router.push('/dashboard')">Kembali ke Dashboard</a-button>
      </div>

      <!-- Assign Role Modal -->
      <a-modal
        v-model:open="assignRoleModalVisible"
        title="Assign Role - Kelola Pengurus"
        :confirm-loading="assignRoleLoading"
        width="900px"
        :footer="null"
      >
        <div class="assign-role-container">
          <!-- Form Assign Role Baru -->
          <div class="assign-new-section">
            <h3 class="section-header">
              <IconifyIcon icon="mdi:account-plus" width="20" style="margin-right: 8px;" />
              Assign Role Baru
            </h3>
            <a-form :model="assignRoleForm" layout="vertical">
              <a-row :gutter="16">
                <a-col :span="12">
                  <a-form-item label="Cari User" required>
                    <a-select
                      v-model:value="assignRoleForm.userId"
                      show-search
                      placeholder="Cari user berdasarkan nama atau email"
                      :filter-option="filterUserOption"
                      :loading="usersLoading"
                      @search="handleUserSearch"
                      allow-clear
                      :disabled="usersLoading"
                    >
                      <a-select-option
                        v-for="user in filteredUsers"
                        :key="user.id"
                        :value="user.id"
                        :disabled="companyUsers.some(u => u.id === user.id)"
                      >
                        {{ user.username }} ({{ user.email }})
                        <span v-if="companyUsers.some(u => u.id === user.id)" class="text-muted"> - Sudah menjadi pengurus</span>
                      </a-select-option>
                    </a-select>
                    <small v-if="usersLoading" class="text-muted">Memuat daftar user...</small>
                    <small v-else-if="allUsers.length === 0 && !usersLoading" class="text-muted">
                      Tidak ada user yang tersedia untuk di-assign.
                    </small>
                  </a-form-item>
                </a-col>
                <a-col :span="12">
                  <a-form-item label="Pilih Role" required>
                    <a-select
                      v-model:value="assignRoleForm.roleId"
                      show-search
                      placeholder="Cari role"
                      :filter-option="filterRoleOption"
                      :loading="rolesLoading"
                      @search="handleRoleSearch"
                      allow-clear
                      :disabled="rolesLoading"
                    >
                      <a-select-option
                        v-for="role in filteredRoles"
                        :key="role.id"
                        :value="role.id"
                      >
                        {{ role.name }}
                      </a-select-option>
                    </a-select>
                    <small v-if="rolesLoading" class="text-muted">Memuat daftar role...</small>
                  </a-form-item>
                </a-col>
              </a-row>
              
              <a-form-item>
                <a-button 
                  type="primary" 
                  :loading="assignRoleLoading" 
                  @click="handleAssignRole" 
                  :disabled="!assignRoleForm.userId || !assignRoleForm.roleId"
                >
                  <IconifyIcon icon="mdi:account-plus" width="16" style="margin-right: 8px;" />
                  Assign Role
                </a-button>
              </a-form-item>
            </a-form>
          </div>

          <a-divider>Pengurus Saat Ini</a-divider>

          <!-- List Pengurus Saat Ini -->
          <div class="current-users-section">
            <h3 class="section-header">
              <IconifyIcon icon="mdi:account-group" width="20" style="margin-right: 8px;" />
              Pengurus Saat Ini
            </h3>
            <a-table
              :columns="userColumns"
              :data-source="companyUsers"
              :loading="usersLoading"
              :pagination="{ pageSize: 10 }"
              row-key="id"
              size="middle"
            >
              <template #bodyCell="{ column, record }">
                <template v-if="column.key === 'role'">
                  <a-tag v-if="record.role" :color="getRoleColor(record.role)">
                    {{ record.role }}
                  </a-tag>
                  <span v-else class="text-muted">-</span>
                </template>
                <template v-if="column.key === 'status'">
                  <a-tag :color="record.is_active ? 'green' : 'red'">
                    {{ record.is_active ? 'Aktif' : 'Tidak Aktif' }}
                  </a-tag>
                </template>
                <template v-if="column.key === 'action'">
                  <a-space>
                    <a-button type="link" size="small" @click="handleEditUserRole(record)">
                      <IconifyIcon icon="mdi:pencil" width="16" />
                      Ubah Role
                    </a-button>
                    <a-button type="link" size="small" danger @click="handleRemoveUser(record)">
                      <IconifyIcon icon="mdi:delete" width="16" />
                      Hapus
                    </a-button>
                  </a-space>
                </template>
              </template>
            </a-table>
          </div>
        </div>
      </a-modal>

      <!-- Edit User Role Modal -->
      <a-modal
        v-model:open="editingUserRoleModalVisible"
        title="Ubah Role Pengurus"
        :confirm-loading="editingRoleLoading"
        @ok="handleSaveUserRole"
        @cancel="handleCancelEditUserRole"
        width="500px"
      >
        <a-form layout="vertical" v-if="editingUserRole">
          <a-form-item label="User">
            <a-input :value="getUserById(editingUserRole.userId)?.username" disabled />
          </a-form-item>
          <a-form-item label="Pilih Role Baru" required>
            <a-select
              v-model:value="editingUserRole.roleId"
              show-search
              placeholder="Cari role"
              :filter-option="filterRoleOption"
              :loading="rolesLoading"
              @search="handleRoleSearch"
            >
              <a-select-option
                v-for="role in filteredRoles"
                :key="role.id"
                :value="role.id"
              >
                {{ role.name }}
              </a-select-option>
            </a-select>
          </a-form-item>
        </a-form>
      </a-modal>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { companyApi, userApi, roleApi, type Company, type BusinessField, type User, type Role, type UserCompanyResponse } from '../api/userManagement'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'
import dayjs from 'dayjs'

const router = useRouter()
const authStore = useAuthStore()

// Computed: Check user roles (from authStore - global role)
const userRole = computed(() => {
  return authStore.user?.role?.toLowerCase() || ''
})

const isSuperAdmin = computed(() => userRole.value === 'superadmin')
const isAdmin = computed(() => userRole.value === 'admin')
const isManager = computed(() => userRole.value === 'manager')
const isStaff = computed(() => userRole.value === 'staff')

// RBAC: Assign Role hanya untuk admin
const canAssignRole = computed(() => isAdmin.value || isSuperAdmin.value)

// RBAC: Edit untuk semua role (staff, manager, admin, superadmin)
const canEdit = computed(() => isAdmin.value || isManager.value || isStaff.value || isSuperAdmin.value)

// RBAC: Add Subsidiary hanya untuk admin
const canAddSubsidiary = computed(() => isAdmin.value || isSuperAdmin.value)

// Check if any menu item is available (to show/hide Options dropdown)
const hasAnyMenuOption = computed(() => canEdit.value || canAssignRole.value || canAddSubsidiary.value)

const company = ref<Company | null>(null)
const currentUserRole = ref<string>('') // Role user di company yang sedang dilihat
const allUserCompanies = ref<UserCompanyResponse[]>([]) // All companies assigned to user with role info
const loading = ref(false)
const loadingCompanies = ref(false) // Loading state for fetching user companies
const activeTab = ref('performance')
const selectedPeriod = ref<string | null>(null)
const showCompanySelector = ref(false) // Show card selector if multiple companies

// Assign Role Modal
const assignRoleModalVisible = ref(false)
const assignRoleLoading = ref(false)
const assignRoleForm = ref({
  userId: undefined as string | undefined,
  roleId: undefined as string | undefined,
})

// Company Users (Pengurus)
const companyUsers = ref<User[]>([])
const allUsers = ref<User[]>([])
const allRoles = ref<Role[]>([])
const usersLoading = ref(false)
const rolesLoading = ref(false)
const userSearchText = ref('')
const roleSearchText = ref('')

// Editing user role
const editingUserRole = ref<{ userId: string; roleId: string | undefined } | null>(null)
const editingUserRoleModalVisible = ref(false)
const editingRoleLoading = ref(false)

const shareholderColumns = [
  { title: 'Jenis', dataIndex: 'type', key: 'type' },
  { title: 'Nama', dataIndex: 'name', key: 'name' },
  { title: 'Nomor Identitas', dataIndex: 'identity_number', key: 'identity_number' },
  { title: 'Persentase', key: 'ownership_percent' },
  { title: 'Jumlah Saham', key: 'share_count' },
  { title: 'Induk Utama', key: 'is_main_parent' },
]

const directorColumns = [
  { title: 'Jabatan', dataIndex: 'position', key: 'position' },
  { title: 'Nama Lengkap', dataIndex: 'full_name', key: 'full_name' },
  { title: 'KTP', dataIndex: 'ktp', key: 'ktp' },
  { title: 'NPWP', dataIndex: 'npwp', key: 'npwp' },
  { title: 'Tanggal Mulai', key: 'start_date' },
  { title: 'Alamat Domisili', dataIndex: 'domicile_address', key: 'domicile_address' },
]

const userColumns = [
  { title: 'Username', dataIndex: 'username', key: 'username' },
  { title: 'Email', dataIndex: 'email', key: 'email' },
  { title: 'Role', key: 'role' },
  { title: 'Status', key: 'status' },
  { title: 'Aksi', key: 'action', width: 200 },
]

// Dummy data untuk charts
const rkapData = ref({
  value: 120000000,
  year: '2025',
  change: 15
})

const opexData = ref({
  value: 80000000,
  quarter: 'Q1 2024',
  change: 5
})

const npatData = ref({
  value: 25000000,
  quarter: 'Q1 2024',
  change: 15
})

// Generate chart data
const generateChartData = (baseValue: number, variance: number = 0.2) => {
  const points = 12
  const data: number[] = []
  for (let i = 0; i < points; i++) {
    const random = (Math.random() - 0.5) * variance
    data.push(baseValue * (1 + random))
  }
  return data
}

const rkapChartData = computed(() => generateChartData(100, 0.3))
const opexChartData = computed(() => generateChartData(80, 0.2))
const npatChartData = computed(() => generateChartData(25, 0.25))

// Generate SVG path untuk chart
const generateChartPath = (data: number[], width: number = 200, height: number = 60) => {
  if (!data || data.length === 0) return ''
  const max = Math.max(...data)
  const min = Math.min(...data)
  const range = max - min || 1
  const stepX = width / (data.length - 1)
  
  const firstValue = data[0] ?? 0
  let path = `M 0 ${height - ((firstValue - min) / range) * height}`
  for (let i = 1; i < data.length; i++) {
    const value = data[i] ?? 0
    const x = i * stepX
    const y = height - ((value - min) / range) * height
    path += ` L ${x} ${y}`
  }
  return path
}

const generateChartFillPath = (data: number[], width: number = 200, height: number = 60) => {
  if (!data || data.length === 0) return ''
  const max = Math.max(...data)
  const min = Math.min(...data)
  const range = max - min || 1
  const stepX = width / (data.length - 1)
  
  let path = `M 0 ${height}`
  const firstValue = data[0] ?? 0
  path += ` L 0 ${height - ((firstValue - min) / range) * height}`
  for (let i = 1; i < data.length; i++) {
    const value = data[i] ?? 0
    const x = i * stepX
    const y = height - ((value - min) / range) * height
    path += ` L ${x} ${y}`
  }
  path += ` L ${width} ${height} Z`
  return path
}

const rkapChartPath = computed(() => generateChartPath(rkapChartData.value))
const rkapChartFillPath = computed(() => generateChartFillPath(rkapChartData.value))
const opexChartPath = computed(() => generateChartPath(opexChartData.value))
const opexChartFillPath = computed(() => generateChartFillPath(opexChartData.value))
const npatChartPath = computed(() => generateChartPath(npatChartData.value))
const npatChartFillPath = computed(() => generateChartFillPath(npatChartData.value))

// Dummy data untuk recent files
const recentFiles = ref([
  {
    key: '1',
    name: 'RUPS_Tahunan_2025',
    type: 'Pdf',
    lastModified: '2 hours ago',
    status: 'complete'
  },
  {
    key: '2',
    name: 'Laporan_Keuangan_Q1_2024',
    type: 'Excel',
    lastModified: '1 day ago',
    status: 'incomplete'
  },
  {
    key: '3',
    name: 'Dokumen_Legal_2024',
    type: 'Pdf',
    lastModified: '3 days ago',
    status: 'complete'
  }
])

// Dummy data untuk recent reports
const recentReports = ref([
  {
    key: '1',
    name: 'Laporan September',
    rkap_percent: 85,
    revenue: '$120M',
    npat: '$25M',
    opex: '$80M'
  },
  {
    key: '2',
    name: 'Laporan Agustus',
    rkap_percent: 82,
    revenue: '$115M',
    npat: '$23M',
    opex: '$78M'
  },
  {
    key: '3',
    name: 'Laporan Juli',
    rkap_percent: 88,
    revenue: '$125M',
    npat: '$27M',
    opex: '$82M'
  }
])

const fileColumns = [
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'Type', key: 'type' },
  { title: 'Last modified', dataIndex: 'lastModified', key: 'lastModified' },
  { title: '', key: 'status' },
  { title: '', key: 'action', width: 30 }
]

const reportColumns = [
  { title: 'Name', dataIndex: 'name', key: 'name' },
  { title: 'RKAP (%)', key: 'rkap_percent' },
  { title: 'Revenue', dataIndex: 'revenue', key: 'revenue' },
  { title: 'NPAT', dataIndex: 'npat', key: 'npat' },
  { title: 'Opex', dataIndex: 'opex', key: 'opex' },
  { title: '', key: 'action', width: 30 }
]

// Computed: sorted companies by role level (highest role first)
const sortedUserCompanies = computed(() => {
  return [...allUserCompanies.value].sort((a, b) => {
    // Sort by role_level (0=superadmin, 1=admin, 2=manager, 3=staff)
    // Semakin kecil level, semakin tinggi role
    return a.role_level - b.role_level
  })
})

// Load all companies assigned to user
const loadUserCompanies = async () => {
  loadingCompanies.value = true
  try {
    allUserCompanies.value = await userApi.getMyCompanies()
    
    // If no companies, show warning
    if (allUserCompanies.value.length === 0) {
      message.warning('Anda belum di-assign ke perusahaan')
      return
    }
    
    // If only 1 company, load it directly
    if (allUserCompanies.value.length === 1) {
      const comp = allUserCompanies.value[0]!
      await loadCompanyDetail(comp.company.id)
      currentUserRole.value = comp.role
      showCompanySelector.value = false
    } else {
      // Multiple companies: SELALU tampilkan selector (sesuai permintaan user)
      showCompanySelector.value = true
      company.value = null
      currentUserRole.value = ''
    }
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal memuat daftar perusahaan: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
    allUserCompanies.value = []
  } finally {
    loadingCompanies.value = false
  }
}

// Load company detail by ID
const loadCompanyDetail = async (companyId: string) => {
  loading.value = true
  try {
    company.value = await companyApi.getById(companyId)
    
    // Find role user di company ini
    const userCompany = allUserCompanies.value.find(c => c.company.id === companyId)
    if (userCompany) {
      currentUserRole.value = userCompany.role
    } else {
      currentUserRole.value = ''
    }
    
    // Generate financial data berdasarkan company ID
    if (company.value) {
      const hash = company.value.id.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
      rkapData.value.value = (100 + (hash % 100)) * 1000000
      rkapData.value.change = 10 + (hash % 10)
      opexData.value.value = (50 + (hash % 50)) * 1000000
      opexData.value.change = 3 + (hash % 5)
      npatData.value.value = (20 + (hash % 30)) * 1000000
      npatData.value.change = 10 + (hash % 10)
      
      // Save selected company to localStorage
      localStorage.setItem('my-company-selected', companyId)
    }
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal memuat data perusahaan: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
    company.value = null
    currentUserRole.value = ''
  } finally {
    loading.value = false
  }
}

// Select company from selector
const selectCompany = async (companyId: string) => {
  await loadCompanyDetail(companyId)
  showCompanySelector.value = false
}

// Show selector again
const showSelector = () => {
  showCompanySelector.value = true
}

// Legacy function for backward compatibility
const loadCompany = async () => {
  await loadUserCompanies()
}

const getMainBusiness = (company: Company): BusinessField | null => {
  if (company.main_business) {
    return company.main_business
  }
  if (company.business_fields && company.business_fields.length > 0) {
    // Find main business field (checking is_main property if it exists)
    const businessFieldsWithMain = company.business_fields as Array<BusinessField & { is_main?: boolean }>
    const mainField = businessFieldsWithMain.find((bf) => bf.is_main)
    return mainField || company.business_fields[0] || null
  }
  return null
}

const formatDate = (date: string | undefined): string => {
  if (!date) return '-'
  return dayjs(date).format('DD MMMM YYYY')
}

const getCompanyLogo = (company: Company): string | undefined => {
  if (company.logo) {
    const apiURL = import.meta.env.VITE_API_URL || 'http://localhost:8080'
    const baseURL = apiURL.replace(/\/api\/v1$/, '')
    return company.logo.startsWith('http') ? company.logo : `${baseURL}${company.logo}`
  }
  return undefined
}

const getCompanyInitial = (name: string | undefined): string => {
  if (!name) return '??'
  const trimmed = name.trim()
  if (!trimmed) return '??'
  const words = trimmed.split(/\s+/).filter(w => w.length > 0)
  if (words.length >= 2) {
    const first = words[0]?.[0]
    const second = words[1]?.[0]
    if (first && second) {
      return (first + second).toUpperCase()
    }
  }
  const firstTwo = trimmed.substring(0, 2)
  return firstTwo ? firstTwo.toUpperCase() : '??'
}

const getIconColor = (name: string): string => {
  const colors: string[] = [
    '#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA07A', '#98D8C8',
    '#F7DC6F', '#BB8FCE', '#85C1E2', '#F8B739', '#52BE80'
  ]
  if (!name) return colors[0]!
  const hash = name.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
  return colors[hash % colors.length]!
}

const getLevelLabel = (level: number): string => {
  switch (level) {
    case 0:
      return 'Holding (Induk)'
    case 1:
      return 'Anak Perusahaan'
    case 2:
      return 'Cucu Perusahaan'
    case 3:
      return 'Cicit Perusahaan'
    default:
      return `Level ${level}`
  }
}

const getLevelColor = (level: number): string => {
  switch (level) {
    case 0:
      return 'red'
    case 1:
      return 'blue'
    case 2:
      return 'green'
    case 3:
      return 'orange'
    default:
      return 'default'
  }
}

const formatCurrency = (value: number): string => {
  if (value >= 1000000000) {
    return `${(value / 1000000000).toFixed(0)}B`
  } else if (value >= 1000000) {
    return `${(value / 1000000).toFixed(0)}M`
  } else if (value >= 1000) {
    return `${(value / 1000).toFixed(0)}K`
  }
  return `${value.toFixed(0)}`
}

const handleManageFiles = () => {
  message.info('Manage files feature coming soon')
}

const handleManageReports = () => {
  message.info('Manage reports feature coming soon')
}

// Export functions
const handleExportPDF = () => {
  message.info('Export PDF feature coming soon')
}

const handleExportExcel = () => {
  message.info('Export Excel feature coming soon')
}

// Menu click handler
const handleMenuClick = ({ key }: { key: string }) => {
  if (key === 'add-subsidiary') {
    handleAddSubsidiary()
  } else if (key === 'edit') {
    handleEdit()
  } else if (key === 'assign-role') {
    openAssignRoleModal()
  }
}

const handleAddSubsidiary = () => {
  router.push('/subsidiaries/new')
}

const handleEdit = () => {
  if (!company.value) {
    message.error('Company tidak ditemukan')
    return
  }
  router.push(`/subsidiaries/${company.value.id}/edit`)
}

// Assign Role functions
const openAssignRoleModal = async () => {
  if (!company.value) {
    message.error('Company tidak ditemukan')
    return
  }
  
  assignRoleModalVisible.value = true
  assignRoleForm.value = {
    userId: undefined,
    roleId: undefined,
  }
  
  // Load users and roles
  // Note: For non-superadmin, users endpoint might return limited results
  // but we still try to load - error will be handled gracefully
  await Promise.all([
    loadUsers(),
    loadRoles()
  ])
}

const loadUsers = async () => {
  if (!company.value) return
  
  usersLoading.value = true
  try {
    // Load all users (backend will filter based on access) - for dropdown selection
    const allUsersData = await userApi.getAll()
    allUsers.value = allUsersData
    
    // Load company users from junction table (supports multiple company assignments)
    try {
      const companyUsersData = await companyApi.getUsers(company.value.id)
      companyUsers.value = companyUsersData
    } catch (error: unknown) {
      // Fallback: if endpoint doesn't exist yet, filter from allUsers
      console.warn('Failed to load company users from endpoint, using fallback:', error)
      companyUsers.value = allUsersData.filter(user => user.company_id === company.value?.id)
    }
  } catch (error: unknown) {
    // Better error handling - check for axios error structure
    const axiosError = error as { 
      response?: { 
        status?: number
        data?: { 
          message?: string
          error?: string
        }
      }
      message?: string
      code?: string
    }
    
    const statusCode = axiosError.response?.status
    const errorMessage = axiosError.response?.data?.message || 
                        axiosError.response?.data?.error || 
                        axiosError.message || 
                        'Unknown error'
    
    // Handle different error scenarios
    if (statusCode === 403 || statusCode === 401) {
      // Permission denied - silently handle, don't show error to user
      console.warn('Access denied to users endpoint (status:', statusCode, '):', errorMessage)
      allUsers.value = []
      companyUsers.value = []
      // Don't show error message for permission issues
    } else if (statusCode === 404) {
      // Not found - might be endpoint issue
      console.warn('Users endpoint not found:', errorMessage)
      allUsers.value = []
      companyUsers.value = []
    } else if (statusCode && statusCode >= 500) {
      // Server error
      console.error('Server error loading users:', errorMessage)
      message.error('Gagal memuat daftar user: Server error')
      allUsers.value = []
      companyUsers.value = []
    } else if (axiosError.code === 'ECONNABORTED' || axiosError.code === 'NETWORK_ERROR') {
      // Network/timeout error
      console.error('Network error loading users:', errorMessage)
      message.error('Gagal memuat daftar user: Masalah koneksi')
      allUsers.value = []
      companyUsers.value = []
    } else {
      // Other errors
      console.error('Error loading users:', error)
      // Only show error if it's not a silent permission issue
      if (statusCode !== 403 && statusCode !== 401) {
        message.error('Gagal memuat daftar user: ' + errorMessage)
      }
      allUsers.value = []
      companyUsers.value = []
    }
  } finally {
    usersLoading.value = false
  }
}

// Alias untuk backward compatibility
const loadCompanyUsers = loadUsers

const loadRoles = async () => {
  rolesLoading.value = true
  try {
    allRoles.value = await roleApi.getAll()
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal memuat daftar role: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
  } finally {
    rolesLoading.value = false
  }
}

// Filtered users and roles
const filteredUsers = computed(() => {
  // Filter out users that are already pengurus
  const availableUsers = allUsers.value.filter(
    user => !companyUsers.value.some(cu => cu.id === user.id)
  )
  
  if (!userSearchText.value) {
    return availableUsers.slice(0, 20) // Limit to 20 for performance
  }
  const search = userSearchText.value.toLowerCase()
  return availableUsers.filter(
    user => 
      user.username.toLowerCase().includes(search) ||
      user.email.toLowerCase().includes(search)
  ).slice(0, 20)
})

const filteredRoles = computed(() => {
  // Filter out superadmin role - hanya untuk developer, bukan untuk user pengguna
  const nonSuperadminRoles = allRoles.value.filter(role => role.name.toLowerCase() !== 'superadmin')
  
  if (!roleSearchText.value) {
    return nonSuperadminRoles
  }
  const search = roleSearchText.value.toLowerCase()
  return nonSuperadminRoles.filter(
    role => role.name.toLowerCase().includes(search)
  )
})

const handleUserSearch = (value: string) => {
  userSearchText.value = value
}

const handleRoleSearch = (value: string) => {
  roleSearchText.value = value
}

const filterUserOption = (input: string, option: unknown) => {
  const opt = option as { value: string }
  const user = allUsers.value.find(u => u.id === opt.value)
  if (!user) return false
  const search = input.toLowerCase()
  return user.username.toLowerCase().includes(search) || 
         user.email.toLowerCase().includes(search)
}

const filterRoleOption = (input: string, option: unknown) => {
  const opt = option as { value: string }
  const role = allRoles.value.find(r => r.id === opt.value)
  if (!role) return false
  return role.name.toLowerCase().includes(input.toLowerCase())
}

const handleAssignRole = async () => {
  if (!company.value || !assignRoleForm.value.userId || !assignRoleForm.value.roleId) {
    message.error('Harap pilih user dan role')
    return
  }
  
  assignRoleLoading.value = true
  try {
    await userApi.assignToCompany(
      assignRoleForm.value.userId,
      company.value.id,
      assignRoleForm.value.roleId
    )
    message.success('User berhasil diassign sebagai pengurus')
    assignRoleForm.value = {
      userId: undefined,
      roleId: undefined,
    }
    // Reload company users
    await loadCompanyUsers()
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(axiosError.response?.data?.message || 'Gagal mengassign user')
  } finally {
    assignRoleLoading.value = false
  }
}

// Edit User Role
const handleEditUserRole = async (user: User) => {
  editingUserRole.value = {
    userId: user.id,
    roleId: user.role_id || undefined,
  }
  editingUserRoleModalVisible.value = true
  await loadRoles()
}

const handleCancelEditUserRole = () => {
  editingUserRoleModalVisible.value = false
  editingUserRole.value = null
}

const handleSaveUserRole = async () => {
  if (!editingUserRole.value || !editingUserRole.value.roleId) {
    message.error('Harap pilih role')
    return
  }
  
  editingRoleLoading.value = true
  try {
    await userApi.update(editingUserRole.value.userId, {
      role_id: editingUserRole.value.roleId,
    })
    message.success('Role pengurus berhasil diubah')
    editingUserRoleModalVisible.value = false
    editingUserRole.value = null
    // Reload company users
    await loadCompanyUsers()
    // Reload company data
    await loadCompany()
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(axiosError.response?.data?.message || 'Gagal mengubah role')
  } finally {
    editingRoleLoading.value = false
  }
}

// Remove User
const handleRemoveUser = async (user: User) => {
  if (!company.value) return
  
  // Show confirmation
  const confirmed = confirm(`Apakah Anda yakin ingin menghapus ${user.username} dari pengurus?`)
  if (!confirmed) return
  
  try {
    // Remove user from company using unassign endpoint (supports multiple company assignments)
    await userApi.unassignFromCompany(user.id, company.value.id)
    message.success('Pengurus berhasil dihapus')
    // Reload company users
    await loadCompanyUsers()
    // Reload company data
    await loadCompany()
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(axiosError.response?.data?.message || 'Gagal menghapus pengurus')
  }
}

const getUserById = (userId: string): User | undefined => {
  return companyUsers.value.find(u => u.id === userId) || allUsers.value.find(u => u.id === userId)
}

const getRoleColor = (role: string): string => {
  const roleLower = role.toLowerCase()
  if (roleLower.includes('admin')) return 'red'
  if (roleLower.includes('manager')) return 'blue'
  if (roleLower.includes('staff')) return 'green'
  return 'default'
}

// Helper functions for company metrics in onboarding cards
const getCompanyRKAP = (companyId: string): number => {
  const hash = companyId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
  return (100 + (hash % 100)) * 1000000
}

const getCompanyRKAPYear = (companyId: string): string => {
  const hash = companyId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
  return String(2024 + (hash % 2))
}

const getCompanyRKAPChange = (companyId: string): number => {
  const hash = companyId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
  return 10 + (hash % 10)
}

const getCompanyOpex = (companyId: string): number => {
  const hash = companyId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
  return (80 + (hash % 40)) * 1000000
}

const getCompanyOpexQuarter = (companyId: string): string => {
  const hash = companyId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
  const quarters = ['Q1 2024', 'Q2 2024', 'Q3 2024', 'Q4 2024']
  return quarters[hash % quarters.length]!
}

const getCompanyOpexChange = (companyId: string): number => {
  const hash = companyId.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0)
  return 3 + (hash % 8)
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

onMounted(() => {
  loadCompany()
})
</script>

<style scoped>
.my-company-layout {
  min-height: 100vh;
}

.detail-content {
  margin: 0 auto;
  padding: 24px;
  max-width: 1400px;
}

.detail-card {
  background: transparent;
}

.loading-container,
.not-found {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  min-height: 400px;
  text-align: center;
}

.not-found p {
  font-size: 16px;
  color: #999;
  margin-bottom: 16px;
}

/* Page Header Container */
.page-header-container {
  padding: 24px;
  background: white;
  border-radius: 8px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

/* Detail Header */
.detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 24px;
  margin-bottom: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid #f0f0f0;
}

.header-left {
  display: flex;
  align-items: flex-start;
  gap: 20px;
  flex: 1;
  min-width: 0;
}

.header-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 12px;
  flex-shrink: 0;
}

.header-actions-top {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-actions-bottom {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-buttons-row {
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.company-icon-large {
  width: 120px;
  height: 120px;
  border-radius: 16px;
  flex-shrink: 0;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.logo-image-large {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.icon-placeholder-large {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 48px;
  font-weight: 700;
  border-radius: 16px;
}

.header-info {
  flex: 1;
  min-width: 0;
}

.company-title {
  font-size: 32px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: #1a1a1a;
  line-height: 1.2;
}

.company-subtitle {
  font-size: 18px;
  color: #666;
  margin: 0 0 16px 0;
  line-height: 1.2;
}

.company-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.meta-item {
  font-size: 14px;
  color: #666;
}

/* Tabs Container */
.tabs-container {
  background: white;
  border-radius: 8px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.tabs-container :deep(.ant-tabs-card) {
  background: transparent;
}

.tabs-container :deep(.ant-tabs-tab) {
  border-radius: 8px 8px 0 0;
}

.tabs-container :deep(.ant-tabs-tab-active) {
  background: white;
}

/* Performance Content */
.performance-content {
  padding: 0;
  background: transparent;
}

.trend-cards-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
  margin-bottom: 32px;
}

.trend-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.trend-card-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.trend-metric {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.metric-value {
  font-size: 32px;
  font-weight: 700;
  color: #1a1a1a;
  line-height: 1.2;
}

.trend-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 14px;
}

.trend-period {
  color: #666;
}

.trend-change {
  padding: 4px 8px;
  border-radius: 4px;
  font-weight: 600;
  font-size: 12px;
}

.trend-change.positive {
  color: #52c41a;
  background: rgba(82, 196, 26, 0.1);
}

.trend-change.negative {
  color: #ff4d4f;
  background: rgba(255, 77, 79, 0.1);
}

.mini-chart-container {
  position: relative;
  width: 100%;
}

.mini-chart {
  width: 100%;
  height: 60px;
}

.chart-fill {
  opacity: 0.6;
}

.chart-line {
  stroke-linecap: round;
  stroke-linejoin: round;
}

.chart-labels {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: #999;
  margin-top: 4px;
}

/* Recent Section */
.recent-section {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.recent-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.card-header-title {
  display: flex;
  align-items: center;
}

/* Profile Content */
.profile-content {
  padding: 0;
  background: transparent;
}

.detail-section {
  width: 100%;
  margin-bottom: 32px;
}

.detail-section:last-child {
  margin-bottom: 0;
}

.section-title {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  color: #1a1a1a;
}

@media (max-width: 1024px) {
  .detail-header {
    flex-direction: column;
    align-items: stretch;
  }

  .header-right {
    align-items: flex-start;
    width: 100%;
  }

  .header-actions-top {
    flex-wrap: wrap;
    width: 100%;
  }

  .trend-cards-row {
    grid-template-columns: 1fr;
  }

  .recent-section {
    grid-template-columns: 1fr;
  }
}

/* Assign Role Modal Styling */
.assign-role-container {
  max-height: 70vh;
  overflow-y: auto;
}

.current-users-section,
.assign-new-section {
  margin-bottom: 24px;
}

.section-header {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  color: #1a1a1a;
}

.text-muted {
  color: #999;
  font-size: 12px;
}

@media (max-width: 768px) {
  .detail-header {
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .header-actions {
    width: 100%;
  }

  .header-actions .ant-btn {
    width: 100%;
  }

  .performance-content,
  .profile-content {
    padding: 16px;
  }
}

/* Company Selector Styles */
.company-selector-container {
  padding: 32px 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.selector-header {
  text-align: center;
  margin-bottom: 32px;
}

.selector-title {
  font-size: 32px;
  font-weight: 700;
  margin: 0 0 8px 0;
  color: #1a1a1a;
}

.selector-description {
  font-size: 16px;
  color: #666;
  margin: 0;
}

.company-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 24px;
  margin-top: 32px;
}

.company-selector-card {
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  height: 100%;
}

.company-selector-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
}

.selector-card-header {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  margin-bottom: 16px;
}

.selector-company-icon {
  width: 80px;
  height: 80px;
  border-radius: 12px;
  flex-shrink: 0;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.selector-logo {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.selector-icon-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 32px;
  font-weight: 700;
  border-radius: 12px;
}

.selector-company-info {
  flex: 1;
  min-width: 0;
}

.selector-company-name {
  font-size: 18px;
  font-weight: 700;
  margin: 0 0 6px 0;
  color: #1a1a1a;
  line-height: 1.3;
}

.selector-company-reg {
  font-size: 13px;
  color: #999;
  margin: 0 0 12px 0;
}

.selector-company-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.selector-card-divider {
  height: 1px;
  background: #f0f0f0;
  margin: 16px 0;
}

.selector-card-content {
  margin-bottom: 16px;
}

.latest-month-header {
  display: flex;
  align-items: center;
  font-size: 12px;
  color: #666;
  margin-bottom: 16px;
}

.metrics-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.metric-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.metric-value {
  font-size: 24px;
  font-weight: 700;
  color: #1a1a1a;
  line-height: 1.2;
}

.metric-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.metric-year,
.metric-quarter {
  font-size: 12px;
  color: #666;
}

.metric-change {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
  color: white;
}

.metric-change.positive {
  background: #52c41a;
}

.metric-change.negative {
  background: #ff4d4f;
}

.metric-label {
  font-size: 12px;
  color: #999;
}

.selector-card-footer {
  margin-top: auto;
  padding-top: 8px;
}

.learn-more-link {
  padding: 0;
  height: auto;
  font-size: 14px;
  color: #1890ff;
  display: flex;
  align-items: center;
}

.learn-more-link:hover {
  color: #40a9ff;
}

@media (max-width: 768px) {
  .company-cards-grid {
    grid-template-columns: 1fr;
  }
  
  .selector-title {
    font-size: 24px;
  }
}

.role-button {
  color: #1890ff;
  font-weight: 500;
  padding: 0;
  height: auto;
}

.role-button:hover {
  color: #40a9ff;
}
</style>
