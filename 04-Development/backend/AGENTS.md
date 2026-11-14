AGENTS.md — Aturan Pengembangan Backend Go

Selalu pakai gaya ponytail (kode seminimal mungkin) di atas wadah context-mode; keduanya digabung dan aktif otomatis.
1. Tujuan
Bertindak sebagai engineer backend Go senior.
Tulis kode yang benar, aman, mudah dirawat, dan layak digunakan pada lingkungan produksi.
Utamakan kesederhanaan, ketepatan, dan konsistensi daripada banyaknya kode. Jangan membuat implementasi hanya agar tugas terlihat selesai.
2. Periksa Proyek Sebelum Mengubah Kode
Sebelum melakukan perubahan:
Baca `AGENTS.md` pada direktori root.
Periksa direktori backend dan file `go.mod`.
Identifikasi versi Go dan dependensi yang sudah digunakan.
Periksa handler, service, repository, model, middleware, konfigurasi, serta pengujian yang relevan.
Ikuti arsitektur dan konvensi penamaan yang sudah diterapkan.
Baca dokumentasi terkait di direktori `rules-direction/` jika direktori tersebut tersedia.
Temukan akar masalah sebelum memperbaiki bug.
Jangan berasumsi proyek menggunakan framework, ORM, database, atau arsitektur tertentu sebelum memeriksa kode yang ada.
3. Aturan Anti-Kode Berlebihan (Anti-Slop)
Buat solusi paling sederhana yang tetap lengkap.
Jangan membuat file, package, interface, atau abstraksi yang tidak diperlukan.
Jangan menggandakan fungsi yang sudah tersedia.
Jangan menulis ulang kode yang sudah berjalan tanpa alasan konkret.
Jangan melakukan refaktor pada modul yang tidak berkaitan dengan tugas.
Jangan menambahkan dependensi jika standard library atau dependensi yang sudah ada sudah mencukupi.
Jangan membuat framework generik untuk satu kebutuhan sederhana.
Jangan menambahkan fitur berdasarkan dugaan atau spekulasi.
Jangan menggunakan data palsu atau respons hardcode pada jalur produksi.
Jangan mengembalikan status berhasil ketika operasi sebenarnya gagal.
Jangan mengabaikan error secara diam-diam.
Jangan meninggalkan implementasi placeholder pada jalur produksi.
Jangan menggunakan state global yang dapat diubah tanpa alasan kuat.
Jangan menambahkan konkurensi jika tidak menyelesaikan masalah nyata.
Jangan menonaktifkan pengujian, error, atau pemeriksaan keamanan hanya agar tugas dinyatakan berhasil.
Setiap abstraksi harus memiliki manfaat yang jelas dan praktis.
4. Arsitektur
Ikuti arsitektur proyek yang sudah ada.
Pisahkan tanggung jawab dengan jelas:
HTTP handler: menangani HTTP, membaca request, dan menyusun response.
Logika bisnis: menangani aturan domain dan perilaku aplikasi.
Akses data: menangani query database dan penyimpanan data.
Middleware: menangani kebutuhan lintas proses HTTP.
Konfigurasi: membaca dan mengelola konfigurasi aplikasi.
Pembagian di atas adalah panduan tanggung jawab, bukan kewajiban untuk membuat package terpisah bagi setiap lapisan.
Jangan memaksakan Clean Architecture, pola repository, framework dependency injection, atau lapisan tambahan ke proyek yang tidak membutuhkannya.
Pastikan arah dependensi jelas. Hindari circular dependency dan keterikatan antarkomponen yang tidak perlu.
5. Standar Penulisan Go
Ikuti idiom Go dan versi Go yang digunakan proyek.
Format file Go yang diubah menggunakan `gofmt`.
Gunakan nama yang bermakna dan ringkas.
Buat fungsi dengan satu tanggung jawab dan ukuran yang wajar.
Utamakan alur kontrol sederhana daripada kode yang terlalu kreatif.
Tangani error secara eksplisit.
Gunakan `%w` saat membungkus error dan error aslinya masih perlu diperiksa.
Hindari nesting berlebihan dan abstraksi yang tidak diperlukan.
Gunakan pointer hanya jika semantiknya memang diperlukan.
Hindari alokasi yang tidak perlu dan optimasi prematur.
Jangan gunakan `panic` untuk error aplikasi biasa.
Jangan gunakan `init()` untuk proses inisialisasi aplikasi yang kompleks.
Hindari variabel global yang tidak diperlukan.
Berikan dokumentasi pada identifier yang diekspor jika sesuai.
Jangan membuat interface hanya agar kode terlihat mudah diuji.
Jaga API publik tetap minimal.
6. Penanganan Error
Jangan membuang error tanpa alasan yang jelas dan dapat dipertanggungjawabkan.
Jangan mengganti operasi yang gagal dengan hasil sukses palsu.
Tambahkan konteks pada error agar lokasi kegagalan lebih mudah diketahui.
Pertahankan error asli jika pemanggil perlu memeriksanya.
Hindari mencatat error yang sama berulang kali di setiap lapisan.
Kembalikan HTTP status code yang sesuai.
Jangan tampilkan stack trace, detail SQL, secret, atau detail implementasi internal pada response API publik.
Pertahankan format response error agar konsisten dengan API yang sudah ada.
7. Aturan HTTP API
Ikuti konvensi route dan format response yang sudah digunakan.
Gunakan HTTP method dan status code yang sesuai.
Validasi request body, query parameter, path parameter, serta header yang relevan.
Terapkan batas ukuran request jika diperlukan.
Berikan response error yang jelas dan konsisten.
Hindari mengekspos model database secara langsung jika hal tersebut menimbulkan keterikatan atau membocorkan field sensitif.
Tangani pembatalan request dan koneksi client yang terputus.
Atur timeout dan batas penggunaan resource yang sesuai.
Hindari operasi blocking yang dapat menghabiskan resource server.
Pertahankan kompatibilitas API kecuali perubahan yang memutus kompatibilitas memang diminta.
Dokumentasikan perubahan kontrak API yang bermakna.
8. Context dan Konkurensi
Teruskan `context.Context` pada operasi yang terkait dengan request.
Jangan menyimpan context khusus request di dalam struct atau state global.
Jangan mengganti context request yang tersedia dengan `context.Background()` tanpa alasan yang valid.
Pastikan setiap goroutine memiliki mekanisme penghentian yang jelas.
Hindari goroutine leak dan pembuatan worker tanpa batas.
Lindungi state bersama yang dapat diubah dari data race.
Gunakan mutex dan channel hanya jika memang sesuai.
Hindari menahan lock ketika melakukan operasi I/O yang tidak perlu.
Jalankan pemeriksaan race detector jika memungkinkan untuk perubahan yang berkaitan dengan konkurensi.
9. Database dan Penyimpanan Data
Ikuti database driver, ORM, dan alur migrasi yang sudah digunakan proyek.
Gunakan parameterized query.
Jangan pernah menggabungkan input yang tidak tepercaya langsung ke dalam SQL.
Tangani error database secara eksplisit.
Tutup `rows` dan resource lain yang sudah diperoleh.
Periksa `rows.Err()` jika relevan.
Gunakan transaksi jika beberapa operasi harus berhasil atau gagal secara atomik.
Jaga batas transaksi tetap jelas dan singkat.
Hindari N+1 query jika relevan.
Terapkan pagination untuk koleksi data yang berpotensi besar.
Tambahkan index berdasarkan pola query dan kebutuhan nyata.
Jangan menjalankan migrasi destruktif tanpa persetujuan eksplisit.
Jangan menghapus kolom, tabel, atau data yang sudah ada secara diam-diam.
Jangan mengarang perubahan skema untuk menutupi bug yang belum ditelusuri.
Tinjau kompatibilitas migrasi dan dampaknya terhadap proses rollback.
10. Keamanan
Validasi dan normalisasi input yang tidak tepercaya.
Terapkan otorisasi pada setiap operasi sensitif.
Jangan mengandalkan validasi frontend sebagai kontrol akses.
Cegah SQL injection, SSRF, path traversal, dan serangan injection lainnya.
Simpan password menggunakan algoritma password hashing yang sudah teruji.
Gunakan library kriptografi yang tepercaya, bukan implementasi kriptografi buatan sendiri.
Jauhkan secret dari source code, log, dan response API.
Jangan mencatat password, access token, session cookie, atau private key.
Atur CORS secara sengaja sesuai kebutuhan.
Terapkan rate limiting jika model ancaman membutuhkannya.
IP klien selalu lewat `middleware.ClientIP(r, cfg.TrustProxyHeaders)`; `TRUST_PROXY_HEADERS=true` hanya di belakang proxy (Railway).
Header keamanan (nosniff, X-Frame-Options, CSP, Referrer-Policy, Permissions-Policy, HSTS khusus HTTPS) dipasang global oleh `middleware.SecurityHeaders`; `/metrics` di luar rantai API.
Setiap rute terautentikasi baru wajib memakai limiter kelompoknya di `routes.go` (`userWriteLimit`, `userReadLimit`, `uploadLimit`, `adminLimit`, per user ID). `MAX_CONCURRENT_REQUESTS` (default 200, 0 = nonaktif) membatasi request bersamaan lewat `middleware.MaxConcurrent` (503 `server_busy`).
Atur timeout yang sesuai untuk HTTP request ke layanan eksternal.
Validasi tujuan redirect dan URL outbound jika relevan.
Gunakan pengaturan cookie dan session yang aman jika digunakan.
Jangan melemahkan kontrol keamanan hanya agar pengujian berhasil.
11. Autentikasi dan Otorisasi
Gunakan kembali mekanisme autentikasi yang sudah ada.
Verifikasi tanda tangan token, masa berlaku, dan claim yang diwajibkan.
Terapkan pemeriksaan izin di sisi server.
Periksa kepemilikan resource jika relevan.
Jangan memercayai user ID atau role yang dikirim oleh client.
Hindari membocorkan keberadaan akun atau resource sensitif jika informasi tersebut harus dirahasiakan.
Jaga response error autentikasi dan otorisasi tetap konsisten.
Jangan pernah membuat sendiri algoritma kriptografi token atau password.
12. Konfigurasi dan Dependensi
Ikuti sistem konfigurasi yang sudah digunakan proyek.
Baca secret dari environment atau sistem pengelolaan secret yang disetujui.
Validasi konfigurasi wajib saat aplikasi melakukan startup.
Berikan kegagalan yang jelas jika konfigurasi penting tidak valid.
Jangan menanam kredensial produksi langsung di kode.
Jangan commit file `.env` yang berisi nilai rahasia asli.
Jangan melakukan upgrade dependensi mayor secara diam-diam.
Jangan mengubah versi Go tanpa alasan yang jelas.
Pertahankan alur pengelolaan dependensi yang sudah digunakan.
Tinjau pemeliharaan, keamanan, dan kebutuhan setiap dependensi baru.
13. Pengujian
Tambahkan pengujian yang memverifikasi perilaku sebenarnya.
Prioritaskan pengujian untuk:
Logika bisnis.
Validasi dan jalur error.
Batas otorisasi.
Perilaku database jika relevan.
HTTP status code dan kontrak response.
Regression test untuk bug yang diperbaiki.
Perilaku konkurensi jika diperlukan.
Aturan pengujian:
Gunakan table-driven test jika membuat pengujian lebih jelas.
Hindari pengujian yang hanya menyalin detail implementasi.
Jangan membuat pengujian lolos dengan melemahkan assertion.
Jangan menghapus pengujian yang sudah ada tanpa alasan.
Jangan melakukan mock terhadap setiap dependensi secara berlebihan.
Gunakan integration test jika unit test tidak cukup untuk memverifikasi perilaku.
Pastikan pengujian deterministik dan terisolasi.
Jalankan pemeriksaan yang tersedia dan relevan dari direktori module yang benar:
```bash
gofmt -w .
go vet ./...
go test ./...
```
Gunakan `gofmt` pada file Go yang relevan; jangan menjalankannya secara membabi buta pada direktori yang berisi file non-Go. Jalankan `go test -race ./...` jika sesuai.
Jangan menyatakan pengujian berhasil jika pengujian tersebut belum dijalankan.
14. Performa
Utamakan kode yang jelas sebelum melakukan micro-optimization.
Hindari query database dan network request yang tidak perlu.
Hindari penggunaan memori dan konkurensi tanpa batas.
Gunakan streaming untuk payload besar jika sesuai.
Gunakan ulang resource secara aman jika memberikan manfaat.
Lakukan profiling sebelum menerapkan optimasi yang kompleks.
Jangan menambahkan caching tanpa strategi invalidasi yang jelas.
Jangan mengorbankan ketepatan demi peningkatan performa yang hanya berdasarkan dugaan.
15. Pengelolaan Perubahan
Sebelum mengedit:
Pahami jalur kode yang relevan.
Identifikasi pemanggil yang terdampak dan kontrak API.
Periksa perubahan pengguna yang belum di-commit.
Setelah mengedit:
Tinjau diff.
Hapus perubahan yang tidak sengaja dan kode yang tidak digunakan.
Jalankan formatter, pengujian, dan pemeriksaan statis yang relevan.
Periksa dampak terhadap kompatibilitas dan keamanan.
Jelaskan kebutuhan migrasi atau perubahan konfigurasi.
Jangan pernah menjalankan perintah Git destruktif atau membuang pekerjaan pengguna tanpa izin eksplisit.
16. Definisi Selesai
Tugas backend baru dapat dinyatakan selesai jika:
Perilaku yang diminta berjalan sesuai tujuan.
Konvensi proyek yang sudah ada tetap dipertahankan.
Validasi input dan otorisasi sudah benar.
Error ditangani dengan tepat.
Pengujian yang relevan sudah ditambahkan atau diperbarui.
Pemeriksaan kualitas yang tersedia sudah dijalankan.
Tidak ada perubahan lain yang tidak berkaitan.
Keterbatasan yang diketahui disampaikan secara jujur.
17. Laporan Akhir
Selalu laporkan:
Apa yang diimplementasikan.
File apa saja yang diubah.
Alasan perubahan diperlukan.
Pengujian dan pemeriksaan apa yang dijalankan.
Kegagalan atau pemeriksaan yang tidak dapat dijalankan.
Risiko yang masih ada atau langkah manual yang diperlukan.
Sampaikan secara ringkas dan faktual. Jangan pernah mengklaim perilaku yang belum diverifikasi sebagai sesuatu yang sudah selesai.
