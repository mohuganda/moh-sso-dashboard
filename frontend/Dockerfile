# ======================
# Stage 1: Build
# ======================
FROM node:20-alpine AS build

WORKDIR /app

COPY . .
RUN npm ci
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
