# GoERP Project Guide (Enterprise Edition)

**GoERP** adalah platform ERP modular berbasis Go yang menggunakan arsitektur *DocType-driven* dan *SaaS-ready*. Sistem ini dirancang untuk skalabilitas masif, integritas finansial deterministik, dan fleksibilitas *Low-Code*.

## 🚀 Quick Start (Production Mode)

1.  **Environment Setup**:
    ```bash
    cp .env.example .env # Sesuaikan kredensial
    ```
2.  **Launch Infrastructure**:
    ```bash
    ./scripts/setup_production.sh
    ```
    *Script ini akan membangun Docker Containers (App, DB, Redis) dan melakukan provisioning Tenant Administrator pertama.*

---

## 🧠 Core Architecture (The Engines)

GoERP bukan sekadar aplikasi, melainkan kumpulan engine cerdas:

### 1. Meta & UI Engine (The Interpreter)
*   **DocType-Driven**: Seluruh skema database, validasi, dan layout UI didefinisikan dalam JSON.
*   **Dynamic Renderer**: Frontend React merender Form, List, dan Dashboard secara otomatis berdasarkan metadata.
*   **Customization**: Mendukung *Custom Fields* dan *Customize Form* per-tenant tanpa menyentuh kode.

### 2. Integrity & Reliability Layer
*   **Transaction Boundary**: Operasi atomik menjamin Header, Child Table, dan Ledger tersimpan utuh.
*   **Concurrency Control**: Menggunakan *Pessimistic Locking* (Submit/Cancel) dan *Optimistic Locking* (Update).
*   **Idempotency Engine**: Proteksi *Double-Posting* menggunakan Header `X-GoERP-Idempotency-Key`.
*   **Numerical Guard**: Enforced precision (2 desimal) pada seluruh kalkulasi finansial.

### 3. Unified Ledger System
*   **FIFO Valuation**: Melacak nilai inventaris secara presisi berdasarkan antrean stok asli.
*   **Perpetual Accounting**: Setiap mutasi fisik (stok) memicu entri akuntansi (GL) secara real-time.
*   **Backdated Reposting**: "Mesin Waktu" yang otomatis menghitung ulang saldo stok dan nilai aset jika ada transaksi tanggal lampau.

### 4. Logic & Rule Engine
*   **Lua Rule Engine**: Konfigurasi diskon, pajak, dan validasi bisnis via UI menggunakan script Lua yang aman.
*   **Client Scripting**: Logika UI (show/hide field, fetch data) menggunakan JavaScript sandbox.
*   **Tax & Compliance**: Engine kalkulasi pajak dinamis (Inclusive/Exclusive) yang terintegrasi ke GL.

---

## 🛠 Developer Guide: Extending GoERP

### Menambah DocType Baru
Tempatkan definisi di `apps/[app]/[module]/doctype/[name]/schema.json`.
Sistem akan otomatis melakukan:
1.  **Auto-Migration**: Membuat tabel dan index di database.
2.  **API Generation**: Membuat endpoint CRUD `/api/v1/resource/[name]`.
3.  **UI Generation**: Membuat form entry dan list view.

### Menggunakan Hooks
Daftarkan fungsi di `hooks.go` modul Anda:
```go
registry.DefaultHookRegistry.Register("SalesInvoice", types.OnSubmit, logic.PostingFungsi)
```

### Scripting API (`frm`)
Di metadata DocType, gunakan JavaScript:
```javascript
return {
  refresh: (frm) => {
    if (frm.doc.total > 1000) frm.set_df_property('discount', 'hidden', false);
  }
}
```

---

## 📊 Project Status & Roadmap

### ✅ Status: PRODUCTION READY
*   **Core**: Meta System, Multi-tenancy, Auth, Permission.
*   **Accounting**: GL, Trial Balance, P&L, Balance Sheet, Fiscal Closing.
*   **Stock**: FIFO, Bin, SLE, Backdated Reposting.
*   **Infrastructure**: Docker, Redis Cache, Job Queue, Tracing.
*   **Integrations**: Webhooks, Document Mapping, Search Engine.

### ⏳ In Progress
*   **Plugin Store**: Runtime instalasi modul pihak ketiga.
*   **Advanced Analytics**: Data Warehouse integration (ClickHouse).

### 🗺 Future Roadmap
*   **AI Decision Layer**: Prediksi stok dan deteksi anomali fraud.
*   **Mobile Native App**: Menggunakan engine yang sama via REST API.

---

## 🔐 Compliance & Security
*   **Data Isolation**: Isolasi ketat antar tenant via `tenant_id` di setiap baris data.
*   **Audit Trail**: Versi dokumen (Versioning) mencatat siapa mengubah apa dan kapan.
*   **Fiscal Guard**: Penguncian periode mencegah manipulasi data tahun lalu yang sudah ditutup.

---
**GoERP Engine v2.0** - *Built for integrity, designed for scale.*
