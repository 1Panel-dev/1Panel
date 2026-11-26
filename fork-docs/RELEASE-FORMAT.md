# Định dạng gói phát hành và luồng cài đặt / nâng cấp của 1Panel v2

Khảo sát ngày 2026-11-27 (T0.3) trên bản phát hành chính thức **v2.5.1** (`resource.1panel.pro`), repo `1Panel-dev/installer` nhánh `v2` (commit `68d2515`, 2026-11-26) và mã nguồn 1Panel tag `v2.5.1`.

Mục tiêu: đọc xong file này là tự dựng được máy chủ phát hành cho XPanel, không cần đọc code.

---

## 0. Tóm tắt cho người dựng máy chủ

Máy chủ phát hành chỉ là **file tĩnh qua HTTPS** (upstream dùng Cloudflare). Không cần API.

```
<RepoURL>  = https://resource.1panel.pro/v2          (bản intl; bản cn: https://resource.fit2cloud.com/1panel/package/v2)

<RepoURL>/stable/latest                                   text, 1 dòng, không xuống dòng: v2.5.1
<RepoURL>/stable/latest.current                           JSON, upstream: {"v2.0":"v2.5.1"}  (xem 3.2)
<RepoURL>/stable/<ver>/release/checksums.txt              sha256sum của mọi tar.gz
<RepoURL>/stable/<ver>/release/1panel-<ver>-linux-<arch>.tar.gz
<RepoURL>/stable/<ver>/release/1panel-<ver>-release-notes Markdown, hiện trong hộp thoại nâng cấp
<RepoURL>/dev/latest                                      v2.5.1-alpha.1   (kênh dev, INSTALL_MODE=dev)
<RepoURL>/dev/<ver>/release/...                           như stable
<RepoURL>/beta/latest, beta/<ver>/release/...             code có dùng; upstream hiện trả 404
<RepoURL>/stable/<ver>/release/...                        <arch>: amd64 arm64 armv7 ppc64le riscv64 s390x

<ResourceURL> = <RepoURL>/resource                         (intl; cn: https://resource.fit2cloud.com/1panel/resource/v2)
<ResourceURL>/scripts/version.txt                          số (timestamp), vd 1744961472
<ResourceURL>/scripts/data.yaml?version=<v>                danh sách script mẫu
<ResourceURL>/scripts/scripts.tar.gz?version=<v>
<ResourceURL>/language/lang.tar.gz                         dự phòng, chỉ tải khi /usr/local/bin/lang thiếu (upstream intl: 404)
<ResourceURL>/geo/GeoIP.mmdb                               dự phòng, chỉ tải khi GeoIP thiếu (upstream intl: 404)
<ResourceURL>/<dir>/version.txt, <dir>/data.yaml           LoadAIResource (agent); v2.5.1 Community không gọi tới

<AppRepoURL> = https://apps.1panel.pro                     (cn: https://apps-assets.fit2cloud.com) — dùng nguyên của upstream ở Phase 1–4
<AppRepoURL>/<mode>/1panel.json.version.txt               số revision
<AppRepoURL>/<mode>/1panel.json.zip                       chỉ mục App Store
<AppRepoURL>/<mode>/1panel/<appKey>/<appVer>/docker-compose.yml
<AppRepoURL>/<mode>/1panel/...                            icon, README... (app_sync_task.go)
```

`<mode>` = `stable` hoặc `dev` (= `base.mode` trong `app.yaml` nhúng ở binary; bản build release là `stable`, xem 3.1).

Quy tắc phát hành một bản mới `<ver>`:
1. Upload toàn bộ `stable/<ver>/release/*` (tar.gz cho mọi arch, `checksums.txt`, `1panel-<ver>-release-notes`).
2. **Sau đó** mới ghi `stable/latest` và `stable/latest.current`. Thứ tự này quan trọng vì installer và nút cập nhật đọc `latest` trước rồi tải ngay.
3. `latest` không được chứa ký tự `<` (code coi đó là trang lỗi HTML và bỏ qua).

---

## 1. Gói phát hành `1panel-<ver>-linux-<arch>.tar.gz`

Tải ngày 2026-11-27: `1panel-v2.5.1-linux-amd64.tar.gz`, 74 052 587 B, SHA256 `42d4ded1bd4a6b5a7a32e0750887fcfc258afeac467a3c5c53f7ec7b74f09038` (khớp `checksums.txt`).

### 1.1 Cấu trúc (thư mục gốc bắt buộc là `1panel-<ver>-linux-<arch>/`)

Code nâng cấp và `quick_start.sh` đều `cd` vào đúng tên này (`strings.ReplaceAll(fileName, ".tar.gz", "")`).

```
1panel-v2.5.1-linux-amd64/
├── 1panel-core                 0755  84 709 536 B   binary core (frontend nhúng sẵn)
├── 1panel-agent                0755  99 823 776 B   binary agent (KVM helper nhúng sẵn)
├── 1pctl                       0755      13 841 B   CLI; = installer/1pctl với ORIGINAL_VERSION=version → v2.5.1
├── install.sh                  0755      44 886 B   = installer/install.sh
├── GeoIP.mmdb                  0644  19 565 776 B   (ngày 2025-03-05)
├── 1panel-core.service         0644         448 B   bản sao y initscript/1panel-core.service (thừa, không được dùng)
├── 1panel-agent.service        0644         449 B   bản sao y initscript/1panel-agent.service
├── initscript/                 0755
│   ├── 1panel-core.service  1panel-agent.service    0644  systemd
│   ├── 1panel-core.init     1panel-agent.init       0644  SysV
│   ├── 1panel-core.openrc   1panel-agent.openrc     0644  OpenRC (Alpine)
│   └── 1panel-core.procd    1panel-agent.procd      0644  OpenWrt
└── lang/                       0755
    └── en.sh fa.sh pt-BR.sh ru.sh zh.sh             0644  chuỗi cho install.sh và 1pctl
```

- Chủ sở hữu trong tar: `runner/runner` (build trên GitHub Actions runner).
- Có thể có thêm (installer xử lý nếu có, gói online không có): `appstore.tar.gz` (App Store offline), `docker/{bin,service/docker.service,conf/daemon.json}` (Docker offline), `iptables-deb/debian{11,12,13}/*.deb`.
- `initscript/*` và `lang/*` lấy từ repo installer **lúc phát hành** (installer HEAD hiện tại khác `lang/zh.sh` so với gói v2.5.1, nên phải ghim commit installer khi build gói).

### 1.2 Binary chính thức khác bản tự build

`go version -m`: go1.27.3, `CGO_ENABLED=0`, `-trimpath`, **`-tags=xpack`**. Bản chính thức có mã Pro/XPack đóng nguồn (chuỗi `xpack.alert.*`, `xpack.VMCountLimit`...) nên lớn hơn bản build từ repo (core 84,7 MB so với 81,7 MB; agent 99,8 MB so với 84,2 MB). XPanel **không** build với tag `xpack` (AGENTS.md quy tắc 5).

### 1.3 `checksums.txt`

Định dạng `sha256sum`: `<hex>  1panel-<ver>-linux-<arch>.tar.gz`, mỗi arch một dòng (upstream có dòng amd64 lặp 2 lần, vô hại). Chỉ installer (`quick_start.sh`) đọc file này; nút nâng cấp trong panel **không** kiểm tra checksum (xem 3.3).

---

## 2. Luồng cài đặt (repo installer, nhánh `v2`)

### 2.1 `quick_start.sh` (người dùng chạy `curl ... | bash` hoặc tải về rồi chạy)

| Bước | Chi tiết |
|---|---|
| Hằng số | `PANEL_EDITION="intl"`, `EDITION_FILE=".selected_edition"` |
| Kiến trúc | `uname -a` → `amd64` / `arm64` / `armv7` / `ppc64le` / `s390x` / `riscv64`; khác thì thoát |
| Kênh | `INSTALL_MODE` (env), mặc định `stable`, chỉ nhận `stable` hoặc `dev` |
| URL 1 | `GET https://resource.1panel.pro/v2/${INSTALL_MODE}/latest` → `VERSION` (rỗng thì thoát) |
| URL 2 | `GET .../v2/${INSTALL_MODE}/${VERSION}/release/checksums.txt` → lấy hash theo tên gói |
| Gói có sẵn | Nếu thư mục hiện tại đã có `1panel-${VERSION}-linux-${arch}.tar.gz` và hash khớp thì không tải lại |
| URL 3 | `curl -LOk .../v2/${INSTALL_MODE}/${VERSION}/release/1panel-${VERSION}-linux-${arch}.tar.gz` (**`-k`: bỏ qua kiểm tra TLS; gói vừa tải về KHÔNG được kiểm tra hash**, chỉ gói có sẵn mới được kiểm) |
| Giải nén | `tar zxf`, `cd 1panel-${VERSION}-linux-${arch}` |
| Edition | `echo intl > .selected_edition` (trong thư mục gói) |
| Chạy | `/bin/bash install.sh "$@"` |

Cả ba URL đều **viết cứng** host `resource.1panel.pro` ⇒ XPanel chỉ cần thay host này (T1.4).

### 2.2 `install.sh`

Tham số/biến môi trường: `PANEL_NON_INTERACTIVE`, `PANEL_LANG` (en, zh, fa, pt-BR, ru), `PANEL_INSTALL_DIR`, `PANEL_PORT`, `PANEL_ENTRANCE`, `PANEL_USERNAME`, `PANEL_PASSWORD`, `PANEL_INSTALL_DOCKER`, `PANEL_DOCKER_MODE` (auto, builtin, online), `PANEL_CONFIGURE_ACCELERATOR`, `PANEL_REPLACE_DAEMON_JSON` (và cờ `--...` tương ứng).

- Ngôn ngữ: danh sách cứng `AVAILABLE_LANGS=("en" "zh" "fa" "pt-BR" "ru")`, đọc `lang/<lang>.sh`; lựa chọn ghi vào `.selected_language`. **Thêm `vi` phải sửa mảng này và thêm `lang/vi.sh`** (T3.x).
- Edition: đọc `.selected_edition`; không có file thì mặc định **`cn`**.
- Thứ tự `main`: `Check_Root → Prepare_System → Set_Dir → Install_Docker → Set_Port → Set_Firewall → Set_Entrance → Set_Username → Set_Password → Init_Panel → Install_AppStore → Get_Ip → Check_Ready → Show_Result`.

**URL install.sh gọi:**

| URL | Khi nào |
|---|---|
| `https://get.docker.com`, mirror docker-ce (aliyun, tencent, 163, cernet), jsdelivr / raw.githubusercontent `docker-install`, `https://download.docker.com` | cài Docker online |
| `https://docker.1panel.live`, `https://hub.1panel.dev`, `https://docker.1ms.run`, `mirror.ccs.tencentyun.com` | ghi registry mirror vào `daemon.json` (khi người dùng đồng ý hoặc máy ở CN) |
| `ipinfo.io/country` | quyết định có cấu hình mirror |
| `https://api64.ipify.org` | lấy IP public để in địa chỉ truy cập |
| Link trong `lang/*.sh`: `https://1panel.pro`, `https://docs.1panel.pro`, `https://github.com/1Panel-dev/1Panel`, Discord | chỉ in ra màn hình (đổi thương hiệu ở T2.1/T1.4) |

**File ghi ra hệ thống** (`init_configure` và `install_and_configure`):

| Đường dẫn | Nội dung |
|---|---|
| `/usr/local/bin/1panel-core`, `/usr/local/bin/1panel-agent` | binary (chmod +x) |
| `/usr/bin/1panel` → `/usr/local/bin/1panel-core` | symlink (luôn tạo lại) |
| `/usr/bin/1panel-core`, `/usr/bin/1panel-agent`, `/usr/bin/1pctl` | symlink (nếu chưa có) |
| `/usr/local/bin/1pctl` | `1pctl` đã `sed` các biến: `BASE_DIR`, `ORIGINAL_PORT`, `ORIGINAL_USERNAME`, `ORIGINAL_PASSWORD` (sau khi cài xong bị thay bằng `**********`), `ORIGINAL_ENTRANCE`, `LANGUAGE`, `PANEL_EDITION`, có thể thêm `CHANGE_USER_INFO=use_existing`. **`ORIGINAL_VERSION` không bị sed: giữ giá trị đã điền sẵn trong gói** |
| `/usr/local/bin/lang/*.sh` | chuỗi cho `1pctl` |
| `$BASE_DIR/1panel/` (mặc định `/opt/1panel`) | thư mục dữ liệu; cài mới thì `rm -rf` nội dung cũ |
| `$BASE_DIR/1panel/geo/GeoIP.mmdb` | |
| `$BASE_DIR/1panel/resource/initscript/` | bản sao `initscript/` (dùng khi nâng cấp/khôi phục) |
| `$BASE_DIR/1panel/resource/offline/appstore.tar.gz` | nếu gói có `appstore.tar.gz` |
| `/etc/systemd/system/1panel-{core,agent}.service` | systemd; `enable` rồi `start` |
| `/etc/init.d/1panel-{core,agent}` | không có systemd: procd (`/etc/rc.common`), OpenRC (`/sbin/openrc-run`) hoặc SysV |
| `/etc/docker/daemon.json` (+ `daemon.json.1panel_bak`) | khi cài Docker / cấu hình mirror |
| `/usr/bin/docker*`, `/etc/systemd/system/docker.service` | Docker offline |
| `<thư mục gói>/install.log`, `.selected_language`, `.selected_edition` | log cài đặt (mật khẩu bị che) |
| `/etc/1panel/agent.sock` | do agent tạo; `Check_Ready` gọi `1pctl restart` nếu thiếu |

### 2.3 Panel biết phiên bản và edition thế nào

1. Lúc build gói: thay `ORIGINAL_VERSION=version` trong `1pctl` thành `ORIGINAL_VERSION=<ver>`. **Đây là nguồn duy nhất của số phiên bản**: binary không chứa phiên bản (`conf/app.yaml` nhúng chỉ có `version: v2.0.0`).
2. `core/init/viper/viper.go` đọc `/usr/local/bin/1pctl` (`ctl_conf.Load`): `BASE_DIR`, `ORIGINAL_*`, `LANGUAGE`, `PANEL_EDITION`. `PANEL_EDITION == "intl"` thì `Edition=intl`, còn lại là `cn`.
3. `core/init/hook/hook.go`: nếu `SystemVersion` trong `core.db` khác `ORIGINAL_VERSION` thì ghi đè. `Edition` lưu trong setting `Edition` (core và agent DB), từ đó chọn host (`RepoURL/ResourceURL/AppRepoURL` trong `core/global/global.go`, `agent/global/global.go`).
4. `1pctl version` → `1panel -l <lang> version` → đọc `SystemVersion` từ `core.db` (cần root).

Hệ quả cho XPanel: gói build phải điền `ORIGINAL_VERSION=v2.5.1-xp.N`; installer phải ghi `PANEL_EDITION=intl` (hoặc T1.3 bỏ phụ thuộc vào edition).

---

## 3. Luồng kiểm tra cập nhật và nâng cấp trong panel

Code: `core/app/service/upgrade.go` (node chính), `agent/utils/version/version.go` (bản sao logic cho agent, dùng ở đa node).

### 3.1 Kênh (`mode`)

`mode = global.CONF.Base.Mode` = `base.mode` của `core/cmd/server/conf/app.yaml` **nhúng vào binary**. Trong repo là `dev`; pipeline phát hành upstream đổi thành `stable` khi build (đã xác nhận: `strings` trên `1panel-core` và `1panel-agent` của gói v2.5.1 thấy `mode: stable`; cả hai binary đều nhúng `app.yaml` riêng). ⇒ Workflow release của XPanel (T1.2) **phải đặt `mode: stable`** trước khi build, nếu không panel sẽ hỏi kênh `dev`. Phiên bản chứa `beta` thì dùng `mode=beta`.

### 3.2 Kiểm tra cập nhật (`SearchUpgrade`)

| Gọi | Dùng làm gì |
|---|---|
| `GET <RepoURL>/stable/latest` | bản mới nhất (`LatestVersion`) |
| `GET <RepoURL>/stable/latest.current` | JSON map `{"<tiền tố>":"<ver>"}`: code tra key = 4 ký tự đầu của bản đang chạy khi minor < 10 (`v2.5.1` → `v2.5`), 5 ký tự khi minor ≥ 10 (`v2.10`). Giá trị là bản vá mới nhất **trong cùng nhánh minor** (`NewVersion`). Upstream hiện chỉ có `{"v2.0":"v2.5.1"}` nên với v2.5.x không tìm thấy key, `NewVersion` rỗng và chỉ `latest` có tác dụng. File phải là JSON hợp lệ, nếu không thì bị bỏ qua (chỉ ghi log) |
| `GET <RepoURL>/beta/latest` | chỉ khi bật Developer Mode |
| `GET <RepoURL>/dev/latest`, `.../beta/latest` | khi `mode=dev` |
| `GET <RepoURL>/<mode>/<ver>/release/1panel-<ver>-release-notes` | nội dung hiển thị |

So sánh phiên bản: `checkVersion` thêm hậu tố `-lts` cho bản **không có dấu `-`** rồi gọi `common.ComparePanelVersion`. Cần kiểm tra kỹ ở T1.5 vì `v2.5.1-xp.1` có `-` còn `v2.5.2` sẽ thành `v2.5.2-lts`.

Trang "Nhật ký thay đổi" (`LoadRelease`) đọc từ `https://docs.fit2cloud.com/1panel/changelog/` hoặc `https://docs.1panel.pro/v2/search/search_index.json` (viết cứng, không qua `RepoURL`; T4.1 xử lý).

### 3.3 Nâng cấp (`Upgrade`)

1. `loadArch()` (`uname -a`), kiểm tra dung lượng trống.
2. Tải `GET <RepoURL>/<mode>/<ver>/release/1panel-<ver>-linux-<arch>.tar.gz` (theo proxy hệ thống nếu có) vào `$BASE_DIR/1panel/tmp/upgrade/<ver>/downloads/`. **Không tải và không kiểm tra `checksums.txt`.**
3. Giải nén; thư mục phải là `1panel-<ver>-linux-<arch>/`.
4. Sao lưu vào `.../upgrade/<ver>/original/`: `/usr/local/bin/{1panel-core,1panel-agent,1pctl,lang/}`, 2 file service (đường dẫn theo init system), `$BASE_DIR/1panel/db/`, `$BASE_DIR/1panel/geo/GeoIP.mmdb`. Ghi `UpgradeLog`.
5. Thay file (lỗi ở bước nào thì rollback tới bước đó):
   - `1panel-core`, `1panel-agent` → `/usr/local/bin/`
   - `1pctl` → `/usr/local/bin/1pctl`, rồi ghi lại `BASE_DIR` và `LANGUAGE` cũ. **Các biến khác lấy theo gói mới**: `ORIGINAL_VERSION` của gói mới thành phiên bản mới, còn `PANEL_EDITION` trở về giá trị trong gói (`cn` nếu gói không sửa). Edition đã lưu trong DB nên vẫn giữ, nhưng gói XPanel nên điền sẵn `PANEL_EDITION=intl`.
   - `initscript/<file service phù hợp>` → thư mục service của init system
   - `lang/` → `/usr/local/bin/lang`
   - `GeoIP.mmdb` → `$BASE_DIR/1panel/geo/`
6. Cập nhật `SystemVersion` (core và agent DB), khởi động lại core + agent. `install.sh` **không** chạy khi nâng cấp.

Nút "Rollback" dùng bản sao lưu ở bước 4.

---

## 4. Việc cần làm cho XPanel (gợi ý cho các task sau)

| Task | Điểm rút ra |
|---|---|
| T1.2 | Gói phải đúng tên và cấu trúc ở mục 1.1; điền `ORIGINAL_VERSION` và `PANEL_EDITION=intl` vào `1pctl`; đặt `base.mode: stable`; KHÔNG dùng `-tags=xpack`; ghim commit installer; upload `release/*` trước, `latest` + `latest.current` sau; có file `release-notes` |
| T1.3 | 3 hàm URL trong `core/global/global.go` và `agent/global/global.go`; nhánh `cn`/enterprise cũng phải trỏ về máy chủ XPanel (hoặc bỏ phụ thuộc edition) |
| T1.4 | `quick_start.sh`: thay host ở 3 URL, bỏ `-k`, kiểm tra hash cả gói vừa tải; `lang/*.sh`: link thương hiệu |
| T1.5 | Kiểm tra `checkVersion`/`ComparePanelVersion` với hậu tố `-xp.N` và `-lts`; `latest.current` của XPanel nên có key `v2.5`; cân nhắc thêm kiểm tra checksum khi nâng cấp |
| T4.1 | `LoadRelease` viết cứng URL docs |
