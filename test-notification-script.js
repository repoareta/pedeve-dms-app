// ============================================
// SCRIPT UNTUK TEST NOTIFIKASI DOKUMEN
// ============================================
// Copy-paste script ini ke browser console dan jalankan

(async function() {
  try {
    console.log('🚀 Memulai test notifikasi...\n')
    
    // ============================================
    // STEP 1: Cek Role User
    // ============================================
    console.log('📋 STEP 1: Mengecek role user...')
    const authUser = JSON.parse(localStorage.getItem('auth_user') || 'null')
    
    if (!authUser) {
      console.error('❌ ERROR: User tidak ditemukan!')
      console.log('Silakan login terlebih dahulu')
      return
    }
    
    console.log('✅ User ditemukan:', authUser.username)
    console.log('   Role:', authUser.role)
    
    if (authUser.role?.toLowerCase() !== 'superadmin') {
      console.error('❌ ERROR: Anda harus login sebagai superadmin!')
      console.log('Role saat ini:', authUser.role)
      console.log('Silakan logout dan login kembali sebagai superadmin')
      return
    }
    
    console.log('✅ Login sebagai superadmin\n')
    
    // ============================================
    // STEP 2: Get CSRF Token
    // ============================================
    console.log('📝 STEP 2: Mengambil CSRF token...')
    
    const csrfResponse = await fetch('http://localhost:8080/api/v1/csrf-token', {
      method: 'GET',
      credentials: 'include',
      headers: {
        'Accept': 'application/json'
      }
    })
    
    if (!csrfResponse.ok) {
      console.error('❌ Gagal mengambil CSRF token:', csrfResponse.status, csrfResponse.statusText)
      const errorText = await csrfResponse.text()
      console.error('Response:', errorText)
      return
    }
    
    const csrfData = await csrfResponse.json()
    console.log('✅ CSRF token berhasil didapat')
    console.log('   Token:', csrfData.csrf_token ? '***' + csrfData.csrf_token.slice(-8) : 'null')
    
    if (!csrfData.csrf_token) {
      console.error('❌ ERROR: CSRF token tidak ditemukan dalam response!')
      console.log('Response:', csrfData)
      return
    }
    
    const csrfToken = csrfData.csrf_token
    console.log('')
    
    // ============================================
    // STEP 3: Get JWT Token (untuk Authorization header)
    // ============================================
    console.log('🔑 STEP 3: Mengambil JWT token...')
    const jwtToken = localStorage.getItem('auth_token')
    
    if (!jwtToken) {
      console.warn('⚠️  WARNING: JWT token tidak ditemukan di localStorage')
      console.log('Menggunakan cookie saja (httpOnly cookie)')
    } else {
      console.log('✅ JWT token ditemukan')
    }
    console.log('')
    
    // ============================================
    // STEP 4: Create Notification untuk Dokumen
    // ============================================
    console.log('📬 STEP 4: Membuat notifikasi untuk dokumen...')
    
    const documentId = '9abf3ec5-9999-46f3-8906-7ba3ca284770'
    console.log('   Document ID:', documentId)
    
    // Prepare headers
    const headers = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'X-CSRF-Token': csrfToken
    }
    
    // Add Authorization header jika ada
    if (jwtToken) {
      headers['Authorization'] = `Bearer ${jwtToken}`
    }
    
    console.log('   Headers:', {
      'Content-Type': headers['Content-Type'],
      'X-CSRF-Token': '***' + csrfToken.slice(-8),
      'Authorization': jwtToken ? 'Bearer ***' : 'Not set (using cookie)'
    })
    
    const response = await fetch('http://localhost:8080/api/v1/development/create-notification-for-document', {
      method: 'POST',
      headers: headers,
      credentials: 'include', // Penting untuk mengirim cookie
      body: JSON.stringify({ 
        document_id: documentId
      })
    })
    
    console.log('   Response status:', response.status, response.statusText)
    
    // Parse response
    const responseText = await response.text()
    let result
    try {
      result = JSON.parse(responseText)
    } catch (e) {
      console.error('❌ ERROR: Response bukan JSON!')
      console.log('Response text:', responseText)
      return
    }
    
    if (!response.ok) {
      console.error('❌ ERROR:', response.status, result)
      
      if (response.status === 403) {
        if (result.error === 'csrf_token_invalid' || result.error === 'csrf_token_missing') {
          console.error('❌ CSRF token invalid atau missing!')
          console.log('Coba refresh halaman dan jalankan script lagi')
        } else {
          console.error('❌ FORBIDDEN: Hanya superadmin yang bisa mengakses endpoint ini!')
          console.log('Pastikan Anda login sebagai superadmin')
        }
      } else if (response.status === 401) {
        console.error('❌ UNAUTHORIZED: Token expired atau tidak valid!')
        console.log('Silakan refresh halaman atau login kembali')
      } else if (response.status === 404) {
        console.error('❌ NOT FOUND: Dokumen tidak ditemukan!')
        console.log('Pastikan document_id benar:', documentId)
      }
      
      return
    }
    
    // Success!
    console.log('✅ BERHASIL! Notifikasi sudah dibuat')
    console.log('')
    console.log('📊 Hasil:')
    console.log('   Message:', result.message)
    console.log('   Notification ID:', result.notification?.id)
    console.log('   Document:', {
      id: result.document?.id,
      name: result.document?.name,
      expiry_date: result.document?.expiry_date,
      days_until_expiry: result.document?.days_until_expiry
    })
    console.log('')
    console.log('💡 Langkah selanjutnya:')
    console.log('   1. Refresh halaman notifikasi (/notifications)')
    console.log('   2. Atau cek bell icon di header')
    console.log('   3. Notifikasi akan muncul untuk uploader dokumen tersebut')
    
  } catch (error) {
    console.error('❌ ERROR:', error)
    console.error('Stack:', error.stack)
  }
})()

