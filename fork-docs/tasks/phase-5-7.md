# Phase 5–7: App Store riêng, đồng bộ upstream, tính năng riêng

---

## T5.1: Mirror App Store

**Phụ thuộc:** T1.3, T1.4. *(Chủ dự án fork `1Panel-dev/appstore` → `xpanel-appstore`.)*

**Việc cần làm**
1. Tìm hiểu cách panel đọc App Store: từ `AppRepoURL()` tải những file nào (danh sách app, `data.yaml`, icon, gói từng phiên bản `.tar.gz`), định dạng ra sao. Đọc code trong `agent/app/service/app*.go` và thư mục `scripts/` của repo appstore. Ghi vào `fork-docs/APPSTORE.md`.
2. Workflow `xp-appstore-sync.yml` (trong repo `xpanel-appstore`): chạy hằng tuần + thủ công; build dữ liệu App Store từ repo (theo đúng cách upstream build, nếu không có script công khai thì sao chép từ `apps.1panel.pro` theo danh sách app, kiểm tra checksum) rồi đẩy lên R2 `apps.<domain>`.
3. Installer (T1.4) ghi `AppRepoURL` = `https://apps.<domain>`.

**Nghiệm thu:** [ ] Máy mới cài từ installer của ta: App Store hiện đủ ứng dụng, cài được OpenResty, MySQL, WordPress. [ ] Tắt truy cập `apps.1panel.pro` (chặn bằng `/etc/hosts`) vẫn cài được.

---

## T5.2: Mô tả ứng dụng tiếng Việt và ứng dụng riêng

**Phụ thuộc:** T5.1, T3.2.

**Việc cần làm:** thêm mô tả `vi` cho 50 ứng dụng phổ biến nhất (theo định dạng đa ngôn ngữ mà `data.yaml` của mỗi app đang dùng); quy trình thêm ứng dụng mới của riêng ta (thư mục, `data.yaml`, compose, icon) viết thành `fork-docs/APPSTORE.md` mục "Thêm ứng dụng".

**Nghiệm thu:** [ ] Panel tiếng Việt hiện mô tả tiếng Việt cho 50 app.

---

## T6.1: Quy trình đồng bộ upstream

**Phụ thuộc:** T1.5, T2.1.

**File:** `scripts/sync/sync-upstream.sh`, `fork-docs/SYNC.md`, `.github/workflows/xp-upstream-watch.yml`.

**Đặc tả**
- `xp-upstream-watch.yml`: chạy hằng ngày, khi upstream có tag `v2.*` mới (không phải `-rc`/`-beta`) → mở issue "Đồng bộ upstream vX.Y.Z" trong repo fork kèm link changelog.
- `sync-upstream.sh <tag>`: `git fetch upstream --tags` → tạo nhánh `sync/<tag>` từ `main` → `git merge <tag>` → nếu xung đột thì dừng, in danh sách file xung đột kèm mục tương ứng trong PATCHES.md → nếu sạch: `make build_all`, `apply.sh --yes` trên bản sao tạm + `--check`, `check-keys.mjs`, chạy test.
- SYNC.md: các bước thủ công còn lại (dịch key mới, test VM theo T6.2, tag `v<tag>-xp.1`, phát hành).
- Làm thử một lần với tag upstream mới nhất có sau v2.5.1.

**Nghiệm thu:** [ ] Đồng bộ thử thành công, ghi lại thời gian và số xung đột vào SYNC.md.

---

## T6.2: Test tự động trên VM

**Phụ thuộc:** T1.4.

**File:** `scripts/test/smoke.sh`, `fork-docs/TEST-MATRIX.md`.

**Đặc tả:** chạy trên VM sạch (root), dùng installer với biến môi trường để không hỏi (kiểm tra các biến mà `install.sh` hỗ trợ cho chế độ không tương tác; nếu không có thì dùng `expect` hoặc ghi vào QUESTIONS.md). Các bước, mỗi bước in PASS/FAIL: cài → lấy thông tin đăng nhập qua `1pctl user-info` → gọi API đăng nhập (đọc tài liệu Swagger trong `core/cmd/server/docs/` để lấy đúng endpoint; hoặc dùng API key nếu `1pctl` tạo được) → cài OpenResty + MySQL qua API App Store → tạo website PHP → `curl` website → tạo DB → backup website → kiểm tra ngôn ngữ `vi` trả về thông báo tiếng Việt. Exit 1 nếu có bước fail.

**Nghiệm thu:** [ ] Pass trên Ubuntu 22.04, Debian 12, Rocky 9; kết quả ghi TEST-MATRIX.md.

---

## T6.3: Điều tra KVM helper

**Phụ thuộc:** T0.2.

**Việc cần làm:** tìm nguồn gốc `1panel-kvm-*` (repo, giấy phép, mã nguồn). Viết kết luận vào `fork-docs/KVM.md` với một trong hai đề xuất: (A) có mã nguồn GPL → build từ nguồn trong CI; (B) không có → build bản fork **không kèm** binary này và ẩn/tắt tính năng VM (tìm cờ build hoặc điểm kiểm tra sẵn có; mô tả patch cần làm). Đổi trạng thái `[?]` chờ chủ dự án chọn.

---

## T6.4: Hướng dẫn chuyển từ 1Panel gốc sang XPanel

**Phụ thuộc:** T1.5.

**Việc cần làm:** kiểm tra máy đang chạy 1Panel bản gốc cùng phiên bản có chuyển sang XPanel được không bằng cách: đổi 3 URL cấu hình (T1.3) + thay binary bằng bản của ta. Viết trang tài liệu "Chuyển từ 1Panel sang XPanel" (tiếng Việt) và lệnh `1pctl`/script hỗ trợ nếu cần (`scripts/release/migrate-from-1panel.sh`, đẩy lên `dl.<domain>/v2/migrate.sh`).

**Nghiệm thu:** [ ] VM chạy 1Panel v2.5.1 gốc chuyển sang `v2.5.1-xp.N`, website và DB còn nguyên, nâng cấp tiếp qua nút được.

---

## Phase 7: Tính năng riêng (chỉ làm khi chủ dự án yêu cầu)

Mỗi tính năng: worker viết thiết kế `fork-docs/design/<tên>.md` (mục tiêu, file upstream phải sửa, ước lượng diff, cách giữ diff nhỏ), chờ duyệt. **Ưu tiên làm thành module/ứng dụng tách rời** (ứng dụng App Store, plugin, hoặc trang riêng gắn vào menu qua một điểm nối duy nhất) thay vì sửa sâu vào code upstream. Ý tưởng ứng viên:

| Mã | Ý tưởng |
|---|---|
| T7.1 | DNS API cho nhà cung cấp tên miền Việt Nam (kiểm tra danh sách provider của thư viện ACME mà 1Panel dùng trước; thêm provider = đóng góp ngược lên thư viện đó) |
| T7.2 | Thông báo qua Telegram/Zalo OA (kiểm tra kênh cảnh báo sẵn có của 1Panel trước) |
| T7.3 | Mẫu website dựng sẵn phổ biến ở VN (WordPress + plugin tiếng Việt, WooCommerce cấu hình VND) dạng ứng dụng App Store |
| T7.4 | Mirror Docker Hub / tài nguyên đặt tại VN để cài nhanh hơn |
