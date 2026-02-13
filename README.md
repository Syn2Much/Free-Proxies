
# Free Proxies

A regularly updated, ready-to-use list of free HTTP and HTTPS proxies. We scrape from every major public source on GitHub, remove duplicates, and verify each proxy before posting. If it's on the list, it was working when we checked.

---

## Grab the List

### HTTP Proxies

```bash
curl -s https://raw.githubusercontent.com/Syn2Much/Free-Proxies/main/http/http.txt
```

### HTTPS Proxies

```bash
curl -s https://raw.githubusercontent.com/Syn2Much/Free-Proxies/main/https/https.txt
```

### Clone the repo

```bash
git clone https://github.com/Syn2Much/Free-Proxies.git
```

## Formats

Each protocol is available in three formats:

| Format | HTTP | HTTPS |
|--------|------|-------|
| Plain text (`ip:port`) | [`http/http.txt`](http/http.txt) | [`https/https.txt`](https/https.txt) |
| CSV (`proxy,country,region`) | [`http/http.csv`](http/http.csv) | [`https/https.csv`](https/https.csv) |
| JSON (full geo + connection info) | [`http/http.json`](http/http.json) | [`https/https.json`](https/https.json) |

## Updates

The list is refreshed on a regular cycle. Check the latest commit timestamp to see when it was last validated. Free proxies are volatile by nature, so some may drop off between updates.

## Support the Project

If this saves you time, drop a star on the repo. It helps us know people are using it and motivates more frequent checks and source expansion.

## Author

**Syn2Much**

- Email: [dev@sinnners.city](mailto:dev@sinnners.city)
- X: [@synacket](https://x.com/synacket)
