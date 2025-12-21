<template>
  <div class="subsidiary-detail-layout">
    <DashboardHeader @logout="handleLogout" />

    <div class="detail-content">
      <!-- Loading State - Full Page Skeleton -->
      <div v-if="loading && !company" class="loading-container">
        <a-spin size="large" />
      </div>

      <!-- Company Detail with Skeleton -->
      <div v-else class="detail-card">
        <div class="back-button-container">
          <a-button type="text" @click="handleBack" class="back-button">
            <IconifyIcon icon="mdi:arrow-left" width="20" style="margin-right: 8px;" />
            Kembali ke Daftar Subsidiary
          </a-button>
        </div>

        <div class="page-header-container" style="min-height: 350px; width: 100%;">
          <!-- Header Section -->
          <div class="detail-header" v-if="company">
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

                <!-- Company Hierarchy -->
                <div v-if="companyHierarchy.length > 1" class="company-hierarchy">
                  <span class="hierarchy-label">Hirarki:</span>
                  <span class="hierarchy-path">
                    <template v-for="(hierarchyCompany, index) in companyHierarchy" :key="hierarchyCompany.id">
                      <span class="hierarchy-item">{{ hierarchyCompany.name }}</span>
                      <span v-if="index < companyHierarchy.length - 1" class="hierarchy-separator">/</span>
                    </template>
                  </span>
                </div>
                
              </div>
            </div>
            <div class="header-actions">
              <a-space>
                <a-button @click="handleExportPDF" :loading="exportLoading"  class="btn-icon-label">
                  <IconifyIcon icon="mdi:file-pdf-box" width="16" style="margin-right: 8px;" />
                  PDF
                </a-button>
                <a-button @click="handleExportExcel" :loading="exportLoading"  class="btn-icon-label">
                  <IconifyIcon icon="mdi:file-excel-box" width="16" style="margin-right: 8px;" />
                  Excel
                </a-button>
                <!-- <a-date-picker v-model:value="selectedPeriod" picker="month" placeholder="Select Periode"
                  format="YYYY-MM" style="width: 150px;" @change="handlePeriodChange" /> -->
                <a-dropdown v-if="hasAnyMenuOption">
                  <template #overlay>
                    <a-menu @click="handleMenuClick">
                      <a-menu-item v-if="canEdit" key="edit">
                        <IconifyIcon icon="mdi:pencil" width="16" style="margin-right: 8px;" />
                        Edit Profile Perusahaan
                      </a-menu-item>
                      <a-menu-item v-if="canAssignRole" key="assign-role">
                        <IconifyIcon icon="mdi:account-plus" width="16" style="margin-right: 8px;" />
                        Assign Role Pengurus
                      </a-menu-item>
                      <a-menu-divider v-if="canDelete && (canEdit || canAssignRole)" />
                      <a-menu-item v-if="canDelete" key="delete" danger>
                        <IconifyIcon icon="mdi:delete" width="16" style="margin-right: 8px;" />
                        Hapus Profile Perusahaan
                      </a-menu-item>
                    </a-menu>
                  </template>
                  <a-button  style="display: flex; align-items: center;" class="btn-icon-label">
                    <IconifyIcon icon="mdi:dots-vertical" width="16" style="margin-right: 8px;" />
                    Pengaturan
                  </a-button>
                </a-dropdown>
              </a-space>
            </div>
          </div>
          <!-- Skeleton for Header -->
          <div v-else-if="loadingHeader" class="detail-header">
            <Skeleton :avatar="{ size: 80 }" :paragraph="{ rows: 3 }" :title="false" active />
          </div>
        </div>

        <!-- Tabs -->
        <div class="tabs-container">
          <a-tabs v-model:activeKey="activeTab" type="card" size="large" @change="handleTabChange">
            <a-tab-pane key="performance" tab="Performance">
              <!-- Performance Tab Content -->
              <div class="performance-content">
                <!-- Filter Periode dengan RangePicker -->
                <a-card class="filter-card" :bordered="false" style="margin-bottom: 24px;">
                  <a-space>
                    <span style="font-weight: 500;">Periode:</span>
                    <a-range-picker
                      v-model:value="periodRange"
                      picker="month"
                      format="MMMM YYYY"
                      :placeholder="['Dari Bulan', 'Sampai Bulan']"
                      style="width: 400px;"
                      @change="handleFinancialPeriodChange"
                    />
                    <a-button type="primary" @click="handleFinancialPeriodChange" :loading="financialComparisonLoading">
                      <IconifyIcon icon="mdi:refresh" width="16" style="margin-right: 4px;" />
                      Refresh
                    </a-button>
                  </a-space>
                  <div style="margin-top: 8px; color: #666; font-size: 12px;">
                    <IconifyIcon icon="mdi:information-outline" width="14" style="margin-right: 4px;" />
                    <span v-if="periodRange && periodRange[0] && periodRange[1]">
                      Data ditampilkan apa adanya dari input: RKAP (nilai tahunan) dan Realisasi (nilai bulanan). 
                      Periode: {{ periodRange[0].format('MMMM YYYY') }} - {{ periodRange[1].format('MMMM YYYY') }}.
                    </span>
                    <span v-else>
                      Silakan pilih periode untuk melihat data.
                    </span>
                  </div>
                </a-card>

                <a-spin :spinning="financialComparisonLoading">
                  <!-- Neraca (Balance Sheet) -->
                  <a-card class="financial-table-card" :bordered="false" style="margin-bottom: 24px;">
                    <template #title>
                      <h3 style="margin: 0; font-size: 18px; font-weight: 600;">
                        Neraca (Balance Sheet) - Periode {{ periodRange && periodRange[0] && periodRange[1] ? `${periodRange[0].format('MMMM YYYY')} - ${periodRange[1].format('MMMM YYYY')}` : 'Pilih Periode' }}
                      </h3>
                    </template>
                    
                    <!-- Chart Utama: Balance Sheet Overview -->
                    <div v-if="financialComparisonLoading">
                      <Skeleton :paragraph="{ rows: 4 }" :title="false" active />
                      <div style="height: 300px; margin-top: 16px;">
                        <Skeleton :paragraph="{ rows: 0 }" :title="false" active />
                      </div>
                    </div>
                    <BalanceSheetOverviewChart
                      v-else
                      :data="balanceSheetOverviewChartData"
                    />
                    
                    <!-- Tabel dalam Accordion (default collapsed) -->
                    <a-collapse v-model:activeKey="balanceSheetTableActiveKey" :bordered="false" style="margin-top: 24px;">
                      <a-collapse-panel key="balance-sheet-table" header="📊 Lihat Tabel Detail">
                        <a-table
                          :columns="balanceSheetColumns"
                          :data-source="balanceSheetMonthlyData"
                          :pagination="false"
                          size="middle"
                          :loading="financialComparisonLoading"
                          :bordered="true"
                          :scroll="{ x: 'max-content' }"
                        >
                          <template #bodyCell="{ column, record }">
                            <template v-if="column.key === 'month'">
                              <strong>{{ record.month }}</strong>
                            </template>
                            <template v-else-if="column.key?.endsWith('_rkap')">
                              {{ getCellValue(column.key, record, balanceSheetItems, 'rkap') }}
                            </template>
                            <template v-else-if="column.key?.endsWith('_realisasi')">
                              {{ getCellValue(column.key, record, balanceSheetItems, 'realisasi') }}
                            </template>
                          </template>
                        </a-table>
                      </a-collapse-panel>
                    </a-collapse>
                  </a-card>

                  <!-- Laba Rugi (Profit & Loss) -->
                  <a-card class="financial-table-card" :bordered="false" style="margin-bottom: 24px;">
                    <template #title>
                      <h3 style="margin: 0; font-size: 18px; font-weight: 600;">
                        Laba Rugi (Profit & Loss) - Periode {{ periodRange && periodRange[0] && periodRange[1] ? `${periodRange[0].format('MMMM YYYY')} - ${periodRange[1].format('MMMM YYYY')}` : 'Pilih Periode' }}
                      </h3>
                    </template>
                    
                    <!-- Chart Utama: Profit Loss Overview -->
                    <div v-if="financialComparisonLoading">
                      <Skeleton :paragraph="{ rows: 4 }" :title="false" active />
                      <div style="height: 300px; margin-top: 16px;">
                        <Skeleton :paragraph="{ rows: 0 }" :title="false" active />
                      </div>
                    </div>
                    <ProfitLossOverviewChart
                      v-else
                      :data="profitLossOverviewChartData"
                    />
                    
                    <!-- Tabel dalam Accordion (default collapsed) -->
                    <a-collapse v-model:activeKey="profitLossTableActiveKey" :bordered="false" style="margin-top: 24px;">
                      <a-collapse-panel key="profit-loss-table" header="📊 Lihat Tabel Detail">
                        <a-table
                          :columns="profitLossColumns"
                          :data-source="profitLossMonthlyData"
                          :pagination="false"
                          size="middle"
                          :loading="financialComparisonLoading"
                          :bordered="true"
                          :scroll="{ x: 'max-content' }"
                        >
                          <template #bodyCell="{ column, record }">
                            <template v-if="column.key === 'month'">
                              <strong>{{ record.month }}</strong>
                            </template>
                            <template v-else-if="column.key?.endsWith('_rkap')">
                              {{ getCellValue(column.key, record, profitLossItems, 'rkap') }}
                            </template>
                            <template v-else-if="column.key?.endsWith('_realisasi')">
                              {{ getCellValue(column.key, record, profitLossItems, 'realisasi') }}
                            </template>
                          </template>
                        </a-table>
                      </a-collapse-panel>
                    </a-collapse>
                  </a-card>

                  <!-- Cashflow -->
                  <a-card class="financial-table-card" :bordered="false" style="margin-bottom: 24px;">
                    <template #title>
                      <h3 style="margin: 0; font-size: 18px; font-weight: 600;">
                        Cashflow - Periode {{ periodRange && periodRange[0] && periodRange[1] ? `${periodRange[0].format('MMMM YYYY')} - ${periodRange[1].format('MMMM YYYY')}` : 'Pilih Periode' }}
                      </h3>
                    </template>
                    
                    <!-- Chart Utama: Cashflow Overview -->
                    <div v-if="financialComparisonLoading">
                      <Skeleton :paragraph="{ rows: 4 }" :title="false" active />
                      <div style="height: 300px; margin-top: 16px;">
                        <Skeleton :paragraph="{ rows: 0 }" :title="false" active />
                      </div>
                    </div>
                    <CashflowOverviewChart
                      v-else
                      :data="cashflowOverviewChartData"
                    />
                    
                    <!-- Tabel dalam Accordion (default collapsed) -->
                    <a-collapse v-model:activeKey="cashflowTableActiveKey" :bordered="false" style="margin-top: 24px;">
                      <a-collapse-panel key="cashflow-table" header="📊 Lihat Tabel Detail">
                        <a-table
                          :columns="cashflowColumns"
                          :data-source="cashflowMonthlyData"
                          :pagination="false"
                          size="middle"
                          :loading="financialComparisonLoading"
                          :bordered="true"
                          :scroll="{ x: 'max-content' }"
                        >
                          <template #bodyCell="{ column, record }">
                            <template v-if="column.key === 'month'">
                              <strong>{{ record.month }}</strong>
                            </template>
                            <template v-else-if="column.key?.endsWith('_rkap')">
                              {{ getCellValue(column.key, record, cashflowItems, 'rkap') }}
                            </template>
                            <template v-else-if="column.key?.endsWith('_realisasi')">
                              {{ getCellValue(column.key, record, cashflowItems, 'realisasi') }}
                            </template>
                          </template>
                        </a-table>
                      </a-collapse-panel>
                    </a-collapse>
                  </a-card>

                  <!-- Rasio Keuangan (%) -->
                  <a-card class="financial-table-card" :bordered="false">
                    <template #title>
                      <h3 style="margin: 0; font-size: 18px; font-weight: 600;">
                        Rasio Keuangan (%) - Periode {{ periodRange && periodRange[0] && periodRange[1] ? `${periodRange[0].format('MMMM YYYY')} - ${periodRange[1].format('MMMM YYYY')}` : 'Pilih Periode' }}
                      </h3>
                    </template>
                    
                    <!-- Chart Utama: Ratio Overview -->
                    <div v-if="financialComparisonLoading">
                      <Skeleton :paragraph="{ rows: 4 }" :title="false" active />
                      <div style="height: 300px; margin-top: 16px;">
                        <Skeleton :paragraph="{ rows: 0 }" :title="false" active />
                      </div>
                    </div>
                    <RatioOverviewChart
                      v-else
                      :data="ratioOverviewChartData"
                    />
                    
                    <!-- Tabel dalam Accordion (default collapsed) -->
                    <a-collapse v-model:activeKey="ratioTableActiveKey" :bordered="false" style="margin-top: 24px;">
                      <a-collapse-panel key="ratio-table" header="📊 Lihat Tabel Detail">
                        <a-table
                          :columns="ratioColumns"
                          :data-source="ratioMonthlyData"
                          :pagination="false"
                          size="middle"
                          :loading="financialComparisonLoading"
                          :bordered="true"
                          :scroll="{ x: 'max-content' }"
                        >
                          <template #bodyCell="{ column, record }">
                            <template v-if="column.key === 'month'">
                              <strong>{{ record.month }}</strong>
                            </template>
                            <template v-else-if="column.key?.endsWith('_rkap')">
                              {{ getCellValue(column.key, record, ratioItems, 'rkap') }}
                            </template>
                            <template v-else-if="column.key?.endsWith('_realisasi')">
                              {{ getCellValue(column.key, record, ratioItems, 'realisasi') }}
                            </template>
                          </template>
                        </a-table>
                      </a-collapse-panel>
                    </a-collapse>
                  </a-card>
                </a-spin>
              </div>
            </a-tab-pane>

            <a-tab-pane key="input-laporan" tab="Input Laporan">
              <!-- Input Laporan Tab Content -->
              <div class="input-laporan-content">
                <a-tabs v-model:activeKey="inputLaporanActiveTab" type="line" class="input-laporan-tabs">
                  <!-- Input RKAP -->
                  <a-tab-pane key="rkap" tab="Input RKAP (Tahunan)">
                    <FinancialReportInputForm
                      :company-id="company?.id || ''"
                      :is-r-k-a-p="true"
                      @saved="handleFinancialReportSaved"
                    />
                  </a-tab-pane>
                  
                  <!-- Input Neraca -->
                  <a-tab-pane key="neraca" tab="Neraca">
                    <FinancialCategoryInput
                      :company-id="company?.id || ''"
                      category="neraca"
                      :items="balanceSheetItems"
                      :can-edit="canEditFinancialData"
                      @saved="handleFinancialReportSaved"
                    />
                  </a-tab-pane>
                  
                  <!-- Input Laba Rugi -->
                  <a-tab-pane key="laba-rugi" tab="Laba Rugi (Profit & Loss)">
                    <FinancialCategoryInput
                      :company-id="company?.id || ''"
                      category="laba-rugi"
                      :items="profitLossItems"
                      :can-edit="canEditFinancialData"
                      @saved="handleFinancialReportSaved"
                    />
                  </a-tab-pane>
                  
                  <!-- Input Cashflow -->
                  <a-tab-pane key="cashflow" tab="Cashflow">
                    <FinancialCategoryInput
                      :company-id="company?.id || ''"
                      category="cashflow"
                      :items="cashflowItems"
                      :can-edit="canEditFinancialData"
                      @saved="handleFinancialReportSaved"
                    />
                  </a-tab-pane>
                  
                  <!-- Input Rasio Keuangan -->
                  <a-tab-pane key="rasio" tab="Rasio Keuangan (%)">
                    <FinancialCategoryInput
                      :company-id="company?.id || ''"
                      category="rasio"
                      :items="ratioItems"
                      :can-edit="canEditFinancialData"
                      @saved="handleFinancialReportSaved"
                    />
                  </a-tab-pane>
                </a-tabs>
              </div>
            </a-tab-pane>

            <a-tab-pane key="profile" tab="Profile">
              <!-- Profile Tab Content -->
              <div class="profile-content">
                <!-- Informasi Dasar -->
                <div class="detail-section">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:information" width="20" style="margin-right: 8px;" />
                    Informasi Dasar
                  </h2>
                  <div class="info-grid">
                    <div class="info-item">
                      <span class="info-label">Nama Lengkap</span>
                      <span class="info-value">{{ company!.name }}</span>
                    </div>
                    <div class="info-item">
                      <span class="info-label">Nama Singkat</span>
                      <span class="info-value">{{ company!.short_name || '-' }}</span>
                    </div>
                    <div class="info-item">
                      <span class="info-label">Kode Perusahaan</span>
                      <span class="info-value">{{ company!.code || '-' }}</span>
                    </div>
                    <div class="info-item">
                      <span class="info-label">Status</span>
                      <span class="info-value">
                        <a-tag :color="company!.status === 'Aktif' ? 'green' : 'red'">{{ company!.status || '-' }}</a-tag>
                      </span>
                    </div>
                    <div class="info-item">
                      <span class="info-label">NPWP</span>
                      <span class="info-value">{{ company!.npwp || '-' }}</span>
                    </div>
                    <div class="info-item">
                      <span class="info-label">NIB</span>
                      <span class="info-value">{{ company!.nib || '-' }}</span>
                    </div>
                    <div v-if="company!.authorized_capital" class="info-item">
                      <span class="info-label">Modal Dasar</span>
                      <span class="info-value">{{ formatCurrency(company!.authorized_capital) }} {{ company!.currency || 'IDR' }}</span>
                    </div>
                    <div v-if="company!.paid_up_capital" class="info-item">
                      <span class="info-label">Modal Disetor</span>
                      <span class="info-value">{{ formatCurrency(company!.paid_up_capital) }} {{ company!.currency || 'IDR' }}</span>
                    </div>
                    <div v-if="company!.description" class="info-item full-width">
                      <span class="info-label">Deskripsi</span>
                      <span class="info-value">{{ company!.description }}</span>
                    </div>
                  </div>
                </div>

                <!-- Informasi Kontak -->
                <div class="detail-section">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:phone" width="20" style="margin-right: 8px;" />
                    Informasi Kontak
                  </h2>
                  <div class="info-grid">
                    <div class="info-item">
                      <span class="info-label">Telepon</span>
                      <span class="info-value">{{ company!.phone || '-' }}</span>
                    </div>
                    <div class="info-item">
                      <span class="info-label">Fax</span>
                      <span class="info-value">{{ company!.fax || '-' }}</span>
                    </div>
                    <div class="info-item">
                      <span class="info-label">Email</span>
                      <span class="info-value">
                        <a v-if="company!.email" :href="`mailto:${company!.email}`">{{ company!.email }}</a>
                        <span v-else>-</span>
                      </span>
                    </div>
                    <div class="info-item">
                      <span class="info-label">Website</span>
                      <span class="info-value">
                        <a v-if="company!.website" :href="company!.website" target="_blank" rel="noopener noreferrer">{{ company!.website }}</a>
                        <span v-else>-</span>
                      </span>
                    </div>
                    <div v-if="company!.address" class="info-item full-width">
                      <span class="info-label">Alamat Perusahaan</span>
                      <span class="info-value">{{ company!.address }}</span>
                    </div>
                    <div v-if="company!.operational_address" class="info-item full-width">
                      <span class="info-label">Alamat Operasional</span>
                      <span class="info-value">{{ company!.operational_address }}</span>
                    </div>
                  </div>
                </div>

                <!-- Struktur Kepemilikan -->
                <div v-if="company!.shareholders && company!.shareholders.length > 0" class="detail-section">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:account-group" width="20" style="margin-right: 8px;" />
                    Struktur Kepemilikan ({{ company!.shareholders.length }})
                  </h2>
                  <a-table
                    :columns="shareholderColumns"
                    :data-source="shareholdersWithSelf"
                    :pagination="false"
                    row-key="id"
                    :scroll="{ x: 'max-content' }"
                    class="striped-table"
                  >
                    <template #bodyCell="{ column, record }">
                      <template v-if="column.key === 'name'">
                        <div>
                          <a
                            v-if="record.shareholder_company_id && !record.is_self"
                            :href="`/subsidiaries/${record.shareholder_company_id}`"
                            target="_blank"
                            style="color: #1890ff; font-weight: 500; text-decoration: none;"
                            @click.stop
                          >
                            {{ record.name }}
                            <IconifyIcon icon="mdi:open-in-new" width="14" style="margin-left: 4px;" />
                          </a>
                          <span v-else style="font-weight: 500; color: #52c41a;">{{ record.name }}</span>
                          <a-tag v-if="record.is_main_parent" color="blue" style="margin-left: 8px;">Induk Utama</a-tag>
                          <a-tag v-if="record.is_self" color="green" style="margin-left: 8px;">Perusahaan Sendiri</a-tag>
                        </div>
                      </template>
                      <template v-else-if="column.key === 'authorized_capital'">
                        <span v-if="record.authorized_capital && record.authorized_capital > 0">
                          {{ formatCurrency(record.authorized_capital) }} {{ company?.currency || 'IDR' }}
                        </span>
                        <span v-else>-</span>
                      </template>
                      <template v-else-if="column.key === 'paid_up_capital'">
                        <span v-if="record.paid_up_capital && record.paid_up_capital > 0">
                          {{ formatCurrency(record.paid_up_capital) }} {{ company?.currency || 'IDR' }}
                        </span>
                        <span v-else>-</span>
                      </template>
                      <template v-else-if="column.key === 'ownership_percent'">
                        <strong>{{ formatOwnershipPercent(record.ownership_percent || 0) }}</strong>
                      </template>
                      <template v-else-if="column.key === 'share_sheet_count'">
                        <span v-if="record.share_sheet_count">{{ formatNumber(record.share_sheet_count) }} lembar</span>
                        <span v-else>-</span>
                      </template>
                      <template v-else-if="column.key === 'share_value_per_sheet'">
                        <span v-if="record.share_value_per_sheet">{{ formatCurrency(record.share_value_per_sheet) }}</span>
                        <span v-else>-</span>
                      </template>
                    </template>
                  </a-table>
                </div>

                <!-- Bidang Usaha -->
                <div v-if="company!.main_business || (company!.business_fields && company!.business_fields.length > 0)" class="detail-section">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:briefcase" width="20" style="margin-right: 8px;" />
                    Bidang Usaha
                  </h2>
                  <div class="info-grid">
                    <div class="info-item">
                      <span class="info-label">Sektor Industri</span>
                      <span class="info-value">{{ getMainBusiness(company!)?.industry_sector || '-' }}</span>
                    </div>
                    <div class="info-item">
                      <span class="info-label">KBLI</span>
                      <span class="info-value">{{ getMainBusiness(company!)?.kbli || '-' }}</span>
                    </div>
                    <div v-if="getMainBusiness(company!)?.main_business_activity" class="info-item full-width">
                      <span class="info-label">Uraian Kegiatan Usaha Utama</span>
                      <span class="info-value">{{ getMainBusiness(company!)?.main_business_activity }}</span>
                    </div>
                    <div v-if="getMainBusiness(company!)?.additional_activities" class="info-item full-width">
                      <span class="info-label">Kegiatan Usaha Tambahan</span>
                      <span class="info-value">{{ getMainBusiness(company!)?.additional_activities }}</span>
                    </div>
                    <div v-if="getMainBusiness(company!)?.start_operation_date" class="info-item">
                      <span class="info-label">Tanggal Mulai Beroperasi</span>
                      <span class="info-value">{{ formatDate(getMainBusiness(company!)?.start_operation_date) }}</span>
                    </div>
                  </div>
                </div>

                <!-- Pengurus/Dewan Direksi -->
                <div v-if="company!.directors && company!.directors.length > 0" class="detail-section" style="margin-top: 20px;">
                  <h2 class="section-title">
                    <IconifyIcon icon="mdi:account-tie" width="20" style="margin-right: 8px;" />
                    Pengurus/Dewan Direksi ({{ company!.directors.length }})
                  </h2>
                  <a-table
                    :columns="directorColumns"
                    :data-source="company!.directors"
                    :pagination="false"
                    row-key="id"
                    :scroll="{ x: 'max-content' }"
                    class="striped-table"
                    :expandable="{
                      rowExpandable: isDirectorRowExpandable,
                      expandRowByClick: true,
                    }"
                  >
                    <template #bodyCell="{ column, record }">
                      <template v-if="column.key === 'full_name'">
                        <strong>{{ record.full_name }}</strong>
                      </template>
                      <template v-else-if="column.key === 'position'">
                        <a-tag v-for="(pos, posIdx) in (Array.isArray(record.position) ? record.position : [record.position])" 
                          :key="posIdx" size="small" color="blue" style="margin-right: 4px;">
                          {{ pos }}
                        </a-tag>
                      </template>
                      <template v-else-if="column.key === 'ktp'">
                        {{ record.ktp || '-' }}
                      </template>
                      <template v-else-if="column.key === 'npwp'">
                        {{ record.npwp || '-' }}
                      </template>
                      <template v-else-if="column.key === 'start_date'">
                        {{ record.start_date ? formatDate(record.start_date) : '-' }}
                      </template>
                      <template v-else-if="column.key === 'end_date'">
                        {{ record.end_date ? formatDate(record.end_date) : '-' }}
                      </template>
                      <template v-else-if="column.key === 'documents'">
                        <a-tag v-if="record.id && getDirectorDocumentsCount(record.id) > 0" color="cyan">
                          <IconifyIcon icon="mdi:attachment" width="14" style="margin-right: 4px;" />
                          {{ getDirectorDocumentsCount(record.id) }} dokumen
                        </a-tag>
                        <span v-else>-</span>
                      </template>
                    </template>
                    <template #expandedRowRender="{ record }">
                      <div v-if="record.id" style="padding: 16px; background-color: #fafafa;">
                        <div v-if="loadingDirectorDocuments[record.id]" style="display: flex; align-items: center;">
                          <a-spin size="small" />
                          <span style="margin-left: 8px; color: #999; font-size: 12px;">Memuat...</span>
                        </div>
                        <div v-else-if="getDirectorDocuments(record.id).length === 0" style="color: #999; font-size: 14px;">
                          Belum ada dokumen
                        </div>
                        <div v-else>
                          <div
                            v-for="doc in getDirectorDocuments(record.id)"
                            :key="doc.id"
                            :style="{
                              display: 'flex',
                              alignItems: 'center',
                              padding: '12px',
                              marginBottom: '8px',
                              border: '1px solid #e8e8e8',
                              borderRadius: '4px',
                              backgroundColor: '#fff',
                            }"
                          >
                            <IconifyIcon 
                              :icon="getDocumentIcon(doc)" 
                              width="18" 
                              style="margin-right: 12px; color: #1890ff; flex-shrink: 0;" 
                            />
                            <span :style="{ flex: 1, color: '#333', fontWeight: 500 }">
                              {{ doc.name || doc.file_name || 'Document' }}
                            </span>
                            <a-tag size="small" style="margin: 0 12px; font-size: 11px; flex-shrink: 0;">
                              {{ getDocumentCategoryLabel(doc) }}
                            </a-tag>
                            <a-space style="flex-shrink: 0;" @click.stop>
                              <a
                                v-if="canPreviewDocument(doc)"
                                style="color: #1890ff; cursor: pointer; padding: 4px;"
                                title="Preview"
                                @click.stop="(e: MouseEvent) => {
                                  e.stopPropagation();
                                  handlePreviewDocument(doc, e);
                                }"
                              >
                                <IconifyIcon icon="mdi:eye" width="16" />
                              </a>
                              <a 
                                :href="getDocumentDownloadUrl(doc.file_path)" 
                                target="_blank"
                                title="Download"
                                style="color: #1890ff; padding: 4px;"
                                @click.stop
                              >
                                <IconifyIcon icon="mdi:download" width="16" />
                              </a>
                              <a
                                style="color: #ff4d4f; cursor: pointer; padding: 4px;"
                                title="Hapus"
                                @click.stop="(e: MouseEvent) => {
                                  e.stopPropagation();
                                  Modal.confirm({
                                    title: 'Hapus dokumen ini?',
                                    content: 'Dokumen akan dihapus secara permanen dari sistem. Tindakan ini tidak dapat dibatalkan.',
                                    okText: 'Ya, Hapus',
                                    cancelText: 'Batal',
                                    onOk: () => handleDeleteDirectorDocument(doc.id, record.id!)
                                  })
                                }"
                              >
                                <IconifyIcon icon="mdi:delete" width="16" />
                              </a>
                            </a-space>
                          </div>
                        </div>
                      </div>
                    </template>
                  </a-table>
                </div>
              </div>
            </a-tab-pane>

            <a-tab-pane key="history" tab="History Perubahan Data">
              <!-- History Tab Content -->
              <div class="history-content">
                <a-table
                  :columns="historyColumns"
                  :data-source="changeHistory"
                  :loading="historyLoading"
                  :pagination="historyPagination"
                  row-key="id"
                  @change="handleHistoryTableChange"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'change_description'">
                      <div class="change-description">
                        <div class="change-header">
                          <strong>{{ record.username || 'User' }}</strong> telah melakukan perubahan data:
                        </div>
                        <ul class="change-list">
                          <li v-for="(change, index) in formatChangeDescription(record)" :key="index">
                            {{ change }}
                          </li>
                        </ul>
                      </div>
                    </template>
                    <template v-if="column.key === 'username'">
                      {{ record.username || '-' }}
                    </template>
                    <template v-if="column.key === 'created_at'">
                      {{ formatDateTime(record.created_at) }}
                    </template>
                  </template>
                </a-table>
              </div>
            </a-tab-pane>
          </a-tabs>
        </div>


      </div>

      <!-- Not Found -->
      <div v-if="!loading && !company" class="not-found">
        <IconifyIcon icon="mdi:alert-circle-outline" width="64" style="color: #ccc; margin-bottom: 16px;" />
        <p>Subsidiary tidak ditemukan</p>
        <a-button type="primary" @click="handleBack">Kembali ke Daftar</a-button>
      </div>

      <!-- Document Preview Modal -->
      <a-modal
        v-model:open="previewModalVisible"
        :title="previewModalTitle"
        :footer="null"
        :width="'100%'"
        :wrap-style="{ top: 0, paddingBottom: 0 }"
        :style="{ top: 0, paddingBottom: 0 }"
        :body-style="{ padding: 0, height: '100vh', display: 'flex', justifyContent: 'center', alignItems: 'center', background: previewModalType === 'image' ? '#000' : previewModalType === 'pdf' ? '#525252' : '#fff' }"
        @cancel="previewModalVisible = false"
      >
        <div v-if="previewModalType === 'image'" style="width: 100vw; height: 100vh; display: flex; justify-content: center; align-items: center; background: #000; position: relative;">
          <img :src="previewModalUrl" :alt="previewModalTitle" style="max-width: 100%; max-height: 100%; object-fit: contain;" />
        </div>
        <div v-else-if="previewModalType === 'pdf'" style="width: 100vw; height: 100vh; display: flex; justify-content: center; align-items: center; background: #525252; position: relative;">
          <iframe :src="previewModalUrl" style="width: 100%; height: 100%; border: none;"></iframe>
        </div>
        <div v-else style="width: 100vw; height: 100vh; display: flex; justify-content: center; align-items: center; flex-direction: column; background: #fff;">
          <IconifyIcon icon="mdi:file-document" width="64" style="color: #1890ff; margin-bottom: 16px;" />
          <p style="margin-bottom: 16px;">{{ previewModalTitle }}</p>
          <a-button type="primary" :href="previewModalUrl" target="_blank">
            <IconifyIcon icon="mdi:download" width="16" style="margin-right: 8px;" />
            Download File
          </a-button>
        </div>
      </a-modal>

      <!-- Assign Role Modal -->
      <a-modal v-model:open="assignRoleModalVisible" title="Assign Role - Kelola Pengurus"
        :confirm-loading="assignRoleLoading" width="900px" :footer="null">
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
                    <a-select v-model:value="assignRoleForm.userId" show-search
                      placeholder="Cari user berdasarkan nama atau email" :filter-option="filterUserOption"
                      :loading="usersLoading" @search="handleUserSearch" allow-clear :disabled="usersLoading">
                      <a-select-option v-for="user in filteredUsers" :key="user.id" :value="user.id"
                        :disabled="companyUsers.some((u: User) => u.id === user.id)">
                        {{ user.username }} ({{ user.email }})
                        <span v-if="isUserAlreadyPengurus(user.id)" class="text-muted"> - Sudah menjadi
                          pengurus</span>
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
                    <a-select v-model:value="assignRoleForm.roleId" show-search placeholder="Cari role"
                      :filter-option="filterRoleOption" :loading="rolesLoading" @search="handleRoleSearch" allow-clear
                      :disabled="rolesLoading">
                      <a-select-option v-for="role in filteredRoles" :key="role.id" :value="role.id">
                        {{ role.name }}
                      </a-select-option>
                    </a-select>
                    <small v-if="rolesLoading" class="text-muted">Memuat daftar role...</small>
                  </a-form-item>
                </a-col>
              </a-row>

              <a-form-item>
                <a-button type="primary" :loading="assignRoleLoading" @click="handleAssignRole"
                  :disabled="!assignRoleForm.userId || !assignRoleForm.roleId">
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
            <a-table :columns="userColumns" :data-source="companyUsers" :loading="usersLoading"
              :pagination="{ pageSize: 10 }" row-key="id" size="middle">
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
      <a-modal v-model:open="editingUserRoleModalVisible" title="Ubah Role Pengurus"
        :confirm-loading="editingRoleLoading" @ok="handleSaveUserRole" @cancel="handleCancelEditUserRole" width="500px">
        <a-form layout="vertical" v-if="editingUserRole">
          <a-form-item label="User">
            <a-input :value="getUserById(editingUserRole.userId)?.username" disabled />
          </a-form-item>
          <a-form-item label="Pilih Role Baru" required>
            <a-select v-model:value="editingUserRole.roleId" show-search placeholder="Cari role"
              :filter-option="filterRoleOption" :loading="rolesLoading" @search="handleRoleSearch">
              <a-select-option v-for="role in filteredRoles" :key="role.id" :value="role.id">
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
import { useRouter, useRoute } from 'vue-router'
import { message, Modal, Skeleton } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import FinancialReportInputForm from '../components/FinancialReportInputForm.vue'
import FinancialCategoryInput from '../components/FinancialCategoryInput.vue'
import BalanceSheetOverviewChart from '../components/BalanceSheetOverviewChart.vue'
import ProfitLossOverviewChart from '../components/ProfitLossOverviewChart.vue'
import CashflowOverviewChart from '../components/CashflowOverviewChart.vue'
import RatioOverviewChart from '../components/RatioOverviewChart.vue'
import { companyApi, userApi, roleApi, type Company, type BusinessField, type User, type Role } from '../api/userManagement'
import reportsApi, { type Report } from '../api/reports'
import { financialReportsApi, type FinancialReport, type FinancialReportComparison } from '../api/financialReports'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'
import dayjs, { type Dayjs } from 'dayjs'
import { auditApi, type UserActivityLog } from '../api/audit'
import documentsApi, { type DocumentItem } from '../api/documents'
import jsPDF from 'jspdf'
import autoTable from 'jspdf-autotable'

// Type definitions for change data structures
interface DirectorChangeData {
  action: 'added' | 'removed'
  position?: string
  full_name?: string
}

interface ShareholderChangeData {
  action: 'added' | 'removed'
  name?: string
  type?: string
}

interface FieldChangeData {
  old: unknown
  new: unknown
}

type ChangeData = DirectorChangeData | ShareholderChangeData | FieldChangeData | Record<string, unknown>

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const company = ref<Company | null>(null)
const loading = ref(false)
const loadingHeader = ref(false)
const loadingReports = ref(false)
const loadingHierarchy = ref(false)
const activeTab = ref('performance')
// selectedPeriod removed - using periodRange instead
const exportLoading = ref(false)
const companyHierarchy = ref<Company[]>([])
const allCompanies = ref<Company[]>([])

// Financial Report state
const financialComparison = ref<FinancialReportComparison | null>(null)
const financialComparisonLoading = ref(false)
// RangePicker untuk periode (default: Januari sampai bulan saat ini)
const periodRange = ref<[Dayjs, Dayjs] | null>([
  dayjs().startOf('year'), // Januari tahun ini
  dayjs(), // Bulan saat ini
])
// Computed untuk backward compatibility dengan fungsi yang masih menggunakan selectedYear, startMonth, endMonth
const selectedYear = computed(() => {
  if (!periodRange.value || !periodRange.value[0]) return dayjs().format('YYYY')
  return periodRange.value[0].format('YYYY')
})
const startMonth = computed(() => {
  if (!periodRange.value || !periodRange.value[0]) return '01'
  return periodRange.value[0].format('MM')
})
const endMonth = computed(() => {
  if (!periodRange.value || !periodRange.value[1]) return dayjs().format('MM')
  return periodRange.value[1].format('MM')
})
const financialReports = ref<FinancialReport[]>([])
const financialReportsLoading = ref(false)
const inputLaporanActiveTab = ref('rkap')

// Accordion state untuk tabel (default collapsed - empty array)
const balanceSheetTableActiveKey = ref<string[]>([])
const profitLossTableActiveKey = ref<string[]>([])
const cashflowTableActiveKey = ref<string[]>([])
const ratioTableActiveKey = ref<string[]>([])


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

// Director documents state
const directorDocumentsMap = ref<Map<string, DocumentItem[]>>(new Map())
const loadingDirectorDocuments = ref<Record<string, boolean>>({})

// Document categories
const documentCategories = [
  { label: 'KTP', value: 'ktp' },
  { label: 'NPWP', value: 'npwp' },
  { label: 'Sertifikat', value: 'certificate' },
  { label: 'Ijazah', value: 'diploma' },
  { label: 'SK Pengangkatan', value: 'appointment_letter' },
  { label: 'Lainnya', value: 'other' },
]

// Shareholder columns
const shareholderColumns = [
  { title: 'Nama Pemegang Saham', key: 'name', width: 250 },
  { title: 'Modal Dasar', key: 'authorized_capital', width: 200 },
  { title: 'Modal Disetor', key: 'paid_up_capital', width: 180 },
  { title: 'Persentase Kepemilikan', key: 'ownership_percent', width: 180, align: 'right' },
  { title: 'Jumlah Lembar Saham', key: 'share_sheet_count', width: 180 },
  { title: 'Nilai per Lembar', key: 'share_value_per_sheet', width: 180 },
]

// State untuk menyimpan data perusahaan pemegang saham yang sudah di-load
const shareholderCompaniesMap = ref<Map<string, Company>>(new Map())

// Load shareholder company data jika belum ada
const loadShareholderCompany = async (companyId: string) => {
  if (shareholderCompaniesMap.value.has(companyId)) {
    return shareholderCompaniesMap.value.get(companyId)
  }
  
  // Cek di allCompanies dulu
  const foundInAll = allCompanies.value.find(c => c.id === companyId)
  if (foundInAll) {
    shareholderCompaniesMap.value.set(companyId, foundInAll)
    return foundInAll
  }
  
  // Jika tidak ada, load dari API
  try {
    const companyData = await companyApi.getById(companyId)
    if (companyData) {
      shareholderCompaniesMap.value.set(companyId, companyData)
      // Juga tambahkan ke allCompanies untuk penggunaan selanjutnya
      if (!allCompanies.value.find(c => c.id === companyId)) {
        allCompanies.value.push(companyData)
      }
      return companyData
    }
  } catch (error) {
    console.error(`Error loading shareholder company ${companyId}:`, error)
  }
  
  return undefined
}

// Computed: Shareholders with self company row
const shareholdersWithSelf = computed(() => {
  if (!company.value) return []
  
  const shareholders = [...(company.value.shareholders || [])]
  
  // Calculate total capital from shareholders (only company shareholders)
  const companyShareholders = shareholders.filter(sh => sh.shareholder_company_id)
  let totalShareholderCapital = 0
  
  companyShareholders.forEach((sh) => {
    // Find company data from allCompanies or shareholderCompaniesMap
    const shareholderCompany = allCompanies.value.find(c => c.id === sh.shareholder_company_id) 
      || shareholderCompaniesMap.value.get(sh.shareholder_company_id!)
    const capital = shareholderCompany?.paid_up_capital || 0
    totalShareholderCapital += capital
  })
  
  const currentCompanyCapital = company.value.paid_up_capital || 0
  
  // Total capital = Modal perusahaan sendiri + Total modal semua pemegang saham
  // Sesuai dengan rumus di form
  const totalCapital = currentCompanyCapital + totalShareholderCapital
  
  // Calculate self ownership percent
  let selfOwnershipPercent = 0
  if (totalCapital > 0) {
    selfOwnershipPercent = (currentCompanyCapital / totalCapital) * 100
  }
  
  // Add self company row at the beginning
  const selfRow = {
    id: `__self__${company.value.id}`,
    name: company.value.name,
    authorized_capital: company.value.authorized_capital || 0,
    paid_up_capital: company.value.paid_up_capital || 0,
    ownership_percent: selfOwnershipPercent,
    shareholder_company_id: company.value.id,
    is_company: true,
    is_main_parent: false,
    is_self: true, // Flag untuk membedakan row perusahaan sendiri
  }
  
  // Calculate ownership percent for each shareholder based on their paid_up_capital
  const shareholdersWithPercent = shareholders.map((sh) => {
    let ownershipPercent = sh.ownership_percent || 0
    
    // Find shareholder company data
    let shareholderCompany: Company | undefined
    if (sh.shareholder_company_id) {
      shareholderCompany = allCompanies.value.find(c => c.id === sh.shareholder_company_id)
        || shareholderCompaniesMap.value.get(sh.shareholder_company_id)
    }
    
    // If shareholder is a company, recalculate based on paid_up_capital
    if (shareholderCompany && totalCapital > 0) {
      const shareholderCapital = shareholderCompany.paid_up_capital || 0
      ownershipPercent = (shareholderCapital / totalCapital) * 100
    }
    
    return {
      ...sh,
      authorized_capital: shareholderCompany?.authorized_capital || 0,
      paid_up_capital: shareholderCompany?.paid_up_capital || 0,
      ownership_percent: ownershipPercent,
    }
  })
  
  return [selfRow, ...shareholdersWithPercent]
})

// Director columns
const directorColumns = [
  { title: 'Nama Lengkap', key: 'full_name', width: 200 },
  { title: 'Jabatan', key: 'position', width: 180 },
  { title: 'KTP', key: 'ktp', width: 150 },
  { title: 'NPWP', key: 'npwp', width: 150 },
  { title: 'Tanggal Mulai', key: 'start_date', width: 130 },
  { title: 'Tanggal Akhir', key: 'end_date', width: 130 },
  { title: 'Dokumen', key: 'documents', width: 120 },
]

// User columns for table
const userColumns = [
  { title: 'Username', dataIndex: 'username', key: 'username' },
  { title: 'Email', dataIndex: 'email', key: 'email' },
  { title: 'Role', key: 'role' },
  { title: 'Status', key: 'status' },
  { title: 'Aksi', key: 'action', width: 200 },
]

// History columns for table
const historyColumns = [
  { title: 'Perubahan', key: 'change_description', width: '60%' },
  { title: 'Diubah Oleh', key: 'username', width: '20%' },
  { title: 'Waktu', key: 'created_at', width: '20%' },
]

// Change history state
const changeHistory = ref<UserActivityLog[]>([])
const historyLoading = ref(false)
const historyPagination = ref({
  current: 1,
  pageSize: 20,
  total: 0,
  showSizeChanger: true,
  showTotal: (total: number) => `Total ${total} perubahan`,
})

// Computed: Check user roles
const userRole = computed(() => {
  return authStore.user?.role?.toLowerCase() || ''
})

const isSuperAdmin = computed(() => userRole.value === 'superadmin')
const isAdministrator = computed(() => userRole.value === 'administrator')
const isAdmin = computed(() => userRole.value === 'admin')
const isManager = computed(() => userRole.value === 'manager')
const isStaff = computed(() => userRole.value === 'staff')

// RBAC: Assign Role untuk admin/superadmin/administrator
const canAssignRole = computed(() => isAdmin.value || isSuperAdmin.value || isAdministrator.value)

// RBAC: Edit Financial Data
// Superadmin & Administrator: bisa edit semua data
// Admin: hanya bisa edit data perusahaan sendiri
const canEditFinancialData = computed(() => {
  if (isSuperAdmin.value || isAdministrator.value) {
    return true // Bisa edit semua
  }
  if (isAdmin.value) {
    // Admin hanya bisa edit data perusahaan sendiri
    // Check jika company yang sedang dilihat adalah milik user
    return authStore.user?.company_id === company.value?.id
  }
  return false
})

// RBAC: Delete untuk admin/superadmin/administrator
const canDelete = computed(() => isAdmin.value || isSuperAdmin.value || isAdministrator.value)

// RBAC: Edit untuk semua role (staff, manager, admin, superadmin, administrator)
const canEdit = computed(() => isAdmin.value || isManager.value || isStaff.value || isSuperAdmin.value || isAdministrator.value)

// Check if any menu item is available (to show/hide Options dropdown)
const hasAnyMenuOption = computed(() => canEdit.value || canAssignRole.value || canDelete.value)

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

// Chart data computed from filtered reports (removed unused computed properties and functions)

// Reports data
const companyReports = ref<Report[]>([])
const reportsLoading = ref(false)

// filteredReports removed (unused)

// Recent reports and reportColumns removed (unused)

const loadCompany = async () => {
  const id = route.params.id as string
  if (!id) {
    message.error('ID perusahaan tidak valid')
    return
  }

  loading.value = true
  loadingHeader.value = true
  try {
    // Load all companies first (needed for shareholder data)
    if (allCompanies.value.length === 0) {
      allCompanies.value = await companyApi.getAll()
    }
    
    // Load company first (required for other operations)
    company.value = await companyApi.getById(id)
    loadingHeader.value = false
    
    // Load reports after company is loaded
    if (company.value) {
      // Load shareholder companies first (needed for ownership calculation)
      if (company.value.shareholders && company.value.shareholders.length > 0) {
        const companyShareholders = company.value.shareholders.filter(sh => sh.shareholder_company_id)
        await Promise.all(
          companyShareholders.map(sh => 
            loadShareholderCompany(sh.shareholder_company_id!).catch(err => {
              console.error(`Error loading shareholder company ${sh.shareholder_company_id}:`, err)
            })
          )
        )
      }
      
      // Run independent API calls in parallel for better performance
      loadingReports.value = true
      loadingHierarchy.value = true
      
      await Promise.all([
        loadCompanyReports(id),
        loadCompanyHierarchy(id),
        loadChangeHistory(),
        loadFinancialReports(id),
      ])
      
      loadingReports.value = false
      loadingHierarchy.value = false
      
      // Load financial comparison (depends on financial reports and selected period)
      await loadFinancialComparison(id)
      
      // Load director documents (can be done in parallel with other operations)
      if (company.value.directors && company.value.directors.length > 0) {
        // Don't await this - let it load in background
        loadAllDirectorDocuments(company.value.directors).catch(err => {
          console.error('Error loading director documents:', err)
        })
      }
    }
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal memuat data perusahaan: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
    loadingHeader.value = false
    loadingReports.value = false
    loadingHierarchy.value = false
  } finally {
    loading.value = false
  }
}

const loadCompanyHierarchy = async (companyId: string) => {
  try {
    // Try to get ancestors from API first
    try {
      const ancestors = await companyApi.getAncestors(companyId)
      // Ancestors API returns from parent to root (excluding current company)
      // We need to reverse to get from root to parent, then add current company
      companyHierarchy.value = [...ancestors].reverse()
      // Add current company at the end
      if (company.value) {
        companyHierarchy.value.push(company.value)
      }
      loadingHierarchy.value = false
      return
    } catch {
      // If API endpoint doesn't exist (404 or other error), build hierarchy manually
      // Don't log as error if it's just 404 (endpoint not implemented yet)
      // Silently handle error and build hierarchy manually
    }

      // Fallback: Build hierarchy manually by loading all companies
      loadingHierarchy.value = false
    if (allCompanies.value.length === 0) {
      allCompanies.value = await companyApi.getAll()
    }

    // Build hierarchy from current company to root
    const hierarchy: Company[] = []
    let currentCompany: Company | undefined = company.value || undefined

    // If company not found in allCompanies, try to find it
    if (!currentCompany) {
      currentCompany = allCompanies.value.find(c => c.id === companyId)
    }

    if (!currentCompany) {
      companyHierarchy.value = []
      return
    }

    // Build hierarchy by traversing parent_id
    const companyMap = new Map<string, Company>()
    allCompanies.value.forEach(c => companyMap.set(c.id, c))

    let current: Company | undefined = currentCompany
    while (current) {
      hierarchy.unshift(current) // Add to beginning to get root -> current order
      
      if (current.parent_id) {
        current = companyMap.get(current.parent_id)
      } else {
        break
      }
    }

    companyHierarchy.value = hierarchy
  } catch (error) {
    console.error('Error loading company hierarchy:', error)
    companyHierarchy.value = []
  }
}

const loadCompanyReports = async (companyId: string) => {
  reportsLoading.value = true
  try {
    const reports = await reportsApi.getByCompanyId(companyId)
    companyReports.value = reports
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    console.error('Gagal memuat data reports:', axiosError.response?.data?.message || axiosError.message || 'Unknown error')
    // Don't show error message to user, just log it
    companyReports.value = []
  } finally {
    reportsLoading.value = false
  }
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

const formatCurrency = (value: number | string | undefined): string => {
  if (value === undefined || value === null) return '-'
  const numValue = typeof value === 'string' ? parseFloat(value) : value
  if (isNaN(numValue)) return '-'
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(numValue)
}

const formatNumber = (value: number | string | undefined): string => {
  if (value === undefined || value === null) return '-'
  const numValue = typeof value === 'string' ? parseFloat(value) : value
  if (isNaN(numValue)) return '-'
  return new Intl.NumberFormat('id-ID').format(numValue)
}

// Format percentage for display - show all significant digits without rounding
const formatOwnershipPercent = (percent: number): string => {
  if (percent === 0) return '0%'
  
  // Use toFixed with high precision (10 decimal places) to ensure all digits are shown
  // This matches the precision used in calculation (10000000000)
  const str = percent.toFixed(10)
  
  // Remove trailing zeros but keep all significant digits
  // This will show values like "80.6451612903%" instead of "80.64516129030000%"
  const trimmed = str.replace(/\.?0+$/, '')
  
  // If after trimming it's empty or just '.', return '0%'
  if (!trimmed || trimmed === '.') return '0%'
  
  return `${trimmed}%`
}

// Load all director documents
const loadAllDirectorDocuments = async (directors: Array<{ id?: string }>) => {
  const directorsWithId = directors.filter(d => d.id)
  if (directorsWithId.length === 0) {
    console.log('No directors with ID found, skipping document load')
    return
  }
  
  console.log(`Loading documents for ${directorsWithId.length} directors`, directorsWithId.map(d => ({ id: d.id })))
  
  const loadPromises = directorsWithId.map(async (director) => {
    if (!director.id) return
    loadingDirectorDocuments.value[director.id] = true
    try {
      console.log(`Fetching documents for director ${director.id}`)
      const response = await documentsApi.listDocumentsPaginated({
        director_id: director.id,
        page: 1,
        page_size: 100,
      })
      console.log(`Loaded ${response.data.length} documents for director ${director.id}`, response.data)
      directorDocumentsMap.value.set(director.id, response.data)
    } catch (error) {
      console.error(`Failed to load documents for director ${director.id}:`, error)
      directorDocumentsMap.value.set(director.id, [])
    } finally {
      loadingDirectorDocuments.value[director.id] = false
    }
  })
  
  await Promise.all(loadPromises)
  console.log('Finished loading director documents', Array.from(directorDocumentsMap.value.entries()))
}

// Get director documents
const getDirectorDocuments = (directorId: string): DocumentItem[] => {
  return directorDocumentsMap.value.get(directorId) || []
}

// Get director documents count
const getDirectorDocumentsCount = (directorId: string): number => {
  return getDirectorDocuments(directorId).length
}

// Get document category label
const getDocumentCategoryLabel = (doc: DocumentItem): string => {
  const meta = doc.metadata as { category?: string } | undefined
  const categoryValue = meta?.category || '-'
  const categoryLabel = documentCategories.find(cat => cat.value === categoryValue)?.label || categoryValue
  return categoryLabel
}

// Get document download URL
const getDocumentDownloadUrl = (filePath: string): string => {
  if (filePath.startsWith('http://') || filePath.startsWith('https://')) {
    return filePath
  }
  const apiURL = import.meta.env.VITE_API_URL?.replace('/api/v1', '') || 'http://localhost:8080'
  return `${apiURL}${filePath}`
}

// Delete director document
const handleDeleteDirectorDocument = async (documentId: string, directorId: string) => {
  try {
    await documentsApi.deleteDocument(documentId)
    
    // Remove dari map
    const docs = directorDocumentsMap.value.get(directorId) || []
    directorDocumentsMap.value.set(
      directorId,
      docs.filter(d => d.id !== documentId)
    )
    
    message.success('Dokumen berhasil dihapus')
    
    // Reload company untuk refresh data
    if (company.value) {
      await loadCompany()
    }
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(axiosError.response?.data?.message || 'Gagal menghapus dokumen')
  }
}

// Document preview modal state
const previewModalVisible = ref(false)
const previewModalTitle = ref('')
const previewModalUrl = ref('')
const previewModalType = ref<'image' | 'pdf' | 'file'>('image')

// Handle document preview
const handlePreviewDocument = (doc: DocumentItem, event?: MouseEvent) => {
  if (event) {
    event.stopPropagation()
    event.preventDefault()
    event.stopImmediatePropagation()
  }
  
  const url = getDocumentDownloadUrl(doc.file_path)
  const fileName = doc.name || doc.file_name || 'Document'
  const fileExt = fileName.split('.').pop()?.toLowerCase() || ''
  const imageExtensions = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg']
  const previewableExtensions = [...imageExtensions, 'pdf']
  
  if (previewableExtensions.includes(fileExt)) {
    previewModalType.value = imageExtensions.includes(fileExt) ? 'image' : 'pdf'
    previewModalTitle.value = fileName
    previewModalUrl.value = url
    previewModalVisible.value = true
  }
}

// Check if document can be previewed
const canPreviewDocument = (doc: DocumentItem): boolean => {
  const fileName = doc.name || doc.file_name || 'Document'
  const fileExt = fileName.split('.').pop()?.toLowerCase() || ''
  const imageExtensions = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg']
  const previewableExtensions = [...imageExtensions, 'pdf']
  return previewableExtensions.includes(fileExt)
}

// Get document icon based on file type
const getDocumentIcon = (doc: DocumentItem): string => {
  const fileName = doc.name || doc.file_name || 'Document'
  const fileExt = fileName.split('.').pop()?.toLowerCase() || ''
  const imageExtensions = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg']
  
  if (imageExtensions.includes(fileExt)) {
    return 'mdi:file-image'
  } else if (fileExt === 'pdf') {
    return 'mdi:file-pdf-box'
  } else {
    return 'mdi:file-document'
  }
}

// Check if director row can be expanded
const isDirectorRowExpandable = (record: { id?: string }): boolean => {
  return record.id ? getDirectorDocumentsCount(record.id) > 0 : false
}


const getCompanyLogo = (company: Company): string | undefined => {
  if (company.logo) {
    const apiURL = import.meta.env.VITE_API_URL || (import.meta.env.DEV ? 'http://localhost:8080' : 'https://api-pedeve-dev.aretaamany.com')
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
  if (level === 0) return 'Holding'
  return `Level ${String(level)}`
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

const handleBack = () => {
  router.push('/subsidiaries')
}

const handleEdit = () => {
  if (company.value) {
    router.push(`/subsidiaries/${company.value.id}/edit`)
  }
}

const handleDelete = async () => {
  if (!company.value) return

  try {
    await companyApi.delete(company.value.id)
    message.success('Subsidiary berhasil dihapus')
    router.push('/subsidiaries')
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(axiosError.response?.data?.message || 'Gagal menghapus subsidiary')
  }
}

const handleMenuClick = ({ key }: { key: string }) => {
  if (key === 'edit') {
    handleEdit()
  } else if (key === 'assign-role') {
    openAssignRoleModal()
  } else if (key === 'delete') {
    // Show Ant Design confirmation modal before delete
    if (company.value) {
      Modal.confirm({
        title: 'Hapus Profile Perusahaan',
        content: `Apakah Anda yakin ingin menghapus "${company.value.name}"? Tindakan ini tidak dapat dibatalkan.`,
        okText: 'Hapus',
        okType: 'danger',
        cancelText: 'Batal',
        onOk: () => {
          handleDelete()
        },
      })
    }
  }
}

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
  } catch {
    message.error('Gagal memuat daftar role')
  } finally {
    rolesLoading.value = false
  }
}

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
    // Reload company data
    await loadCompany()
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
  return companyUsers.value.find((u: User) => u.id === userId) || allUsers.value.find((u: User) => u.id === userId)
}

const isUserAlreadyPengurus = (userId: string): boolean => {
  return companyUsers.value.some((u: User) => u.id === userId)
}

const getRoleColor = (role: string): string => {
  const roleLower = role.toLowerCase()
  if (roleLower.includes('admin')) return 'red'
  if (roleLower.includes('manager')) return 'blue'
  if (roleLower.includes('staff')) return 'green'
  return 'default'
}

// handleManageReports removed (unused)

// Handle period change - removed (unused)
// const handlePeriodChange = () => {
//   // Filter is automatically applied via computed property
//   // No need to reload data, just let computed properties react
// }

// Export PDF dengan data dari Performance Table
const handleExportPDF = async () => {
  if (!company.value) {
    message.error('Company tidak ditemukan')
    return
  }

  // Validasi periode
  if (!periodRange.value || !periodRange.value[0] || !periodRange.value[1]) {
    message.warning('Silakan pilih periode terlebih dahulu')
    return
  }

  try {
    exportLoading.value = true

    // Create PDF document
    const doc = new jsPDF('landscape', 'mm', 'a4')
    const pageWidth = doc.internal.pageSize.getWidth()
    const pageHeight = doc.internal.pageSize.getHeight()
    let yPosition = 20

    // Helper function untuk menambahkan header section
    const addHeader = (title: string) => {
      if (yPosition > pageHeight - 40) {
        doc.addPage()
        yPosition = 20
      }
      doc.setFontSize(16)
      doc.setFont('helvetica', 'bold')
      doc.text(title, 14, yPosition)
      yPosition += 10
    }

    // Helper function untuk menambahkan tabel dengan merged headers
    const addTable = (title: string, items: Array<{ key: string; label: string; field: string; isRatio: boolean }>, data: Array<Record<string, unknown>>) => {
      if (data.length === 0) return

      addHeader(title)

      // Prepare table data dengan struktur merged headers (2 level headers)
      // Level 1: Bulan + item labels (merged untuk RKAP dan Realisasi)
      interface HeaderCell {
        content: string
        rowSpan?: number
        colSpan?: number
      }
      const topHeaders: HeaderCell[] = [{ content: 'Bulan', rowSpan: 2 }]
      items.forEach((item) => {
        topHeaders.push({ content: item.label, colSpan: 2 })
      })

      // Level 2: RKAP dan Realisasi untuk setiap item
      const subHeaders: string[] = []
      items.forEach(() => {
        subHeaders.push('RKAP', 'Realisasi')
      })

      const tableData: (string | number)[][] = []
      data.forEach((row) => {
        const rowData: (string | number)[] = [row.month as string]
        items.forEach((item) => {
          const rkapValue = getCellValue(`${item.key}_rkap`, row, items, 'rkap')
          const realisasiValue = getCellValue(`${item.key}_realisasi`, row, items, 'realisasi')
          rowData.push(rkapValue, realisasiValue)
        })
        tableData.push(rowData)
      })

      // Add table dengan autoTable (mendukung merged headers)
      autoTable(doc, {
        head: [topHeaders, subHeaders],
        body: tableData,
        startY: yPosition,
        theme: 'striped',
        headStyles: {
          fillColor: [66, 139, 202],
          textColor: 255,
          fontStyle: 'bold',
          fontSize: 9,
          halign: 'center',
          valign: 'middle',
        },
        bodyStyles: {
          fontSize: 8,
          textColor: [0, 0, 0],
          halign: 'right',
          valign: 'middle',
        },
        alternateRowStyles: {
          fillColor: [245, 245, 245],
        },
        columnStyles: {
          0: { cellWidth: 40, halign: 'left', fontStyle: 'bold' },
        },
        margin: { left: 14, right: 14, top: 10 },
        styles: {
          cellPadding: 4,
          overflow: 'linebreak',
          cellWidth: 'auto',
          lineWidth: 0.1,
          lineColor: [200, 200, 200],
        },
        didDrawPage: (data) => {
          if (data && data.cursor) {
            yPosition = data.cursor.y + 10
          }
        },
      })

      // Get final Y position from autoTable
      interface AutoTableResult {
        lastAutoTable?: {
          finalY?: number
        }
      }
      const docWithAutoTable = doc as unknown as AutoTableResult
      const finalY = docWithAutoTable.lastAutoTable?.finalY
      if (finalY) {
        yPosition = finalY + 15
      } else {
        yPosition += 50 // Fallback jika finalY tidak tersedia
      }
    }

    // Cover page
    doc.setFontSize(20)
    doc.setFont('helvetica', 'bold')
    doc.text('Laporan Keuangan', pageWidth / 2, 50, { align: 'center' })
    
    doc.setFontSize(16)
    doc.setFont('helvetica', 'normal')
    doc.text(company.value.name, pageWidth / 2, 60, { align: 'center' })
    
    doc.setFontSize(12)
    const periodText = `${periodRange.value[0].format('MMMM YYYY')} - ${periodRange.value[1].format('MMMM YYYY')}`
    doc.text(`Periode: ${periodText}`, pageWidth / 2, 70, { align: 'center' })
    
    doc.setFontSize(10)
    const exportDate = dayjs().format('DD MMMM YYYY HH:mm')
    doc.text(`Dicetak pada: ${exportDate}`, pageWidth / 2, 80, { align: 'center' })

    // Start tables on new page for better layout
    doc.addPage()
    yPosition = 20

    // Add tables untuk setiap kategori
    addTable(
      'Neraca (Balance Sheet)',
      balanceSheetItems,
      balanceSheetMonthlyData.value
    )

    addTable(
      'Laba Rugi (Profit & Loss)',
      profitLossItems,
      profitLossMonthlyData.value
    )

    addTable(
      'Cashflow',
      cashflowItems,
      cashflowMonthlyData.value
    )

    addTable(
      'Rasio Keuangan (%)',
      ratioItems,
      ratioMonthlyData.value
    )

    // Generate filename
    let filename = `Laporan_Keuangan_${company.value.name.replace(/\s+/g, '_')}`
    if (periodRange.value && periodRange.value[0] && periodRange.value[1]) {
      filename += `_${periodRange.value[0].format('YYYY-MM')}_${periodRange.value[1].format('YYYY-MM')}`
    }
    filename += '.pdf'

    // Save PDF
    doc.save(filename)

    message.success('Export PDF berhasil')
  } catch (error: unknown) {
    console.error('Export PDF error:', error)
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal export PDF: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
  } finally {
    exportLoading.value = false
  }
}

// Export Excel
const handleExportExcel = async () => {
  if (!company.value) {
    message.error('Company tidak ditemukan')
    return
  }

  // Check if period range is selected
  if (!periodRange.value || !periodRange.value[0] || !periodRange.value[1]) {
    message.warning('Silakan pilih periode terlebih dahulu di tab Performance')
    return
  }

  try {
    exportLoading.value = true
    
    const startPeriod = periodRange.value[0].format('YYYY-MM')
    const endPeriod = periodRange.value[1].format('YYYY-MM')

    const blob = await financialReportsApi.exportPerformanceExcel(
      company.value.id,
      startPeriod,
      endPeriod
    )
    
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')

    // Generate filename dengan filter info
    const filename = `Performance_${company.value.name.replace(/\s+/g, '_')}_${startPeriod}_${endPeriod}.xlsx`

    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    message.success('Export Excel berhasil')
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal export Excel: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
  } finally {
    exportLoading.value = false
  }
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}

// Load change history for company - include company and financial_report resources
const loadChangeHistory = async () => {
  if (!company.value) return

  historyLoading.value = true
  try {
    // Load company, financial_report, and document changes
    // Note: financial_report and document logs need to be filtered by company_id in details JSON
    const [companyResponse, financialResponse, documentResponse] = await Promise.all([
      auditApi.getUserActivityLogs({
        resource: 'company',
        resource_id: company.value.id,
        page: 1, // Load all for filtering
        pageSize: 1000, // Large page size to get all company logs
      }),
      auditApi.getUserActivityLogs({
        resource: 'financial_report',
        // Don't filter by resource_id - we'll filter by company_id in details
        page: 1,
        pageSize: 1000, // Large page size to get all financial logs
      }),
      auditApi.getUserActivityLogs({
        resource: 'document',
        // Don't filter by resource_id - we'll filter by company_id in details
        page: 1,
        pageSize: 1000, // Large page size to get all document logs
      }),
    ])
    
    // Filter financial_report and document logs by company_id in details
    const companyId = company.value.id
    const filteredFinancialLogs = financialResponse.data.filter((log) => {
      if (!log.details) return false
      try {
        const details = JSON.parse(log.details)
        return details.company_id === companyId
      } catch {
        return false
      }
    })
    
    const filteredDocumentLogs = documentResponse.data.filter((log) => {
      if (!log.details) return false
      try {
        const details = JSON.parse(log.details)
        // Include documents that have company_id matching this company
        // This includes director attachments which have company_id in details
        return details.company_id === companyId
      } catch {
        return false
      }
    })
    
    // Merge and sort by created_at (newest first)
    const allLogs = [...companyResponse.data, ...filteredFinancialLogs, ...filteredDocumentLogs].sort((a, b) => {
      return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
    })
    
    // Apply pagination to merged results
    const startIndex = (historyPagination.value.current - 1) * historyPagination.value.pageSize
    const endIndex = startIndex + historyPagination.value.pageSize
    const paginatedLogs = allLogs.slice(startIndex, endIndex)
    
    changeHistory.value = paginatedLogs
    historyPagination.value.total = allLogs.length
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(err.response?.data?.message || err.message || 'Gagal memuat history perubahan')
  } finally {
    historyLoading.value = false
  }
}

// Handle history table change (pagination)
const handleHistoryTableChange = (pagination: { current: number; pageSize: number }) => {
  historyPagination.value.current = pagination.current
  historyPagination.value.pageSize = pagination.pageSize
  loadChangeHistory()
}

// Format change description from audit log details - returns array of change strings
const formatChangeDescription = (log: UserActivityLog): string[] => {
  if (!log.details) {
    if (log.resource === 'financial_report') {
      return [`${getActionLabel(log.action)} pada laporan keuangan`]
    }
    if (log.resource === 'document') {
      return [`${getActionLabel(log.action)} dokumen`]
    }
    return [`${getActionLabel(log.action)} pada data perusahaan`]
  }

  try {
    const details = JSON.parse(log.details)
    const changes: string[] = []

    // Handle financial_report resource
    if (log.resource === 'financial_report') {
      const reportType = details.type || 'Laporan Keuangan'
      const period = details.period || details.year || 'tidak diketahui'
      const year = details.year || 'tidak diketahui'
      
      if (log.action === 'create') {
        changes.push(`membuat ${reportType} baru untuk periode ${period} (${year})`)
        // Show all initial values if available
        if (details.changes && typeof details.changes === 'object') {
          for (const [field, changeData] of Object.entries(details.changes)) {
            if (changeData && typeof changeData === 'object' && 'new' in changeData) {
              const fieldData = changeData as FieldChangeData
              const fieldLabel = getFieldLabel(field, 'financial_report')
              const newValue = formatFieldValue(fieldData.new)
              if (newValue && newValue !== '0' && newValue !== '0.00') {
                changes.push(`  - ${fieldLabel}: ${newValue}`)
              }
            }
          }
        }
      } else if (log.action === 'update') {
        changes.push(`mengubah ${reportType} untuk periode ${period} (${year}):`)
        // Show all field changes
        if (details.changes && typeof details.changes === 'object') {
          for (const [field, changeData] of Object.entries(details.changes)) {
            if (changeData && typeof changeData === 'object' && 'old' in changeData && 'new' in changeData) {
              const fieldData = changeData as FieldChangeData
              const fieldLabel = getFieldLabel(field, 'financial_report')
              const oldValue = formatFieldValue(fieldData.old)
              const newValue = formatFieldValue(fieldData.new)
              changes.push(`  - ${fieldLabel}: dari "${oldValue}" menjadi "${newValue}"`)
            }
          }
        }
        if (changes.length === 1) {
          // No field changes detected, show generic message
          changes.push('  - Data telah diubah')
        }
      } else if (log.action === 'delete') {
        changes.push(`menghapus ${reportType} untuk periode ${period} (${year})`)
      }
    }
    // Handle different action types for company
    else if (log.action === 'update' || log.action === 'update_company') {
      // Extract changes from details
      if (details.changes && typeof details.changes === 'object') {
        for (const [field, changeData] of Object.entries(details.changes)) {
          if (changeData && typeof changeData === 'object') {
            const data = changeData as ChangeData
            // Handle added directors/shareholders
            if (field.startsWith('director_added_') && 'action' in data && data.action === 'added') {
              const directorData = data as DirectorChangeData
              const position = directorData.position || 'tidak diketahui'
              const fullName = directorData.full_name || 'tidak diketahui'
              changes.push(`menambahkan pengurus baru: ${position} - ${fullName}`)
            } else if (field.startsWith('director_removed_') && 'action' in data && data.action === 'removed') {
              const directorData = data as DirectorChangeData
              const position = directorData.position || 'tidak diketahui'
              const fullName = directorData.full_name || 'tidak diketahui'
              changes.push(`menghapus pengurus: ${position} - ${fullName}`)
            } else if (field.startsWith('shareholder_added_') && 'action' in data && data.action === 'added') {
              const shareholderData = data as ShareholderChangeData
              const name = shareholderData.name || 'tidak diketahui'
              const type = shareholderData.type ? ` (${shareholderData.type})` : ''
              changes.push(`menambahkan pemegang saham baru: ${name}${type}`)
            } else if (field.startsWith('shareholder_removed_') && 'action' in data && data.action === 'removed') {
              const shareholderData = data as ShareholderChangeData
              const name = shareholderData.name || 'tidak diketahui'
              const type = shareholderData.type ? ` (${shareholderData.type})` : ''
              changes.push(`menghapus pemegang saham: ${name}${type}`)
            } else if ('old' in data && 'new' in data) {
              // Regular field change
              const fieldData = data as FieldChangeData
              const fieldLabel = getFieldLabel(field)
              const oldValue = formatFieldValue(fieldData.old)
              const newValue = formatFieldValue(fieldData.new)
              changes.push(`${fieldLabel} dari sebelumnya "${oldValue}" menjadi "${newValue}"`)
            }
          }
        }
      } else if (details.field && details.old_value !== undefined && details.new_value !== undefined) {
        // Single field change format
        const fieldLabel = getFieldLabel(details.field)
        const oldValue = formatFieldValue(details.old_value)
        const newValue = formatFieldValue(details.new_value)
        changes.push(`${fieldLabel} dari sebelumnya "${oldValue}" menjadi "${newValue}"`)
      }
    } else if (log.action === 'create' || log.action === 'create_company') {
      changes.push('membuat data perusahaan baru')
    } else if (log.action === 'delete' || log.action === 'delete_company') {
      changes.push('menghapus data perusahaan')
    }
    // Handle document resource
    else if (log.resource === 'document') {
      const fileName = details.file_name || details.document_name || 'dokumen'
      const documentName = details.document_name || fileName
      
      if (log.action === 'delete' || log.action === 'delete_document') {
        if (details.document_type === 'director_attachment') {
          changes.push(`menghapus dokumen attachment pengurus: ${documentName}`)
          if (details.director_id) {
            changes.push(`  - ID Pengurus: ${details.director_id}`)
          }
        } else {
          changes.push(`menghapus dokumen: ${documentName}`)
        }
        if (details.folder_id) {
          changes.push(`  - Folder ID: ${details.folder_id}`)
        }
      } else if (log.action === 'create' || log.action === 'create_document') {
        if (details.document_type === 'director_attachment') {
          changes.push(`mengunggah dokumen attachment pengurus: ${documentName}`)
          if (details.director_id) {
            changes.push(`  - ID Pengurus: ${details.director_id}`)
          }
        } else {
          changes.push(`mengunggah dokumen: ${documentName}`)
        }
      } else if (log.action === 'update' || log.action === 'update_document') {
        if (details.document_type === 'director_attachment') {
          changes.push(`memperbarui dokumen attachment pengurus: ${documentName}`)
          if (details.director_id) {
            changes.push(`  - ID Pengurus: ${details.director_id}`)
          }
        } else {
          changes.push(`memperbarui dokumen: ${documentName}`)
        }
      }
    }

    if (changes.length > 0) {
      return changes
    }
  } catch (error) {
    console.error('Error parsing change details:', error)
  }

  // Fallback based on resource type
  if (log.resource === 'financial_report') {
    return [`${getActionLabel(log.action)} pada laporan keuangan`]
  }
  if (log.resource === 'document') {
    return [`${getActionLabel(log.action)} dokumen`]
  }
  return [`${getActionLabel(log.action)} pada data perusahaan`]
}

// Get field label in Indonesian
const getFieldLabel = (field: string, resource?: string): string => {
  // Handle financial_report fields
  if (resource === 'financial_report') {
    const financialFieldLabels: Record<string, string> = {
      // Neraca
      current_assets: 'Aset Lancar',
      non_current_assets: 'Aset Tidak Lancar',
      short_term_liabilities: 'Liabilitas Jangka Pendek',
      long_term_liabilities: 'Liabilitas Jangka Panjang',
      equity: 'Ekuitas',
      // Laba Rugi
      revenue: 'Revenue',
      operating_expenses: 'Beban Usaha',
      operating_profit: 'Laba Usaha',
      other_income: 'Pendapatan Lain-Lain',
      tax: 'Tax',
      net_profit: 'Laba Bersih',
      // Cashflow
      operating_cashflow: 'Arus kas bersih dari operasi',
      investing_cashflow: 'Arus kas bersih dari investasi',
      financing_cashflow: 'Arus kas bersih dari pendanaan',
      ending_balance: 'Saldo Akhir',
      // Rasio
      roe: 'ROE (Return on Equity)',
      roi: 'ROI (Return on Investment)',
      current_ratio: 'Rasio Lancar',
      cash_ratio: 'Rasio Kas',
      ebitda: 'EBITDA',
      ebitda_margin: 'EBITDA Margin',
      net_profit_margin: 'Net Profit Margin',
      operating_profit_margin: 'Operating Profit Margin',
      debt_to_equity: 'Debt to Equity',
      // Metadata
      year: 'Tahun',
      period: 'Periode',
      is_rkap: 'Jenis Laporan',
      remark: 'Keterangan',
    }
    return financialFieldLabels[field] || field
  }
  
  // Handle company fields (existing logic)
  // Handle director fields with index pattern: director_{index}_{field}
  const directorMatch = field.match(/^director_(\d+)_(.+)$/)
  if (directorMatch && directorMatch[1] && directorMatch[2]) {
    const index = parseInt(directorMatch[1], 10) + 1 // Convert to 1-based for display
    const subField = directorMatch[2]
    const subFieldLabels: Record<string, string> = {
      position: 'jabatan',
      full_name: 'nama lengkap',
      ktp: 'nomor KTP',
      npwp: 'nomor NPWP',
      start_date: 'tanggal awal jabatan',
      domicile_address: 'alamat domisili',
    }
    const subFieldLabel = subFieldLabels[subField] || subField
    return `pengurus ${index} - ${subFieldLabel}`
  }

  // Handle shareholder fields with index pattern: shareholder_{index}_{field}
  const shareholderMatch = field.match(/^shareholder_(\d+)_(.+)$/)
  if (shareholderMatch && shareholderMatch[1] && shareholderMatch[2]) {
    const index = parseInt(shareholderMatch[1], 10) + 1 // Convert to 1-based for display
    const subField = shareholderMatch[2]
    const subFieldLabels: Record<string, string> = {
      name: 'nama',
      type: 'jenis pemegang saham',
      identity_number: 'nomor identitas',
      ownership_percent: 'persentase kepemilikan',
      share_sheet_count: 'jumlah lembar saham',
      share_value_per_sheet: 'nilai rupiah per lembar',
    }
    const subFieldLabel = subFieldLabels[subField] || subField
    return `pemegang saham ${index} - ${subFieldLabel}`
  }

  const fieldLabels: Record<string, string> = {
    name: 'nama perusahaan',
    short_name: 'nama singkat',
    description: 'deskripsi',
    npwp: 'NPWP',
    nib: 'NIB',
    status: 'status',
    phone: 'telepon',
    fax: 'fax',
    email: 'email',
    website: 'website',
    address: 'alamat perusahaan',
    operational_address: 'alamat operasional',
    code: 'kode perusahaan',
    parent_id: 'perusahaan induk',
    authorized_capital: 'modal dasar',
    paid_up_capital: 'modal disetor',
    // Shareholder fields (without index)
    shareholder_name: 'nama pemegang saham',
    shareholder_type: 'jenis pemegang saham',
    shareholder_identity_number: 'nomor identitas pemegang saham',
    shareholder_ownership_percent: 'persentase kepemilikan',
    shareholder_share_sheet_count: 'jumlah lembar saham',
    shareholder_share_value_per_sheet: 'nilai rupiah per lembar',
    // Director fields (without index)
    director_position: 'jabatan pengurus',
    director_full_name: 'nama lengkap pengurus',
    director_ktp: 'nomor KTP pengurus',
    director_npwp: 'nomor NPWP pengurus',
    director_start_date: 'tanggal awal jabatan pengurus',
    director_domicile_address: 'alamat domisili pengurus',
    // Business field
    business_industry_sector: 'sektor industri',
    business_kbli: 'KBLI',
    business_main_activity: 'uraian kegiatan usaha utama',
    business_additional_activities: 'kegiatan usaha tambahan',
    business_start_operation_date: 'tanggal mulai beroperasi',
  }
  return fieldLabels[field] || field
}

// Format field value for display
const formatFieldValue = (value: unknown): string => {
  if (value === null || value === undefined || value === '') {
    return '(kosong)'
  }
  if (typeof value === 'boolean') {
    return value ? 'Ya' : 'Tidak'
  }
  // Format large numbers (financial values)
  if (typeof value === 'number') {
    if (value >= 1000000000) {
      return `Rp ${(value / 1000000000).toFixed(2)}M`
    } else if (value >= 1000000) {
      return `Rp ${(value / 1000000).toFixed(2)}Jt`
    } else if (value >= 1000) {
      return `Rp ${(value / 1000).toFixed(2)}Rb`
    } else if (value < 100 && value > 0) {
      // Likely a ratio/percentage
      return `${value.toFixed(2)}%`
    }
    return value.toLocaleString('id-ID')
  }
  if (typeof value === 'object') {
    return JSON.stringify(value)
  }
  return String(value)
}

// Get action label in Indonesian
const getActionLabel = (action: string): string => {
  const actionLabels: Record<string, string> = {
    create: 'pembuatan',
    update: 'perubahan',
    delete: 'penghapusan',
    create_company: 'pembuatan',
    update_company: 'perubahan',
    delete_company: 'penghapusan',
  }
  return actionLabels[action] || action
}

// Format date and time
const formatDateTime = (dateString: string): string => {
  if (!dateString) return '-'
  return dayjs(dateString).format('DD MMM YYYY, HH:mm:ss')
}

// Load financial reports for company
const loadFinancialReports = async (companyId: string) => {
  financialReportsLoading.value = true
  try {
    financialReports.value = await financialReportsApi.getByCompanyId(companyId)
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    console.error('Failed to load financial reports:', err.response?.data?.message || err.message)
    financialReports.value = []
  } finally {
    financialReportsLoading.value = false
  }
}

// Load financial comparison (RKAP vs Realisasi YTD)
const loadFinancialComparison = async (companyId: string) => {
  if (!selectedYear.value || !startMonth.value || !endMonth.value) return
  
  // Use endMonth for comparison (YTD up to end month)
  financialComparisonLoading.value = true
  try {
    financialComparison.value = await financialReportsApi.getComparison(companyId, selectedYear.value, endMonth.value)
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } }; message?: string }
    // Don't show error if no data found (RKAP or Realisasi might not exist yet)
    if (err.response?.data?.message && !err.response.data.message.includes('not found')) {
      console.error('Failed to load financial comparison:', err.response.data.message)
    }
    financialComparison.value = null
  } finally {
    financialComparisonLoading.value = false
  }
}


// Handle period change for financial comparison
const handleFinancialPeriodChange = async () => {
  // Validate date range
  if (!periodRange.value || !periodRange.value[0] || !periodRange.value[1]) {
    message.warning('Silakan pilih periode yang valid')
    return
  }
  
  // Validate that start month is before or equal to end month
  if (periodRange.value[0].isAfter(periodRange.value[1])) {
    message.warning('Bulan awal harus lebih kecil atau sama dengan bulan akhir')
    return
  }
  
  // Validate that both months are in the same year
  if (periodRange.value[0].format('YYYY') !== periodRange.value[1].format('YYYY')) {
    message.warning('Periode harus dalam tahun yang sama')
    // Auto-correct: set end month to same year as start month
    periodRange.value = [periodRange.value[0], periodRange.value[0].endOf('year')]
  }
  
  if (company.value) {
    // Reload financial reports to get monthly data for the selected year
    await loadFinancialReports(company.value.id)
    await loadFinancialComparison(company.value.id)
  }
}

// Format currency helper - uses company currency setting
const formatCurrencyValue = (value: number | string | undefined): string => {
  if (value === undefined || value === null) return '-'
  const numValue = typeof value === 'string' ? parseFloat(value) : value
  if (isNaN(numValue)) return '-'
  
  // Get currency from company, default to IDR (Rupiah)
  const currency = company.value?.currency || 'IDR'
  const absValue = Math.abs(numValue)
  const sign = numValue < 0 ? '-' : ''
  
  // Format based on currency
  if (currency === 'USD') {
    // USD format: $32B, $129M, $5K
    if (absValue >= 1000000000) {
      return `${sign}$${(absValue / 1000000000).toFixed(2)}B`
    } else if (absValue >= 1000000) {
      return `${sign}$${(absValue / 1000000).toFixed(2)}M`
    } else if (absValue >= 1000) {
      return `${sign}$${(absValue / 1000).toFixed(2)}K`
    }
    return `${sign}$${absValue.toLocaleString('en-US')}`
  } else {
    // IDR format: Rp 129M, Rp 32B, Rp 5K
    if (absValue >= 1000000000) {
      return `${sign}Rp ${(absValue / 1000000000).toFixed(2)}B`
    } else if (absValue >= 1000000) {
      return `${sign}Rp ${(absValue / 1000000).toFixed(2)}Jt`
    } else if (absValue >= 1000) {
      return `${sign}Rp ${(absValue / 1000).toFixed(2)}Rb`
    }
    return `${sign}Rp ${absValue.toLocaleString('id-ID')}`
  }
}


// Format ratio helper (for ratios like current_ratio, debt_to_equity)
const formatRatioValue = (value: number | string | undefined): string => {
  if (value === undefined || value === null) return '-'
  const numValue = typeof value === 'string' ? parseFloat(value) : value
  if (isNaN(numValue)) return '-'
  return numValue.toFixed(2)
}

// Handle tab change - load data when tab is selected
const handleTabChange = (activeKey: string) => {
  if (activeKey === 'history' && company.value) {
    loadChangeHistory()
  } else if (activeKey === 'performance' && company.value) {
    loadFinancialComparison(company.value.id)
  } else if (activeKey === 'input-laporan' && company.value) {
    loadFinancialReports(company.value.id)
  }
}

// Handle financial report saved
const handleFinancialReportSaved = async () => {
  if (company.value) {
    await loadFinancialReports(company.value.id)
    await loadFinancialComparison(company.value.id)
    message.success('Laporan keuangan berhasil disimpan')
  }
}



// Helper function to get formatted cell value
const getCellValue = (
  columnKey: string | undefined,
  record: Record<string, unknown>,
  items: Array<{ key: string; isRatio: boolean }>,
  valueType: 'rkap' | 'realisasi' | 'difference'
): string => {
  if (!columnKey) return '-'
  
  // Extract item key from column key (e.g., "current_assets_rkap" -> "current_assets")
  const itemKey = columnKey.replace(`_${valueType}`, '')
  const item = items.find(i => i.key === itemKey)
  
  if (!item) return '-'
  
  const value = record[columnKey]
  
  if (value === undefined || value === null) return '-'
  
  // Convert to number safely
  if (typeof value !== 'number' && typeof value !== 'string') {
    return '-'
  }
  
  const numValue = typeof value === 'number' ? value : parseFloat(value)
  
  if (isNaN(numValue)) return '-'
  
  if (item.isRatio) {
    return formatRatioValue(numValue)
  } else {
    return formatCurrencyValue(numValue)
  }
}

// Generate columns with merged headers for financial items
const generateMergedColumns = (items: Array<{ key: string; label: string; field: string; isRatio: boolean }>) => {
  const baseColumns = [
    {
      title: 'Bulan',
      key: 'month',
      dataIndex: 'month',
      width: 120,
      fixed: 'left' as const,
      align: 'left' as const,
    },
  ]

  const itemColumns = items.map((item) => ({
    title: item.label,
    key: item.key,
    align: 'center' as const,
    children: [
      {
        title: 'RKAP',
        key: `${item.key}_rkap`,
        dataIndex: `${item.key}_rkap`,
        align: 'right' as const,
        width: 120,
      },
      {
        title: 'Realisasi',
        key: `${item.key}_realisasi`,
        dataIndex: `${item.key}_realisasi`,
        align: 'right' as const,
        width: 120,
      },
    ],
  }))

  return [...baseColumns, ...itemColumns]
}

// Table columns sudah di-generate secara dinamis dengan merged headers

// Generate monthly data with all items in one row
const generateMonthlyDataWithAllItems = (items: Array<{ key: string; label: string; field: string; isRatio: boolean }>) => {
  if (!financialReports.value || !selectedYear.value || !startMonth.value || !endMonth.value) return []
  
  const allMonths = ['01', '02', '03', '04', '05', '06', '07', '08', '09', '10', '11', '12']
  const monthNames = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember']
  
  const startIndex = allMonths.indexOf(startMonth.value)
  const endIndex = allMonths.indexOf(endMonth.value)
  
  if (startIndex === -1 || endIndex === -1 || startIndex > endIndex) return []
  
  const filteredMonths = allMonths.slice(startIndex, endIndex + 1)
  
  // Get RKAP for the year
  const rkap = financialReports.value.find(r => r.is_rkap && r.year === selectedYear.value)
  const rkapRecord = rkap ? (rkap as unknown as Record<string, unknown>) : undefined
  
  // Generate data for each month
  return filteredMonths.map((month) => {
    const period = `${selectedYear.value}-${month}`
    const realisasi = financialReports.value.find(r => !r.is_rkap && r.period === period)
    const realisasiRecord = realisasi ? (realisasi as unknown as Record<string, unknown>) : undefined
    
    const monthIndex = allMonths.indexOf(month)
    const rowData: Record<string, unknown> = {
      key: month,
      month: monthNames[monthIndex] || month,
    }
    
    // Add data for each item - tampilkan data apa adanya tanpa perhitungan
    items.forEach((item) => {
      // RKAP: tampilkan nilai tahunan langsung (tidak dibagi 12)
      const rkapAnnualValue = rkapRecord ? ((rkapRecord[item.field] as number | undefined) ?? 0) : 0
      
      // Realisasi: tampilkan nilai bulanan apa adanya
      const realisasiValue = realisasiRecord ? ((realisasiRecord[item.field] as number | undefined) ?? 0) : 0
      
      // Tampilkan data apa adanya - tidak ada perhitungan otomatis
      rowData[`${item.key}_rkap`] = rkapAnnualValue
      rowData[`${item.key}_realisasi`] = realisasiValue
    })
    
    return rowData
  })
}

// Define items for each category
const balanceSheetItems = [
  { key: 'current_assets', label: 'A. Aset Lancar', field: 'current_assets', isRatio: false },
  { key: 'non_current_assets', label: 'B. Aset Tidak Lancar', field: 'non_current_assets', isRatio: false },
  { key: 'short_term_liabilities', label: 'C. Liabilitas Jangka Pendek', field: 'short_term_liabilities', isRatio: false },
  { key: 'long_term_liabilities', label: 'D. Liabilitas Jangka Panjang', field: 'long_term_liabilities', isRatio: false },
  { key: 'equity', label: 'E. Ekuitas', field: 'equity', isRatio: false },
]

const profitLossItems = [
  { key: 'revenue', label: 'A. Revenue', field: 'revenue', isRatio: false },
  { key: 'operating_expenses', label: 'B. Beban Usaha', field: 'operating_expenses', isRatio: false },
  { key: 'operating_profit', label: 'C. Laba Usaha', field: 'operating_profit', isRatio: false },
  { key: 'other_income', label: 'D. Pendapatan Lain-Lain', field: 'other_income', isRatio: false },
  { key: 'tax', label: 'E. Tax', field: 'tax', isRatio: false },
  { key: 'net_profit', label: 'F. Laba Bersih', field: 'net_profit', isRatio: false },
]

const cashflowItems = [
  { key: 'operating_cashflow', label: 'A. Arus kas bersih dari operasi', field: 'operating_cashflow', isRatio: false },
  { key: 'investing_cashflow', label: 'B. Arus kas bersih dari investasi', field: 'investing_cashflow', isRatio: false },
  { key: 'financing_cashflow', label: 'C. Arus kas bersih dari pendanaan', field: 'financing_cashflow', isRatio: false },
  { key: 'ending_balance', label: 'D. Saldo Akhir', field: 'ending_balance', isRatio: false },
]

const ratioItems = [
  { key: 'roe', label: 'ROE (Return on Equity)', field: 'roe', isRatio: true },
  { key: 'roi', label: 'ROI (Return on Investment)', field: 'roi', isRatio: true },
  { key: 'current_ratio', label: 'Rasio Lancar', field: 'current_ratio', isRatio: true },
  { key: 'cash_ratio', label: 'Rasio Kas', field: 'cash_ratio', isRatio: true },
  { key: 'ebitda', label: 'EBITDA', field: 'ebitda', isRatio: false },
  { key: 'ebitda_margin', label: 'EBITDA Margin', field: 'ebitda_margin', isRatio: true },
  { key: 'net_profit_margin', label: 'Net Profit Margin', field: 'net_profit_margin', isRatio: true },
  { key: 'operating_profit_margin', label: 'Operating Profit Margin', field: 'operating_profit_margin', isRatio: true },
  { key: 'debt_to_equity', label: 'Debt to Equity', field: 'debt_to_equity', isRatio: true },
]

// Computed columns and data for merged table structure
const balanceSheetColumns = computed(() => generateMergedColumns(balanceSheetItems))
const balanceSheetMonthlyData = computed(() => generateMonthlyDataWithAllItems(balanceSheetItems))

const profitLossColumns = computed(() => generateMergedColumns(profitLossItems))
const profitLossMonthlyData = computed(() => generateMonthlyDataWithAllItems(profitLossItems))

const cashflowColumns = computed(() => generateMergedColumns(cashflowItems))
const cashflowMonthlyData = computed(() => generateMonthlyDataWithAllItems(cashflowItems))

const ratioColumns = computed(() => generateMergedColumns(ratioItems))
const ratioMonthlyData = computed(() => generateMonthlyDataWithAllItems(ratioItems))

// Chart data for Balance Sheet Overview (main chart)
const balanceSheetOverviewChartData = computed(() => {
  if (!financialReports.value || !selectedYear.value || !startMonth.value || !endMonth.value) return []
  
  const allMonths = ['01', '02', '03', '04', '05', '06', '07', '08', '09', '10', '11', '12']
  const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des']
  
  const startIndex = allMonths.indexOf(startMonth.value)
  const endIndex = allMonths.indexOf(endMonth.value)
  
  if (startIndex === -1 || endIndex === -1 || startIndex > endIndex) return []
  
  const filteredMonths = allMonths.slice(startIndex, endIndex + 1)
  
  // Get RKAP for the year
  const rkap = financialReports.value.find(r => r.is_rkap && r.year === selectedYear.value)
  const rkapRecord = rkap ? (rkap as unknown as Record<string, unknown>) : undefined
  
  // Calculate monthly values for Total Assets, Total Liabilities, and Equity
  return filteredMonths.map((month) => {
    const period = `${selectedYear.value}-${month}`
    const realisasi = financialReports.value.find(r => !r.is_rkap && r.period === period)
    const realisasiRecord = realisasi ? (realisasi as unknown as Record<string, unknown>) : undefined
    
    const monthIndex = allMonths.indexOf(month)
    
    // Calculate Total Assets = current_assets + non_current_assets
    // RKAP: tampilkan nilai tahunan langsung (tidak dibagi 12)
    const rkapTotalAssets = rkapRecord 
      ? ((rkapRecord['current_assets'] as number | undefined) ?? 0) + ((rkapRecord['non_current_assets'] as number | undefined) ?? 0)
      : 0
    
    const realisasiTotalAssets = realisasiRecord
      ? ((realisasiRecord['current_assets'] as number | undefined) ?? 0) + ((realisasiRecord['non_current_assets'] as number | undefined) ?? 0)
      : 0
    
    // Calculate Total Liabilities = short_term_liabilities + long_term_liabilities
    const rkapTotalLiabilities = rkapRecord
      ? ((rkapRecord['short_term_liabilities'] as number | undefined) ?? 0) + ((rkapRecord['long_term_liabilities'] as number | undefined) ?? 0)
      : 0
    
    const realisasiTotalLiabilities = realisasiRecord
      ? ((realisasiRecord['short_term_liabilities'] as number | undefined) ?? 0) + ((realisasiRecord['long_term_liabilities'] as number | undefined) ?? 0)
      : 0
    
    // Equity
    const rkapEquity = rkapRecord ? ((rkapRecord['equity'] as number | undefined) ?? 0) : 0
    const realisasiEquity = realisasiRecord ? ((realisasiRecord['equity'] as number | undefined) ?? 0) : 0
    
    return {
      label: monthNames[monthIndex] || month,
      totalAssets: {
        rkap: rkapTotalAssets, // Nilai tahunan langsung
        realisasi: realisasiTotalAssets,
      },
      totalLiabilities: {
        rkap: rkapTotalLiabilities, // Nilai tahunan langsung
        realisasi: realisasiTotalLiabilities,
      },
      equity: {
        rkap: rkapEquity, // Nilai tahunan langsung
        realisasi: realisasiEquity,
      },
    }
  })
})

// Chart data for Profit Loss Overview (Revenue vs Net Profit)
const profitLossOverviewChartData = computed(() => {
  if (!financialReports.value || !selectedYear.value || !startMonth.value || !endMonth.value) return []
  
  const allMonths = ['01', '02', '03', '04', '05', '06', '07', '08', '09', '10', '11', '12']
  const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des']
  
  const startIndex = allMonths.indexOf(startMonth.value)
  const endIndex = allMonths.indexOf(endMonth.value)
  
  if (startIndex === -1 || endIndex === -1 || startIndex > endIndex) return []
  
  const filteredMonths = allMonths.slice(startIndex, endIndex + 1)
  
  // Get RKAP for the year
  const rkap = financialReports.value.find(r => r.is_rkap && r.year === selectedYear.value)
  const rkapRecord = rkap ? (rkap as unknown as Record<string, unknown>) : undefined
  
  return filteredMonths.map((month) => {
    const period = `${selectedYear.value}-${month}`
    const realisasi = financialReports.value.find(r => !r.is_rkap && r.period === period)
    const realisasiRecord = realisasi ? (realisasi as unknown as Record<string, unknown>) : undefined
    
    const monthIndex = allMonths.indexOf(month)
    
    // Revenue - RKAP: nilai tahunan langsung
    const rkapRevenue = rkapRecord ? ((rkapRecord['revenue'] as number | undefined) ?? 0) : 0
    const realisasiRevenue = realisasiRecord ? ((realisasiRecord['revenue'] as number | undefined) ?? 0) : 0
    
    // Net Profit - RKAP: nilai tahunan langsung
    const rkapNetProfit = rkapRecord ? ((rkapRecord['net_profit'] as number | undefined) ?? 0) : 0
    const realisasiNetProfit = realisasiRecord ? ((realisasiRecord['net_profit'] as number | undefined) ?? 0) : 0
    
    return {
      label: monthNames[monthIndex] || month,
      revenue: {
        rkap: rkapRevenue, // Nilai tahunan langsung
        realisasi: realisasiRevenue,
      },
      netProfit: {
        rkap: rkapNetProfit, // Nilai tahunan langsung
        realisasi: realisasiNetProfit,
      },
    }
  })
})

// Chart data for Cashflow Overview (Net Cashflow vs Ending Balance)
const cashflowOverviewChartData = computed(() => {
  if (!financialReports.value || !selectedYear.value || !startMonth.value || !endMonth.value) return []
  
  const allMonths = ['01', '02', '03', '04', '05', '06', '07', '08', '09', '10', '11', '12']
  const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des']
  
  const startIndex = allMonths.indexOf(startMonth.value)
  const endIndex = allMonths.indexOf(endMonth.value)
  
  if (startIndex === -1 || endIndex === -1 || startIndex > endIndex) return []
  
  const filteredMonths = allMonths.slice(startIndex, endIndex + 1)
  
  // Get RKAP for the year
  const rkap = financialReports.value.find(r => r.is_rkap && r.year === selectedYear.value)
  const rkapRecord = rkap ? (rkap as unknown as Record<string, unknown>) : undefined
  
  return filteredMonths.map((month) => {
    const period = `${selectedYear.value}-${month}`
    const realisasi = financialReports.value.find(r => !r.is_rkap && r.period === period)
    const realisasiRecord = realisasi ? (realisasi as unknown as Record<string, unknown>) : undefined
    
    const monthIndex = allMonths.indexOf(month)
    
    // Net Cashflow = Operating + Investing + Financing
    // RKAP: nilai tahunan langsung
    const rkapOperating = rkapRecord ? ((rkapRecord['operating_cashflow'] as number | undefined) ?? 0) : 0
    const rkapInvesting = rkapRecord ? ((rkapRecord['investing_cashflow'] as number | undefined) ?? 0) : 0
    const rkapFinancing = rkapRecord ? ((rkapRecord['financing_cashflow'] as number | undefined) ?? 0) : 0
    const rkapNetCashflow = rkapOperating + rkapInvesting + rkapFinancing
    
    const realisasiOperating = realisasiRecord ? ((realisasiRecord['operating_cashflow'] as number | undefined) ?? 0) : 0
    const realisasiInvesting = realisasiRecord ? ((realisasiRecord['investing_cashflow'] as number | undefined) ?? 0) : 0
    const realisasiFinancing = realisasiRecord ? ((realisasiRecord['financing_cashflow'] as number | undefined) ?? 0) : 0
    const realisasiNetCashflow = realisasiOperating + realisasiInvesting + realisasiFinancing
    
    // Ending Balance - RKAP: nilai tahunan langsung
    const rkapEndingBalance = rkapRecord ? ((rkapRecord['ending_balance'] as number | undefined) ?? 0) : 0
    const realisasiEndingBalance = realisasiRecord ? ((realisasiRecord['ending_balance'] as number | undefined) ?? 0) : 0
    
    return {
      label: monthNames[monthIndex] || month,
      netCashflow: {
        rkap: rkapNetCashflow, // Nilai tahunan langsung
        realisasi: realisasiNetCashflow,
      },
      endingBalance: {
        rkap: rkapEndingBalance, // Nilai tahunan langsung
        realisasi: realisasiEndingBalance,
      },
    }
  })
})

// Chart data for Ratio Overview (ROE, ROI, Current Ratio, Debt-to-Equity)
const ratioOverviewChartData = computed(() => {
  if (!financialReports.value || !selectedYear.value || !startMonth.value || !endMonth.value) return []
  
  const allMonths = ['01', '02', '03', '04', '05', '06', '07', '08', '09', '10', '11', '12']
  const monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun', 'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des']
  
  const startIndex = allMonths.indexOf(startMonth.value)
  const endIndex = allMonths.indexOf(endMonth.value)
  
  if (startIndex === -1 || endIndex === -1 || startIndex > endIndex) return []
  
  const filteredMonths = allMonths.slice(startIndex, endIndex + 1)
  
  // Get RKAP for the year
  const rkap = financialReports.value.find(r => r.is_rkap && r.year === selectedYear.value)
  const rkapRecord = rkap ? (rkap as unknown as Record<string, unknown>) : undefined
  
  return filteredMonths.map((month) => {
    const period = `${selectedYear.value}-${month}`
    const realisasi = financialReports.value.find(r => !r.is_rkap && r.period === period)
    const realisasiRecord = realisasi ? (realisasi as unknown as Record<string, unknown>) : undefined
    
    const monthIndex = allMonths.indexOf(month)
    
    // ROE
    const rkapROE = rkapRecord ? ((rkapRecord['roe'] as number | undefined) ?? 0) : 0
    const realisasiROE = realisasiRecord ? ((realisasiRecord['roe'] as number | undefined) ?? 0) : 0
    
    // ROI
    const rkapROI = rkapRecord ? ((rkapRecord['roi'] as number | undefined) ?? 0) : 0
    const realisasiROI = realisasiRecord ? ((realisasiRecord['roi'] as number | undefined) ?? 0) : 0
    
    // Current Ratio
    const rkapCurrentRatio = rkapRecord ? ((rkapRecord['current_ratio'] as number | undefined) ?? 0) : 0
    const realisasiCurrentRatio = realisasiRecord ? ((realisasiRecord['current_ratio'] as number | undefined) ?? 0) : 0
    
    // Debt to Equity
    const rkapDebtToEquity = rkapRecord ? ((rkapRecord['debt_to_equity'] as number | undefined) ?? 0) : 0
    const realisasiDebtToEquity = realisasiRecord ? ((realisasiRecord['debt_to_equity'] as number | undefined) ?? 0) : 0
    
    return {
      label: monthNames[monthIndex] || month,
      roe: {
        rkap: rkapROE,
        realisasi: realisasiROE,
      },
      roi: {
        rkap: rkapROI,
        realisasi: realisasiROI,
      },
      currentRatio: {
        rkap: rkapCurrentRatio,
        realisasi: realisasiCurrentRatio,
      },
      debtToEquity: {
        rkap: rkapDebtToEquity,
        realisasi: realisasiDebtToEquity,
      },
    }
  })
})

onMounted(() => {
  loadCompany()
})
</script>

<style scoped>
.subsidiary-detail-layout {
  min-height: 100vh;
}

.back-button-container {
  margin-top: 100px;
  margin-bottom: -100px;
}

.back-button-container .back-button {
  margin-bottom: 24px;
  padding: 0;
  height: auto;
  display: flex;
  align-items: center;
  padding: 5px 8px;
  margin: 24px;
}

.detail-content {
  margin: 0 auto;
}

.back-button {
  margin-bottom: 24px;
  padding: 0;
  height: auto;
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

/* Detail card styles moved to child components */

.detail-header {
  display: flex;
  align-items: flex-start;
  gap: 24px;
  width: 100%;
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
}

.company-subtitle {
  font-size: 16px;
  color: #666;
  margin: -14px 0 16px 0;
}

.company-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
  /* background: orange; */
  margin-top: -15px;
}

.meta-item {
  font-size: 14px;
  color: #666;
}

.company-hierarchy {
  /* background: red; */
  width: 100%;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: -10px;
  padding-top: 0px;
  /* border-top: 1px solid #f0f0f0; */
}

.hierarchy-label {
  font-size: 12px;
  font-weight: 500;
  color: #333;
}

.hierarchy-path {
  display: flex;
  align-items: center;
  gap: 4px;
  flex-wrap: wrap;
}

.hierarchy-item {
  font-size: 12px;
  color: #333;
}

.hierarchy-separator {
  font-size: 14px;
  color: #999;
  margin: 0 2px;
}

.header-actions {
  flex-shrink: 0;
}

/* Ensure all buttons and date picker in header-actions have consistent height */
.header-actions :deep(.ant-btn) {
  height: 40px !important;
  min-height: 40px !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
}

.header-actions :deep(.ant-picker) {
  height: 40px !important;
}

.header-actions :deep(.ant-picker-input) {
  height: 40px !important;
}

.header-actions :deep(.ant-picker-input > input) {
  height: 38px !important;
  line-height: 38px !important;
}

/* Tabs Container */
.tabs-container {
  margin-top: 50px;
  padding: 24px;
  /* background: red; */
}

.tabs-container :deep(.ant-tabs-card) {
  background: transparent;
  background: white;
  margin-top: -50px;
  padding: 24px;
  border-radius: 
  15px;
}

.tabs-container :deep(.ant-tabs-tab) {
  border-radius: 8px 8px 0 0;
}

.tabs-container :deep(.ant-tabs-tab-active) {
  background: white;
}

/* Performance Content */
.performance-content {
  padding: 0px 0 24px 0;
  /* background: orange; */
  border-radius: 0 8px 8px 8px;
}

/* Financial Table Card - untuk table dengan banyak kolom */
.financial-table-card {
  overflow-x: auto;
}

.financial-table-card :deep(.ant-table-wrapper) {
  overflow-x: auto;
}

.financial-table-card :deep(.ant-table) {
  min-width: 100%;
}

.financial-table-card :deep(.ant-table-thead > tr > th) {
  white-space: nowrap;
  text-align: center;
  font-weight: 600;
  background: #fafafa;
}

.financial-table-card :deep(.ant-table-thead > tr > th[colspan]) {
  text-align: center;
  background: #e6f7ff;
  font-weight: 600;
}

.financial-table-card :deep(.ant-table-tbody > tr > td) {
  white-space: nowrap;
}

/* Mini Charts Grid */
.mini-charts-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-top: 24px;
}

.mini-chart-card {
  background: white;
  border-radius: 8px;
}

.mini-chart-card :deep(.ant-card-body) {
  padding: 16px;
}

.mini-chart-card :deep(.ant-card-head) {
  padding: 12px 16px;
  min-height: auto;
  border-bottom: 1px solid #f0f0f0;
}

.mini-chart-card :deep(.ant-card-head-title) {
  padding: 0;
}

/* Responsive untuk mini charts */
@media (max-width: 1200px) {
  .mini-charts-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .mini-charts-grid {
    grid-template-columns: 1fr;
  }
}

/* Input Laporan Content */
.input-laporan-content {
  padding: 24px;
  background: white;
  border-radius: 0 8px 8px 8px;
  position: relative;
  z-index: 1;
}

.input-laporan-tabs {
  position: relative;
  z-index: 1;
}

.input-laporan-tabs :deep(.ant-tabs-card) {
  background: transparent;
  margin-top: 0;
  padding: 0;
}

.input-laporan-tabs :deep(.ant-tabs-nav) {
  margin-bottom: 16px;
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
  width: 100%;
}

.trend-card-content :deep(.ant-card-body) {
  padding: 24px;
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
  margin: 0;
  padding: 0;
  overflow: visible;
  display: block;
}

.mini-chart {
  width: 100% !important;
  height: 60px;
  display: block;
  margin: 0;
  padding: 0;
  max-width: 100%;
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
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.recent-card {
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}

.full-width {
  width: 100%;
}

.card-header-title {
  display: flex;
  align-items: center;
}

/* Profile Content */
.profile-content {
  padding: 24px;
  background: white;
  border-radius: 0 8px 8px 8px;
}

/* Info Grid Layout */
.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
  margin-top: 16px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.info-item.full-width {
  grid-column: 1 / -1;
}

.info-label {
  font-size: 13px;
  font-weight: 500;
  color: #666;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.info-value {
  font-size: 15px;
  color: #1a1a1a;
  word-break: break-word;
}

.info-value a {
  color: #1890ff;
  text-decoration: none;
}

.info-value a:hover {
  text-decoration: underline;
}

/* Shareholders List */
.shareholders-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 16px;
}

.shareholder-card {
  background: #f8f9fa;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 20px;
  transition: all 0.3s ease;
}

.shareholder-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  border-color: #1890ff;
}

.shareholder-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 2px solid #e0e0e0;
}

.shareholder-name {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
  display: flex;
  align-items: center;
  gap: 8px;
}

.shareholder-badge {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 6px 14px;
  border-radius: 20px;
  font-weight: 600;
  font-size: 14px;
}

.shareholder-details {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  padding: 8px 0;
}

.info-row.full-width {
  flex-direction: column;
  gap: 4px;
}

/* Directors List - Compact Style */
.directors-list-compact {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 12px;
}

.director-item-compact {
  background: #fafafa;
  border: 1px solid #e8e8e8;
  border-radius: 6px;
  padding: 12px 16px;
  transition: all 0.2s ease;
}

.director-item-compact:hover {
  background: #f5f5f5;
  border-color: #d9d9d9;
}

.director-main-info {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.director-name-compact {
  flex: 0 0 200px;
  min-width: 200px;
}

.director-name-compact strong {
  font-size: 15px;
  color: #1a1a1a;
  display: block;
  margin-bottom: 4px;
}

.director-positions-compact {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.director-info-compact {
  flex: 1;
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: center;
  font-size: 13px;
  color: #666;
}

.info-compact {
  white-space: nowrap;
}

.info-compact.full-width-compact {
  width: 100%;
  white-space: normal;
  margin-top: 4px;
}

.info-compact strong {
  color: #333;
  margin-right: 4px;
}

/* Director Documents - Compact */
.director-documents-compact {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px solid #e8e8e8;
}

.documents-header-compact {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 500;
  color: #666;
}

.documents-label-compact {
  color: #666;
}

.documents-loading-compact {
  padding: 8px;
  text-align: center;
}

.documents-list-compact {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.document-item-compact {
  display: flex;
  align-items: center;
  padding: 6px 10px;
  background: #fff;
  border: 1px solid #e8e8e8;
  border-radius: 4px;
  transition: all 0.2s ease;
  font-size: 13px;
}

.document-item-compact:hover {
  background: #f5f5f5;
  border-color: #d9d9d9;
}

.document-name-compact {
  flex: 1;
  color: #1a1a1a;
  margin-right: 8px;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.document-actions {
  margin-left: auto;
  display: flex;
  gap: 4px;
  align-items: center;
}

.document-action-btn {
  color: #1890ff;
  cursor: pointer;
  padding: 4px 6px;
  display: flex;
  align-items: center;
  transition: all 0.2s ease;
  border-radius: 4px;
}

.document-action-btn:hover {
  background: #e6f7ff;
  color: #40a9ff;
}

.document-action-btn.delete-btn {
  color: #ff4d4f;
}

.document-action-btn.delete-btn:hover {
  background: #fff1f0;
  color: #ff7875;
}

.documents-empty-compact {
  padding: 8px;
  text-align: center;
  color: #999;
  font-size: 12px;
}

.text-muted {
  color: #999;
}

/* Responsive */
@media (max-width: 768px) {
  .info-grid {
    grid-template-columns: 1fr;
  }
  
  .shareholder-header,
  .director-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .info-row {
    flex-direction: column;
    gap: 4px;
  }
  
  /* Compact directors responsive */
  .director-main-info {
    flex-direction: column;
    gap: 8px;
  }
  
  .director-name-compact {
    flex: 1;
    min-width: auto;
  }
  
  .director-info-compact {
    width: 100%;
  }
  
  .document-item-compact {
    flex-wrap: wrap;
  }
  
  .document-name-compact {
    min-width: 150px;
  }
  
  .document-actions {
    margin-left: 0;
    margin-top: 4px;
  }
}

.detail-sections {
  display: flex;
  flex-direction: column;
  gap: 32px;
  margin-top: 24px;
}

.detail-section {
  width: 100%;
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
  .trend-cards-row {
    grid-template-columns: 1fr;
  }

  .recent-section {
    flex-direction: column;
  }
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

/* Assign Role Modal Styles */
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

.change-description {
  max-width: 100%;
}

.change-header {
  margin-bottom: 8px;
  color: #1a1a1a;
}

.change-list {
  margin: 0;
  padding-left: 20px;
  list-style-type: disc;
}

.change-list li {
  margin-bottom: 4px;
  line-height: 1.6;
  color: #666;
}

/* Striped Table Styles */
:deep(.striped-table) {
  .ant-table-tbody > tr:nth-child(even) {
    background-color: #fafafa;
  }
  
  .ant-table-tbody > tr:hover {
    background-color: #e6f7ff !important;
    cursor: pointer;
  }
  
  .ant-table-tbody > tr:nth-child(even):hover {
    background-color: #e6f7ff !important;
  }
}

/* Expanded Row Styles for Director Documents */
:deep(.ant-table-expanded-row) {
  background-color: #f9f9f9;
  
  .ant-table-expanded-row-level-1 {
    background-color: #fff;
  }
}

.document-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
