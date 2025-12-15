<template>
  <div class="subsidiary-form-layout">
    <DashboardHeader @logout="handleLogout" />

    <div class="subsidiary-form-wrapper">
      <!-- Page Header Section -->
      <div class="page-header-container">
        <div class="page-header">
          <h1 class="page-title">{{ isEditMode ? 'Edit Perusahaan' : 'Tambah Perusahaan Baru' }}</h1>
            <p class="page-description">
              {{ isEditMode 
                ? 'Perbarui informasi perusahaan, struktur kepemilikan, bidang usaha, dan data pengurus perusahaan.'
                : 'Tambah perusahaan baru ke dalam sistem. Lengkapi informasi identitas, struktur kepemilikan, bidang usaha, dan data pengurus.' 
              }}
            </p>
        </div>
      </div>

      <!-- Loading Overlay -->
      <a-spin :spinning="loading" tip="Menyimpan data perusahaan, harap tunggu..." style="min-height: 400px;">
        <div class="form-content">
        <a-card class="form-card">
        <!-- Progress Steps -->
        <a-steps :current="currentStep" class="form-steps">
          <a-step title="Identitas Perusahaan" />
          <a-step title="Struktur Kepemilikan" />
          <a-step title="Bidang Usaha" />
          <a-step title="Pengurus/Dewan Direksi" />
        </a-steps>

        <!-- Step 1: Identitas Perusahaan -->
        <div v-if="currentStep === 0" class="step-content">
          <h2 class="step-title">Company Information</h2>

          <a-divider />

          <a-form :label-col="{ span: 24 }" :wrapper-col="{ span: 24 }">
          <!-- Informasi Dasar -->
          <div class="form-section">
            <h3 class="section-title">
              <IconifyIcon icon="mdi:information" width="20" style="margin-right: 8px;" />
              Informasi Dasar
            </h3>
            <a-row :gutter="[12, 0]">

              <a-col :xs="24" :md="6">
                <a-form-item label="Nama Lengkap" required>
                  <a-input v-model:value="formData.name" placeholder="Nama lengkap perusahaan" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="6">
                <a-form-item label="Nama Singkat">
                  <a-input v-model:value="formData.short_name" placeholder="Nama singkat perusahaan" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="6">
                <a-form-item label="NPWP">
                  <a-input v-model:value="formData.npwp" placeholder="Nomor Pokok Wajib Pajak" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="6">
                <a-form-item label="NIB">
                  <a-input v-model:value="formData.nib" placeholder="Nomor Induk Berusaha" />
                </a-form-item>
              </a-col>
              <a-col :xs="16">
                <a-form-item label="Deskripsi">
                  <a-textarea v-model:value="formData.description" :rows="1" placeholder="Deskripsi perusahaan" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="8">
                <a-form-item label="Status">
                  <a-select v-model:value="formData.status" placeholder="Pilih status">
                    <a-select-option value="Aktif">Aktif</a-select-option>
                    <a-select-option value="Tidak Aktif">Tidak Aktif</a-select-option>
                  </a-select>
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Logo">
                  <a-upload
                    :file-list="logoFileList"
                    :before-upload="handleLogoUpload"
                    @remove="handleLogoRemove"
                    accept="image/png,image/jpeg,image/jpg"
                    :max-count="1"
                    list-type="picture-card"
                  >
                    <div v-if="logoFileList.length < 1">
                      <IconifyIcon icon="mdi:plus" width="24" />
                      <div style="margin-top: 8px">Upload</div>
                    </div>
                  </a-upload>
                  <div v-if="logoFileList.length > 0" style="margin-top: 8px; color: #666; font-size: 12px">
                    Format: PNG, JPG, JPEG | Maks: 5MB
                  </div>
                </a-form-item>
              </a-col>
              
            </a-row>
          </div>

          <!-- Informasi Kontak -->
          <div class="form-section">
            <h3 class="section-title">
              <IconifyIcon icon="mdi:phone" width="20" style="margin-right: 8px;" />
              Informasi Kontak
            </h3>
            <a-row :gutter="[12, 0]">
              <a-col :xs="24" :md="6">
                <a-form-item label="Telp">
                  <a-input v-model:value="formData.phone" placeholder="Nomor telepon" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="6">
                <a-form-item label="Fax">
                  <a-input v-model:value="formData.fax" placeholder="Nomor fax" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="6">
                <a-form-item label="Email">
                  <a-input v-model:value="formData.email" type="email" placeholder="Email perusahaan" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="6">
                <a-form-item label="Website">
                  <a-input v-model:value="formData.website" placeholder="Website perusahaan" />
                </a-form-item>
              </a-col>
            </a-row>
          </div>

          <!-- Alamat Perusahaan -->
          <div class="form-section">
            <h3 class="section-title">
              <IconifyIcon icon="mdi:map-marker" width="20" style="margin-right: 8px;" />
              Alamat Perusahaan
            </h3>
            <a-row :gutter="[12, 0]">
              <a-col :xs="12">
                <a-form-item label="Alamat Perusahaan">
                  <a-textarea v-model:value="formData.address" :rows="3" placeholder="Alamat perusahaan" />
                </a-form-item>
              </a-col>
              <a-col :xs="12">
                <a-form-item label="Alamat Operasional">
                  <a-textarea v-model:value="formData.operational_address" :rows="3" placeholder="Alamat operasional" />
                </a-form-item>
              </a-col>
            </a-row>
          </div>
          </a-form>
        </div>

        <!-- Step 2: Struktur Kepemilikan -->
        <div v-if="currentStep === 1" class="step-content">
          <h2 class="step-title">
            <IconifyIcon icon="mdi:account-group" width="24" style="margin-right: 8px;" />
            Struktur Kepemilikan
          </h2>

          <a-divider />

          <a-form :label-col="{ span: 24 }" :wrapper-col="{ span: 24 }">
            <!-- Informasi Dasar -->
            <div class="form-section">
              <h3 class="section-title">
                <IconifyIcon icon="mdi:information" width="20" style="margin-right: 8px;" />
                Informasi Dasar
              </h3>
              <a-row :gutter="[12, 0]">
                <a-col v-if="!route.params.id" :xs="24" :md="12">
                  <a-form-item label="Kode Perusahaan" required>
                    <a-input v-model:value="formData.code" placeholder="Kode perusahaan (unik)" />
                  </a-form-item>
                </a-col>
                <a-col :xs="24" :md="12">
                  <a-form-item label="Perusahaan Induk">
                    <a-select
                      v-model:value="formData.parent_id"
                      placeholder="Pilih perusahaan induk (opsional - bisa di-setup nanti)"
                      allow-clear
                      :disabled="isCompanyCapitalGreaterThanShareholders"
                    >
                      <a-select-option
                        v-for="company in availableCompanies"
                        :key="company.id"
                        :value="company.id"
                        :disabled="route.params.id === company.id"
                      >
                        {{ company.name }} ({{ getLevelLabel(company.level) }})
                      </a-select-option>
                    </a-select>
                    <div v-if="!isCompanyCapitalGreaterThanShareholders" style="margin-top: 4px; color: #666; font-size: 12px">
                      Kosongkan untuk sekarang, bisa di-setup nanti secara manual
                    </div>
                    <a-alert
                      v-if="isCompanyCapitalGreaterThanShareholders"
                      type="info"
                      show-icon
                      style="margin-top: 8px; font-size: 12px"
                    >
                      <template #message>
                        <div style="line-height: 1.5;">
                          <strong>Tidak ada perusahaan induk</strong><br/>
                          Modal Disetor perusahaan ini ({{ (formData.paid_up_capital || 0).toLocaleString('id-ID') }}) lebih besar dari total Modal Disetor semua pemegang saham ({{ getTotalShareholderCapital().toLocaleString('id-ID') }}). 
                          Perusahaan ini dianggap independen.<br/>
                          <strong>Kepemilikan sendiri: {{ currentCompanyOwnershipPercent }}%</strong> dari total modal semua modal disetor dari perusahaan yang berkontribusi.
                        </div>
                      </template>
                    </a-alert>
                  </a-form-item>
                </a-col>
                <a-col :xs="24" :md="6">
                  <a-form-item label="Modal Dasar">
                    <a-input-number
                      v-model:value="formData.authorized_capital"
                      :min="0"
                      :precision="0"
                      style="width: 100%"
                      placeholder="Modal Dasar"
                      :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                      :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                    />
                  </a-form-item>
                </a-col>
                <a-col :xs="24" :md="6">
                  <a-form-item label="Modal Disetor">
                    <a-input-number
                      v-model:value="formData.paid_up_capital"
                      :min="0"
                      :precision="0"
                      style="width: 100%"
                      placeholder="Modal Disetor"
                      :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                      :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                    />
                    <div v-if="formData.paid_up_capital && formData.shareholders.some(sh => sh.shareholder_company_id)" style="margin-top: 4px; font-size: 12px; color: #666;">
                      <a-popover
                        :title="null"
                        trigger="hover"
                        placement="top"
                      >
                        <template #content>
                          <div style="font-size: 12px; line-height: 1.6;">
                            <div style="margin-bottom: 8px;">
                              <strong>Persentase Kepemilikan Sendiri:</strong>
                            </div>
                            <div style="margin-bottom: 4px;">
                              <strong>{{ currentCompanyOwnershipPercent }}%</strong>
                            </div>
                            <div style="margin-bottom: 4px;">
                              Modal Disetor perusahaan sendiri: {{ formData.currency || 'IDR' }} {{ (formData.paid_up_capital || 0).toLocaleString('id-ID') }}
                            </div>
                            <div style="margin-bottom: 4px;">
                              Total Modal Disetor semua pemegang saham: {{ formData.currency || 'IDR' }} {{ getTotalShareholderCapital().toLocaleString('id-ID') }}
                            </div>
                            <div style="margin-bottom: 4px;">
                              Total modal: {{ formData.currency || 'IDR' }} {{ ((formData.paid_up_capital || 0) + getTotalShareholderCapital()).toLocaleString('id-ID') }}
                            </div>
                            <div style="margin-top: 8px; padding-top: 8px; border-top: 1px solid #e8e8e8;">
                              <strong>Rumus perhitungan:</strong><br/>
                              Total modal = Modal perusahaan sendiri + Total modal semua pemegang saham<br/>
                              Persentase kepemilikan sendiri = (Modal perusahaan sendiri ÷ Total modal) × 100%
                            </div>
                          </div>
                        </template>
                        <template #default>
                          <span style="cursor: help; text-decoration: underline; color: #1890ff;">
                            Kepemilikan sendiri: {{ currentCompanyOwnershipPercent }}%
                          </span>
                        </template>
                      </a-popover>
                    </div>
                  </a-form-item>
                </a-col>
                <a-col :xs="24" :md="12">
                  <a-form-item label="Mata Uang" required>
                    <a-radio-group v-model:value="formData.currency" button-style="solid">
                      <a-radio-button value="IDR">
                        <IconifyIcon icon="mdi:cash" width="16" style="margin-right: 4px;" />
                        Rupiah (IDR)
                      </a-radio-button>
                      <a-radio-button value="USD">
                        <IconifyIcon icon="mdi:currency-usd" width="16" style="margin-right: 4px;" />
                        Dollar (USD)
                      </a-radio-button>
                    </a-radio-group>
                    <div style="margin-top: 4px; font-size: 12px; color: #8c8c8c;">
                      Mata uang yang digunakan untuk laporan keuangan perusahaan ini. Semua nilai di laporan akan ditampilkan sesuai mata uang yang dipilih.
                    </div>
                  </a-form-item>
                </a-col>
              </a-row>
          </div>

          <!-- Pemegang Saham -->
          <div class="form-section">
            <h3 class="section-title">
              <IconifyIcon icon="mdi:information" width="20" style="margin-right: 8px;" />
              Pemegang Saham
            </h3>
            <a-table
              :columns="shareholderColumns"
              :data-source="formData.shareholders"
              :pagination="false"
              row-key="id"
            >
              <template #bodyCell="{ column, record, index }">
                <template v-if="column.key === 'type'">
                  <div class="shareholder-type-select-wrapper">
                    <a-select
                      v-model:value="record.type"
                      mode="multiple"
                      placeholder="Pilih atau ketik jenis pemegang saham baru"
                      show-search
                      allow-clear
                      :filter-option="false"
                      :loading="loadingShareholderTypes"
                      :search-value="shareholderTypeSearchValue"
                      :max-tag-count="1"
                      :max-tag-placeholder="(omittedValues: string[]) => `+${omittedValues.length} lainnya`"
                      style="width: 100%"
                      @search="handleShareholderTypeSearch"
                      @change="handleShareholderTypeChange(record, $event)"
                      @select="handleShareholderTypeSelect(record, $event)"
                    >
                      <a-select-option
                        v-for="shareholderType in filteredShareholderTypes"
                        :key="shareholderType.id"
                        :value="shareholderType.name"
                        :disabled="!shareholderType.is_active && !record.type.includes(shareholderType.name)"
                      >
                        <span :style="{ opacity: shareholderType.is_active ? 1 : 0.6 }">
                          {{ shareholderType.name }}
                          <a-tag v-if="!shareholderType.is_active" color="default" size="small" style="margin-left: 8px;">
                            Tidak Aktif
                          </a-tag>
                        </span>
                      </a-select-option>
                      <a-select-option
                        v-if="shareholderTypeSearchValue && !filteredShareholderTypes.find((st: ShareholderType) => st.name.toLowerCase() === shareholderTypeSearchValue.toLowerCase()) && canManageShareholderTypes"
                        :value="shareholderTypeSearchValue"
                        style="color: #1890ff;"
                      >
                        <IconifyIcon icon="mdi:plus-circle" style="margin-right: 4px;" />
                        Buat "{{ shareholderTypeSearchValue }}"
                      </a-select-option>
                    </a-select>
                    <!-- <div v-if="canManageShareholderTypes" class="shareholder-type-hint">
                      <IconifyIcon icon="mdi:information-outline" style="margin-right: 4px;" />
                      <span>Ketik nama baru dan tekan Enter untuk membuat jenis pemegang saham baru. Klik icon hapus di dropdown untuk menghapus jenis pemegang saham.</span>
                    </div> -->
                  </div>
                </template>
                <template v-if="column.key === 'name'">
                  <!-- Always show dropdown first, if individual mode then show input below -->
                  <a-select
                    :value="record.shareholder_company_id || '__individual__'"
                    style="width: 100%"
                    placeholder="Pilih perusahaan atau individu/eksternal"
                    show-search
                    :filter-option="(input: string, option: { label?: string; children?: Array<{ children?: unknown }> }) => {
                      const label = option?.label || (option?.children?.[0]?.children ? String(option.children[0].children) : '')
                      return label.toLowerCase().includes(input.toLowerCase())
                    }"
                    @change="(value: string | null) => handleShareholderCompanyChange(record, value)"
                  >
                    <a-select-option value="__individual__" style="color: #1890ff; font-weight: 500;">
                      <IconifyIcon icon="mdi:account-plus" width="16" style="margin-right: 8px;" />
                      Individu/Eksternal (Input Manual)
                    </a-select-option>
                    <a-select-option
                      v-for="company in getAvailableShareholderCompanies()"
                      :key="company.id"
                      :value="company.id"
                    >
                      {{ company.name }} {{ company.code ? `(${company.code})` : '' }}
                    </a-select-option>
                  </a-select>
                  <!-- Show input text if individual/external mode -->
                  <a-input 
                    v-if="!record.isCompany"
                    v-model:value="record.name" 
                    placeholder="Nama pemegang saham (Individu/Eksternal)"
                    style="margin-top: 8px;"
                  />
                </template>
                <template v-if="column.key === 'identity_number'">
                  <a-input 
                    v-model:value="record.identity_number" 
                    :placeholder="record.isCompany ? 'Nomor Identitas (Otomatis dari NPWP/NIB)' : 'Nomor Identitas (KTP/NPWP)'"
                    :disabled="record.isCompany && (record.shareholder_company_id ? true : false)"
                  />
                  <div v-if="record.isCompany && !record.identity_number" style="margin-top: 4px; font-size: 11px; color: #8c8c8c;">
                    NPWP/NIB perusahaan belum tersedia. Isi manual jika diperlukan.
                  </div>
                </template>
                <template v-if="column.key === 'ownership_percent'">
                  <a-popover
                    v-if="record.isCompany"
                    :title="null"
                    trigger="hover"
                    placement="top"
                  >
                    <template #content>
                      <div style="font-size: 12px; line-height: 1.6;">
                        <div style="margin-bottom: 8px;">
                          <strong>Persentase ini dihitung otomatis berdasarkan Modal Disetor:</strong>
                        </div>
                        <div style="margin-bottom: 4px;">
                          Modal Disetor perusahaan pemegang saham: {{ getShareholderCapitalInfo(record) }}
                        </div>
                        <div style="margin-bottom: 4px;">
                          Modal Disetor perusahaan sendiri: {{ formData.currency || 'IDR' }} {{ (formData.paid_up_capital || 0).toLocaleString('id-ID') }}
                        </div>
                        <div style="margin-bottom: 4px;">
                          Total Modal Disetor semua pemegang saham: {{ formData.currency || 'IDR' }} {{ getTotalShareholderCapital().toLocaleString('id-ID') }}
                        </div>
                        <div style="margin-bottom: 4px;">
                          Total modal: {{ formData.currency || 'IDR' }} {{ ((formData.paid_up_capital || 0) + getTotalShareholderCapital()).toLocaleString('id-ID') }}
                        </div>
                        <div style="margin-top: 8px; padding-top: 8px; border-top: 1px solid #e8e8e8;">
                          <strong>Rumus perhitungan:</strong><br/>
                          Total modal = Modal perusahaan sendiri + Total modal semua pemegang saham<br/>
                          Persentase = (Modal Disetor perusahaan pemegang saham ÷ Total modal) × 100%
                        </div>
                      </div>
                    </template>
                    <template #default>
                      <a-input-number
                        v-model:value="record.ownership_percent"
                        :min="0"
                        :max="100"
                        :precision="10"
                        :disabled="record.isCompany"
                        style="width: 100%"
                        placeholder="%"
                      />
                    </template>
                  </a-popover>
                  <a-input-number
                    v-else
                    v-model:value="record.ownership_percent"
                    :min="0"
                    :max="100"
                    :precision="10"
                    :disabled="record.isCompany"
                    style="width: 100%"
                    placeholder="%"
                  />
                  <div v-if="record.isCompany && record.ownership_percent === 0" style="margin-top: 4px; font-size: 11px; color: #ff4d4f;">
                    <IconifyIcon icon="mdi:alert-circle" width="12" />
                    Perusahaan belum memiliki Modal Disetor
                  </div>
                </template>
                <template v-if="column.key === 'share_sheet_count'">
                  <a-input-number
                    v-model:value="record.share_sheet_count"
                    :min="0"
                    :precision="0"
                    style="width: 100%"
                    placeholder="Jumlah lembar saham"
                    :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                    :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                  />
                </template>
                <template v-if="column.key === 'share_value_per_sheet'">
                  <a-input-number
                    v-model:value="record.share_value_per_sheet"
                    :min="0"
                    :precision="0"
                    style="width: 100%"
                    placeholder="Nilai Rupiah per lembar"
                    :formatter="(value: number | undefined) => value ? `${value}`.replace(/\B(?=(\d{3})+(?!\d))/g, ',') : ''"
                    :parser="(value: string) => value.replace(/\$\s?|(,*)/g, '')"
                  />
                </template>
                <template v-if="column.key === 'actions'">
                  <a-button type="link" danger size="small" @click="removeShareholder(index)">
                    <IconifyIcon icon="mdi:delete" width="16" />
                  </a-button>
                </template>
              </template>
            </a-table>
            <a-button type="dashed" style="width: 100%; margin-top: 16px;" @click="addShareholder">
              <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
              Tambah Pemegang Saham
            </a-button>
          </div>
          </a-form>
        </div>

        <!-- Step 3: Bidang Usaha -->
        <div v-if="currentStep === 2" class="step-content">
          <h2 class="step-title">
            <IconifyIcon icon="mdi:briefcase" width="24" style="margin-right: 8px;" />
            Bidang Usaha
          </h2>

          <a-divider />

          <a-form :label-col="{ span: 24 }" :wrapper-col="{ span: 24 }">
          <!-- Utama -->
          <div class="form-section">
            <h3 class="section-title">Utama</h3>
            <a-row :gutter="[12, 0]">
              <a-col :xs="24" :md="12">
                <a-form-item label="Sektor Industri">
                  <a-input v-model:value="formData.main_business.industry_sector" placeholder="Sektor industri" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="KBLI">
                  <a-input v-model:value="formData.main_business.kbli" placeholder="Klasifikasi Baku Lapangan Usaha Indonesia" />
                </a-form-item>
              </a-col>
              <a-col :xs="24">
                <a-form-item label="Uraian Kegiatan Usaha Utama">
                  <a-textarea v-model:value="formData.main_business.main_business_activity" :rows="4" placeholder="Uraian kegiatan usaha utama" />
                </a-form-item>
              </a-col>
            </a-row>
          </div>

          <!-- Lain-lain -->
          <div class="form-section">
            <h3 class="section-title">Lain-lain</h3>
            <a-row :gutter="[12, 0]">
              <a-col :xs="24">
                <a-form-item label="Kegiatan Usaha Tambahan">
                  <a-textarea v-model:value="formData.main_business.additional_activities" :rows="3" placeholder="Kegiatan usaha tambahan" />
                </a-form-item>
              </a-col>
              <a-col :xs="24" :md="12">
                <a-form-item label="Tanggal Mulai Beroperasi">
                  <a-date-picker
                    v-model:value="formData.main_business.start_operation_date"
                    style="width: 100%"
                    format="DD MMMM YYYY"
                    placeholder="Pilih tanggal"
                  />
                </a-form-item>
              </a-col>
            </a-row>
          </div>
          </a-form>
        </div>

        <!-- Step 4: Pengurus/Dewan Direksi -->
        <div v-if="currentStep === 3" class="step-content">
          <h2 class="step-title">
            <IconifyIcon icon="mdi:account-tie" width="24" style="margin-right: 8px;" />
            Pengurus/Dewan Direksi
          </h2>

          <a-divider />

          <a-form :label-col="{ span: 24 }" :wrapper-col="{ span: 24 }">
          <div class="form-section">
            <h3 class="section-title">
              <IconifyIcon icon="mdi:information" width="20" style="margin-right: 8px;" />
              Data Individu
            </h3>
            <a-table
              :columns="directorColumns"
              :data-source="formData.directors"
              :pagination="false"
              row-key="id"
            >
              <template #bodyCell="{ column, record, index }">
                <template v-if="column.key === 'position'">
                  <div class="director-position-select-wrapper">
                    <a-select
                      v-model:value="record.position"
                      mode="multiple"
                      placeholder="Pilih atau ketik jabatan baru"
                      show-search
                      allow-clear
                      :filter-option="false"
                      :loading="loadingDirectorPositions"
                      :search-value="directorPositionSearchValue"
                      :max-tag-count="1"
                      :max-tag-placeholder="(omittedValues: string[]) => `+${omittedValues.length} lainnya`"
                      style="width: 100%"
                      @search="handleDirectorPositionSearch"
                      @change="handleDirectorPositionChange(record, $event)"
                      @select="handleDirectorPositionSelect(record, $event)"
                    >
                      <a-select-option
                        v-for="directorPosition in filteredDirectorPositions"
                        :key="directorPosition.id"
                        :value="directorPosition.name"
                        :disabled="!directorPosition.is_active && !record.position.includes(directorPosition.name)"
                      >
                        <span :style="{ opacity: directorPosition.is_active ? 1 : 0.6 }">
                          {{ directorPosition.name }}
                          <a-tag v-if="!directorPosition.is_active" color="default" size="small" style="margin-left: 8px;">
                            Tidak Aktif
                          </a-tag>
                        </span>
                      </a-select-option>
                      <a-select-option
                        v-if="directorPositionSearchValue && !filteredDirectorPositions.find((dp: DirectorPosition) => dp.name.toLowerCase() === directorPositionSearchValue.toLowerCase()) && canManageDirectorPositions"
                        :value="directorPositionSearchValue"
                        style="color: #1890ff;"
                      >
                        <IconifyIcon icon="mdi:plus-circle" style="margin-right: 4px;" />
                        Buat "{{ directorPositionSearchValue }}"
                      </a-select-option>
                    </a-select>
                  </div>
                </template>
                <template v-if="column.key === 'full_name'">
                  <a-input v-model:value="record.full_name" placeholder="Nama lengkap" />
                </template>
                <template v-if="column.key === 'ktp'">
                  <a-input v-model:value="record.ktp" placeholder="Nomor KTP" />
                </template>
                <template v-if="column.key === 'npwp'">
                  <a-input v-model:value="record.npwp" placeholder="Nomor NPWP" />
                </template>
                <template v-if="column.key === 'start_date'">
                  <a-date-picker
                    v-model:value="record.start_date"
                    style="width: 100%"
                    format="DD MMMM YYYY"
                    placeholder="Tanggal awal jabatan"
                  />
                </template>
                <template v-if="column.key === 'domicile_address'">
                  <a-input v-model:value="record.domicile_address" placeholder="Alamat domisili" />
                </template>
                <template v-if="column.key === 'actions'">
                  <a-space>
                    <a-badge :count="getDirectorDocumentCount(record, index)" :number-style="{ backgroundColor: '#52c41a' }">
                      <a-button type="link" size="small" @click="handleAttachFiles(index)" title="Attach Files">
                        <IconifyIcon icon="mdi:attachment" width="16" />
                      </a-button>
                    </a-badge>
                    <a-button type="link" danger size="small" @click="removeDirector(index)">
                      <IconifyIcon icon="mdi:delete" width="16" />
                    </a-button>
                  </a-space>
                </template>
              </template>
            </a-table>
            <a-button type="dashed" style="width: 100%; margin-top: 16px;" @click="addDirector">
              <IconifyIcon icon="mdi:plus" width="16" style="margin-right: 8px;" />
              Tambah +
            </a-button>
          </div>
          </a-form>
        </div>

        <!-- Modal Upload Files untuk Director -->
        <a-modal
          v-model:open="attachFilesModalVisible"
          title="Upload Dokumen Individu"
          :confirm-loading="uploadingFiles"
          :width="700"
          :footer="null"
          @cancel="handleCloseAttachModal"
        >
          <div v-if="selectedDirectorIndex !== null">
            <div style="margin-bottom: 16px;">
              <p><strong>Individu:</strong> {{ formData.directors[selectedDirectorIndex]?.full_name || '-' }}</p>
              <p><strong>Jabatan:</strong> {{ formData.directors[selectedDirectorIndex]?.position?.join(', ') || '-' }}</p>
            </div>
            
            <a-form-item label="Kategori Dokumen" required>
              <a-select
                v-model:value="attachFilesForm.documentCategory"
                placeholder="Pilih kategori dokumen"
                :options="documentCategories"
              />
            </a-form-item>

            <a-form-item label="Upload File" required>
              <a-upload
                v-model:file-list="attachFilesForm.fileList"
                :before-upload="handleBeforeUpload"
                :custom-request="handleCustomUpload"
                multiple
                :accept="'.docx,.xlsx,.xls,.pptx,.ppt,.pdf,.jpg,.jpeg,.png'"
              >
                <a-button>
                  <IconifyIcon icon="mdi:upload" width="16" style="margin-right: 8px;" />
                  Pilih File
                </a-button>
                <template #tip>
                  <div style="color: #666; font-size: 12px; margin-top: 8px;">
                    Format yang diizinkan: DOCX, Excel (XLSX/XLS), PowerPoint (PPTX/PPT), PDF, dan gambar (JPG/JPEG/PNG)
                  </div>
                </template>
              </a-upload>
            </a-form-item>

            <!-- Tabel Dokumen yang Sudah Di-upload -->
            <a-divider style="margin: 16px 0;" />
            <div style="margin-bottom: 16px;">
              <h4 style="margin-bottom: 12px;">
                Dokumen yang Sudah Di-upload ({{ currentDirectorDocuments.length }})
                <span v-if="getPendingFilesCount(selectedDirectorIndex) > 0" style="color: #ff9800; font-size: 13px; font-weight: normal; margin-left: 8px;">
                  (+ {{ getPendingFilesCount(selectedDirectorIndex) }} file menunggu upload)
                </span>
              </h4>
              <a-table
                :columns="documentTableColumns"
                :data-source="currentDirectorDocuments"
                :pagination="{ pageSize: 5 }"
                size="small"
                :loading="loadingDirectorDocuments"
                row-key="id"
              >
                <template #bodyCell="{ column, record }">
                  <template v-if="column.key === 'category'">
                    <span>
                      {{ getDocumentCategoryLabel(record) }}
                    </span>
                  </template>
                  <template v-else-if="column.key === 'size'">
                    <span>{{ formatFileSize(record.size) }}</span>
                  </template>
                  <template v-else-if="column.key === 'created_at'">
                    <span>{{ record.created_at ? dayjs(record.created_at).format('DD/MM/YYYY HH:mm') : '-' }}</span>
                  </template>
                  <template v-else-if="column.key === 'action'">
                    <a :href="getDocumentDownloadUrl(record.file_path)" target="_blank" style="color: #1890ff;">
                      Download
                    </a>
                  </template>
                </template>
                <template #emptyText>
                  <a-empty description="Belum ada dokumen yang di-upload" style="margin: 24px 0;" />
                </template>
              </a-table>
              
              <!-- Pending Files List -->
              <div v-if="selectedDirectorIndex !== null && getPendingFilesCount(selectedDirectorIndex) > 0" style="margin-top: 12px;">
                <div style="margin-bottom: 8px; font-weight: 600; color: #ff9800;">File Menunggu Upload:</div>
                <div v-for="(file, fileIdx) in getPendingFilesForModal(selectedDirectorIndex)" :key="`pending-${selectedDirectorIndex}-${fileIdx}`" 
                  style="display: flex; align-items: center; padding: 8px 12px; background: #fff7e6; border: 1px solid #ffd591; border-radius: 4px; margin-bottom: 6px;">
                  <IconifyIcon icon="mdi:clock-outline" width="16" style="margin-right: 8px; color: #ff9800;" />
                  <span style="flex: 1; font-size: 14px;">{{ file.name }}</span>
                  <a-tag size="small" color="orange">Menunggu Upload</a-tag>
                </div>
              </div>
            </div>

            <!-- Footer dengan layout yang lebih rapi -->
            <div style="margin-top: 24px; padding-top: 16px; border-top: 1px solid #f0f0f0; display: flex; justify-content: space-between; align-items: center;">
              <div style="flex: 1;">
                <small style="color: #666; font-size: 12px;">
                  File akan diupload saat form disubmit (klik Finish/Update)
                </small>
              </div>
              <a-space>
                <a-button @click="handleCloseAttachModal">Batal</a-button>
                <a-button type="primary" @click="handleFinishAttachFiles" :disabled="attachFilesForm.fileList.length === 0 || !attachFilesForm.documentCategory">
                  Simpan ke Daftar ({{ attachFilesForm.fileList.length }})
                </a-button>
              </a-space>
            </div>
          </div>
        </a-modal>

        <!-- Navigation Buttons -->
        <div class="form-actions">
          <a-space>
            <a-button @click="handleCancel">Cancel</a-button>
            <a-button v-if="currentStep > 0" @click="prevStep">
              <IconifyIcon icon="mdi:arrow-left" width="16" style="margin-right: 4px;" />
              Previous
            </a-button>
            <a-button v-if="currentStep < 3" type="primary" @click="nextStep">
              Next
              <IconifyIcon icon="mdi:arrow-right" width="16" style="margin-left: 4px;" />
            </a-button>
            <a-button
              v-if="currentStep === 3" 
              type="primary"
              class="finish-button"
              @click="handleSubmit"
              :loading="loading"
              :disabled="loading"
            >
              <template v-if="!loading">
                Finish
                <IconifyIcon icon="mdi:arrow-right" width="16" style="margin-left: 4px;" />
              </template>
              <template v-else>
                Menyimpan...
              </template>
            </a-button>
          </a-space>
        </div>
      </a-card>
      </div>
    </a-spin>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, nextTick, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { message } from 'ant-design-vue'
import DashboardHeader from '../components/DashboardHeader.vue'
import { companyApi, uploadApi, shareholderTypesApi, directorPositionsApi, type Company, type Shareholder, type ShareholderType, type DirectorPosition, type BusinessField, type Director } from '../api/userManagement'
import { useAuthStore } from '../stores/auth'
import { Icon as IconifyIcon } from '@iconify/vue'
import apiClient from '../api/client'
import documentsApi, { type DocumentFolder, type DocumentItem } from '../api/documents'
import dayjs from 'dayjs'
import 'dayjs/locale/id'
import type { UploadFile } from 'ant-design-vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const currentStep = ref(0)
const loading = ref(false)
const availableCompanies = ref<Company[]>([])
const logoFileList = ref<Array<{ uid: string; name: string; status?: string; url?: string }>>([])
const uploadingLogo = ref(false)
const hasRootHolding = ref(false)

// Shareholder types (master data)
const shareholderTypes = ref<ShareholderType[]>([])
const loadingShareholderTypes = ref(false)
const shareholderTypeSearchValue = ref('')

// Director positions (master data)
const directorPositions = ref<DirectorPosition[]>([])
const loadingDirectorPositions = ref(false)
const directorPositionSearchValue = ref('')

// Attach files modal state
const attachFilesModalVisible = ref(false)
const selectedDirectorIndex = ref<number | null>(null)
const uploadingFiles = ref(false)
const attachFilesForm = ref<{
  fileList: UploadFile[]
  documentCategory: string | undefined
}>({
  fileList: [],
  documentCategory: undefined,
})

// Store temporary files that will be uploaded on form submit
// Format: { tempDirectorIndex: number, files: File[], category: string }[]
const pendingDirectorFiles = ref<Array<{
  tempDirectorIndex: number
  directorId?: string // Director ID jika sudah ada (untuk existing directors)
  fullName?: string // Full name untuk matching jika ID belum ada
  ktp?: string // KTP untuk matching jika ID belum ada
  files: File[]
  category: string
}>>([])

// Store uploaded documents with temporary director reference (for backward compatibility with updateDirectorDocumentRelations)
// Format: { tempDirectorIndex: number, documentIds: string[], category: string }[]
const pendingDirectorDocuments = ref<Array<{
  tempDirectorIndex: number
  documentIds: string[]
  category: string
}>>([])

// Store fetched documents per director (directorId -> DocumentItem[])
const directorDocumentsMap = ref<Map<string, DocumentItem[]>>(new Map())

// Loading state for fetching director documents
const loadingDirectorDocuments = ref(false)

// Current director documents displayed in modal
const currentDirectorDocuments = ref<DocumentItem[]>([])

// Document categories for individual documents
const documentCategories = [
  { label: 'KTP', value: 'ktp' },
  { label: 'NPWP', value: 'npwp' },
  { label: 'Sertifikat', value: 'certificate' },
  { label: 'Ijazah', value: 'diploma' },
  { label: 'SK Pengangkatan', value: 'appointment_letter' },
  { label: 'Lainnya', value: 'other' },
]

// Check if user can manage shareholder types (superadmin/administrator only)
const canManageShareholderTypes = computed(() => {
  return authStore.user?.role?.toLowerCase() === 'superadmin' || authStore.user?.role?.toLowerCase() === 'administrator'
})

// Check if user can manage director positions (superadmin/administrator only)
const canManageDirectorPositions = computed(() => {
  return authStore.user?.role?.toLowerCase() === 'superadmin' || authStore.user?.role?.toLowerCase() === 'administrator'
})

// Check if edit mode
const isEditMode = computed(() => {
  return !!route.params.id
})

const formData = ref({
  // Step 1: Identitas Perusahaan
  name: '',
  short_name: '',
  description: '',
  npwp: '',
  nib: '',
  status: 'Aktif',
  logo: '',
  phone: '',
  fax: '',
  email: '',
  website: '',
  address: '',
  operational_address: '',
  code: '',
  parent_id: undefined as string | undefined,
  authorized_capital: undefined as number | undefined, // Modal Dasar
  paid_up_capital: undefined as number | undefined, // Modal Disetor
  currency: 'IDR' as string, // Mata uang: IDR (default) atau USD
  
  // Step 2: Struktur Kepemilikan
  shareholders: [] as Array<{
    id?: string
    shareholder_company_id?: string | null // ID perusahaan pemegang saham (nullable)
    type: string[] // Changed to array for multiple selection (tag system)
    name: string
    identity_number: string
    ownership_percent: number // 10 digit desimal (calculated automatically)
    share_sheet_count?: number
    share_value_per_sheet?: number
    is_main_parent: boolean
    isCompany?: boolean // Flag untuk membedakan perusahaan dari sistem vs individu/eksternal
  }>,
  
  // Step 3: Bidang Usaha
  main_business: {
    industry_sector: '',
    kbli: '',
    main_business_activity: '',
    additional_activities: '',
    start_operation_date: null as dayjs.Dayjs | null,
  },
  
  // Step 4: Pengurus/Dewan Direksi
  directors: [] as Array<{
    id?: string
    position: string[] // Changed to array for multiple selection (tag system)
    full_name: string
    ktp: string
    npwp: string
    start_date: dayjs.Dayjs | null
    domicile_address: string
  }>,
})

const shareholderColumns = [
  { title: 'Jenis Pemegang Saham', key: 'type', width: 250 },
  { title: 'Nama Pemegang Saham', key: 'name', width: 200 },
  { title: 'Nomor Identitas', key: 'identity_number', width: 180 },
  { title: 'Persentase Kepemilikan', key: 'ownership_percent', width: 150 },
  { title: 'Jumlah Lembar Saham', key: 'share_sheet_count', width: 180 },
  { title: 'Nilai Rupiah per Lembar', key: 'share_value_per_sheet', width: 200 },
  { title: 'Aksi', key: 'actions', width: 80 },
]

const directorColumns = [
  { title: 'Jabatan', key: 'position', width: 150 },
  { title: 'Nama Lengkap', key: 'full_name', width: 200 },
  { title: 'Nomor Identitas (KTP)', key: 'ktp', width: 150 },
  { title: 'NPWP', key: 'npwp', width: 150 },
  { title: 'Tanggal Awal Jabatan', key: 'start_date', width: 150 },
  { title: 'Alamat Domisili', key: 'domicile_address', width: 200 },
  { title: 'Aksi', key: 'actions', width: 80 },
]

const addShareholder = () => {
  formData.value.shareholders.push({
    type: [],
    shareholder_company_id: null,
    name: '',
    identity_number: '',
    ownership_percent: 0,
    share_sheet_count: undefined,
    share_value_per_sheet: undefined,
    is_main_parent: false,
    isCompany: false, // Default: individu/eksternal
  })
  // Recalculate ownership percentages after adding
  calculateOwnershipPercentages()
}

const removeShareholder = (index: number) => {
  formData.value.shareholders.splice(index, 1)
  // Recalculate ownership percentages after removing
  calculateOwnershipPercentages()
}

// Calculate ownership percentages automatically based on paid_up_capital
// Only for shareholders that are companies (have shareholder_company_id)
// If current company's paid_up_capital > total shareholder capital, include it in calculation
const calculateOwnershipPercentages = () => {
  // Get all company shareholders (those with shareholder_company_id)
  const companyShareholders = formData.value.shareholders.filter(sh => sh.shareholder_company_id)
  
  // Calculate total paid_up_capital from all company shareholders
  let totalShareholderCapital = 0
  const shareholderCapitals: Map<number, number> = new Map()
  
  companyShareholders.forEach((sh, index) => {
    const company = availableCompanies.value.find(c => c.id === sh.shareholder_company_id)
    const capital = company?.paid_up_capital || 0
    shareholderCapitals.set(index, capital)
    totalShareholderCapital += capital
  })
  
  // Get current company's paid_up_capital
  const currentCompanyCapital = formData.value.paid_up_capital || 0
  
  // If current company's capital is greater than total shareholder capital, include it in total
  const includeCurrentCompanyInTotal = currentCompanyCapital > totalShareholderCapital && totalShareholderCapital > 0
  const totalCapital = includeCurrentCompanyInTotal 
    ? currentCompanyCapital + totalShareholderCapital 
    : totalShareholderCapital
  
  // Calculate percentage for each company shareholder (10 decimal places)
  if (totalCapital > 0) {
    companyShareholders.forEach((sh, index) => {
      const capital = shareholderCapitals.get(index) || 0
      const percentage = (capital / totalCapital) * 100
      // Round to 10 decimal places
      sh.ownership_percent = Math.round(percentage * 10000000000) / 10000000000
    })
  } else {
    // If total capital is 0, set all to 0
    companyShareholders.forEach(sh => {
      sh.ownership_percent = 0
    })
  }
  
  // Update parent company based on highest percentage
  updateParentCompanyBasedOnPercent()
}

// Handle change when user selects company or switches to individual/external
const handleShareholderCompanyChange = (record: typeof formData.value.shareholders[number], value: string | null) => {
  if (value === null || value === '') {
    // Switch to individual/external mode
    record.shareholder_company_id = null
    record.isCompany = false
    record.name = ''
    record.identity_number = ''
    record.ownership_percent = 0 // For individual/external, percentage is manual
    
    // Recalculate ownership percentages and update parent company
    calculateOwnershipPercentages()
  } else if (value === '__individual__') {
    // User explicitly selected "Individu/Eksternal" option
    record.shareholder_company_id = null
    record.isCompany = false
    record.name = ''
    record.identity_number = ''
    record.ownership_percent = 0
    
    // Recalculate ownership percentages and update parent company
    calculateOwnershipPercentages()
  } else {
    // User selected a company from the list
    const selectedCompany = availableCompanies.value.find(c => c.id === value)
    if (selectedCompany) {
      record.shareholder_company_id = value
      record.isCompany = true
      record.name = selectedCompany.name
      // Auto-fill identity_number from NPWP or NIB
      record.identity_number = selectedCompany.npwp || selectedCompany.nib || ''
      
      // Recalculate ownership percentages and update parent company
      calculateOwnershipPercentages()
      
      // Show alert if company doesn't have paid_up_capital
      if (!selectedCompany.paid_up_capital || selectedCompany.paid_up_capital === 0) {
        message.warning(`Perusahaan "${selectedCompany.name}" belum memiliki data Modal Disetor. Persentase kepemilikan akan dihitung sebagai 0%.`)
      }
    }
  }
}

// Get filtered companies for shareholder dropdown (exclude current company being edited)
const getAvailableShareholderCompanies = () => {
  const currentCompanyId = route.params.id as string
  return availableCompanies.value.filter(c => c.id !== currentCompanyId)
}

// Get shareholder capital info for popover
const getShareholderCapitalInfo = (record: typeof formData.value.shareholders[number]) => {
  if (!record.shareholder_company_id) {
    return '-'
  }
  
  const shareholderCompany = availableCompanies.value.find(c => c.id === record.shareholder_company_id)
  if (!shareholderCompany) {
    return '-'
  }
  
  const capital = shareholderCompany.paid_up_capital || 0
  // Use currency from shareholder company if available, otherwise use form currency
  const currency = shareholderCompany.currency || formData.value.currency || 'IDR'
  const formattedCapital = capital.toLocaleString('id-ID')
  
  return `${currency} ${formattedCapital}`
}

// Calculate total capital from all company shareholders
const getTotalShareholderCapital = () => {
  const companyShareholders = formData.value.shareholders.filter(sh => sh.shareholder_company_id)
  let totalCapital = 0
  
  companyShareholders.forEach((sh) => {
    const company = availableCompanies.value.find(c => c.id === sh.shareholder_company_id)
    if (company) {
      const capital = company.paid_up_capital || 0
      totalCapital += capital
    }
  })
  
  return totalCapital
}

// Check if current company's paid_up_capital is greater than total shareholder capital
const isCompanyCapitalGreaterThanShareholders = computed(() => {
  const currentCompanyCapital = formData.value.paid_up_capital || 0
  const totalShareholderCapital = getTotalShareholderCapital()
  return currentCompanyCapital > totalShareholderCapital && totalShareholderCapital > 0
})

// Calculate ownership percentage of current company (kepemilikan sendiri)
// Always calculated: (Modal perusahaan sendiri / Total modal) × 100%
// Total modal = Modal perusahaan sendiri + Total modal semua pemegang saham
const currentCompanyOwnershipPercent = computed(() => {
  const currentCompanyCapital = formData.value.paid_up_capital || 0
  const totalShareholderCapital = getTotalShareholderCapital()
  const totalCapital = currentCompanyCapital + totalShareholderCapital
  
  if (totalCapital === 0) {
    return 0
  }
  
  const percentage = (currentCompanyCapital / totalCapital) * 100
  // Round to 2 decimal places for display
  return Math.round(percentage * 100) / 100
})

// Auto-update parent company based on highest ownership percentage
// This will always update to the company with highest percentage (only for company shareholders, not individuals)
// BUT: If current company's paid_up_capital > total shareholder capital, set parent_id to null
const updateParentCompanyBasedOnPercent = () => {
  if (formData.value.shareholders.length === 0) {
    // If no shareholders, don't change parent_id (keep existing or null)
    return
  }
  
  // Check if current company's capital is greater than total shareholder capital
  const currentCompanyCapital = formData.value.paid_up_capital || 0
  const totalShareholderCapital = getTotalShareholderCapital()
  
  // If current company's capital is greater than total shareholder capital, set parent to undefined
  if (currentCompanyCapital > totalShareholderCapital && totalShareholderCapital > 0) {
    if (formData.value.parent_id !== undefined) {
      formData.value.parent_id = undefined
      console.log('Auto-cleared parent company: Current company capital is greater than total shareholder capital', {
        currentCompanyCapital,
        totalShareholderCapital
      })
    }
    return
  }
  
  // Only consider company shareholders (those with shareholder_company_id)
  const companyShareholders = formData.value.shareholders.filter(sh => sh.shareholder_company_id && sh.ownership_percent > 0)
  
  if (companyShareholders.length === 0) {
    // If no company shareholders with percentage > 0, set to undefined
    if (formData.value.parent_id !== undefined) {
      formData.value.parent_id = undefined
    }
    return
  }
  
  // Find shareholder with highest percentage (only from company shareholders)
  const maxPercentShareholder = companyShareholders.reduce((max, sh) => {
    return sh.ownership_percent > (max?.ownership_percent || 0) ? sh : max
  }, companyShareholders[0])
  
  // Auto-update to the company with highest percentage
  if (maxPercentShareholder && maxPercentShareholder.shareholder_company_id) {
    const newParentId = maxPercentShareholder.shareholder_company_id
    const previousParentId = formData.value.parent_id
    if (previousParentId !== newParentId) {
      formData.value.parent_id = newParentId
      console.log('Auto-updated parent company:', {
        previousParentId: previousParentId,
        newParentId: newParentId,
        shareholderName: maxPercentShareholder.name,
        percentage: maxPercentShareholder.ownership_percent
      })
    }
  }
}

const addDirector = () => {
  formData.value.directors.push({
    position: [], // Changed to array for multiple selection
    full_name: '',
    ktp: '',
    npwp: '',
    start_date: null,
    domicile_address: '',
  })
}

const removeDirector = (index: number) => {
  formData.value.directors.splice(index, 1)
  
  // Remove pending files untuk director yang dihapus
  pendingDirectorFiles.value = pendingDirectorFiles.value.filter(
    (pf) => pf.tempDirectorIndex !== index
  )
  
  // Update index untuk pending files yang index-nya lebih besar (karena array bergeser)
  pendingDirectorFiles.value = pendingDirectorFiles.value.map((pf) => {
    if (pf.tempDirectorIndex > index) {
      return {
        ...pf,
        tempDirectorIndex: pf.tempDirectorIndex - 1,
      }
    }
    return pf
  })
  
  // Remove pending documents untuk director yang dihapus (legacy support)
  pendingDirectorDocuments.value = pendingDirectorDocuments.value.filter(
    (pd) => pd.tempDirectorIndex !== index
  )
  
  // Update index untuk pending documents yang index-nya lebih besar
  pendingDirectorDocuments.value = pendingDirectorDocuments.value.map((pd) => {
    if (pd.tempDirectorIndex > index) {
      return {
        ...pd,
        tempDirectorIndex: pd.tempDirectorIndex - 1,
      }
    }
    return pd
  })
}

// Attach files handlers
// Load all director documents for form (pre-load saat edit mode)
const loadAllDirectorDocumentsForForm = async (directors: Array<{ id?: string }>) => {
  const loadPromises = directors
    .filter(d => d.id)
    .map(async (director) => {
      if (!director.id) return
      try {
        const response = await documentsApi.listDocumentsPaginated({
          director_id: director.id,
          page: 1,
          page_size: 100,
        })
        directorDocumentsMap.value.set(director.id, response.data)
      } catch (error) {
        console.error(`Failed to load documents for director ${director.id}:`, error)
        directorDocumentsMap.value.set(director.id, [])
      }
    })
  
  await Promise.all(loadPromises)
}

const handleAttachFiles = async (index: number) => {
  selectedDirectorIndex.value = index
  attachFilesForm.value.fileList = []
  attachFilesForm.value.documentCategory = undefined
  attachFilesModalVisible.value = true
  
  // Fetch existing documents for this director
  const director = formData.value.directors[index]
  console.log('Opening attach modal for director:', { index, director: { id: director?.id, name: director?.full_name } })
  
  if (director?.id) {
    loadingDirectorDocuments.value = true
    try {
      // Always fetch fresh data to ensure we have the latest documents
      console.log(`Fetching documents for director ${director.id}`)
      const response = await documentsApi.listDocumentsPaginated({
        director_id: director.id,
        page: 1,
        page_size: 100, // Fetch all documents for this director
      })
      console.log(`Fetched ${response.data.length} documents for director ${director.id}`, response.data)
      currentDirectorDocuments.value = response.data
      directorDocumentsMap.value.set(director.id, response.data)
    } catch (error) {
      console.error('Failed to fetch director documents:', error)
      // Fallback to cached data if fetch fails
      const cachedDocs = directorDocumentsMap.value.get(director.id) || []
      console.log(`Using cached documents for director ${director.id}: ${cachedDocs.length} docs`)
      currentDirectorDocuments.value = cachedDocs
    } finally {
      loadingDirectorDocuments.value = false
    }
  } else {
    console.log('Director does not have ID yet, cannot fetch documents')
    // For new directors without ID yet, documents will be fetched after director ID is assigned
    currentDirectorDocuments.value = []
  }
}

const handleCloseAttachModal = () => {
  attachFilesModalVisible.value = false
  selectedDirectorIndex.value = null
  attachFilesForm.value.fileList = []
  attachFilesForm.value.documentCategory = undefined
  currentDirectorDocuments.value = []
}

// Validasi file sebelum upload
const handleBeforeUpload = (file: File): boolean => {
  // Validasi tipe file
  const allowedExtensions = ['.docx', '.xlsx', '.xls', '.pptx', '.ppt', '.pdf', '.jpg', '.jpeg', '.png']
  const fileName = file.name.toLowerCase()
  const ext = fileName.substring(fileName.lastIndexOf('.'))
  
  if (!allowedExtensions.includes(ext)) {
    message.error(`Format file ${file.name} tidak diizinkan. Hanya DOCX, Excel (XLSX/XLS), PowerPoint (PPTX/PPT), PDF, dan gambar (JPG/JPEG/PNG) yang diperbolehkan.`)
    return false
  }
  
  // Validasi ukuran (max 50MB untuk dokumen)
  const maxSize = 50 * 1024 * 1024 // 50MB
  if (file.size > maxSize) {
    message.error(`Ukuran file ${file.name} melebihi 50MB`)
    return false
  }
  
  // Return false untuk prevent auto upload, kita akan upload manual
  return false
}

// Custom upload handler (tidak langsung upload, hanya add ke fileList)
const handleCustomUpload = () => {
  // Tidak melakukan apapun di sini, file sudah ditambahkan ke fileList
  // Upload akan dilakukan saat user klik button "Upload"
}

// Helper function untuk menghitung jumlah dokumen per direktur
const getDirectorDocumentCount = (director: typeof formData.value.directors[0], index: number): number => {
  // Jika direktur sudah punya ID, ambil dari map (existing documents)
  if (director?.id) {
    const docs = directorDocumentsMap.value.get(director.id)
    if (docs) {
      // Tambahkan pending files count
      const pendingFiles = pendingDirectorFiles.value.find(pf => pf.tempDirectorIndex === index)
      return docs.length + (pendingFiles ? pendingFiles.files.length : 0)
    }
  }
  
  // Jika belum punya ID, cek pending files berdasarkan index
  const pendingFiles = pendingDirectorFiles.value.find(pf => pf.tempDirectorIndex === index)
  return pendingFiles ? pendingFiles.files.length : 0
}

// Helper functions untuk tabel dokumen
const documentTableColumns = [
  { title: 'Nama File', dataIndex: 'name', key: 'name' },
  { title: 'Kategori', key: 'category' },
  { title: 'Ukuran', key: 'size' },
  { title: 'Tanggal Upload', key: 'created_at' },
  { title: 'Aksi', key: 'action' },
]

const getDocumentCategoryLabel = (record: DocumentItem): string => {
  const meta = record.metadata as { category?: string } | undefined
  const categoryValue = meta?.category || '-'
  const categoryLabel = documentCategories.find(cat => cat.value === categoryValue)?.label || categoryValue
  return categoryLabel
}

const formatFileSize = (size: number): string => {
  const sizeInMB = (size / (1024 * 1024)).toFixed(2)
  return `${sizeInMB} MB`
}

const getDocumentDownloadUrl = (filePath: string): string => {
  if (filePath.startsWith('http://') || filePath.startsWith('https://')) {
    return filePath
  }
  const apiURL = import.meta.env.VITE_API_URL?.replace('/api/v1', '') || 'http://localhost:8080'
  return `${apiURL}${filePath}`
}

// Helper untuk mendapatkan jumlah pending files untuk director index
const getPendingFilesCount = (directorIndex: number | null): number => {
  if (directorIndex === null) return 0
  const pendingFiles = pendingDirectorFiles.value.find(pf => pf.tempDirectorIndex === directorIndex)
  return pendingFiles ? pendingFiles.files.length : 0
}

// Helper untuk mendapatkan pending files di modal
const getPendingFilesForModal = (directorIndex: number | null): File[] => {
  if (directorIndex === null) return []
  const pendingFiles = pendingDirectorFiles.value.find(pf => pf.tempDirectorIndex === directorIndex)
  return pendingFiles ? pendingFiles.files : []
}

// Simpan files ke temporary state (tidak langsung upload)
const handleFinishAttachFiles = () => {
  if (selectedDirectorIndex.value === null) {
    message.error('Individu tidak ditemukan')
    return
  }
  
  if (attachFilesForm.value.fileList.length === 0) {
    message.error('Pilih minimal satu file untuk diupload')
    return
  }
  
  if (!attachFilesForm.value.documentCategory) {
    message.error('Pilih kategori dokumen')
    return
  }
  
  // Extract files from UploadFile objects
  const files: File[] = []
  for (const uploadFile of attachFilesForm.value.fileList) {
    if (uploadFile.originFileObj) {
      files.push(uploadFile.originFileObj)
    }
  }
  
  if (files.length === 0) {
    message.error('Tidak ada file yang valid untuk disimpan')
    return
  }
  
  // Cek apakah sudah ada pending files untuk director index ini
  const existingPendingIndex = pendingDirectorFiles.value.findIndex(
    (pf) => pf.tempDirectorIndex === selectedDirectorIndex.value
  )
  
  if (existingPendingIndex >= 0) {
    // Merge dengan existing files (append)
    const existingPending = pendingDirectorFiles.value[existingPendingIndex]
    const director = formData.value.directors[selectedDirectorIndex.value]
    if (existingPending) {
      existingPending.files.push(...files)
      existingPending.category = attachFilesForm.value.documentCategory
      // Update identifier jika belum ada atau berubah
      if (!existingPending.directorId && director?.id) {
        existingPending.directorId = director.id
      }
      if (!existingPending.fullName && director?.full_name) {
        existingPending.fullName = director.full_name
      }
      if (!existingPending.ktp && director?.ktp) {
        existingPending.ktp = director.ktp
      }
      message.success(`Berhasil menambahkan ${files.length} file ke daftar (total: ${existingPending.files.length} file). File akan diupload saat form disubmit.`)
    }
  } else {
    // Tambah baru
    const director = formData.value.directors[selectedDirectorIndex.value]
    pendingDirectorFiles.value.push({
      tempDirectorIndex: selectedDirectorIndex.value,
      directorId: director?.id, // Simpan director ID jika sudah ada
      fullName: director?.full_name, // Simpan full name untuk matching
      ktp: director?.ktp, // Simpan KTP untuk matching
      files: files,
      category: attachFilesForm.value.documentCategory,
    })
    message.success(`Berhasil menambahkan ${files.length} file ke daftar. File akan diupload saat form disubmit.`)
  }
  
  // Close modal
  handleCloseAttachModal()
}

const handleLogoUpload = async (file: File): Promise<boolean> => {
  // Validasi ukuran (max 5MB)
  const maxSize = 5 * 1024 * 1024
  if (file.size > maxSize) {
    message.error('Ukuran file melebihi 5MB')
    return false
  }

  // Validasi format
  const allowedTypes = ['image/png', 'image/jpeg', 'image/jpg']
  if (!allowedTypes.includes(file.type)) {
    message.error('Format file tidak diizinkan. Hanya PNG, JPG, dan JPEG yang diperbolehkan')
    return false
  }

  uploadingLogo.value = true
  try {
    const response = await uploadApi.uploadLogo(file)
    formData.value.logo = response.url
    // Cek apakah URL sudah full URL (dari GCP Storage) atau relative (local storage)
    let logoUrl: string
    if (response.url.startsWith('http://') || response.url.startsWith('https://')) {
      // Full URL dari GCP Storage, langsung pakai
      logoUrl = response.url
    } else {
      // Relative URL dari local storage, tambahkan baseURL
      const apiURL = import.meta.env.VITE_API_URL || (import.meta.env.DEV ? 'http://localhost:8080' : 'https://api-pedeve-dev.aretaamany.com')
      const baseURL = apiURL.replace(/\/api\/v1$/, '') // Hapus /api/v1 jika ada
      logoUrl = `${baseURL}${response.url}`
    }
    logoFileList.value = [{
      uid: '-1',
      name: file.name,
      status: 'done',
      url: logoUrl,
    }]
    message.success('Logo berhasil diupload')
    return false // Prevent default upload
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error(axiosError.response?.data?.message || 'Gagal upload logo')
    return false
  } finally {
    uploadingLogo.value = false
  }
}

const handleLogoRemove = (): void => {
  formData.value.logo = ''
  logoFileList.value = []
}

const nextStep = () => {
  if (currentStep.value < 3) {
    currentStep.value++
  }
}

const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

const handleCancel = () => {
  router.push('/subsidiaries')
}

const handleSubmit = async () => {
  // Prevent multiple submissions - check di awal dan set loading immediately
  if (loading.value) {
    console.warn('Submit already in progress, ignoring duplicate call')
    return
  }

  // Tidak ada validasi wajib untuk parent_id - bisa kosong dan di-setup nanti

  // Recalculate ownership percentages before submit to ensure all calculations are up-to-date
  calculateOwnershipPercentages()

  // Set loading IMMEDIATELY sebelum async operations untuk prevent race condition
  loading.value = true
  
  // Disable button immediately (redundant check tapi lebih aman)
  const submitButton = document.querySelector('.finish-button') as HTMLButtonElement
  if (submitButton) {
    submitButton.disabled = true
  }
  
  try {
    // Prepare data untuk API - menggunakan snake_case sesuai JSON tag
    // Log currency sebelum submit untuk debugging
    console.log('Submitting company data with currency:', {
      currency: formData.value.currency,
      formDataCurrency: formData.value.currency,
      isEditMode: !!route.params.id
    })

    const submitData = {
      name: formData.value.name,
      short_name: formData.value.short_name,
      description: formData.value.description,
      code: formData.value.code || `COMP-${Date.now()}`,
      npwp: formData.value.npwp,
      nib: formData.value.nib,
      status: formData.value.status,
      logo: formData.value.logo,
      phone: formData.value.phone,
      fax: formData.value.fax,
      email: formData.value.email,
      website: formData.value.website,
      address: formData.value.address,
      operational_address: formData.value.operational_address,
      // Set parent_id: always send, even if null (to allow removing parent)
      // CRITICAL: Always send parent_id field, even if null, so backend knows to update it
      parent_id: formData.value.parent_id ? formData.value.parent_id : null,
      authorized_capital: formData.value.authorized_capital !== undefined ? formData.value.authorized_capital : null,
      paid_up_capital: formData.value.paid_up_capital !== undefined ? formData.value.paid_up_capital : null,
      currency: formData.value.currency || 'IDR', // Default IDR jika tidak ada
      shareholders: formData.value.shareholders.map(sh => ({
        shareholder_company_id: sh.shareholder_company_id || null, // Send company_id if exists, null if individual/external
        type: Array.isArray(sh.type) ? sh.type.join(', ') : (sh.type || ''), // Convert array to comma-separated string for backend
        name: sh.name,
        identity_number: sh.identity_number || '',
        ownership_percent: sh.ownership_percent || 0, // For companies, this is calculated automatically; for individuals, it's manual
        share_sheet_count: sh.share_sheet_count !== undefined ? sh.share_sheet_count : null,
        share_value_per_sheet: sh.share_value_per_sheet !== undefined ? sh.share_value_per_sheet : null,
        is_main_parent: false,
      })),
      main_business: (formData.value.main_business.industry_sector || formData.value.main_business.kbli) ? {
        industry_sector: formData.value.main_business.industry_sector,
        kbli: formData.value.main_business.kbli,
        main_business_activity: formData.value.main_business.main_business_activity,
        additional_activities: formData.value.main_business.additional_activities,
        start_operation_date: formData.value.main_business.start_operation_date?.format('YYYY-MM-DD') || null,
      } : null,
      directors: formData.value.directors.map(d => ({
        position: Array.isArray(d.position) ? d.position.join(', ') : (d.position || ''), // Convert array to comma-separated string for backend
        full_name: d.full_name,
        ktp: d.ktp,
        npwp: d.npwp,
        start_date: d.start_date?.format('YYYY-MM-DD') || null,
        domicile_address: d.domicile_address,
      })),
    }

    let savedCompanyId: string
    
    if (route.params.id) {
      // Edit mode - use full update endpoint
      console.log('Updating company with data:', JSON.stringify(submitData, null, 2))
      const response = await apiClient.put(`/companies/${route.params.id}/full`, submitData)
      console.log('Update response:', response.data)
      message.success('Perusahaan berhasil diupdate')
      savedCompanyId = route.params.id as string
      
      // Reload company data setelah update untuk memastikan currency ter-update dan mendapatkan directors dengan ID
      const updatedCompany = await companyApi.getById(savedCompanyId)
      console.log('Reloaded company after update:', {
        currency: updatedCompany.currency,
        fullCompany: updatedCompany
      })
      formData.value.currency = updatedCompany.currency || 'IDR'
      
      // Upload pending files dan update director_id untuk pending documents
      if (updatedCompany.directors) {
        // Upload pending files (baru - files yang disimpan temporary)
        if (pendingDirectorFiles.value.length > 0) {
          await uploadPendingDirectorFiles(savedCompanyId, updatedCompany.directors)
        }
        // Update existing documents yang sudah diupload (legacy support)
        if (pendingDirectorDocuments.value.length > 0) {
          await updateDirectorDocumentRelations(savedCompanyId, updatedCompany.directors)
        }
      }
      
      // Success - redirect ke halaman detail subsidiaries setelah delay kecil
      await new Promise(resolve => setTimeout(resolve, 100))
      router.push(`/subsidiaries/${savedCompanyId}`)
    } else {
      // Create mode - use full create endpoint
      console.log('Creating company with data:', JSON.stringify(submitData, null, 2))
      const response = await apiClient.post('/companies/full', submitData)
      console.log('Create response:', response.data)
      message.success('Perusahaan berhasil dibuat')
      savedCompanyId = response.data.id
      
      // Upload pending files dan update director_id untuk pending documents
      // Reload company untuk mendapatkan directors dengan ID
      const newCompany = await companyApi.getById(savedCompanyId)
      if (newCompany.directors) {
        // Upload pending files (baru - files yang disimpan temporary)
        if (pendingDirectorFiles.value.length > 0) {
          await uploadPendingDirectorFiles(savedCompanyId, newCompany.directors)
        }
        // Update existing documents yang sudah diupload (legacy support)
        if (pendingDirectorDocuments.value.length > 0) {
          await updateDirectorDocumentRelations(savedCompanyId, newCompany.directors)
        }
      }
      
      // Success - redirect ke halaman list subsidiaries setelah delay kecil
      await new Promise(resolve => setTimeout(resolve, 100))
      router.push('/subsidiaries')
    }
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    message.error('Gagal menyimpan: ' + (axiosError.response?.data?.message || axiosError.message || 'Unknown error'))
  } finally {
    // Re-enable button dan reset loading state
    loading.value = false
    const submitButton = document.querySelector('.finish-button') as HTMLButtonElement
    if (submitButton) {
      submitButton.disabled = false
    }
  }
}

// Upload pending files untuk directors
const uploadPendingDirectorFiles = async (companyId: string, directors: Director[]) => {
  if (pendingDirectorFiles.value.length === 0) {
    return
  }
  
  try {
    // Cari folder perusahaan
    const folders = await documentsApi.listFolders()
    let companyFolder = folders.find((f) => {
      const folderWithCompany = f as DocumentFolder & { company_id?: string | null }
      return folderWithCompany.company_id === companyId
    })
    
    if (!companyFolder) {
      // Jika folder belum ada, buat folder baru
      const company = await companyApi.getById(companyId)
      companyFolder = await documentsApi.createFolder(company.name)
    }
    
    // Upload files untuk setiap pending file group
    for (const pendingFileGroup of pendingDirectorFiles.value) {
      // Cari director berdasarkan ID terlebih dahulu (lebih akurat)
      let director: Director | undefined
      
      if (pendingFileGroup.directorId) {
        // Cari berdasarkan ID jika ada
        director = directors.find(d => d.id === pendingFileGroup.directorId)
      }
      
      // Jika tidak ditemukan berdasarkan ID, cari berdasarkan identifier (full_name + ktp)
      if (!director && pendingFileGroup.fullName) {
        director = directors.find(d => {
          const nameMatch = d.full_name === pendingFileGroup.fullName
          const ktpMatch = !pendingFileGroup.ktp || d.ktp === pendingFileGroup.ktp
          return nameMatch && ktpMatch
        })
      }
      
      // Fallback: gunakan index jika ID/identifier matching gagal
      if (!director) {
        const directorIndex = pendingFileGroup.tempDirectorIndex
        if (directorIndex >= 0 && directorIndex < directors.length) {
          director = directors[directorIndex]
        }
      }
      
      if (!director || !director.id) {
        console.warn(`Director not found for pending file group:`, pendingFileGroup)
        continue
      }
      
      // Upload setiap file dengan director_id
      for (const file of pendingFileGroup.files) {
        try {
          await documentsApi.uploadDocument({
            file: file,
            folder_id: companyFolder.id,
            director_id: director.id,
            title: file.name,
            status: 'active',
            metadata: {
              category: pendingFileGroup.category,
              director_id: director.id,
              director_name: director.full_name,
            },
          })
        } catch (error: unknown) {
          const err = error as { message?: string }
          console.error(`Failed to upload file ${file.name}:`, err.message || 'Unknown error')
          // Continue dengan file lain meskipun ada error
        }
      }
    }
    
    // Clear pending files setelah berhasil upload
    pendingDirectorFiles.value = []
  } catch (error: unknown) {
    const err = error as { message?: string }
    console.error('Failed to upload pending director files:', err.message || 'Unknown error')
    message.warning('Beberapa file mungkin belum terupload. Silakan cek kembali di halaman dokumen.')
  }
}

// Update director_id untuk documents yang sudah diupload
const updateDirectorDocumentRelations = async (companyId: string, directors: Director[]) => {
  if (pendingDirectorDocuments.value.length === 0) {
    return
  }
  
  try {
    // Cari folder perusahaan
    const folders = await documentsApi.listFolders()
    let companyFolder = folders.find((f) => {
      const folderWithCompany = f as DocumentFolder & { company_id?: string | null }
      return folderWithCompany.company_id === companyId
    })
    
    if (!companyFolder) {
      // Jika folder belum ada, buat folder baru
      const company = await companyApi.getById(companyId)
      companyFolder = await documentsApi.createFolder(company.name)
    }
    
    // Update director_id untuk setiap pending document
    for (const pendingDoc of pendingDirectorDocuments.value) {
      // Cari director berdasarkan index (legacy support untuk pendingDirectorDocuments)
      // Note: pendingDirectorDocuments sudah deprecated, gunakan pendingDirectorFiles
      const directorIndex = pendingDoc.tempDirectorIndex
      let director: Director | undefined
      
      if (directorIndex >= 0 && directorIndex < directors.length) {
        director = directors[directorIndex]
      }
      
      if (!director || !director.id) {
        console.warn(`Director not found at index ${directorIndex}`, pendingDoc)
        continue
      }
      
      // Update setiap document dengan director_id
      for (const docId of pendingDoc.documentIds) {
        try {
          await documentsApi.updateDocument(docId, {
            director_id: director.id,
            metadata: {
              category: pendingDoc.category,
              director_id: director.id,
              director_name: director.full_name,
            },
          })
        } catch (error: unknown) {
          const err = error as { message?: string }
          console.error(`Failed to update document ${docId}:`, err.message || 'Unknown error')
          // Continue dengan document lain meskipun ada error
        }
      }
    }
    
    // Clear pending documents setelah berhasil update
    pendingDirectorDocuments.value = []
  } catch (error: unknown) {
    const err = error as { message?: string }
    console.error('Failed to update director document relations:', err.message || 'Unknown error')
    message.warning('Beberapa dokumen mungkin belum terhubung dengan individu. Silakan cek kembali di halaman dokumen.')
  }
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
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

const loadAvailableCompanies = async () => {
  try {
    availableCompanies.value = await companyApi.getAll()
    // Check if there's a root holding (parent_id = null, level = 0)
    hasRootHolding.value = availableCompanies.value.some(c => c.parent_id === null || c.parent_id === undefined)
  } catch (error) {
    console.error('Failed to load companies:', error)
  }
}

const loadCompanyData = async () => {
  if (route.params.id) {
    loading.value = true
    try {
      // Load available companies dulu untuk mencari main_parent_company
      if (availableCompanies.value.length === 0) {
        await loadAvailableCompanies()
      }
      const company: Company = await companyApi.getById(route.params.id as string)
      
      // Populate form data
      formData.value.name = company.name || ''
      formData.value.short_name = company.short_name || ''
      formData.value.description = company.description || ''
      formData.value.npwp = company.npwp || ''
      formData.value.nib = company.nib || ''
      formData.value.status = company.status || 'Aktif'
      formData.value.logo = company.logo || ''
      // Set logo file list jika ada logo
      if (company.logo) {
        let logoUrl: string
        if (company.logo.startsWith('http')) {
          logoUrl = company.logo
        } else {
          // Get base URL tanpa /api/v1 untuk static files
          const apiURL = import.meta.env.VITE_API_URL || (import.meta.env.DEV ? 'http://localhost:8080' : 'https://api-pedeve-dev.aretaamany.com')
          const baseURL = apiURL.replace(/\/api\/v1$/, '') // Hapus /api/v1 jika ada
          logoUrl = `${baseURL}${company.logo}`
        }
        logoFileList.value = [{
          uid: '-1',
          name: company.logo.split('/').pop() || 'logo',
          status: 'done',
          url: logoUrl,
        }]
      } else {
        logoFileList.value = []
      }
      formData.value.phone = company.phone || ''
      formData.value.fax = company.fax || ''
      formData.value.email = company.email || ''
      formData.value.website = company.website || ''
      formData.value.address = company.address || ''
      formData.value.operational_address = company.operational_address || ''
      formData.value.code = company.code || ''
      formData.value.parent_id = company.parent_id
      formData.value.authorized_capital = company.authorized_capital || undefined
      formData.value.paid_up_capital = company.paid_up_capital || undefined
      // Load currency from company, default to IDR if not available
      formData.value.currency = company.currency || 'IDR'
      console.log('Loaded company currency:', {
        companyId: company.id,
        currencyFromAPI: company.currency,
        currencySet: formData.value.currency,
        fullCompany: company
      })
      formData.value.shareholders = (company.shareholders || []).map((sh: Shareholder) => ({
        id: sh.id,
        shareholder_company_id: sh.shareholder_company_id || null,
        type: sh.type ? (typeof sh.type === 'string' ? sh.type.split(',').map(t => t.trim()).filter(t => t) : []) : [],
        name: sh.name,
        identity_number: sh.identity_number || '',
        ownership_percent: sh.ownership_percent || 0,
        share_sheet_count: sh.share_sheet_count || undefined,
        share_value_per_sheet: sh.share_value_per_sheet || undefined,
        is_main_parent: sh.is_main_parent ?? false,
        isCompany: !!sh.shareholder_company_id, // Set flag based on whether shareholder_company_id exists
      }))
      
      // Recalculate ownership percentages after loading (this will also update parent company)
      // Use nextTick to ensure Vue has finished updating the reactive data
      await nextTick()
      calculateOwnershipPercentages()
      // Transform business_fields array to main_business (ambil yang is_main = true atau yang pertama)
      if (company.business_fields && company.business_fields.length > 0) {
        const businessFieldsWithMain = company.business_fields as Array<BusinessField & { is_main?: boolean }>
        const mainBusiness = businessFieldsWithMain.find((bf) => bf.is_main) || company.business_fields[0]
        if (mainBusiness) {
          formData.value.main_business.industry_sector = mainBusiness.industry_sector || ''
          formData.value.main_business.kbli = mainBusiness.kbli || ''
          formData.value.main_business.main_business_activity = mainBusiness.main_business_activity || ''
          formData.value.main_business.additional_activities = mainBusiness.additional_activities || ''
          formData.value.main_business.start_operation_date = mainBusiness.start_operation_date ? dayjs(mainBusiness.start_operation_date) : null
        }
      } else if (company.main_business) {
        // Fallback untuk kompatibilitas jika ada main_business langsung
        formData.value.main_business.industry_sector = company.main_business.industry_sector || ''
        formData.value.main_business.kbli = company.main_business.kbli || ''
        formData.value.main_business.main_business_activity = company.main_business.main_business_activity || ''
        formData.value.main_business.additional_activities = company.main_business.additional_activities || ''
        formData.value.main_business.start_operation_date = company.main_business.start_operation_date ? dayjs(company.main_business.start_operation_date) : null
      }
      formData.value.directors = (company.directors || []).map((d: Director) => ({
        ...d,
        position: d.position ? (typeof d.position === 'string' ? d.position.split(',').map(t => t.trim()).filter(t => t) : []) : [], // Convert comma-separated string to array
        start_date: d.start_date ? dayjs(d.start_date) : null,
      }))
      
      // Pre-load documents for all directors after directors are loaded
      if (formData.value.directors && formData.value.directors.length > 0) {
        console.log('Pre-loading documents for directors in edit mode', formData.value.directors.map(d => ({ id: d.id, name: d.full_name })))
        await loadAllDirectorDocumentsForForm(formData.value.directors)
      }
    } catch {
      message.error('Gagal memuat data perusahaan')
    } finally {
      loading.value = false
    }
  }
}

// Shareholder type functions (similar to document types)
const loadShareholderTypes = async (includeInactive = false) => {
  loadingShareholderTypes.value = true
  try {
    shareholderTypes.value = await shareholderTypesApi.getShareholderTypes(includeInactive)
  } catch (error) {
    console.error('Failed to load shareholder types:', error)
    message.error('Gagal memuat jenis pemegang saham')
  } finally {
    loadingShareholderTypes.value = false
  }
}

const handleShareholderTypeSearch = (value: string) => {
  shareholderTypeSearchValue.value = value
}

const handleShareholderTypeChange = async (record: { type: string[] }, values: string[]) => {
  // Validate: check if any of the selected values are inactive
  const invalidValues: string[] = []
  const currentValues = record.type || []
  
  for (const value of values) {
    const shareholderType = shareholderTypes.value.find((st: ShareholderType) => st.name === value)
    if (shareholderType && !shareholderType.is_active) {
      const wasAlreadySelected = currentValues.includes(value)
      if (!wasAlreadySelected) {
        invalidValues.push(value)
      }
    }
  }
  
  // Remove invalid (inactive) values
  if (invalidValues.length > 0) {
    message.warning(`Jenis pemegang saham berikut tidak aktif dan tidak dapat dipilih: ${invalidValues.join(', ')}`)
    values = values.filter(v => !invalidValues.includes(v))
  }
  
  // Handle when new value is added (not in existing types)
  const newValues = values.filter((v: string) => {
    const exists = shareholderTypes.value.find((st: ShareholderType) => 
      st.name === v || st.name.toLowerCase() === v.toLowerCase()
    )
    return !exists && !currentValues.includes(v)
  })
  
  const failedValues: string[] = []
  const processedValues = [...values]
  
  for (const newValue of newValues) {
    if (canManageShareholderTypes.value && newValue.trim()) {
      try {
        await handleShareholderTypeCreate(newValue.trim())
        await loadShareholderTypes(false)
        const created = shareholderTypes.value.find((st: ShareholderType) => 
          st.name.toLowerCase() === newValue.trim().toLowerCase()
        )
        if (!created) {
          failedValues.push(newValue)
          const index = processedValues.indexOf(newValue)
          if (index !== -1) {
            processedValues.splice(index, 1)
          }
        } else {
          // Replace the new value with the created one (exact name from DB)
          const index = processedValues.indexOf(newValue)
          if (index !== -1) {
            processedValues[index] = created.name
          }
        }
      } catch (error) {
        console.error(`Failed to create shareholder type "${newValue}":`, error)
        failedValues.push(newValue)
        const index = processedValues.indexOf(newValue)
        if (index !== -1) {
          processedValues.splice(index, 1)
        }
      }
    } else {
      failedValues.push(newValue)
      const index = processedValues.indexOf(newValue)
      if (index !== -1) {
        processedValues.splice(index, 1)
      }
    }
  }
  
  // Normalize values: use exact names from database
  const normalizedValues = processedValues.map((v: string) => {
    const exists = shareholderTypes.value.find((st: ShareholderType) => 
      st.name === v || st.name.toLowerCase() === v.toLowerCase()
    )
    return exists ? exists.name : v
  })
  
  record.type = normalizedValues
  
  if (failedValues.length > 0) {
    message.warning(`Jenis pemegang saham berikut gagal dibuat atau tidak ditemukan: ${failedValues.join(', ')}`)
  }
}

const handleShareholderTypeSelect = async (record: { type: string[] }, value: string) => {
  // Check if this is a new value that needs to be created
  const existingType = shareholderTypes.value.find((st: ShareholderType) => 
    st.name === value || st.name.toLowerCase() === value.toLowerCase()
  )
  
  if (!existingType && canManageShareholderTypes.value && value.trim()) {
    // This is a new value, create it
    try {
      await handleShareholderTypeCreate(value.trim())
      await loadShareholderTypes(false)
      const created = shareholderTypes.value.find((st: ShareholderType) => 
        st.name.toLowerCase() === value.trim().toLowerCase()
      )
      if (created && !record.type.includes(created.name)) {
        record.type.push(created.name)
      }
    } catch (error) {
      console.error(`Failed to create shareholder type "${value}":`, error)
      // Remove the value if creation failed
      record.type = record.type.filter((t: string) => t !== value)
    }
  } else if (existingType) {
    // Existing type selected
    if (!existingType.is_active) {
      const isAlreadySelected = record.type.includes(existingType.name)
      if (!isAlreadySelected) {
        message.warning(`Jenis pemegang saham "${value}" tidak aktif dan tidak dapat dipilih.`)
        record.type = record.type.filter((t: string) => t !== value)
        return
      }
    }
    
    // Use the exact name from database (case-sensitive)
    if (!record.type.includes(existingType.name)) {
      record.type.push(existingType.name)
    }
    // Remove the value if it's different from the database name
    if (value !== existingType.name) {
      record.type = record.type.filter((t: string) => t !== value)
    }
  }
  
  shareholderTypeSearchValue.value = ''
}

const handleShareholderTypeCreate = async (value: string) => {
  if (!canManageShareholderTypes.value) {
    message.warning('Hanya superadmin dan administrator yang dapat membuat jenis pemegang saham baru')
    return
  }

  const trimmedValue = value.trim()
  if (!trimmedValue) {
    message.warning('Nama jenis pemegang saham tidak boleh kosong')
    return
  }

  const existing = shareholderTypes.value.find(
    (st: ShareholderType) => st.name.toLowerCase() === trimmedValue.toLowerCase()
  )
  if (existing) {
    message.warning(`Jenis pemegang saham "${trimmedValue}" sudah ada`)
    return
  }

  try {
    loadingShareholderTypes.value = true
    const newShareholderType = await shareholderTypesApi.createShareholderType(trimmedValue)
    
    if (!newShareholderType || !newShareholderType.id) {
      throw new Error('Invalid response from server: shareholder type not created in database')
    }
    
    await loadShareholderTypes(false)
    
    const verified = shareholderTypes.value.find((st: ShareholderType) => 
      st.id === newShareholderType.id || st.name.toLowerCase() === trimmedValue.toLowerCase()
    )
    if (!verified) {
      throw new Error('Shareholder type was not saved to database')
    }
    
    message.success(`Jenis pemegang saham "${trimmedValue}" berhasil dibuat dan disimpan ke database`)
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    const errorMessage = axiosError.response?.data?.message || axiosError.message || 'Gagal membuat jenis pemegang saham'
    console.error(`Failed to create shareholder type:`, error)
    message.error(errorMessage)
    throw error
  } finally {
    loadingShareholderTypes.value = false
    shareholderTypeSearchValue.value = ''
  }
}


// Filter shareholder types based on search and active status
const filteredShareholderTypes = computed(() => {
  const selectedShareholderTypeNames = formData.value.shareholders.flatMap(sh => sh.type || [])
  
  let filtered = shareholderTypes.value.filter((st: ShareholderType) => {
    if (st.is_active) return true
    return selectedShareholderTypeNames.includes(st.name)
  })
  
  if (shareholderTypeSearchValue.value) {
    const searchLower = shareholderTypeSearchValue.value.toLowerCase()
    filtered = filtered.filter((st: ShareholderType) => 
      st.name.toLowerCase().includes(searchLower)
    )
  }
  
  return filtered
})

// Director position functions (similar to shareholder types)
const loadDirectorPositions = async (includeInactive = false) => {
  loadingDirectorPositions.value = true
  try {
    directorPositions.value = await directorPositionsApi.getDirectorPositions(includeInactive)
  } catch (error) {
    console.error('Failed to load director positions:', error)
    message.error('Gagal memuat jabatan pengurus')
  } finally {
    loadingDirectorPositions.value = false
  }
}

const handleDirectorPositionSearch = (value: string) => {
  directorPositionSearchValue.value = value
}

const handleDirectorPositionChange = async (record: { position: string[] }, values: string[]) => {
  record.position = values || []
  
  // Check for new values that don't exist in master data (case-insensitive)
  const newValues = values.filter(v => {
    const trimmed = v.trim()
    if (!trimmed) return false
    return !directorPositions.value.find((dp: DirectorPosition) => 
      dp.name.toLowerCase() === trimmed.toLowerCase()
    )
  })
  
  if (newValues.length > 0 && canManageDirectorPositions.value) {
    // Create new director positions sequentially to avoid race conditions
    for (const newValue of newValues) {
      try {
        await handleDirectorPositionCreate(newValue)
        // Reload positions after creation to ensure we have the latest data
        await loadDirectorPositions(true) // Include inactive
      } catch (error) {
        console.error(`Failed to create director position "${newValue}":`, error)
        // Don't remove the value if it's a duplicate - backend will return existing
        // Only remove if it's a real error
        const errorMessage = (error as { response?: { data?: { message?: string } }; message?: string })?.response?.data?.message || ''
        if (!errorMessage.includes('sudah ada') && !errorMessage.includes('duplicate')) {
          record.position = record.position.filter((p: string) => p !== newValue)
        }
      }
    }
  }
}

const handleDirectorPositionSelect = async (record: { position: string[] }, value: string) => {
  // Check if this is a new value that needs to be created (case-insensitive)
  const trimmedValue = value.trim()
  const exists = directorPositions.value.find((dp: DirectorPosition) => 
    dp.name.toLowerCase() === trimmedValue.toLowerCase()
  )
  if (!exists && canManageDirectorPositions.value && trimmedValue) {
    try {
      await handleDirectorPositionCreate(trimmedValue)
      // Reload positions after creation to ensure we have the latest data
      await loadDirectorPositions(true) // Include inactive
    } catch (error) {
      console.error(`Failed to create director position "${value}":`, error)
      // Don't remove the value if it's a duplicate - backend will return existing
      // Only remove if it's a real error
      const errorMessage = (error as { response?: { data?: { message?: string } }; message?: string })?.response?.data?.message || ''
      if (!errorMessage.includes('sudah ada') && !errorMessage.includes('duplicate')) {
        record.position = record.position.filter((p: string) => p !== value)
      }
    }
  }
}

const handleDirectorPositionCreate = async (name: string) => {
  if (!canManageDirectorPositions.value) {
    message.warning('Hanya superadmin dan administrator yang dapat membuat jabatan pengurus baru')
    return
  }

  const trimmedValue = name.trim()
  if (!trimmedValue) {
    return
  }

  // Check if already exists (case-insensitive) before creating
  const existing = directorPositions.value.find(
    (dp: DirectorPosition) => dp.name.toLowerCase() === trimmedValue.toLowerCase()
  )
  if (existing) {
    // If exists but inactive, it will be reactivated by backend
    // If exists and active, backend will return error which we'll handle
    if (existing.is_active) {
      message.info(`Jabatan pengurus "${trimmedValue}" sudah ada`)
      return
    }
  }

  loadingDirectorPositions.value = true
  try {
    const newDirectorPosition = await directorPositionsApi.createDirectorPosition(trimmedValue)
    
    // Check if already in list (might have been added by another call or reactivated)
    const alreadyExists = directorPositions.value.find(
      (dp: DirectorPosition) => dp.id === newDirectorPosition.id || dp.name.toLowerCase() === trimmedValue.toLowerCase()
    )
    if (!alreadyExists) {
      directorPositions.value.push(newDirectorPosition)
    } else {
      // Update existing if it was reactivated
      const index = directorPositions.value.findIndex(
        (dp: DirectorPosition) => dp.id === newDirectorPosition.id || dp.name.toLowerCase() === trimmedValue.toLowerCase()
      )
      if (index !== -1) {
        directorPositions.value[index] = newDirectorPosition
      }
    }
    
    message.success(`Jabatan pengurus "${trimmedValue}" berhasil dibuat dan disimpan ke database`)
  } catch (error: unknown) {
    const axiosError = error as { response?: { data?: { message?: string } }; message?: string }
    const errorMessage = axiosError.response?.data?.message || axiosError.message || 'Gagal membuat jabatan pengurus'
    
    // If error is about duplicate, try to reload positions to get the existing one
    if (errorMessage.includes('sudah ada') || errorMessage.includes('duplicate')) {
      try {
        await loadDirectorPositions(true) // Include inactive
        const existingPosition = directorPositions.value.find(
          (dp: DirectorPosition) => dp.name.toLowerCase() === trimmedValue.toLowerCase()
        )
        if (existingPosition) {
          message.info(`Jabatan pengurus "${trimmedValue}" sudah ada`)
          return // Don't throw error, just return
        }
      } catch (loadError) {
        console.error('Failed to reload director positions:', loadError)
      }
    }
    
    console.error(`Failed to create director position:`, error)
    message.error(errorMessage)
    throw error
  } finally {
    loadingDirectorPositions.value = false
    directorPositionSearchValue.value = ''
  }
}

// Filter director positions based on search and active status
const filteredDirectorPositions = computed(() => {
  const selectedDirectorPositionNames = formData.value.directors.flatMap(d => d.position || [])
  
  let filtered = directorPositions.value.filter((dp: DirectorPosition) => {
    if (dp.is_active) return true
    return selectedDirectorPositionNames.includes(dp.name)
  })
  
  if (directorPositionSearchValue.value) {
    const searchLower = directorPositionSearchValue.value.toLowerCase()
    filtered = filtered.filter((dp: DirectorPosition) => 
      dp.name.toLowerCase().includes(searchLower)
    )
  }
  
  return filtered
})

// Watch for changes in shareholders to recalculate ownership percentages reactively
watch(
  () => formData.value.shareholders,
  () => {
    // Recalculate ownership percentages when shareholders change
    calculateOwnershipPercentages()
  },
  { deep: true }
)

// Watch for changes in availableCompanies to recalculate when company data updates
watch(
  () => availableCompanies.value,
  () => {
    // Recalculate ownership percentages when available companies data changes
    // This ensures that if a shareholder company's paid_up_capital changes, percentages update
    calculateOwnershipPercentages()
  },
  { deep: true }
)

// Watch for changes in paid_up_capital (Modal Disetor) of the current company being edited
// This ensures reactive updates when modal disetor changes
watch(
  () => formData.value.paid_up_capital,
  () => {
    // Recalculate ownership percentages when modal disetor changes
    // Note: This affects calculation if this company is a shareholder in other companies
    calculateOwnershipPercentages()
  }
)

onMounted(async () => {
  await Promise.all([
    loadAvailableCompanies(),
    loadShareholderTypes(false),
    loadDirectorPositions(false),
  ])
  await loadCompanyData()
  // After loading company, check if any shareholder type exists in active types
  const hasInactiveType = formData.value.shareholders.some(sh => 
    sh.type.some((t: string) => 
      !shareholderTypes.value.find((st: ShareholderType) => st.name === t)
    )
  )
  if (hasInactiveType) {
    await loadShareholderTypes(true) // Include inactive to show the company's types
  }
  // After loading company, check if any director position exists in active positions
  const hasInactivePosition = formData.value.directors.some(d => 
    d.position.some((p: string) => 
      !directorPositions.value.find((dp: DirectorPosition) => dp.name === p)
    )
  )
  if (hasInactivePosition) {
    await loadDirectorPositions(true) // Include inactive to show the company's positions
  }
})
</script>

<style scoped>
.subsidiary-form-layout {
  min-height: 100vh;
  background: #f5f5f5;
}

.subsidiary-form-wrapper {
  width: 100%;
}

.page-header-container {
  /* background: #fff;
  border-bottom: 1px solid #e8e8e8; */
  /* padding: 24px; */
  /* margin-bottom: 0; */
  max-width: 1200px;
}

.page-header {
  margin: 0 auto !important;
  display: flex !important;
  flex-direction: column !important;
  justify-content: space-between !important;
  align-items: flex-start !important;
}


/* .header-left {
  flex: 1;
}

.page-title {
  font-size: 24px;
  font-weight: 600;
  margin: 0 0 8px 0;
  color: #1a1a1a;
  line-height: 1.4;
}

.page-description {
  margin: 0;
  color: #666;
  font-size: 14px;
  line-height: 1.5;
} */

.form-content {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
}

.form-card {
  background: white;
  border-radius: 8px;
  /* box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1); */
}

.form-steps {
  margin-bottom: 32px;
}

.step-content {
  min-height: 400px;
}

.step-title {
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
}

.form-section {
  margin-bottom: 16px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  color: #035CAB;
}

.form-actions {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid #e8e8e8;
  display: flex;
  justify-content: flex-end;
}

/* Label di atas input */
.form-card :deep(.ant-form-item-label) {
  display: block;
  text-align: left;
  margin-bottom: 4px;
  padding: 0;
}

.form-card :deep(.ant-form-item-label > label) {
  height: auto;
  line-height: 1.5;
  font-weight: 500;
  color: rgba(0, 0, 0, 0.85);
}

.form-card :deep(.ant-form-item-label > label.ant-form-item-required:not(.ant-form-item-required-mark-optional)::before) {
  margin-right: 4px;
}

.form-card :deep(.ant-form-item-control) {
  flex: 1;
}

/* Ensure all inputs, selects, date pickers, and buttons have consistent height (40px) */
/* Text input dan input dengan affix wrapper */
.form-card :deep(input.ant-input:not([type="textarea"])),
.form-card :deep(.ant-input-affix-wrapper) {
  height: 40px !important;
}

.form-card :deep(.ant-input-affix-wrapper .ant-input) {
  height: 100% !important;
}

/* Select */
.form-card :deep(.ant-select-selector) {
  height: 40px !important;
}

.form-card :deep(.ant-select-selection-item),
.form-card :deep(.ant-select-selection-placeholder) {
  line-height: 38px !important;
}

/* Date Picker */
.form-card :deep(.ant-picker),
.form-card :deep(.ant-picker-input) {
  height: 40px !important;
}

.form-card :deep(.ant-picker-input > input) {
  height: 38px !important;
  line-height: 38px !important;
}

/* Textarea tetap fleksibel, tidak perlu height 40px */
.form-card :deep(textarea.ant-input) {
  height: auto !important;
  min-height: auto !important;
}

/* Buttons height */
.form-card :deep(.ant-btn) {
  height: 40px !important;
  min-height: 40px !important;
  display: flex !important;
  align-items: center !important;
  justify-content: center !important;
}

/* Small buttons tetap kecil */
.form-card :deep(.ant-btn-sm) {
  height: 24px !important;
  min-height: 24px !important;
}

/* Form item margin bottom untuk merapatkan */
.form-card :deep(.ant-form-item) {
  margin-bottom: 8px !important;
}

@media (max-width: 768px) {
  .form-content {
    padding: 16px;
  }
  
  .form-steps {
    margin-bottom: 24px;
  }
  
  .step-content {
    min-height: 300px;
  }
}

.shareholder-type-select-wrapper {
  width: 100%;
}

.shareholder-type-select-wrapper :deep(.ant-select-selector) {
  min-height: 40px;
  /* padding: 2px 4px; */
  /* background: orange !important; */
  height: auto !important;
}

.shareholder-type-select-wrapper :deep(.ant-select-selection-item) {
  /* margin: 2px 4px; */
  height: auto;
  line-height: 28px !important;
  padding: 0 8px;
  background: #f0f0f0;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
}

.shareholder-type-select-wrapper :deep(.ant-select-selection-item-content) {
  display: inline-block;
  margin-right: 4px;
}

.shareholder-type-select-wrapper :deep(.ant-select-selection-placeholder) {
  line-height: 36px;
}

.shareholder-type-hint {
  margin-top: 6px;
  font-size: 12px;
  color: #1890ff;
  line-height: 1.5;
  display: flex;
  align-items: flex-start;
}

.shareholder-type-hint span {
  flex: 1;
}

.director-position-select-wrapper {
  width: 100%;
}

.director-position-select-wrapper :deep(.ant-select-selector) {
  min-height: 40px;
  height: auto !important;
}

.director-position-select-wrapper :deep(.ant-select-selection-item) {
  height: auto;
  line-height: 28px !important;
  padding: 0 8px;
  background: #f0f0f0;
  border: 1px solid #d9d9d9;
}

.director-position-select-wrapper :deep(.ant-select-selection-item-content) {
  display: inline-block;
  margin-right: 4px;
}

.director-position-select-wrapper :deep(.ant-select-selection-placeholder) {
  line-height: 36px;
}
</style>
