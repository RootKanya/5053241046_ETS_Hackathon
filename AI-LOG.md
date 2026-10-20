# AI-LOG

Catat setiap penggunaan AI yang menghasilkan kode. Jika tidak memakai AI sama sekali, tulis satu baris: **tidak memakai AI**.

Log yang jujur tidak mengurangi nilai, sebanyak apa pun AI dipakai. Yang mengurangi nilai adalah log yang tidak sesuai dengan kode, atau isi log yang tidak bisa dijelaskan.

## Struktur Proyek & WSL Setup
- Prompt: "okay now i want to cleate the file folder structire can you give me the wsl command to do so"
- Dipakai: Perintah Bash/WSL (`mkdir -p`, `go mod init`, `touch`) untuk membuat folder dan file proyek di WSL.
- Diubah: Langsung dieksekusi di terminal WSL untuk menginisialisasi repositori dari awal.

## Konstruksi README.md & Ceklis Fitur
- Prompt: " it all come out from the file i gave please make it verry simple this is only an hour im tired also make sure this [template README]"
- Dipakai: Ringkasan README.md beserta daftar ceklis fitur B1-B7 dan status Persona B.
- Diubah: Disesuaikan dengan status pengerjaan aktual untuk siap dimasukkan ke repositori GitHub.

## Penjelasan Aturan Bisnis & Konsep Sumber Daya
- Prompt: "sumberdaya is a human or the proyektor and cable ?"
- Dipakai: Pemahaman konsep domain model antara `sumber_daya` (aset/alat) dan `peminjam` (subjek mahasiswa).
- Diubah: Memastikan pemisahan data struct dan URL handler antara aset kampus dan identitas mahasiswa.

## Panduan Pengujian API dengan Postman
- Prompt: "okay now i wanna use postman to test connection how ??"
- Dipakai: Langkah-langkah pengujian endpoint `sumber-daya`, `peminjaman`, dan `kuota` di Postman.
- Diubah: Mengeksekusi request secara berurutan mulai dari POST sumber daya hingga pengecekan kuota khas Persona B.

## Perbaikan URL & Format Request Postman
- Prompt: "like this or sumberdaya is the link"
- Dipakai: Format penulisan URL endpoint `http://localhost:8080/sumber-daya` pada request bar Postman.
- Diubah: Mengganti teks `sumber-daya` menjadi URL lengkap dengan HTTP method `POST` dan header `Content-Type: application/json`.

## Troubleshooting Error 404 pada Endpoint DELETE (Peminjaman ID)
- Prompt: "still, also is peminjaman the same as sumberdaya create"
- Dipakai: Analisis penyebab error `404 Not Found` pada `DELETE /peminjaman/5`.
- Diubah: Memahami bahwa `sumber_daya` dan `peminjaman` menggunakan counter ID terpisah di memori, lalu membuat data peminjaman terlebih dahulu via `POST /peminjaman` sebelum menghapusnya.

## Perbaikan HTTP Method & Ekspor Koleksi Postman
- Prompt: "is this all it where do i need to doanload or attch link postmena to my proses"
- Dipakai: Langkah ekspor koleksi Postman menjadi file JSON (`postman_collection.json`).
- Diubah: Memperbaiki HTTP Method pada koleksi (mengubah `Create Peminjaman` ke `POST` dan `Delete Peminjaman` ke `DELETE`) sebelum diekspor ke repositori.

## Change Request A: Dashboard Ringkasan Harian
- Prompt: "hmm i want to add a new sruct below wdyt"
- Dipakai: Struct RingkasanResponse, GetDashboardSummary() di store.go, dan handleGetRingkasan() di handler.go.
- Diubah: Menyesuaikan perhitungan total_unit_slot (jumlah unit * durasi slot) dan memastikan status 400 jika tanggal kosong/invalid.