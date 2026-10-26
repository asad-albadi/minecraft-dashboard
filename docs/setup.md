# Setup, accounts and security

The dashboard is an independent service. Do not edit, restart or recreate the Minecraft container to install it. Use already enabled private status/RCON interfaces and read-only mounts. Existing world/player files are never written.

## Named accounts

From the repository root, run `cd backend && go run ./cmd/setup-accounts -directory ../.dashboard`. The command creates **admin** and **viewer** accounts with separate cryptographically random passwords in `initial-credentials.txt` and Argon2id hashes in `accounts.json`. It refuses to overwrite either file. Move the initial passwords to your password manager; keep these files out of Git. No account is a Minecraft operator account.

Without local Go, build the tools using the Go Docker toolchain and run the resulting binaries locally:

```sh
mkdir -p .dashboard
docker run --rm -v "$PWD/backend:/src" -v "$PWD/.dashboard:/out" -w /src golang:1.22-alpine sh -c 'CGO_ENABLED=0 go build -o /out/setup-accounts ./cmd/setup-accounts && CGO_ENABLED=0 go build -o /out/hash-password ./cmd/hash-password'
.dashboard/setup-accounts -directory .dashboard
mkdir -p .dashboard/data
chmod 700 .dashboard .dashboard/data
```

Set `DASHBOARD_ACCOUNTS_SOURCE` to the absolute account-file path, `DASHBOARD_DATA_SOURCE` to the absolute dashboard-only data directory, and `DASHBOARD_UID`/`DASHBOARD_GID` to their owner IDs (`id -u`, `id -g`). The container remains non-root. Account file permissions must be 0600 or 0400. Never change ownership of Minecraft files to satisfy dashboard permissions; provision read-only access separately.

Account JSON is an array of `{ "username": "chosen-name", "passwordHash": "<Argon2id hash>", "role": "admin" }` records. Only `admin` and `viewer` are accepted, maximum 100 accounts. To change passwords/names, generate a hash with `go run ./cmd/hash-password` (password from stdin, 12–1024 bytes) and edit the private JSON. Restart only the dashboard backend to reload accounts and revoke all sessions. There is no web account-management endpoint.

Generate `SESSION_SECRET` with `openssl rand -base64 48`. It is a signing key, never a user's password. Old access-secret authentication has no fallback. Missing/invalid account configuration fails closed with HTTP 503; unauthenticated requests return 401. Sessions expire after eight hours and logout revokes the active session. Five login attempts per source IP per minute are allowed; behind the included proxy this is intentionally a shared deployment limit. Authentication does not trust arbitrary forwarded IP headers.

## Origin and transport

Compose binds HTTP to localhost and defaults `DASHBOARD_COOKIE_SECURE=false` for that local deployment only. For any HTTPS reverse proxy set it to `true` and set `DASHBOARD_PUBLIC_ORIGIN` to the exact origin, such as `https://dashboard.example.com` (no trailing slash). Cookies are HttpOnly/SameSite=Strict. Mutating requests require matching Origin and JSON Content-Type, including login/logout. Keep RCON private. Do not attach a Docker socket, even read-only.

## Sources and permissions

`MINECRAFT_HOST` is a hostname/IP without scheme or port; `MINECRAFT_PORT` defaults to 25565. `MINECRAFT_NETWORK` names an existing private Docker network. Paths in `.env` are host paths mounted read-only. Support includes standard `world/stats` and `world/advancements` plus the legacy `world/players/...` layout. Missing values remain Unknown/Unavailable. Public avatar requests send Minecraft names to mc-heads.net; local artwork is the fallback.

World directory size is cached for five minutes. Server status is cached for 15 seconds. Log polling reads up to 256 KiB per second, drops individual lines larger than 16 KiB, and retains at most 500 console lines. Profile JSON files are limited to 4 MiB; retained log history scanning is bounded and cached. Complete lifetime player history is not guaranteed.

Dashboard-owned `history.json` retains the latest 500 recognized activity events and 500 player messages. Server HH:MM:SS timestamps retain their source timezone; UTC observation timestamps are separate and do not pretend to be event dates. Restart preserves history but does not replay downtime logs. Audit keeps `audit.jsonl` and one rotated file, up to 1 MiB each. Back up this directory separately; it contains player activity and potentially chat text. Corrupt history is not overwritten automatically.

## Optional administrative integrations

Both `DASHBOARD_ENABLE_CHAT` and `DASHBOARD_ENABLE_CONSOLE` default false. Explicit opt-in requires a configured admin account, private RCON and writable dashboard-only audit storage. Chat uses escaped JSON `tellraw`; console permits only `list`, `time query daytime`, and `time query day`. Viewer requests are denied server-side. Ten combined actions per admin per minute are allowed. Audit intent is synced before sending; transport/post-send audit failure can mean unknown delivery, so do not blindly retry. A successful RCON reply is not proof a player read a message.

World mutations, arbitrary RCON, player administration, container lifecycle and backup writes/restores remain disabled; no configuration flag bypasses that boundary. Login/logout audit failures are warned in backend logs without locking users out. Command/chat audit failures prevent sending.
