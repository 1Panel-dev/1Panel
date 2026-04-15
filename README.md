<p align="center"><a href="https://1panel.pro"><img src="https://resource.1panel.pro/img/1panel-logo.png" alt="1Panel" width="300" /></a></p>

<h3 align="center">The open-source VPS control panel with native AI agent support</h3>

<p align="center">
  Trusted by <strong>2,000,000+</strong> self-hosters worldwide
</p>

<p align="center">
  <a href="https://trendshift.io/repositories/2462" target="_blank"><img src="https://trendshift.io/api/badge/repositories/2462" alt="1Panel-dev%2F1Panel | Trendshift" style="width: 240px; height: auto;" /></a>
</p>

<p align="center">
  <a href="https://www.gnu.org/licenses/gpl-3.0.html"><img src="https://shields.io/github/license/1Panel-dev/1Panel?color=%231890FF" alt="License: GPL v3"></a>
  <a href="https://app.codacy.com/gh/1Panel-dev/1Panel"><img src="https://app.codacy.com/project/badge/Grade/da67574fd82b473992781d1386b937ef" alt="Codacy"></a>
  <a href="https://discord.gg/bUpUqWqdRr"><img src="https://img.shields.io/discord/1318846410149335080?logo=discord&labelColor=%20%235462eb&logoColor=%20%23f5f5f5&color=%20%235462eb" alt="Discord"></a>
  <a href="https://github.com/1Panel-dev/1Panel/releases"><img src="https://img.shields.io/github/v/release/1Panel-dev/1Panel" alt="GitHub release"></a>
  <a href="https://github.com/1Panel-dev/1Panel"><img src="https://img.shields.io/github/stars/1Panel-dev/1Panel?color=%231890FF&style=flat-square" alt="Stars"></a>
</p>

<p align="center">
  <a href="/README.md"><img alt="English" src="https://img.shields.io/badge/English-d9d9d9"></a>
  <a href="/docs/README.zh-Hans.md"><img alt="中文(简体)" src="https://img.shields.io/badge/中文(简体)-d9d9d9"></a>
  <a href="/docs/README.ja.md"><img alt="日本語" src="https://img.shields.io/badge/日本語-d9d9d9"></a>
  <a href="/docs/README.pt-br.md"><img alt="Português (Brasil)" src="https://img.shields.io/badge/Português (Brasil)-d9d9d9"></a>
  <a href="/docs/README.ar.md"><img alt="العربية" src="https://img.shields.io/badge/العربية-d9d9d9"></a>
  <a href="/docs/README.de.md"><img alt="Deutsch" src="https://img.shields.io/badge/Deutsch-d9d9d9"></a>
  <a href="/docs/README.es.md"><img alt="Español" src="https://img.shields.io/badge/Español-d9d9d9"></a>
  <a href="/docs/README.fr.md"><img alt="français" src="https://img.shields.io/badge/français-d9d9d9"></a>
  <a href="/docs/README.ko.md"><img alt="한국어" src="https://img.shields.io/badge/한국어-d9d9d9"></a>
  <a href="/docs/README.id.md"><img alt="Bahasa Indonesia" src="https://img.shields.io/badge/Bahasa Indonesia-d9d9d9"></a>
  <a href="/docs/README.zh-Hant.md"><img alt="中文(繁體)" src="https://img.shields.io/badge/中文(繁體)-d9d9d9"></a>
  <a href="/docs/README.tr.md"><img alt="Türkçe" src="https://img.shields.io/badge/Türkçe-d9d9d9"></a>
  <a href="/docs/README.ru.md"><img alt="Русский" src="https://img.shields.io/badge/Русский-d9d9d9"></a>
  <a href="/docs/README.ms.md"><img alt="Bahasa Melayu" src="https://img.shields.io/badge/Bahasa Melayu-d9d9d9"></a>
</p>

---

## What is 1Panel?

1Panel is a modern, open-source VPS control panel — and the only one with **native AI agent support**. Run Ollama models, deploy OpenClaw agents, and manage your entire server stack from one clean web interface. No CLI memorization required.

👉 Watch the [2-minute introduction](https://www.youtube.com/watch?v=Jl_wqp-XA08)

## Why 1Panel?

| | 1Panel | cPanel / Plesk | aaPanel | Webmin |
|--|--------|----------------|---------|--------|
| Free & open source | ✅ | ❌ | Partial | ✅ |
| Native AI agent runtime | ✅ | ❌ | ❌ | ❌ |
| One-click app marketplace | ✅ 165+ apps | ❌ | ✅ | ❌ |
| Modern UI (post-2020) | ✅ | ❌ | Partial | ❌ |
| Docker / container management | ✅ | ❌ | ❌ | ❌ |
| Active development | ✅ | ✅ | ✅ | Slow |

## Key Features

- **AI Agent Runtime**: Deploy Ollama LLMs, spin up OpenClaw personal agents, and monitor GPU utilization — all from the dashboard. No separate AI stack to manage.
- **One-Click Website Deployment**: Launch production-ready websites with automatic domain binding, SSL provisioning, and Nginx config — zero manual setup.
- **App Marketplace**: 165+ trusted open-source apps (Nextcloud, Bitwarden, Umami, NocoBase, and more) installed and updated with a single click.
- **Docker & Container Management**: Create, start, stop, and inspect containers, images, networks, and volumes through a visual UI — no CLI juggling.
- **Security Out of the Box**: Firewall rules, fail2ban, container isolation, WAF, and audit logs — configured and running from day one.
- **Backup & Restore**: Schedule automated backups to AWS S3, Cloudflare R2, or local storage. Restore any snapshot in one click.

## Quick Start

> **Requirements:** Linux VPS (Debian / Ubuntu / CentOS / Rocky), 1 GB RAM, internet access.  
> Takes ~60 seconds.

```bash
bash -c "$(curl -sSL https://resource.1panel.pro/v2/quick_start.sh)"
```

After installation, open `http://<your-server-ip>:<port>/<security-path>` in your browser.  
Run `1pctl user-info` via SSH if you need to retrieve your access credentials.

## Screenshot

![1Panel UI](https://resource.1panel.pro/img/overview_en_v2.png)

## Pro Edition

1Panel OSS is free forever. Pro adds features built for teams and production workloads:

| Feature | OSS | Pro |
|---------|:---:|:---:|
| One-click app installs | ✅ | ✅ |
| AI agents (OpenClaw) | 1 agent | Unlimited |
| WAF & advanced security | Basic | ✅ |
| Website tamper protection | ❌ | ✅ |
| Website uptime monitoring | ❌ | ✅ |
| Multi-node management | ❌ | ✅ |
| Custom logo & theme | ❌ | ✅ |
| Priority support | ❌ | ✅ |

**From $80/year.** [Compare plans & start 30-day free trial →](https://1panel.pro/pricing)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=1Panel-dev/1Panel&type=Date)](https://star-history.com/#1Panel-dev/1Panel&Date)

## Community & Support

- **Discord** — [Join the community](https://discord.gg/bUpUqWqdRr) for help, feature requests, and show-and-tell
- **Docs** — [1panel.pro/docs](https://1panel.pro/docs)
- **Issues** — [GitHub Issues](https://github.com/1Panel-dev/1Panel/issues) for bug reports

## Security

Found a vulnerability? Please read [SECURITY.md](/SECURITY.md) before disclosing.

## License

Licensed under the [GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.html).
