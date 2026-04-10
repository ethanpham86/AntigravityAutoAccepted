# Sổ Tay Lệnh Git Bỏ Túi Đầy Đủ 🚀

*(Tài liệu này là đồ dùng cá nhân nội bộ, đã được cách ly thành công qua phễu `.gitignore` - an toàn 100% không bao giờ lo lỡ tay Push lên Public)*

---

## 1. Khởi tạo kho chứa mới (Tạo Repository từ số 0)
Khi bạn vừa code xong 1 dự án trắng tinh và muốn đưa toàn bộ mã nguồn đẩy lên Github:

```bash
# Biến thư mục hiện tại thành kho chứa Git
git init

# Gom tất cả file vào giỏ hàng chuẩn bị thanh toán
git add .

# Đóng gói và gắn tên cho giỏ hàng
git commit -m "Khởi tạo mã nguồn ban đầu"

# Khai báo tên Cành chính bắt buộc là 'main'
git branch -M main

# Chỉ đường link tới cái kho trống trơn bạn vừa thiết lập trên Github
git remote add origin https://github.com/TênNick/TênDựÁn.git

# Đẩy cục nợ lên mạng (Chỉ lần đẩy đầu tiên mới cần cờ '-u')
git push -u origin main
```

---

## 2. Tải Code từ mạng về máy (Clone & Pull)
Nếu bạn chuyển qua máy tính ở công ty hoặc lập trình trên laptop mới, bạn lấy code về kiểu gì?

```bash
# 1. CLONE (Chỉ dùng lấy 1 lần duy nhất để tải toàn bộ Repo + Thư mục trên mạng về máy)
git clone https://github.com/TênNick/TênDựÁn.git 

# 2. PULL (Cập nhật và KÉO Code mới)
# Dùng rát nhiều hàng ngày. Ví dụ tối bạn làm ở nhà Push lên, sáng lên Công ty chưa thấy code mới cập nhật, bạn gõ lệnh này để Git nó Tải bù khoảng cách.
git pull origin main
```

---

## 3. Quy trình bảo lưu Code hàng ngày (Add, Commit, Push)
Mỗi ngày khi bạn Code xong 1 chức năng mới (hoặc fix xong bug xịn), hãy gõ thứ tự 3 lệnh thần thánh này để bảo vệ Code ngay lập tức:

```bash
# B1: Gom hết thay đổi
git add .
    # Mẹo: Nếu chỉ mún gom đúng 1 Cụm file thì gõ: git add config.yaml main.go

# B2: Đóng gói và ghi chú lại cho nhớ
git commit -m "Tui vừa làm xong hệ thống AutoClick v2"

# B3: Quăng lên kho trên GitHub để lưu Trữ!
git push origin main
```

---

## 4. Phân thân Tách Nhánh (Branching)
Tại sao phải chia nhánh ngọn cành? Khi dự án đang chạy rất ổn định (trên kênh `main`), đột nhiên sếp bắt bạn làm chức năng mới nhưng bạn SỢ RỦI RO code hỏng, không cho phần mềm gốc chạy được nữa. Lúc này Nhánh là chân ái!

```bash
# 1. Xem bạn đang đứng ở Cành (Branch) nào:
git branch

# 2. Tách bộ code hiện tại ra 1 cành MỚI để quậy phá riêng (vd tên là dev-click)
git checkout -b dev-click

# 3. [TUỆT VỜI] Sau khi bạn gõ lệnh trên, bạn đứng ở môi trường riêng dev-click. Bạn tha hồ Đổi Code, Phá Code. Cành 'main' ban đầu sẽ KHÔNG HỀ HẤN GÌ. 

# 4. Khi phá code xong thấy Ngon, bạn có thể chạy Push nhánh mới này lên GitHub (thay vì Push main)
git push origin dev-click

# 5. [Đổi qua đổi lại] - Để nhảy lại bộ code cũ bên nhánh main (để coi lại code hồi chiều), gõ cực lẹ:
git checkout main
```

---

## 5. Trộn Nhánh Code (Merge)
Sau khi quậy phá trên nhánh `dev-click` thành công rực rỡ, tính năng ăn trọn tiền. Bạn muốn Trộn toàn bộ cái hay của nó Bỏ lọt vào lại nhánh `main` để Live!

```bash
# B1: Bắt buộc bạn phải bay về đứng ở căn cứ đích (tức là nhánh main)
git checkout main

# B2: Quát Git kéo râu của thằng `dev-click` hoà trộn vào tụi 'main' đi!
git merge dev-click

# B3: Đẩy lại nhánh main rực rỡ lên Github!
git push origin main
```

---

## 6. Tuyệt Chiêu Xử Lý Sự Cố (Troubleshooting)

### Sự Cố A: Lỡ quăng file nhạy cảm nên mún phi tang xoá khỏi Git?
```bash
# Lệnh gỡ bám đuôi khỏi Git nhưng VẪN GIỮ LẠI FILE TRONG MÁY TÍNH
git rm --cached ten-file-lo-mat-khau.txt

# Xong nhớ ném tên file đó vào góc `.gitignore` để bịt miệng thằng Git báo đòi đưa nó lên mạng lần nữa.
```

### Sự Cố B: Lỗi 403 / 401 bị chặn Permission (Chứng nhận tài khoản lỗi)
Windows Credential chập mạch tự nạp Account cũ? Ta dùng phép thuật bắt Git bật lại ô cửa Login:
```bash
# Cấu hình lại đầu ống nối Remote URL (Nối cứng Tên Nick vào luôn)
git remote set-url origin https://ethanpham86@github.com/ethanpham86/AntigravityAutoAccepted.git

# Ép lôi bảng Popup bắt Trình Duyệt hiện ra xác thực
$env:GCM_INTERACTIVE="Always"  # (Mã trên PowerShell ngầm)
git push origin main
```
