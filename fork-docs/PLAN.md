# Kế hoạch: fork 1Panel thành bản riêng (tiếng Việt, thương hiệu riêng)

> Khảo sát dựa trên mã nguồn 1Panel **v2.5.1** (tag ngày 12/11/2026).

## 1. Vì sao fork thay vì viết lại

1Panel là mã nguồn mở và đã có sẵn gần như mọi tính năng của một panel quản lý server:

| | 1Panel v2 có sẵn |
|---|---|
| Kiến trúc | `core` (đăng nhập, giao diện, quản lý nhiều máy) + `agent` (thao tác trên máy), Go + Gin + GORM/SQLite |
| Giao diện | Vue 3 + Vite + Element Plus, 12 ngôn ngữ (chưa có tiếng Việt) |
| Chức năng | Website (OpenResty), SSL/ACME + DNS API, database (MySQL, PostgreSQL, Redis, MongoDB), file, cron, backup, firewall, SSH, Fail2ban, ClamAV, terminal, container/compose, runtime PHP/Node/Go/Python/Java, giám sát, snapshot, AI/MCP… |
| App Store | 269 ứng dụng dạng Docker Compose (repo `1Panel-dev/appstore`) |
| Installer | `1Panel-dev/installer` nhánh `v2`, CLI `1pctl` |
| Tài liệu | `1Panel-dev/docs` (MkDocs, nhánh `en/v2`) |

Vì vậy việc của chúng ta là **đóng gói lại**: build từ source, đổi thương hiệu, thêm tiếng Việt, tự phát hành, tự cập nhật, và giữ đồng bộ với upstream. Không viết lại tính năng nào đã có.

## 2. Pháp lý: GPL-3.0 (cả 4 repo)

Được phép: dùng, sửa, phát hành lại, kể cả thương mại. Bắt buộc:
- Giữ nguyên `LICENSE` và mọi dòng bản quyền của FIT2CLOUD trong mã nguồn.
- Bản phát hành phải kèm (hoặc chỉ rõ chỗ lấy) **toàn bộ mã nguồn** đã sửa, cũng theo GPL-3.0. Repo fork để **public** trên GitHub là cách đơn giản nhất.
- Ghi rõ là bản đã sửa đổi (README + trang "Giới thiệu" trong panel: "Dựa trên 1Panel của FIT2CLOUD, giấy phép GPL-3.0").

Không được:
- **Dùng tên và logo "1Panel"** cho sản phẩm của mình. GPL không cấp quyền thương hiệu (mục 7e của giấy phép). Phải đổi tên, logo và tên miền.
- Dùng tính năng **Pro/XPack**: phần này đóng mã, không có trong repo (`PANEL_XPACK=false`). Bản fork chỉ có tính năng bản Community.

Tên tạm trong tài liệu này: **XPanel**. Mọi chỗ dùng tên đều lấy từ 1 file cấu hình (`branding/brand.env`), nên đổi tên sau rất dễ.

## 3. Nguyên tắc số 1: diff với upstream càng nhỏ càng tốt

Upstream ra bản mới khoảng 2–3 tuần một lần. Nếu sửa trực tiếp hàng nghìn chỗ "1Panel" trong code, mỗi lần đồng bộ sẽ là địa ngục xung đột. Vì vậy:

1. **Không sửa chuỗi thương hiệu trong repo.** Đổi thương hiệu chạy bằng script `scripts/brand/apply.sh` trong CI, trên bản checkout tạm, ngay trước khi build. Repo vẫn giữ chữ "1Panel" như upstream.
2. **Giữ nguyên tên nội bộ:** module Go `github.com/1Panel-dev/1Panel/...`, binary `1panel-core` / `1panel-agent`, thư mục cài `/opt/1panel`, CLI `1pctl`, service `1panel-core`/`1panel-agent`. Đổi những thứ này thì diff phình to mà người dùng ít khi thấy. Script brand chỉ thêm alias CLI `xp` → `1pctl`. Có thể xem lại sau khi ổn định.
3. **Patch vào code chỉ ở vài điểm có chủ đích**, mỗi patch là 1 commit riêng có tiền tố `[fork]` để dễ rebase:
   - URL máy chủ cập nhật / tài nguyên / app store đọc từ cấu hình (`global.go`).
   - Đăng ký ngôn ngữ `vi` (frontend + backend).
   - Ẩn quảng cáo bản Pro.
4. **File thêm mới** đặt trong thư mục riêng (`branding/`, `scripts/brand/`, `scripts/release/`, `.github/workflows/xp-*.yml`, `fork-docs/`) để không bao giờ xung đột.
5. Bản dịch tiếng Việt **gửi PR ngược lên upstream**. Được nhận thì từ đó không phải giữ diff phần dịch nữa.

## 4. Kiến trúc 1Panel cần nắm

```
Trình duyệt ──HTTPS──▶ 1panel-core  (port, lối vào bảo mật, đăng nhập, phục vụ giao diện,
                         │           chuyển tiếp API tới agent; quản lý nhiều node)
                         └──unix socket/HTTPS──▶ 1panel-agent (website, DB, file, container...)
                                                   └─ Docker: mọi app trong App Store chạy bằng compose
Dữ liệu: /opt/1panel/db/core.db, agent.db (SQLite) · CLI: 1pctl · config: /usr/local/bin/1pctl + /opt/1panel/conf
```

Máy chủ bên ngoài mà panel gọi tới (đã gom trong `core/global/global.go` và `agent/global/global.go`):

| Hàm | Upstream (bản intl) | Dùng cho |
|---|---|---|
| `RepoURL()` | `https://resource.1panel.pro/v2` | kiểm tra & tải bản cập nhật panel |
| `ResourceURL()` | `https://resource.1panel.pro/v2/resource` | tài nguyên phụ (GeoIP, script...) |
| `AppRepoURL()` | `https://apps.1panel.pro` | dữ liệu App Store |

Ngoài ra còn: tải binary KVM helper từ `resource.fit2cloud.com` khi build, các link tài liệu, diễn đàn, GitHub trong giao diện.

## 5. Hạ tầng của bản fork

| Thứ | Đề xuất |
|---|---|
| Mã nguồn | GitHub, 4 repo fork public: `xpanel` (từ 1Panel), `xpanel-installer`, `xpanel-docs`, `xpanel-appstore` (giai đoạn sau) |
| Máy chủ phát hành | Cloudflare R2 + tên miền riêng, ví dụ `https://dl.<domain>/v2/stable/latest`, `/v2/stable/<ver>/release/<file>`. Bố cục thư mục **giống hệt** upstream để `quick_start.sh` và nút cập nhật chỉ cần đổi host |
| Tài liệu | Cloudflare Pages, `https://docs.<domain>` |
| App Store | Giai đoạn 1 dùng luôn App Store của upstream (`apps.1panel.pro`). Giai đoạn 5 mới tự host bản mirror |
| Phiên bản | `v<upstream>-xp.<n>`, ví dụ `v2.5.1-xp.1`. Hàm `ComparePanelVersion` tách theo `.` và `-` nên `v2.5.1-xp.1 < v2.5.1-xp.2 < v2.5.2`; T1.5 sẽ có unit test xác nhận |
| Môi trường dev | WSL2 Ubuntu 24.04, repo ở `~/xpanel` (ổ ext4 của WSL, nhanh hơn nhiều lần so với `/mnt/d` khi build). Sao lưu/đồng bộ bằng GitHub, push sau mỗi task. Không dùng Nextcloud cho repo |

## 6. Lộ trình

| Phase | Nội dung | Ước lượng |
|---|---|---|
| 0 | Fork, build được từ source, hiểu gói phát hành | 3–4 ngày |
| 1 | CI, phát hành lên R2, installer riêng, cập nhật trong panel | 1–1,5 tuần |
| 2 | Đổi thương hiệu bằng script, ẩn quảng cáo Pro | 1 tuần |
| 3 | Tiếng Việt: giao diện (~8.400 dòng), backend, installer; gửi PR upstream | 1,5–2 tuần |
| 4 | Tài liệu tiếng Việt | 1 tuần cho phần chính, sau đó làm dần |
| 5 | App Store riêng (tuỳ chọn) | 1 tuần |
| 6 | Quy trình đồng bộ upstream, test VM tự động | 3–4 ngày, rồi chạy định kỳ |
| 7 | Tính năng riêng (cần thiết kế, làm khi có nhu cầu thật) | |

**Mốc dùng được (MVP) = xong Phase 0–3:** người dùng cài bằng một dòng lệnh từ máy chủ của bạn, panel mang tên riêng, giao diện tiếng Việt, cập nhật được qua nút trong panel.

Chi tiết từng task cho worker: `tasks/` (đọc `AGENTS.md` trước).

## 7. Rủi ro

| Rủi ro | Cách xử lý |
|---|---|
| Upstream đổi cấu trúc làm script brand hoặc patch hỏng | Script brand có bước kiểm tra (đếm chuỗi còn sót, kiểm tra build); CI chạy trên mỗi tag upstream mới trước khi merge |
| Bản dịch lệch khi upstream thêm key | Script so key `en` ↔ `vi` chạy trong CI; key thiếu tự rơi về tiếng Anh (vue-i18n fallback) |
| KVM helper là binary tải sẵn từ máy chủ FIT2CLOUD, có thể không có mã nguồn | T6.3 điều tra giấy phép; nếu không rõ thì build kèm cờ tắt tính năng VM |
| App Store upstream đổi điều khoản / chặn | Phase 5: mirror từ repo GPL `appstore` |
| Phải công bố mã nguồn | Repo public ngay từ đầu, link trong trang Giới thiệu |
| Go 1.27.3 và Node/Vite 8 mới, worker có thể dùng bản cũ | Ghi rõ phiên bản trong T0.2, CI cài đúng phiên bản |
