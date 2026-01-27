# ---------- Build stage ----------
FROM node:22-alpine AS builder
WORKDIR /app

# Install dependencies
COPY ui/package.json ui/yarn.lock ./
RUN yarn install --frozen-lockfile

# Copy source
COPY ui .

# Build production bundle (uses config.production.json)
RUN yarn build

# ---------- Runtime stage ----------
FROM nginx:alpine

# Copy built static files
COPY --from=builder /app/build /usr/share/nginx/html

# Copy nginx proxy config
COPY docker/ui/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
