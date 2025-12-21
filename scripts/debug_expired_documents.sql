-- Query untuk melihat dokumen yang seharusnya terdeteksi sebagai expired
-- Jalankan query ini di database untuk debugging

-- 1. Lihat semua dokumen dengan expiry_date (tidak NULL)
SELECT 
    id,
    name,
    file_name,
    expiry_date,
    uploader_id,
    folder_id,
    created_at,
    updated_at,
    CASE 
        WHEN expiry_date IS NULL THEN 'NULL'
        WHEN expiry_date < CURRENT_DATE THEN 'SUDAH EXPIRED'
        WHEN expiry_date = CURRENT_DATE THEN 'EXPIRED HARI INI'
        WHEN expiry_date <= CURRENT_DATE + INTERVAL '14 days' THEN 'AKAN EXPIRED DALAM 14 HARI'
        ELSE 'MASIH JAUH'
    END as status_expiry
FROM documents
WHERE expiry_date IS NOT NULL
ORDER BY expiry_date ASC;

-- 2. Lihat dokumen yang expired hari ini sampai 14 hari ke depan (threshold default)
SELECT 
    id,
    name,
    file_name,
    expiry_date,
    uploader_id,
    folder_id,
    CURRENT_DATE as today,
    expiry_date - CURRENT_DATE as days_until_expiry,
    CASE 
        WHEN expiry_date < CURRENT_DATE THEN 'SUDAH EXPIRED'
        WHEN expiry_date = CURRENT_DATE THEN 'EXPIRED HARI INI'
        ELSE 'AKAN EXPIRED'
    END as status
FROM documents
WHERE expiry_date IS NOT NULL 
    AND expiry_date <= CURRENT_DATE + INTERVAL '14 days'
ORDER BY expiry_date ASC;

-- 3. Lihat dokumen yang expired hari ini sampai 10 hari ke depan (sesuai permintaan user)
SELECT 
    id,
    name,
    file_name,
    expiry_date,
    uploader_id,
    folder_id,
    CURRENT_DATE as today,
    expiry_date - CURRENT_DATE as days_until_expiry,
    CASE 
        WHEN expiry_date < CURRENT_DATE THEN 'SUDAH EXPIRED'
        WHEN expiry_date = CURRENT_DATE THEN 'EXPIRED HARI INI'
        ELSE 'AKAN EXPIRED'
    END as status
FROM documents
WHERE expiry_date IS NOT NULL 
    AND expiry_date <= CURRENT_DATE + INTERVAL '10 days'
ORDER BY expiry_date ASC;

-- 4. Cek apakah ada dokumen dengan expiry_date di metadata tapi NULL di kolom expiry_date
-- Extract expired_date dari metadata JSON
SELECT 
    id,
    name,
    file_name,
    expiry_date as expiry_date_column,
    metadata->>'expired_date' as expired_date_from_metadata,
    metadata->>'expiry_date' as expiry_date_from_metadata_alt,
    metadata,
    CASE 
        WHEN expiry_date IS NULL THEN 'expiry_date NULL'
        ELSE 'expiry_date ADA'
    END as expiry_status,
    CASE 
        WHEN expiry_date IS NOT NULL THEN 'Dari Kolom'
        WHEN metadata->>'expired_date' IS NOT NULL THEN 'Dari Metadata (expired_date)'
        WHEN metadata->>'expiry_date' IS NOT NULL THEN 'Dari Metadata (expiry_date)'
        ELSE 'TIDAK ADA'
    END as source_of_expiry_date
FROM documents
WHERE metadata IS NOT NULL
    AND (metadata::text LIKE '%expired_date%' OR metadata::text LIKE '%expiry_date%')
ORDER BY created_at DESC
LIMIT 20;

-- 5. Hitung total dokumen dengan expiry_date (dari kolom ATAU metadata)
-- Query ini menghitung dokumen yang memiliki expiry_date di kolom
SELECT 
    'Dari Kolom expiry_date' as source,
    COUNT(*) as total_documents_with_expiry,
    COUNT(CASE WHEN expiry_date < CURRENT_DATE THEN 1 END) as already_expired,
    COUNT(CASE WHEN expiry_date = CURRENT_DATE THEN 1 END) as expired_today,
    COUNT(CASE WHEN expiry_date > CURRENT_DATE AND expiry_date <= CURRENT_DATE + INTERVAL '14 days' THEN 1 END) as expiring_in_14_days,
    COUNT(CASE WHEN expiry_date > CURRENT_DATE + INTERVAL '14 days' THEN 1 END) as expiring_later
FROM documents
WHERE expiry_date IS NOT NULL;

-- 6. Hitung total dokumen dengan expired_date di metadata (bukan di kolom)
-- Extract dan parse expired_date dari metadata JSON
WITH documents_with_metadata_expiry AS (
    SELECT 
        id,
        name,
        file_name,
        expiry_date,
        -- Coba extract dari expired_date dulu, jika tidak ada coba expiry_date
        COALESCE(
            (metadata->>'expired_date')::date,
            (metadata->>'expiry_date')::date,
            NULL
        ) as expired_date_from_metadata
    FROM documents
    WHERE metadata IS NOT NULL
        AND expiry_date IS NULL  -- Hanya yang tidak punya expiry_date di kolom
        AND (
            metadata->>'expired_date' IS NOT NULL 
            OR metadata->>'expiry_date' IS NOT NULL
        )
)
SELECT 
    'Dari Metadata (expired_date/expiry_date)' as source,
    COUNT(*) as total_documents_with_expiry,
    COUNT(CASE WHEN expired_date_from_metadata < CURRENT_DATE THEN 1 END) as already_expired,
    COUNT(CASE WHEN expired_date_from_metadata = CURRENT_DATE THEN 1 END) as expired_today,
    COUNT(CASE WHEN expired_date_from_metadata > CURRENT_DATE AND expired_date_from_metadata <= CURRENT_DATE + INTERVAL '14 days' THEN 1 END) as expiring_in_14_days,
    COUNT(CASE WHEN expired_date_from_metadata > CURRENT_DATE + INTERVAL '14 days' THEN 1 END) as expiring_later
FROM documents_with_metadata_expiry;

-- 7. Gabungan: Hitung total dokumen dengan expiry_date (dari kolom ATAU metadata)
WITH all_expiry_dates AS (
    SELECT 
        id,
        name,
        file_name,
        -- Gunakan expiry_date dari kolom jika ada, jika tidak ambil dari metadata
        COALESCE(
            expiry_date,
            (metadata->>'expired_date')::date,
            (metadata->>'expiry_date')::date,
            NULL
        ) as effective_expiry_date
    FROM documents
    WHERE expiry_date IS NOT NULL 
        OR (metadata IS NOT NULL AND (
            metadata->>'expired_date' IS NOT NULL 
            OR metadata->>'expiry_date' IS NOT NULL
        ))
)
SELECT 
    'TOTAL (Kolom + Metadata)' as source,
    COUNT(*) as total_documents_with_expiry,
    COUNT(CASE WHEN effective_expiry_date < CURRENT_DATE THEN 1 END) as already_expired,
    COUNT(CASE WHEN effective_expiry_date = CURRENT_DATE THEN 1 END) as expired_today,
    COUNT(CASE WHEN effective_expiry_date > CURRENT_DATE AND effective_expiry_date <= CURRENT_DATE + INTERVAL '14 days' THEN 1 END) as expiring_in_14_days,
    COUNT(CASE WHEN effective_expiry_date > CURRENT_DATE + INTERVAL '14 days' THEN 1 END) as expiring_later
FROM all_expiry_dates
WHERE effective_expiry_date IS NOT NULL;

-- 8. Detail dokumen yang akan expired (dari kolom ATAU metadata) dalam 14 hari ke depan
WITH all_expiry_dates AS (
    SELECT 
        id,
        name,
        file_name,
        expiry_date as expiry_date_column,
        (metadata->>'expired_date')::date as expired_date_from_metadata,
        (metadata->>'expiry_date')::date as expiry_date_from_metadata_alt,
        -- Gunakan expiry_date dari kolom jika ada, jika tidak ambil dari metadata
        COALESCE(
            expiry_date,
            (metadata->>'expired_date')::date,
            (metadata->>'expiry_date')::date,
            NULL
        ) as effective_expiry_date,
        CASE 
            WHEN expiry_date IS NOT NULL THEN 'Kolom expiry_date'
            WHEN metadata->>'expired_date' IS NOT NULL THEN 'Metadata expired_date'
            WHEN metadata->>'expiry_date' IS NOT NULL THEN 'Metadata expiry_date'
            ELSE 'TIDAK ADA'
        END as source
    FROM documents
    WHERE expiry_date IS NOT NULL 
        OR (metadata IS NOT NULL AND (
            metadata->>'expired_date' IS NOT NULL 
            OR metadata->>'expiry_date' IS NOT NULL
        ))
)
SELECT 
    id,
    name,
    file_name,
    effective_expiry_date,
    effective_expiry_date - CURRENT_DATE as days_until_expiry,
    source,
    CASE 
        WHEN effective_expiry_date < CURRENT_DATE THEN 'SUDAH EXPIRED'
        WHEN effective_expiry_date = CURRENT_DATE THEN 'EXPIRED HARI INI'
        WHEN effective_expiry_date <= CURRENT_DATE + INTERVAL '14 days' THEN 'AKAN EXPIRED DALAM 14 HARI'
        ELSE 'MASIH JAUH'
    END as status
FROM all_expiry_dates
WHERE effective_expiry_date IS NOT NULL
    AND effective_expiry_date <= CURRENT_DATE + INTERVAL '14 days'
ORDER BY effective_expiry_date ASC;
