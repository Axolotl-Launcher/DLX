# GitHub Actions production deployment

## Release flow

1. Push a reviewed commit to `main`.
2. **CI** runs Go and Sponsor Web checks.
3. **Publish images** creates immutable GHCR images tagged with the commit SHA.
4. Open **Actions → Deploy production → Run workflow**, paste that successful commit SHA, and approve the protected `production` environment.
5. The server pulls only that immutable image tag, starts it without building, and requires Gateway `/readyz` before reporting success.

## One-time GitHub configuration

Create the `production` Environment in **Settings → Environments** and set **Required reviewers**. Add these Environment secrets:

| Secret | Value |
| --- | --- |
| `DEPLOY_HOST` | `198.44.26.50` |
| `DEPLOY_PORT` | `1200` |
| `DEPLOY_USER` | `deploy` |
| `DEPLOY_SSH_PRIVATE_KEY` | the dedicated deployment private key generated during setup |

The server needs a GitHub Container Registry token with **read:packages** only. Store it locally on the server in Docker's credential store using `docker login ghcr.io`; never put it in this repository or Actions secrets.

## Server-only data

`/opt/axolotl-dlx/secrets/` remains only on the server. It contains database, Session, Pepper, internal DLX, and Afdian secrets. CI/CD must never upload, replace, print, or commit these files.

## Rollback

Run **Deploy production** again with a previously published successful 40-character Git SHA. The deployment script uses that immutable tag rather than `main` or `latest`.
