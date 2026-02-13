<div align="center">

# 🌐 Free Proxies

**A continuously updated, validated list of free HTTP & HTTPS proxies.**

Scraped → Deduplicated → Verified → Published — every 30 minutes, 24/7.

[![GitHub Stars](https://img.shields.io/github/stars/Syn2Much/Free-Proxies?style=for-the-badge&logo=github&color=yellow)](https://github.com/Syn2Much/Free-Proxies/stargazers)
[![Last Commit](https://img.shields.io/github/last-commit/Syn2Much/Free-Proxies?style=for-the-badge&logo=github&color=blue)](https://github.com/Syn2Much/Free-Proxies/commits/main)
[![License](https://img.shields.io/github/license/Syn2Much/Free-Proxies?style=for-the-badge&color=green)](LICENSE)

</div>

---

## ⚡ Quick Start

Grab the latest proxies instantly:

```bash
# HTTP proxies
curl -s https://raw.githubusercontent.com/Syn2Much/Free-Proxies/main/http/http.txt

# HTTPS proxies
curl -s https://raw.githubusercontent.com/Syn2Much/Free-Proxies/main/https/https.txt
```

Or clone the full repo:

```bash
git clone https://github.com/Syn2Much/Free-Proxies.git
```

---

## 🔧 How It Works

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Scrape      │ ──▶ │  Deduplicate  │ ──▶ │   Validate    │ ──▶ │   Publish     │
│  Major GitHub │     │  Remove dupes │     │  Health check │     │  Push to repo │
│  proxy sources│     │  & malformed  │     │  each proxy   │     │  every 30 min │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
```

- **Sources** — Aggregated from top GitHub proxy lists (only recently updated list)
- **Deduplication** — Identical and malformed entries are stripped
- **Validation** — Every proxy is checked for connectivity before publishing
- **Schedule** — Full pipeline runs every 30 minutes, around the clock

---

## 📦 Available Formats

| Format | HTTP | HTTPS |
|:-------|:-----|:------|
| **Plain Text** (`ip:port`) | [`http/http.txt`](http/http.txt) | [`https/https.txt`](https/https.txt) |
| **CSV** | [`http/http.csv`](http/http.csv) | [`https/https.csv`](https/https.csv) |
| **JSON** | [`http/http.json`](http/http.json) | [`https/https.json`](https/https.json) |

---

## 📡 Usage Examples

**Python (requests)**
```python
import requests

proxies = requests.get(
    "https://raw.githubusercontent.com/Syn2Much/Free-Proxies/main/http/http.txt"
).text.strip().splitlines()

for proxy in proxies[:5]:
    try:
        r = requests.get("https://httpbin.org/ip", proxies={"http": f"http://{proxy}"}, timeout=5)
        print(f"[✓] {proxy} → {r.json()['origin']}")
    except:
        print(f"[✗] {proxy}")
```

**Bash (one-liner)**
```bash
while read proxy; do
  curl -s --proxy "http://$proxy" --max-time 5 https://httpbin.org/ip && echo " ← $proxy"
done < <(curl -s https://raw.githubusercontent.com/Syn2Much/Free-Proxies/main/http/http.txt | head -10)
```

---

## ⚠️ Disclaimer

Free proxies are inherently volatile — expect some to go offline between refresh cycles. These are intended for **testing, research, and development purposes**. Do not route sensitive traffic through untrusted proxies.

---

## ⭐ Support

If this saves you time, star the repo — it helps with visibility and keeps the project going.

---

## 👤 Author

**Syn2Much**

[![Email](https://img.shields.io/badge/Email-dev%40sinnners.city-red?style=flat-square&logo=gmail)](mailto:dev@sinnners.city)
[![X](https://img.shields.io/badge/@synacket-black?style=flat-square&logo=x)](https://x.com/synacket)
