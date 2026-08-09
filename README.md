<p align="center"><a href="https://1panel.pro"><img src="https://resource.1panel.pro/img/1panel-logo.png" alt="1Panel" width="300" /></a></p>
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
  <a href="/docs/README.es-es.md"><img alt="Español" src="https://img.shields.io/badge/Español-d9d9d9"></a>
  <a href="/docs/README.fr.md"><img alt="français" src="https://img.shields.io/badge/français-d9d9d9"></a>
  <a href="/docs/README.ko.md"><img alt="한국어" src="https://img.shields.io/badge/한국어-d9d9d9"></a>
  <a href="/docs/README.id.md"><img alt="Bahasa Indonesia" src="https://img.shields.io/badge/Bahasa Indonesia-d9d9d9"></a>
  <a href="/docs/README.zh-Hant.md"><img alt="中文(繁體)" src="https://img.shields.io/badge/中文(繁體)-d9d9d9"></a>
  <a href="/docs/README.tr.md"><img alt="Türkçe" src="https://img.shields.io/badge/Türkçe-d9d9d9"></a>
  <a href="/docs/README.ru.md"><img alt="Русский" src="https://img.shields.io/badge/Русский-d9d9d9"></a>
  <a href="/docs/README.ms.md"><img alt="Bahasa Melayu" src="https://img.shields.io/badge/Bahasa Melayu-d9d9d9"></a>
  <a href="/docs/README.fa.md"><img alt="Persian" src="https://img.shields.io/badge/%D9%81%D8%A7%D8%B1%D8%B3%DB%8C-d9d9d9"></a>
  <a href="/docs/README.lo.md"><img alt="ພາສາລາວ" src="https://img.shields.io/badge/%E0%BA%9E%E0%BA%B2%E0%BA%AA%E0%BA%B2%E0%BA%A5%E0%BA%B2%E0%BA%A7-d9d9d9"></a>
</p>

---

## What is 1Panel?

1Panel is a modern, open-source Linux server management panel and a lightweight AI management platform. Through an intuitive web interface, it provides users with comprehensive, one-stop server management capabilities:
- **Full-Stack AI Management**: Offers a unified management platform from bare metal to agents (Metal-to-Agent). It integrates an AI gateway, and Skills Hub, while supporting centralized management of agents and models.
- **Efficient Visual Operations**: Easily manage Linux servers through a web-based GUI, streamlining tasks such as host monitoring, file management, database management, and container management.
- **Rapid Website Deployment**: Deeply integrates with popular website builders like WordPress and . It enables one-click domain binding and SSL certificate configuration, significantly lowering the barrier to website creation.
- **Curated App Store**: Features a built-in store of high-quality open-source applications, providing one-click installation and upgrade services to effortlessly extend server capabilities.
- **Enterprise-Grade Security**: Deploys applications based on container technology to effectively minimize vulnerability exposure. It also provides security features such as WAF and log auditing to ensure comprehensive server protection.
- **One-Click Data Backup**: Supports one-click backup and restoration, and integrates with various cloud storage solutions to ensure data security and prevent loss.

## Why 1Panel?

| | 1Panel | cPanel / Plesk | aaPanel | Webmin |
|--|--------|----------------|---------|--------|
| Free & open source | ✅ | ❌ | Partial | ✅ |
| AI management | ✅ | ❌ | ❌ | ❌ |
| One-click app marketplace | ✅ 165+ apps | ❌ | ✅ | ❌ |
| Modern UI (post-2020) | ✅ | ❌ | Partial | ❌ |
| Docker / container management | ✅ | ❌ | ❌ | ❌ |
| Active development | ✅ | ✅ | ✅ | Slow |

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

1Panel OSS is free forever. 1Panel Pro and Ent adds features built for teams and production workloads:

| Feature | OSS | Pro | Ent |
|---------|:---:|:---:|:---:|
| One-click app installs | ✅ | ✅ | ✅ |
| AI agents (OpenClaw) | 1 agent | Unlimited | ✅ |
| WAF & advanced security | Basic | ✅ | ✅ |
| Website tamper protection | ❌ | ✅ | ✅ |
| Website uptime monitoring | ❌ | ✅ | ✅ |
| Multi-node management | ❌ | ✅ | ✅ |
| Custom logo & theme | ❌ | ✅ | ✅ |
| KVM Web UI | ❌ | ❌ | ✅ |
| AI Gateway | ❌ | ❌ | ✅ |
| Priority support | ❌ | ❌ | ✅ |

**From $80/year.** [Compare plans & start 30-day free trial →](https://1panel.pro/pricing)

## Community & Support

- **Discord** — [Join the community](https://discord.gg/bUpUqWqdRr) for help, feature requests, and show-and-tell
- **Docs** — [1panel.pro/docs](https://1panel.pro/docs)
- **Issues** — [GitHub Issues](https://github.com/1Panel-dev/1Panel/issues) for bug reports

## Security

Found a vulnerability? Please read [SECURITY.md](/SECURITY.md) before disclosing.

## License

Licensed under the [GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.html).
