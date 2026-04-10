# Changelog

All notable changes to this project will be documented in this file.

## [v2.5] - 2026-04-10
### Added
- Thêm tính năng **Sweep Click**: Bot chuyển từ chế độ "Click 1 Hit" thành cỗ máy quét cạn "Click All Matches". Triệt phá 1 lúc nhiều nút bấm liền kề.
- Áp dụng **Toạ độ cản trùng (Coordinates Deduplication)**: Bán kính 20px xung quanh Click.
- Gắn hệ thống File Physical Logging ra `autoclick.log` chuyên nghiệp phục vụ môi trường Production (`internal/logger`).
- **Auto-Learner Guard**: Còi báo động khi user sử dụng Full Screenshot làm rác bộ nhớ Keywords thay cho ảnh cropper.

### Fixed
- Vá lỗi bắt toạ độ ma UIPI / Scaling bằng cách thiết kế hệ quy chiếu điểm mù tuyệt đối trong `engine.go`.
- Vá rác Console: Khống chế hoàn toàn luồng Warning Memory Leak (ObjectCache) của thư viện C++ Leptonica (Tesseract) thông qua cờ `debug_file=NUL`.
- Xoá triệt để cảnh báo `context.Canceled` khi thoát trình Graceful Shutdown `Ctrl+C`.

---

## [v2.1] - 2026-04-09
### Changed
- Gỡ bỏ CGO. Sử dụng 100% Thuần Windows Native Package `syscall` & `unsafe` để trỏ vào `user32` + `gdi32`.
- Keyword Matching Algorithm: Chuyển sang Levenshtein Distance Threshold với Exact Substring Rule. Bỏ hẳn kiểu bắt chữ hời hợt `strings.Contains` gây "Cướp chuột loạn xạ".

## [v1.0] - Pilot
### Added
- Tính năng Selector Image.
- Vòng lặp Loop Scan đơn lẻ.
- Initial codebase for `AutoClickAccepted`.
