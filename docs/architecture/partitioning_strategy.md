# 🛡️ GoERP Database Partitioning Strategy

Untuk memastikan performa sistem tetap stabil saat data mencapai jutaan baris, GoERP mengadopsi strategi **Declarative Partitioning** pada PostgreSQL.

## 1. Tabel Target
Tabel berikut wajib dipartisi karena memiliki volume transaksi tinggi:
- `tabGLEntry` (General Ledger)
- `tabStockLedgerEntry` (Mutasi Stok)
- `tabActivityLog` (Audit Trail)

## 2. Strategi Partisi: Range by Date
Partisi akan dibagi berdasarkan kolom `posting_date` dengan interval **bulanan**.

### Contoh Implementasi SQL:
```sql
-- 1. Buat tabel induk (Template)
CREATE TABLE tabGLEntry (
    name VARCHAR(255),
    tenant_id VARCHAR(255),
    account VARCHAR(255),
    debit DECIMAL(18,4),
    credit DECIMAL(18,4),
    posting_date DATE NOT NULL,
    -- ... kolom lainnya
    PRIMARY KEY (name, posting_date) -- Kolom partisi wajib masuk PK
) PARTITION BY RANGE (posting_date);

-- 2. Buat partisi bulanan
CREATE TABLE tabGLEntry_y2026m04 PARTITION OF tabGLEntry
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

CREATE TABLE tabGLEntry_y2026m05 PARTITION OF tabGLEntry
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
```

## 3. Otomasi Partisi
GoERP menggunakan **Background Job** (`PartitionMaintenanceJob`) yang berjalan setiap akhir bulan untuk:
1. Membuat partisi baru untuk bulan depan.
2. Memindahkan data lama (Cold Storage) ke tabel arsip jika sudah lebih dari 5 tahun.

## 4. Keuntungan Operasional
- **Query Pruning**: PostgreSQL hanya akan memindai partisi yang relevan dengan filter tanggal.
- **Fast Deletion**: Menghapus data tahun lama cukup dengan `DROP PARTITION`, jauh lebih cepat daripada `DELETE FROM`.
- **Vacuum Optimization**: Vacuuming hanya dilakukan pada partisi yang aktif berubah.
