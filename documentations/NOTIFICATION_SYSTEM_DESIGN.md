# 📧 Sistem Notifikasi - Document Expiration Alert

## 🎯 Overview
Sistem notifikasi untuk mengingatkan user tentang dokumen yang akan expired dalam waktu tertentu (dapat dikonfigurasi oleh admin).

## 📋 Fitur
1. **Email Notification** - Mengirim email ke user tentang dokumen yang akan expired
2. **In-App Notification** - Notifikasi real-time di dalam aplikasi
3. **Configurable Threshold** - Admin dapat mengatur batas minimal (hari/minggu/bulan)

## 🏗️ Arsitektur

### 1. Database Schema

#### Tabel: `notification_settings`
```sql
CREATE TABLE notification_settings (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    email_enabled BOOLEAN DEFAULT true,
    in_app_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

#### Tabel: `notifications`
```sql
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    type VARCHAR(50) NOT NULL, -- 'document_expiry', 'system', dll
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    resource_type VARCHAR(50), -- 'document', 'report', dll
    resource_id UUID,
    is_read BOOLEAN DEFAULT false,
    created_at TIMESTAMP,
    read_at TIMESTAMP
);
```

#### Tabel: `notification_configs`
```sql
CREATE TABLE notification_configs (
    id UUID PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL, -- 'document_expiry_threshold_days'
    config_value VARCHAR(255) NOT NULL, -- '30' (hari)
    description TEXT,
    updated_by UUID REFERENCES users(id),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

#### Modifikasi: `documents` table
```sql
ALTER TABLE documents ADD COLUMN expiry_date TIMESTAMP;
ALTER TABLE documents ADD COLUMN expiry_notified BOOLEAN DEFAULT false;
```

### 2. Backend Components

#### A. Email Service
**File**: `backend/internal/infrastructure/email/email.go`

**Opsi 1: SendGrid (Recommended)**
```go
package email

import (
    "github.com/sendgrid/sendgrid-go"
    "github.com/sendgrid/sendgrid-go/helpers/mail"
)

type EmailService struct {
    client *sendgrid.Client
    fromEmail string
}

func NewEmailService(apiKey, fromEmail string) *EmailService {
    return &EmailService{
        client: sendgrid.NewSendClient(apiKey),
        fromEmail: fromEmail,
    }
}

func (s *EmailService) SendDocumentExpiryNotification(toEmail, userName string, documentName string, daysUntilExpiry int) error {
    subject := fmt.Sprintf("⚠️ Dokumen '%s' Akan Expired dalam %d Hari", documentName, daysUntilExpiry)
    
    htmlContent := fmt.Sprintf(`
        <h2>Peringatan: Dokumen Akan Expired</h2>
        <p>Halo %s,</p>
        <p>Dokumen <strong>%s</strong> akan expired dalam <strong>%d hari</strong>.</p>
        <p>Silakan perbarui atau perpanjang dokumen tersebut.</p>
        <a href="https://pedeve-dev.aretaamany.com/documents">Lihat Dokumen</a>
    `, userName, documentName, daysUntilExpiry)
    
    from := mail.NewEmail("Pedeve DMS", s.fromEmail)
    to := mail.NewEmail(userName, toEmail)
    message := mail.NewSingleEmail(from, subject, to, "", htmlContent)
    
    _, err := s.client.Send(message)
    return err
}
```

**Opsi 2: SMTP (Generic)**
```go
package email

import (
    "net/smtp"
    "fmt"
)

type SMTPConfig struct {
    Host     string
    Port     int
    Username string
    Password string
    From     string
}

type EmailService struct {
    config SMTPConfig
}

func (s *EmailService) SendDocumentExpiryNotification(toEmail, userName string, documentName string, daysUntilExpiry int) error {
    auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
    
    subject := fmt.Sprintf("⚠️ Dokumen '%s' Akan Expired dalam %d Hari", documentName, daysUntilExpiry)
    body := fmt.Sprintf("Halo %s,\n\nDokumen %s akan expired dalam %d hari.\n\nSilakan perbarui dokumen tersebut.", userName, documentName, daysUntilExpiry)
    
    msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s", toEmail, subject, body))
    
    addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
    return smtp.SendMail(addr, auth, s.config.From, []string{toEmail}, msg)
}
```

#### B. Notification Service
**File**: `backend/internal/usecase/notification_usecase.go`

```go
package usecase

import (
    "time"
    "github.com/repoareta/pedeve-dms-app/backend/internal/domain"
    "github.com/repoareta/pedeve-dms-app/backend/internal/infrastructure/email"
)

type NotificationUseCase interface {
    CheckExpiringDocuments() error
    GetUserNotifications(userID string, unreadOnly bool) ([]domain.Notification, error)
    MarkAsRead(notificationID, userID string) error
    GetNotificationConfig(key string) (string, error)
    UpdateNotificationConfig(key, value, updatedBy string) error
}

type notificationUseCase struct {
    docRepo repository.DocumentRepository
    notifRepo repository.NotificationRepository
    emailService *email.EmailService
}

func (uc *notificationUseCase) CheckExpiringDocuments() error {
    // 1. Ambil threshold dari config
    thresholdDays, err := uc.GetNotificationConfig("document_expiry_threshold_days")
    if err != nil {
        thresholdDays = "30" // default 30 hari
    }
    
    days, _ := strconv.Atoi(thresholdDays)
    thresholdDate := time.Now().AddDate(0, 0, days)
    
    // 2. Query dokumen yang akan expired dalam threshold
    expiringDocs, err := uc.docRepo.GetExpiringDocuments(thresholdDate)
    if err != nil {
        return err
    }
    
    // 3. Untuk setiap dokumen, kirim notifikasi
    for _, doc := range expiringDocs {
        // Skip jika sudah pernah di-notify
        if doc.ExpiryNotified {
            continue
        }
        
        // Dapatkan user yang terkait (uploader atau assigned users)
        users, err := uc.getDocumentRelatedUsers(doc.ID)
        if err != nil {
            continue
        }
        
        for _, user := range users {
            daysUntilExpiry := int(doc.ExpiryDate.Sub(time.Now()).Hours() / 24)
            
            // Create in-app notification
            notification := &domain.Notification{
                ID: uuid.GenerateUUID(),
                UserID: user.ID,
                Type: "document_expiry",
                Title: fmt.Sprintf("Dokumen '%s' Akan Expired", doc.Name),
                Message: fmt.Sprintf("Dokumen '%s' akan expired dalam %d hari", doc.Name, daysUntilExpiry),
                ResourceType: "document",
                ResourceID: doc.ID,
                IsRead: false,
                CreatedAt: time.Now(),
            }
            uc.notifRepo.Create(notification)
            
            // Send email notification (jika enabled)
            if user.EmailNotificationEnabled {
                uc.emailService.SendDocumentExpiryNotification(
                    user.Email,
                    user.Username,
                    doc.Name,
                    daysUntilExpiry,
                )
            }
        }
        
        // Mark document as notified
        uc.docRepo.MarkExpiryNotified(doc.ID)
    }
    
    return nil
}
```

#### C. Background Job (Cron/Scheduler)
**File**: `backend/internal/infrastructure/scheduler/scheduler.go`

```go
package scheduler

import (
    "github.com/robfig/cron/v3"
    "github.com/repoareta/pedeve-dms-app/backend/internal/usecase"
)

type Scheduler struct {
    cron *cron.Cron
    notificationUC usecase.NotificationUseCase
}

func NewScheduler(notificationUC usecase.NotificationUseCase) *Scheduler {
    return &Scheduler{
        cron: cron.New(),
        notificationUC: notificationUC,
    }
}

func (s *Scheduler) Start() {
    // Check expiring documents setiap hari jam 9 pagi
    s.cron.AddFunc("0 9 * * *", func() {
        s.notificationUC.CheckExpiringDocuments()
    })
    
    s.cron.Start()
}

func (s *Scheduler) Stop() {
    s.cron.Stop()
}
```

#### D. Server-Sent Events (SSE) Endpoint
**File**: `backend/internal/delivery/http/notification_handler.go`

```go
func (h *NotificationHandler) StreamNotifications(c *fiber.Ctx) error {
    userID := c.Locals("user_id").(string)
    
    c.Set("Content-Type", "text/event-stream")
    c.Set("Cache-Control", "no-cache")
    c.Set("Connection", "keep-alive")
    
    // Create channel untuk notifications
    notifChan := make(chan domain.Notification)
    
    // Start goroutine untuk check notifications
    go func() {
        ticker := time.NewTicker(5 * time.Second) // Check setiap 5 detik
        defer ticker.Stop()
        
        for {
            select {
            case <-ticker.C:
                notifications, _ := h.notificationUC.GetUserNotifications(userID, true)
                for _, notif := range notifications {
                    notifChan <- notif
                }
            case <-c.Context().Done():
                return
            }
        }
    }()
    
    // Stream notifications ke client
    for {
        select {
        case notif := <-notifChan:
            data, _ := json.Marshal(notif)
            fmt.Fprintf(c, "data: %s\n\n", data)
            c.Response().SetBodyStreamWriter(func(w *bufio.Writer) {
                w.Flush()
            })
        case <-c.Context().Done():
            return nil
        }
    }
}
```

### 3. Frontend Components

#### A. Notification Service
**File**: `frontend/src/services/notificationService.ts`

```typescript
import { EventSourcePolyfill } from 'event-source-polyfill'
import apiClient from '../api/client'

export interface Notification {
  id: string
  user_id: string
  type: string
  title: string
  message: string
  resource_type?: string
  resource_id?: string
  is_read: boolean
  created_at: string
  read_at?: string
}

class NotificationService {
  private eventSource: EventSource | null = null
  private listeners: ((notification: Notification) => void)[] = []

  connect(userId: string) {
    const token = localStorage.getItem('auth_token')
    
    this.eventSource = new EventSourcePolyfill(
      `${import.meta.env.VITE_API_URL}/api/v1/notifications/stream`,
      {
        headers: {
          Authorization: `Bearer ${token}`,
        },
      }
    )

    this.eventSource.onmessage = (event) => {
      const notification: Notification = JSON.parse(event.data)
      this.listeners.forEach(listener => listener(notification))
    }

    this.eventSource.onerror = (error) => {
      console.error('SSE connection error:', error)
      // Fallback to polling
      this.startPolling(userId)
    }
  }

  private startPolling(userId: string) {
    setInterval(async () => {
      try {
        const response = await apiClient.get<Notification[]>(
          `/notifications?user_id=${userId}&unread_only=true`
        )
        response.data.forEach(notif => {
          this.listeners.forEach(listener => listener(notif))
        })
      } catch (error) {
        console.error('Polling error:', error)
      }
    }, 10000) // Poll setiap 10 detik
  }

  onNotification(callback: (notification: Notification) => void) {
    this.listeners.push(callback)
  }

  disconnect() {
    if (this.eventSource) {
      this.eventSource.close()
      this.eventSource = null
    }
  }

  async markAsRead(notificationId: string) {
    await apiClient.put(`/notifications/${notificationId}/read`)
  }

  async getNotifications(unreadOnly = false) {
    const response = await apiClient.get<Notification[]>(
      `/notifications?unread_only=${unreadOnly}`
    )
    return response.data
  }
}

export default new NotificationService()
```

#### B. Notification Component
**File**: `frontend/src/components/NotificationBell.vue`

```vue
<template>
  <a-badge :count="unreadCount" :offset="[10, 0]">
    <a-button type="text" @click="showDrawer = true">
      <IconifyIcon icon="mdi:bell" width="24" />
    </a-button>
  </a-badge>
  
  <a-drawer
    v-model:open="showDrawer"
    title="Notifikasi"
    placement="right"
    :width="400"
  >
    <a-list
      :data-source="notifications"
      :loading="loading"
    >
      <template #renderItem="{ item }">
        <a-list-item
          :class="{ 'unread': !item.is_read }"
          @click="handleNotificationClick(item)"
        >
          <a-list-item-meta>
            <template #title>
              {{ item.title }}
            </template>
            <template #description>
              {{ item.message }}
              <div class="notification-time">
                {{ formatTime(item.created_at) }}
              </div>
            </template>
          </a-list-item-meta>
        </a-list-item>
      </template>
    </a-list>
  </a-drawer>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import notificationService, { Notification } from '../services/notificationService'
import dayjs from 'dayjs'

const notifications = ref<Notification[]>([])
const loading = ref(false)
const showDrawer = ref(false)

const unreadCount = computed(() => {
  return notifications.value.filter(n => !n.is_read).length
})

onMounted(async () => {
  // Load existing notifications
  loading.value = true
  notifications.value = await notificationService.getNotifications()
  loading.value = false
  
  // Connect to SSE
  const userId = localStorage.getItem('user_id')
  if (userId) {
    notificationService.connect(userId)
    notificationService.onNotification((notif) => {
      notifications.value.unshift(notif)
      // Show browser notification
      if ('Notification' in window && Notification.permission === 'granted') {
        new Notification(notif.title, {
          body: notif.message,
          icon: '/iconPertamina.png',
        })
      }
    })
  }
})

onUnmounted(() => {
  notificationService.disconnect()
})

const handleNotificationClick = async (notif: Notification) => {
  if (!notif.is_read) {
    await notificationService.markAsRead(notif.id)
    notif.is_read = true
  }
  
  // Navigate to resource if available
  if (notif.resource_type === 'document' && notif.resource_id) {
    router.push(`/documents/${notif.resource_id}`)
  }
}

const formatTime = (date: string) => {
  return dayjs(date).fromNow()
}
</script>
```

## 🔧 Environment Variables

```bash
# Email Configuration (SendGrid)
SENDGRID_API_KEY=your-sendgrid-api-key
EMAIL_FROM=noreply@pedeve-dev.aretaamany.com

# ATAU SMTP Configuration
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your-sendgrid-api-key
EMAIL_FROM=noreply@pedeve-dev.aretaamany.com
```

## 📦 Dependencies

### Backend
```go
// go.mod
require (
    github.com/sendgrid/sendgrid-go v3.14.0+incompatible  // Untuk SendGrid
    // ATAU
    // net/smtp sudah built-in di Go
    
    github.com/robfig/cron/v3 v3.0.1  // Untuk scheduler
)
```

### Frontend
```json
{
  "dependencies": {
    "event-source-polyfill": "^2.0.2"  // Untuk SSE support di semua browser
  }
}
```

## 🚀 Implementation Steps

1. **Setup Email Service**
   - Daftar SendGrid (gratis 100 email/hari)
   - Dapatkan API key
   - Implement email service di backend

2. **Database Migration**
   - Tambahkan kolom `expiry_date` dan `expiry_notified` ke `documents`
   - Buat tabel `notifications`, `notification_settings`, `notification_configs`

3. **Backend Implementation**
   - Implement notification usecase
   - Implement scheduler untuk check expiring documents
   - Implement SSE endpoint untuk real-time notifications

4. **Frontend Implementation**
   - Implement notification service dengan SSE
   - Buat notification bell component
   - Integrate ke dashboard header

5. **Testing**
   - Test email delivery
   - Test in-app notifications
   - Test scheduler

## 💰 Cost Estimation

- **SendGrid**: Gratis (100 email/hari) atau $15/bulan (40k email)
- **GCP Compute**: Tidak ada biaya tambahan (menggunakan VM yang sudah ada)
- **Database**: Tidak ada biaya tambahan (menggunakan Cloud SQL yang sudah ada)

## ⚠️ Considerations

1. **Email Deliverability**: SendGrid memiliki deliverability yang baik, mengurangi risiko masuk spam
2. **Rate Limiting**: SendGrid free tier: 100 email/hari, pastikan tidak melebihi
3. **SSE Fallback**: Implement polling sebagai fallback jika SSE tidak tersedia
4. **Browser Notifications**: Perlu permission dari user, optional feature
5. **Notification Cleanup**: Hapus notifikasi lama (>30 hari) untuk menjaga performa

