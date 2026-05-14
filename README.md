# Pandaria

Pandaria drives **Chrome** and **[Lightpanda][lp]** against the same URL through a
shared man-in-the-middle HTTP proxy, archives every response on disk, and replays
those archives on later runs. The result is a deterministic harness for comparing
how a full-fat browser and a JavaScript-runtime-focused headless engine see the
same page — and for iterating against a frozen snapshot of the network without
hitting the origin.

## Cache layout

URLs are mapped to filesystem paths so a query string never collides with a
bare path:

```
/tmp/pandaria/<engine>/sites/<host>/<path-without-ext>[.<sha1(query)>]<.ext>
```

Trailing-slash URLs become `__index.html` so directories and pages can coexist.
Content-Type on replay is derived from the file extension via `mime`.

## Install

Requires Go 1.26+.

```sh
git clone <repo> pandaria
cd pandaria
go build -o pandaria .
```

For the Lightpanda side you also need a running Lightpanda CDP server
listening on `LPAddr` (default `127.0.0.1:9222`). See the commented-out
exec snippet in `driver/driver_lightpanda.go` for the canonical invocation.

## Usage

Fetch a URL with both engines, capturing everything to disk:

```sh
./pandaria fetch https://example.com
```

Run only the proxy + Chrome (skip Lightpanda):

```sh
./pandaria fetch https://example.com --enable-lp=false
```

Replay from cache only (no fresh fetches — Chrome headful so you can poke at
it):

```sh
./pandaria fetch https://example.com --chrome-headless=false
```

The process runs until you `Ctrl-C`. On exit it tears down both CDP contexts
cleanly.

## CLI reference

| Flag                     | Default            | Description                                     |
| ------------------------ | ------------------ | ----------------------------------------------- |
| `<url>` (positional)     | —                  | URL to fetch.                                   |
| `--http-proxy-addr`      | `127.0.0.1:3000`   | Bind address for the MITM proxy.                |
| `--lp-addr`              | `127.0.0.1:9222`   | Lightpanda CDP endpoint to attach to.           |
| `--enable-http-proxy`    | `true`             | Run the proxy server.                           |
| `--enable-lp`            | `true`             | Drive a Lightpanda instance.                    |
| `--enable-chrome`        | `true`             | Drive a Chrome instance.                        |
| `--chrome-headless`      | `true`             | Run Chrome headless.                            |

## Notes & caveats

- The proxy assumes one of two `User-Agent`s (`Chrome` or `Lightpanda`); any
  other traffic that lands on it will panic. This is intentional — a third
  client routed through the same port would silently corrupt the corpus.
- Chrome is started with `--disable-web-security` and
  `--ignore-certificate-errors` whenever the proxy is enabled, since goproxy
  signs MITM certs on the fly. Don't reuse the user-data-dir
  (`/tmp/pandaria/chrome-userdata-dir`) for anything you care about.
- Caches are never invalidated automatically. Delete `/tmp/pandaria/` to force
  a fresh capture.

[lp]: https://lightpanda.io
