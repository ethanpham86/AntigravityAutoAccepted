# AutoClickAccepted v3.2.0 Hybrid

Bot tự động phát hiện và click nút trên màn hình Windows. **Không phải virus** — đây là công cụ hỗ trợ click tự động, giống AutoHotKey nhưng thông minh hơn vì nhận dạng được hình ảnh và chữ.

## Tính Năng Chính

| # | Tính năng | Mô tả |
|---|----------|-------|
| 🖼️ | **Template Matching** | So khớp pixel 1:1 với ảnh mẫu trong `img/`. Nhanh, chính xác, không cần Tesseract |
| 🔍 | **OCR Fallback** | Nếu template không khớp → dùng Tesseract OCR đọc chữ trên màn hình (optional) |
| 📚 | **Auto-Learn** | Tự học keyword từ ảnh `img/` khi khởi động, merge vào `config.yaml` |
| ⏯️ | **Pause/Resume** | Nhấn **F6** để tạm dừng/tiếp tục, **F7** để dừng hẳn — hoạt động toàn hệ thống |
| 📝 | **Clear Logging** | Mỗi click ghi rõ: click vào **cái gì**, **ở đâu**, **bằng cách nào** |
| 🖥️ | **Portable** | Copy folder sang máy Windows khác là chạy. Config tự resolve theo vị trí exe |
| 🎯 | **DPI Aware** | Tọa độ chính xác trên màn hình High-DPI |
| 🔇 | **Background Click** | Click ẩn bằng `PostMessage` — không cướp chuột vật lý |

## Yêu Cầu Hệ Thống

| Yêu cầu | Bắt buộc? | Chi tiết |
|---------|-----------|---------|
| Windows 10/11 | ✅ Bắt buộc | Dùng Windows API (GDI, User32) |
| Go 1.21+ | ✅ Bắt buộc | Chỉ cần khi build từ source |
| Tesseract OCR 5.0+ | ❌ Tuỳ chọn | Không có → chỉ chạy template matching. Cài từ `dependencies/` |

## Cài Đặt

### Cách 1: Dùng exe có sẵn
```
1. Copy cả folder (gồm autoclick.exe, config.yaml, img/) sang máy đích
2. Chạy autoclick.exe
```

### Cách 2: Build từ source
```bash
go mod tidy
go build -o autoclick.exe ./cmd/main.go
```

### Cách 3: Build GUI (không hiện console)
```bash
go build -ldflags "-H windowsgui" -o autoclick_gui.exe ./cmd/main.go
```

### Cài Tesseract OCR (tuỳ chọn)
```
Chạy: dependencies/tesseract-ocr-w64-setup-5.5.0.20241111.exe
→ Thêm vào PATH hoặc cài vào C:\Program Files\Tesseract-OCR\
```

## Cấu Hình (`config.yaml`)

```yaml
# Từ khoá cần click (không phân biệt hoa thường)
# Ưu tiên từ trên xuống dưới — từ khoá đầu tiên được click trước
keywords:
  - "ACCEPTED"
  - "Allow Once"
  - "Run"
  - "Retry"
  - "OK"

# Thời gian giữa các lần quét (ms). 1500ms = 1.5 giây
scan_interval_ms: 1500

# Ngưỡng tin cậy OCR (0-100). Thấp = bắt nhiều hơn nhưng dễ nhầm
confidence_threshold: 40

# Mức log: debug, info, error
log_level: "info"

# Background click = true → không cướp chuột vật lý
use_background_click: true

# Vùng quét cố định (bỏ qua chọn tay). Format: "x,y,width,height"
# Bỏ comment dòng dưới nếu muốn:
# scan_region: "100,200,800,600"
```

## Cách Sử Dụng

### Bước 1: Chuẩn bị ảnh mẫu (nếu dùng Template Matching)
1. Dùng **Snipping Tool** (Win + Shift + S)
2. Cắt **sát viền** nút cần click (VD: nút "Accept", "OK")
3. Lưu `.png` vào folder `img/`
4. Bot sẽ tự học khi khởi động

### Bước 2: Khởi chạy
```bash
autoclick.exe                           # Mặc định
autoclick.exe -config myconfig.yaml     # Config tuỳ chỉnh
autoclick.exe -debug                    # Debug mode (lưu ảnh capture)
autoclick.exe -interval 500             # Ghi đè scan interval
autoclick.exe -imgdir my_buttons        # Dùng folder ảnh khác
```

### Bước 3: Chọn vùng quét
- Nếu không có `scan_region` trong config → màn hình sẽ mờ đi, kéo chuột khoanh vùng
- Nếu có `scan_region` → bỏ qua, chạy luôn

### Bước 4: Điều khiển khi đang chạy

| Phím | Hành động |
|------|----------|
| **F6** | Tạm dừng / Tiếp tục (Toggle) |
| **F7** | Dừng hẳn (Graceful shutdown) |
| **Ctrl+C** | Dừng hẳn (Graceful shutdown) |

> 💡 **F6/F7 hoạt động toàn hệ thống** — không cần focus vào console window.

## Log & Giám Sát

### Format log mỗi click
```
2026/04/19 09:45:01 [CLICK] ✓ "ACCEPTED" @ (450,320) | OCR conf=95% | BG
2026/04/19 09:45:03 [CLICK] ✓ "Accept Button" @ (120,480) | Template 92% | BG
```

Giải thích:
- `"ACCEPTED"` — **tên nút/keyword** đã click
- `@ (450,320)` — **toạ độ** click trên màn hình
- `OCR conf=95%` hoặc `Template 92%` — phương thức + độ tin cậy
- `BG` = Background click, `Physical` = click vật lý

### Trạng thái Pause/Resume
```
2026/04/19 09:50:00 [INFO]  🔴 PAUSED — Nhấn F6 để tiếp tục
2026/04/19 09:52:00 [INFO]  🟢 RESUMED — Bot đang quét lại
```

### Thống kê khi thoát
```
📊 Final Statistics:
=====================
Total Scans : 342
Total Clicks: 18
Total Errors: 0
--- Clicks by Keyword ---
  ACCEPTED             : 12 clicks
  Allow Once           : 4 clicks
  OK                   : 2 clicks
=====================
```

### File log
- Console: hiện real-time
- File: `autoclick.log` (cùng thư mục exe)

## Cấu Trúc Dự Án

```
AutoClickAccepted/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── capture/             # Chụp màn hình bằng Windows GDI
│   ├── clicker/             # Click chuột (Background + Physical)
│   ├── engine/              # Vòng lặp quét chính + Pause/Resume
│   ├── hotkey/              # Global Hotkey (F6/F7)
│   ├── learner/             # Tự học keyword từ img/
│   ├── logger/              # Log ra console + file
│   ├── matcher/             # Template matching (pixel SAD)
│   ├── ocr/                 # Tesseract OCR wrapper
│   └── selector/            # Chọn vùng quét bằng chuột
├── img/                     # Ảnh mẫu nút cần click
├── dependencies/            # Tesseract installer
├── config.yaml              # Cấu hình chính
├── autoclick.log            # File log
├── Makefile                 # Build commands
├── ARCHITECTURE.md          # Kiến trúc kỹ thuật chi tiết
└── doc/
    ├── ARCHITECTURE.md      # Kiến trúc (tiếng Việt)
    ├── USAGE.md             # Hướng dẫn sử dụng
    └── CHANGELOG.md         # Lịch sử thay đổi
```

## Makefile

```bash
make build       # Build console exe
make build-gui   # Build GUI exe (không hiện console)
make run         # Chạy trực tiếp
make test        # Chạy unit tests
make clean       # Xoá exe
```

## Lưu Ý Quan Trọng

- **Đây không phải virus.** Tool dùng Windows API công khai (GDI, User32, PostMessage) giống như AutoHotKey.
- **Ảnh mẫu phải cắt sát nút.** Nếu screenshot nguyên màn hình → bot sẽ cảnh báo và bỏ qua.
- **Template Matching ưu tiên trước OCR.** Nếu ảnh khớp → click ngay, không chạy Tesseract (tiết kiệm CPU).
- **Chạy với Admin** nếu cần click vào ứng dụng có quyền đặc biệt (UAC dialogs).
- **Background Click** có thể không hoạt động với một số app (VD: game dùng DirectX). Khi đó bot tự fallback sang Physical Click.
