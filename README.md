vps-manager/
├── cmd/
│   └── server/
│       └── main.go                 # Điểm vào chính của ứng dụng
├── internal/
│   ├── api/
│   │   ├── handlers.go             # HTTP handlers cho REST API
│   │   └── routes.go               # Định nghĩa routes
│   ├── docker/
│   │   ├── client.go               # Docker client wrapper
│   │   ├── container.go            # Quản lý container
│   │   └── volume.go               # Quản lý volumes
│   ├── models/
│   │   └── vps.go                  # Data models
│   └── config/
│       └── config.go               # Cấu hình ứng dụng
├── scripts/
│   └── init.sh                     # Script khởi tạo
├── data/                           # Thư mục lưu trữ volumes
├── go.mod
├── go.sum
├── Dockerfile                      # Dockerfile cho ứng dụng manager
├── docker-compose.yml              # Docker Compose để chạy toàn bộ
└── README.md
