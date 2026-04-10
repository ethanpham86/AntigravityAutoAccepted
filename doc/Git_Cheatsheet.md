# Bách Khoa Toàn Thư Git Tương Tác Code (Thực Chiến Dành Cho Dev) 🚀

Cuốn cẩm nang này được thiết kế theo đúng quy trình **Bước-theo-Bước (Step-by-step)** từ lúc bạn bật máy lên Code, sửa Code, cho tới lúc quăng Code lên server GitHub. Ở đây có tất cả mọi lệnh xử lý từ cơ bản đến cấp cứu mã nguồn.

---

## KHÚC 1: KHỞI ĐỘNG VÀ BẮT ĐẦU DỰ ÁN 🌅
*(Chỉ dùng khi chuẩn bị nhận 1 task mới hoặc khởi động dự án mới)*

**1. Nạp dự án có sẵn từ mạng về máy tính để làm (Clone)**
```bash
git clone https://github.com/TênNick/TênDựÁn.git
```
**2. Tạo dự án mới toanh từ số 0 (Nếu bạn tự viết Code)**
```bash
git init
git branch -M main
git remote add origin https://github.com/TênNick/TênDựÁn.git
```
**3. Kéo Code đồng bộ mới nhất (Đầu giờ sáng mở máy lên)**
```bash
# Trước khi gõ code, LUÔN LUÔN phải cập nhật xem tụi đồng nghiệp có đẩy gì mới lên chưa
git pull origin main
```

---

## KHÚC 2: RẼ NHÁNH CODE AN TOÀN (GIAI ĐOẠN ĐANG CODE) 🌿
*(Luật ngầm: TUYỆT ĐỐI không Code trực tiếp trên nhánh `main`. Nhánh `main` là thánh địa của dự án sống. Bạn phải rẽ nhánh để chế cháo riêng).*

**1. Tạo không gian nháp mới (Tạo Branch)**
```bash
# Lệnh này vừa Mở 1 nhánh mới tên là 'tinh-nang-a' VÀ Tự động bế thân xác bạn nhảy chui vào nhánh đó.
git checkout -b tinh-nang-a
```
**2. Kiểm tra xem mình đang đứng ở nhánh nào? (Đề phòng code nhầm địa chỉ)**
```bash
git branch
# Kết quả sẽ in ra dấu (*) xanh lá cây ngay tại nhánh mặt tiền bạn đang đứng.
```

---

## KHÚC 3: KIỂM TRA ĐỊA HÌNH KHI ĐANG VIẾT CODE 🔍
*(Sếp yêu cầu kiểm tra kỹ xem file nào bị sửa, có code gì lỗi lòi ra không trước khi quăng lên kho)*

**1. Xem tình trạng Toàn Trận (File nào sửa, file nào bị xoá)**
```bash
git status
```
**2. [SIÊU QUAN TRỌNG] Xem rốt cuộc mình vừa gõ cái dòng gì trong file?**
```bash
# Xem sự khác biệt giữa code cũ và code bạn vừa sửa TRƯỚC KHI ấn Commit.
git diff
```
**3. Coi lịch sử gia phả (Xem dạo này thằng nào vứt code gì lên)**
```bash
# In ra danh sách log ngắn gọn nhất
git log -n 5 --oneline
```

---

## KHÚC 4: ĐÓNG GÓI CHỐT KẾ TOÁN (ADD & COMMIT) 📦
*(Thực hiện mỗi chiều khi Task xong, hoặc lúc chuẩn bị nghỉ trưa mún khoá Code)*

**1. Quăng tất cả các file vừa sửa vào rổ**
```bash
git add .
    # Hoặc làm màu chỉ nhặt 1 file: git add con-bug.go
```
**2. Bấm nút Chốt Đơn (Gắn băng rôn mô tả)**
```bash
git commit -m "Tính năng: Tôi vừa gắn xong bộ Auto-Learner càn quét chữ"
```
**3. Ném toàn bộ thành quả mẻ hàng lên Github!**
```bash
# [Nếu bạn đang sửa trên nhánh phụ tinh-nang-a thì thay chữ main bằng tinh-nang-a]
git push origin tinh-nang-a
```

---

## KHÚC 5: TRỘN CODE VÀ GỘP NHÁNH THÀNH CÔNG (MERGE KẾT QUẢ) ⚔️
*(Khi mọi nhánh phụ bạn quậy đã ngon, sếp bảo "Gộp tết vào main cho anh")*

**1. Nhảy ngược về nhánh cốt lõi (main)**
```bash
git checkout main
# Cập nhật code mới nhất từ thiên hạ kẻo đá nhau
git pull origin main  
```
**2. Bê toàn bộ râu ria cụm tính năng ở cành kia dán vào main**
```bash
# Trộn code nhánh tinh-nang-a đè thẳng vào nơi bạn rẽ
git merge tinh-nang-a
```
*(Nếu thành công tuyệt đối, gõ)*
```bash
git push origin main
```
> ⚠️ **NẾU CÓ CONFLICT (Xung Đột)**: Sẽ có dòng chữ đỏ `Merge conflict in...`. Lúc này mở VSCode lên, nó sẽ bôi đỏ 2 mảng code bị đụng nhau. Gọt xoá bớt code lỗi -> lưu file lại -> Nhập 3 lệnh Khúc 4 `add` -> `commit` -> `push` là qua ải sinh tử!

---

## KHÚC 6: BIỆT ĐỘI CẤP CỨU & HOÀN TÁC (REVERT / STASH / RESET) 🚑
*(Đời không như là mơ, đôi khi bạn Xóa nhầm file hoặc push nhầm mã độc)*

**1. Chặn lại sự dang dở (Cất tạm Code đi)**
Đang code nửa chừng thì sếp réo vào họp, bắt nhảy sang nhánh khác fix gấp bug? Không thể Commit vì mã chưa chín?
```bash
# Cất gọn mớ hỗn độn này vào ngăn kéo ngầm (như nhét quần áo bẩn vào tủ để dọn phòng)
git stash

# Bây giờ phòng sạch rồi, bạn tha hồ chuyển nhánh: git checkout nhánh-cứu-hộ

# Khi cấp cứu mã xong, về lại nhánh gốc, muốn LUỘC lại cục dang dở kia ra code tiếp:
git stash pop
```

**2. Vứt hết những gì vừa gõ hôm nay (Undo tất cả)**
Chưa kịp Commit mà gõ code rác chán quá muốn vứt hết làm lại từ đầu:
```bash
git restore . 
# Code sẽ Tự Động thụt lùi y chang trạng thái của lần Commit cuối cùng gần nhất
```

**3. Khôi phục 1 mảng Commit (Hối hận vì lỡ gõ `git commit`)**
Lỡ tay ấn Enter chốt nạp mạng 1 lần nhưng quên chưa bỏ 1 file Config vào?
```bash
# Giết bỏ lần Commit vừa xong, lôi đống file về màn hình chờ để bạn nhặt thêm đồ
git reset --soft HEAD~1
```

**4. Hành hình đứt đoạn 1 Đống Rác (Rút lại Mã đã bị Quăng lên Github)**
Sếp chửi đống code mà bạn TỰ HÀO ném lên nhánh `main` hôm qua là rác, buộc xoá sạch không dấu vết:
```bash
# Sinh ra phản lực Nghịch đảo: Tự sinh ra 1 lần Commit Cụ Lội Ngược Dòng đâm lủng cái code cũ xoá đi!
git revert <Mã-Hash-Của-Đống-Commit-Lỗi>
# (Lấy Mã hash 7 chữ số bằng lệnh: git log --oneline)
```

**5. Lỡ public file nhạy cảm và muốn xoá nó khỏi Github ngay**
```bash
# Xoá khỏi sự kìm kẹp Tracking Git nhưng GIỮ LẠI cho Local máy tính
git rm --cached ten-file-mat-khau-bank.json
# Nhét nó vào file .gitignore liền trước khi Commit
```
