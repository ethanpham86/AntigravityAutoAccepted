# Cấu Trúc Kiến Trúc (Architecture)

## Tổng Quan (Overview)
`AutoClickAccepted` được xây dựng hoàn toàn bằng Go (Zero-CGO dependency cho Windows API) theo cấu trúc thư mục tiêu chuẩn, đảm bảo tính clean-code, tách biệt rạch ròi giữa các Module chức năng.

Dự án áp dụng mô hình **Event-Loop Pipeline**: `[Capture] -> [OCR] -> [Cognitive/Learner] -> [Sweep Match] -> [Syscall Clicker]`.

---

## Các Module Cốt Lõi (Core Modules)

### 1. `internal/capture` (Screen Capture)
Sử dụng `user32.dll` và `gdi32.dll` (GDI) để chụp ảnh màn hình tĩnh tốc độ siêu cao (~10ms/frame).  
Biến đổi không gian hình học đa màn hình xử lý luôn các toạ độ âm (Negative Coordinates). Cuối cùng upscale hình ảnh bằng `image/draw` để tăng độ nét báo cáo cho OCR.

### 2. `internal/ocr` (Optical Character Recognition)
Kết nối gián tiếp (os/exec) sang tiến trình CLI của `tesseract.exe`.
- Cấu hình `--psm 11` (Sparse Text) ép engine lôi ra tối đa các chữ cái trong khung UI.
- Thắt cổ chai Memory Leak của Tesseract C++ Engine bằng tuỳ biến `-c debug_file=NUL` cản luồng rác stdout Console.

### 3. `internal/learner` (Auto-Learner)
Mô đun tự học thông minh: tự quét các hình ảnh PNG trong folder `img/` khi khởi động, thu thập các chữ cái thông qua cơ chế OCR và bổ sung động vào cấu trúc `Keywords`.
*Luật sinh tồn:* Nó đánh giá file ảnh rác bằng len(words) > 10. Chống người dùng đẩy Full Screenshot vào.

### 4. `internal/engine` (The Brain)
Loop Timer quản lý toàn bộ nhịp điệu Quét (Scan) và Nhấn (Click). 
- **Sweep Click (Multi-Targeting):** Phát hiện 1 list tất cả các Hit. 
- **Toạ độ toạ trúng (Dedup):** `dx*dx + dy*dy < 400` giới hạn bán kính quét trùng pixel.
- Gửi các gói toạ độ tuyệt đối (Absolute Screen Coords) xuống tầng Syscall. Xử lý triệt để false-positive qua thuật toán Levenshtein Distance & Word Boundary.

### 5. `internal/clicker` (Syscall Injector)
Giao tiếp tầng thấp với Hệ điều hành Windows `SetCursorPos` và `SendInput`.
Mô phỏng MOUSEEVENTF_LEFTDOWN | MOUSEEVENTF_LEFTUP 1 cách trung thực nhất. Vượt 100% UIPI Restriction của Windows nếu User chạy bằng đặc quyền Admin.

### 6. `internal/logger` (Observability)
Mô hình `io.MultiWriter` xuất đồng thời lên:
- `os.Stdout` (Màn hình Console cho dev xem Real-time)
- `autoclick.log` (File vật lý lưu vết log, phục vụ thanh tra bằng ELK/Kibana).

---
## Pipeline Tương Tác
```mermaid
graph TD
    A[Main Loop Timer] --> B[Capture Screen Region GDI]
    B --> C[Tesseract OCR Processing]
    C --> D[Identify Words + Confidence]
    D --> E[Engine: Exact / Fuzzy Match]
    E --> F[Engine: Sweep Dedup 20px]
    F --> G[Syscall: SendInput (Click)]
    G --> H[Wait 500ms Delay]
    H --> F
```
