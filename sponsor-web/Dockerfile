FROM node:22.14.0-alpine AS build
WORKDIR /app
COPY sponsor-web/package.json sponsor-web/pnpm-lock.yaml sponsor-web/.npmrc ./
# Pin the package manager as well as the reviewed lockfile. pnpm 10 does not
# apply pnpm 11's wall-clock release-age policy to an already pinned lockfile.
RUN corepack enable && corepack install --global pnpm@10.28.1 && pnpm install --frozen-lockfile
COPY sponsor-web/ ./
RUN pnpm run build
FROM caddy:2.8.4-alpine
COPY --from=build /app/dist /usr/share/caddy
EXPOSE 80
