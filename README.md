## Tugas ke-4 Sistem Terdistribusi

Tugas ini bertujuan mengimplementasikan message queue dan gRPC ke dalam bentuk aplikasi chat sederhana. Aplikasi ini terdiri dari beberapa services sebagai berikut:

| Service             | Tech Stack | Node            | Replikasi |
| ------------------- | ---------- | --------------- | --------- |
| Message queue       | RabbitMQ   | Manager         | 1         |
| Website             | React Vite | Manager, worker | 2         |
| API                 | Golang     | Manager, worker | 2         |
| Async Worker        | Golang     | Manager, worker | 2         |
| Notification Server | Golang     | Manager, worker | 2         |
| Database            | Postgres   | Manager         | 1         |

Cara kerja aplikasi:
