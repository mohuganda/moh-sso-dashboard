# ======================
# Stage 1: Build
# ======================
FROM node:22-alpine AS build

WORKDIR /app
ENV FRONTEND_ASSET_BASE_URL=/portal
ENV FRONTEND_BASE_PATH=/portal
ENV NODE_OPTIONS=--max-old-space-size=4096

COPY . .
RUN npm ci --no-audit --no-fund --fetch-retries=5 --fetch-retry-mintimeout=20000
RUN node scripts/use-runtime-config.mjs production
RUN npm run build:docker

# ======================
# Stage 2: Production (Static server)
# ======================
FROM node:20-alpine

WORKDIR /app

RUN npm install -g serve

COPY --from=build /app/dist ./dist

EXPOSE 3000

CMD ["serve", "-s", "dist", "-l", "3000"]
