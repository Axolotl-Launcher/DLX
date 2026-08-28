# Sponsor Gateway (foundation)

This directory is intentionally isolated from DLX: it owns sponsor identity, entitlement, API-key validation, and the DeepL-compatible public API while DLX remains an internal translator.

## Implemented in this foundation

- POST /v1/translate accepts Launcher-compatible JSON (text, source_lang, target_lang) and returns a translations array.
- Bearer API keys use the axl_live_<id>_<secret> shape. Only a peppered HMAC-SHA-256 digest is stored/compared in constant time; the required pepper is supplied through a secret file.
- The route checks key status and the permanent lifetime threshold (990 fen) before contacting DLX, limits bodies to 2 MiB/text to 10,000 runes, returns a non-sensitive request ID, applies a local fixed-window limit when configured, and does not log text or credentials.
- GET /healthz, GET /readyz, and GET /v1/account are present.
- The initial PostgreSQL schema is in migrations/001_initial.sql; it stores integer fen, one active key per user, and intentionally has no translation-body column.

## Not production-ready yet

The executable currently wires an intentionally denying in-memory Store. A PostgreSQL-backed store, Redis rate limiter/usage aggregation, login/session endpoints, Afdian client/webhooks/sync worker, admin functions, and sponsor-web UI must be implemented before a deployment can issue or accept real user keys. Do not deploy it as an active public sponsor service yet.

compose.production.yaml is a topology starting point: it publishes only Caddy (80/443) and keeps DLX/PostgreSQL/Redis on the Compose network. Pin the built images to reviewed digests in CI before release and create the untracked secret files under secrets/ only on the host.

## Runtime configuration

The gateway now fails closed unless all of the following are configured: `DATABASE_URL` (or `DATABASE_URL_FILE`), `REDIS_ADDR`, `API_KEY_PEPPER` (or `_FILE`), and `DLX_INTERNAL_TOKEN` (or `_FILE`). On startup it pings PostgreSQL and Redis and applies embedded, transactional schema migrations before listening. The PostgreSQL URL should be held in the Docker secret `database_url`; it must never be committed or logged.

Redis uses a per-key, atomic fixed-window rate bucket. PostgreSQL persists only aggregate daily usage (request count, input character count, error count).

## Recovery-code login

This deployment uses the requested simplified flow: after the server verifies an Afdian order and binds the provider identity, it generates an `axl_login_...` recovery code. The full code is shown once only. The database stores an HMAC digest, never the plaintext. A later login exchanges this code for a signed, HttpOnly session.

Order numbers are never sufficient on their own: the claim endpoint must re-query Afdian, verify that the order belongs to Axolotl, and enforce the one-user-per-Afdian-identity constraint before a code is issued. There is intentionally no automatic recovery path; lost codes require manual support and identity re-verification.
