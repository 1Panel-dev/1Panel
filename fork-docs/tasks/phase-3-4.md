# Phase 3–4: Tiếng Việt và tài liệu

## Quy ước dịch (áp dụng cho mọi task phase này)
- Tạo `fork-docs/GLOSSARY-vi.md` ở T3.1 **trước khi** dịch, rồi dùng thống nhất. Gợi ý: website = website, domain = tên miền, container = container, image = image, compose = compose, snapshot = bản chụp, cron job = tác vụ định kỳ, firewall = tường lửa, reverse proxy = reverse proxy, runtime = môi trường chạy, node = node, panel = panel, app store = cửa hàng ứng dụng, backup = sao lưu, restore = khôi phục, certificate = chứng chỉ, log = nhật ký.
- Dịch từ **tiếng Anh** (`en`), không dịch từ tiếng Trung. Giữ nguyên placeholder (`{0}`, `{name}`, `%s`, `{{ .Name }}`), thẻ HTML, ký tự xuống dòng.
- Văn phong: ngắn gọn, giọng trung tính, viết hoa chữ đầu câu, không viết hoa mọi chữ.
- Chữ "1Panel" trong bản dịch **giữ nguyên** (brand script sẽ đổi lúc build), để bản dịch có thể gửi PR lên upstream.

---

## T3.1: Dịch giao diện frontend

**Phụ thuộc:** T0.2.

**File:** `frontend/src/lang/modules/vi.ts` (mới), `fork-docs/GLOSSARY-vi.md`.

**Việc cần làm**
1. Copy cấu trúc từ `en.ts` (khoảng 8.400 dòng). Dịch theo từng nhóm key cấp 1 (commons, menu, home, website, database, container, file, cronjob, setting, ...). Mỗi nhóm một commit `T3.1: vi <nhóm>` để dễ review.
2. Không để sót key nào; key chưa chắc chắn ghi `// TODO-vi` cuối dòng và liệt kê trong báo cáo.
3. Dịch xong, mở từng trang trong panel (đã đăng ký ngôn ngữ ở T3.2) để kiểm tra chữ bị tràn nút/cột; rút gọn bản dịch nếu tràn.

**Nghiệm thu:** [ ] `check-keys.mjs` (T3.4) báo 0 key thiếu, 0 key thừa. [ ] Không còn `TODO-vi` hoặc đã được chủ dự án duyệt. [ ] Ảnh chụp 10 màn hình chính ở tiếng Việt lưu vào `fork-docs/screenshots/vi/`.

---

## T3.2: Đăng ký ngôn ngữ `vi` `[fork]`

**Phụ thuộc:** T3.1 (có thể làm song song với file dịch còn dở).

**File sửa** (đã khảo sát ở v2.5.1, kiểm tra lại khi làm):
- `frontend/src/lang/index.ts`: thêm `vi: () => import('./modules/vi')`.
- Danh sách chọn ngôn ngữ: `frontend/src/App.vue`, `frontend/src/views/login/components/login-form.vue`, `frontend/src/views/setting/panel/index.vue`, `frontend/src/views/share/index.vue`, `frontend/src/lang/fu.ts`, `frontend/src/utils/app-store.ts` (tìm theo chuỗi `'es-ES'` để thấy hết chỗ cần thêm). Nhãn hiển thị: `Tiếng Việt`.
- Locale của Element Plus (`element-plus/es/locale/lang/vi`) và thư viện ngày giờ đang dùng (dayjs/moment, kiểm tra trong code) cho `vi`.
- `core/i18n/i18n.go`, `agent/i18n/i18n.go`: thêm `"vi": "lang/vi.yaml"` vào map.
- `core/app/dto/auth.go`: thêm `vi` vào `validate:"... oneof=..."` của `Language`. Tìm thêm các `oneof` tương tự trong core/agent.
- App Store: kiểm tra `utils/app-store.ts` chọn mô tả ứng dụng theo ngôn ngữ thế nào; `vi` không có mô tả thì rơi về `en`.

**Nghiệm thu:** [ ] Chọn Tiếng Việt ở trang login và trong Cài đặt đều được, F5 vẫn giữ. [ ] Không lỗi console. [ ] Ghi PATCHES.md đủ danh sách file.

---

## T3.3: Dịch thông báo backend

**Phụ thuộc:** T3.2.

**File:** `core/i18n/lang/vi.yaml`, `agent/i18n/lang/vi.yaml` (dịch từ `en.yaml`, ~600 và ~820 dòng).

**Nghiệm thu:** [ ] Gây vài lỗi có chủ đích (sai mật khẩu, tạo website trùng tên miền, xoá DB đang dùng) → thông báo tiếng Việt. [ ] `check-keys.mjs` báo đủ key.

---

## T3.4: Script kiểm tra key dịch

**Phụ thuộc:** T3.1.

**File:** `scripts/i18n/check-keys.mjs`.

**Đặc tả:** so `vi` với `en` cho: `frontend/src/lang/modules/{en,vi}.ts` (import động file TS bằng `tsx` hoặc parse AST bằng `typescript`), `core/i18n/lang/{en,vi}.yaml`, `agent/i18n/lang/{en,vi}.yaml`, installer `lang/{en,vi}.sh` (khi chạy trong repo installer, có cờ `--installer <path>`). In key thiếu/thừa, placeholder lệch (`{0}` có trong en mà không có trong vi), exit 1 nếu có lỗi. Chạy trong CI (T1.1).

**Nghiệm thu:** [ ] Cố tình xoá 1 key và sửa 1 placeholder → script báo đúng 2 lỗi.

---

## T3.5: Gửi bản dịch lên upstream

**Phụ thuộc:** T3.1–T3.4.

**Việc cần làm:** đọc `docs/TRANSLATION.md` của upstream, làm đúng hướng dẫn của họ (có thể họ yêu cầu khác cách ta đăng ký ở T3.2). Tạo nhánh từ `upstream/dev-v2` (hoặc nhánh họ chỉ định), áp các file dịch + đăng ký ngôn ngữ, mở PR. Bản dịch giữ chữ "1Panel" (đúng quy ước ở trên). Ghi link PR vào STATUS.md. Khi PR được merge, ở lần đồng bộ kế tiếp (T6.1) bỏ các patch T3.2 trùng lặp khỏi PATCHES.md.

**Nghiệm thu:** [ ] PR đã mở, có mô tả bằng tiếng Anh.

---

## T4.1: Dựng trang tài liệu riêng

**Phụ thuộc:** T2.1.

**Repo:** `xpanel-docs`, nhánh `xp` tạo từ `en/v2`.

**Việc cần làm**
1. Build thử: `pip install -r requirements.txt && mkdocs build`. Ghi lại phiên bản Python/MkDocs dùng được.
2. Thêm tiếng Việt: thư mục `docs/vi/` (hoặc cơ chế đa ngôn ngữ mà theme/plugin hiện tại hỗ trợ; kiểm tra `mkdocs.yml` và `overrides/`, nếu cần plugin mới thì hỏi trong QUESTIONS.md). Tiếng Việt là ngôn ngữ mặc định của site, tiếng Anh giữ làm phụ.
3. Đổi thương hiệu trong `mkdocs.yml` (`site_name`, `site_url`, `site_author`, logo, favicon, link mạng xã hội) bằng giá trị từ `brand.env`; ghi chú "Dựa trên tài liệu 1Panel, GPL-3.0" ở footer.
4. Thay lệnh cài đặt trong tài liệu bằng lệnh của ta (T1.4).
5. Cloudflare Pages build từ nhánh `xp`, tên miền `docs.<domain>`.

**Nghiệm thu:** [ ] `docs.<domain>` truy cập được, có tìm kiếm, chọn được ngôn ngữ.

---

## T4.2: Dịch tài liệu ưu tiên

**Phụ thuộc:** T4.1.

**Thứ tự:** trang chủ → `installation/` (toàn bộ) → hướng dẫn tạo website + SSL → database → App Store → backup/khôi phục → FAQ → các trang còn lại. Ảnh chụp màn hình thay bằng ảnh panel tiếng Việt khi có (`fork-docs/screenshots/vi/`). Mỗi trang 1 commit.

**Nghiệm thu:** [ ] Người mới cài và tạo được website WordPress chỉ bằng tài liệu tiếng Việt (chủ dự án thử).
