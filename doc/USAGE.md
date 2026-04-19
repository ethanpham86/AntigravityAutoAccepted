# Hướng Dẫn Sử Dụng — AutoClickAccepted v3.2.0

## ⚡ Quick Start (3 Bước)

1. Copy folder (gồm `autoclick.exe`, `config.yaml`, `img/`) sang máy đích
2. Chỉnh `config.yaml` nếu cần (keywords, scan_region)
3. Chạy `autoclick.exe`

---

## ⚙️ Cấu Hình (`config.yaml`)

File `config.yaml` nằm cạnh exe. Bot tự tìm đúng vị trí khi chạy.

```yaml
# Từ khoá cần click (không phân biệt hoa thường)
# Ưu tiên từ TRÊN xuống DƯỚI — từ khoá đầu tiên click trước
keywords:
  - "ACCEPTED"
  - "Allow Once"
  - "Run"
  - "Retry"
  - "OK"

# Thời gian giữa các lần quét (ms). 1500ms = 1.5 giây
# Nhỏ hơn = nhanh hơn nhưng tốn CPU
scan_interval_ms: 1500

# Ngưỡng tin cậy OCR (0-100)
# Thấp = bắt nhiều hơn nhưng dễ nhầm. 40 là cân bằng.
confidence_threshold: 40

# Mức log: debug, info, error
log_level: "info"

# Background click = true → không cướp chuột vật lý (dùng PostMessage)
# = false → click vật lý (di chuyển chuột)
use_background_click: true

# Vùng quét cố định (bỏ qua chọn tay khi khởi động)
# Format: "x,y,width,height" (pixel)
# Bỏ comment nếu muốn dùng:
# scan_region: "100,200,800,600"
```

---

## 📸 Chuẩn Bị Ảnh Mẫu (Auto-Learn)

Bot tự học nút cần click từ ảnh trong folder `img/`:

### Cách làm:
1. Mở ứng dụng có nút cần click
2. Dùng **Snipping Tool** (`Win + Shift + S`)
3. **CẮT SÁT VIỀN** nút bấm (VD: nút "Accept", "OK", "Allow Once")
4. Lưu `.png` vào folder `img/`
5. Lần chạy kế tiếp → bot tự nạp

### Quy tắc:
| Đúng ✅ | Sai ❌ |
|---------|--------|
| Cắt sát nút bấm (≤ 200x200 pixel) | Screenshot nguyên màn hình |
| Nền nút rõ ràng | Ảnh mờ, nhòe |
| File `.png` hoặc `.jpg` | File `.gif`, `.svg` |

> ⚠️ Nếu ảnh quá lớn (> 200x200), bot sẽ cảnh báo trong log và chỉ dùng OCR, không dùng template matching.

---

## 🖥️ Khởi Chạy

### Chạy cơ bản
```bash
autoclick.exe
```

### Chạy với options
```bash
autoclick.exe -config myconfig.yaml     # Config tuỳ chỉnh
autoclick.exe -debug                    # Debug mode (lưu ảnh capture vào debug/)
autoclick.exe -interval 500             # Ghi đè scan interval (500ms)
autoclick.exe -imgdir my_buttons        # Dùng folder ảnh khác
```

### Build từ source
```bash
make build       # Console mode
make build-gui   # GUI mode (không hiện console)
make run         # Chạy trực tiếp
make test        # Unit tests
```

---

## 🔲 Chọn Vùng Quét

### Cách 1: Chọn bằng chuột (mặc định)
- Màn hình sẽ mờ đi (opacity 30%)
- **Kéo chuột** khoanh vùng khu vực hay xuất hiện nút
- Khoanh **càng hẹp** → bot chạy **càng nhanh** (ít pixel cần quét hơn)

### Cách 2: Cấu hình sẵn trong config.yaml
```yaml
scan_region: "100,200,800,600"
# x=100, y=200, width=800px, height=600px
```

> 💡 Khi chọn vùng bằng chuột xong, bot sẽ gợi ý câu lệnh để thêm vào config:
> ```
> scan_region: "100,200,800,600"
> ```

---

## ⏯️ Điều Khiển Khi Đang Chạy

| Phím | Hành động | Ghi chú |
|------|----------|---------|
| **F6** | ⏯️ Tạm dừng / Tiếp tục | Hoạt động toàn hệ thống — không cần focus console |
| **F7** | 🛑 Dừng hẳn | Graceful shutdown, in thống kê |
| **Ctrl+C** | 🛑 Dừng hẳn | Graceful shutdown, in thống kê |

### Trạng thái trong log:
```
2026/04/19 09:50:00 [INFO]  🔴 PAUSED — Nhấn F6 để tiếp tục
2026/04/19 09:52:00 [INFO]  🟢 RESUMED — Bot đang quét lại
```

---

## 📝 Log & Giám Sát

### File log
- **Console:** Hiện real-time (nếu có console window)
- **File:** `autoclick.log` — nằm cạnh exe, append mỗi lần chạy

### Format mỗi click
```
2026/04/19 09:45:01 [CLICK] ✓ "ACCEPTED" @ (450,320) | OCR conf=95% | BG
2026/04/19 09:45:03 [CLICK] ✓ "Accept Button" @ (120,480) | Template 92% | BG
```

| Phần | Ý nghĩa |
|------|---------|
| `"ACCEPTED"` | **Tên keyword** đã click |
| `@ (450,320)` | **Toạ độ** trên màn hình |
| `OCR conf=95%` | Phương thức OCR, độ tin cậy 95% |
| `Template 92%` | Phương thức template match, 92% pixel giống |
| `BG` | Background click (không cướp chuột) |
| `Physical` | Click vật lý (di chuột) |

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

### Grep log nhanh
```bash
# Xem tất cả click events:
findstr "[CLICK]" autoclick.log

# Xem chỉ "ACCEPTED":
findstr "ACCEPTED" autoclick.log

# Xem errors:
findstr "[ERROR]" autoclick.log
```

---

## 🚀 Mang Sang Máy Khác (Portability)

### Cần copy:
```
autoclick.exe          # Chương trình chính
config.yaml            # Cấu hình
img/                   # Ảnh mẫu nút bấm
```

### Không cần:
- Go compiler (đã compile sẵn)
- Tesseract (optional — nếu không có thì chạy template-matching-only)

### Cài Tesseract (tuỳ chọn):
```
Chạy: dependencies/tesseract-ocr-w64-setup-5.5.0.20241111.exe
Thêm vào PATH hoặc cài vào: C:\Program Files\Tesseract-OCR\
```

---

## ❓ FAQ / Troubleshooting

| Vấn đề | Giải pháp |
|--------|----------|
| Bot không click được | Kiểm tra vùng quét đã đúng chưa. Thử `-debug` để xem capture |
| "Tesseract not found" | Chỉ warning, bot vẫn chạy bằng template matching |
| Click không trúng | DPI scaling? Chạy lại, bot tự `SetDPIAware` |
| F6/F7 không hoạt động | Có instance khác đang chiếm hotkey? Kill process cũ |
| Background click không tác dụng | Một số app (game) chặn PostMessage → bot tự fallback Physical |
| Ảnh mẫu không match | Cắt sát hơn, đúng resolution, ≤ 200x200 pixel |
