# ======================
# Stage 1: Build
# ======================
FROM node:20-alpine AS build

WORKDIR /app

# Install ALL deps (including dev deps)
COPY package*.json ./
RUN npm ci

# Copy all source code
COPY . .

# Build the app
RUN npm run build

# ======================
# Stage 2: Production (Nginx)
# ======================
FROM nginx:1.27-alpine

# Remove default nginx website
RUN rm -rf /usr/share/nginx/html/*

# Copy build output
COPY --from=build /app/dist /usr/share/nginx/html

COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
