# Hướng dẫn Sử Dụng (Usage Manual)

## ⚡ Setup Nhanh (Quick Start)
1. Cài đặt [Tesseract OCR](https://github.com/UB-Mannheim/tesseract/wiki).
2. Tải mã nguồn và biên dịch:
   ```bash
   go build -o bin/autoclick.exe ./cmd/main.go
   ```
3. Chạy file `bin/autoclick.exe` (Run As Administrator nếu tương tác với app đòi quyền system).

---

## ⚙️ Cấu Hình (config.yaml)
Tool sẽ đọc file `config.yaml` đặt cùng thư mục để vận hành:

```yaml
# Thời gian nghỉ ngủ giữa các lần quét càn màn hình (Đơn vị: milliseconds)
scan_interval_ms: 1500

# Ngưỡng tin cậy của Tesseract (0 - 100). Đặt 15-30 nếu UI là Dark Mode mờ.
confidence_threshold: 40

# Mức độ in Log: debug, info, error
log_level: "info"

# Từ khoá Nút Bấm ưu tiên từ trên xuống dưới. Nút nào ưu tiên phải thả lên TOP.
keywords:
  - "ACCEPTED"
  - "Allow Once"
  - "Run"
  - "Retry"
```

## 📸 Khoanh vùng quét (Region Selector)
Khi khởi chạy, màn hình sẽ mờ đi. Bạn hãy xài chuột **Kéo & Khoanh Vùng** vào khu vực có khả năng xuất hiện thông báo / Nút bấm nhất. 
> Việc khoanh vùng càng Hẹp sẽ giúp Tool tiết kiệm CPU và tăng cấp số nhân (x10) tốc độ phân tích chữ!

## 🧠 Tính năng Tự Học (Auto-Learner)
Mệt mỏi với việc nhập chữ thủ công vào `config.yaml`?
1. Sử dụng Snipping Tool của Windows (Ctrl + Shift + S).
2. **CẮT SÁT RẠT** nội dung của cái Nút đó (VD: nút `Ok`, `Deploy`).
3. Lưu tấm ảnh đó thành `*.png` quăng thẳng vào thư mục `img/`.
4. Lần chạy tiếp theo, Auto-Learner sẽ đọc và nạp não tự động chữ cái đó vào danh sách săn mồi!
*(Lưu ý chặn hệ thống: Tool sẽ réo còi Báo Động vào Log nếu bạn quăng nguyên cả màn hình bự chảng vào làm nhiễu hệ thống)*.

## 🛑 Cách Thoát Khỏi Ma Trận
Tại môi trường Console (Terminal), bấm phím `Ctrl + C`. Tool sẽ thông báo Graceful Shutdown và in Cập nhật Thống Kê toàn phiên `Session Summary` cực kì mượt mà. Đừng hoảng sát tắt nút `X`, hãy cho nó cơ hội in báo cáo!
