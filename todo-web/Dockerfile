# syntax=docker/dockerfile:1.6

FROM node:20-alpine AS build
WORKDIR /app

# Install deps first (better layer caching)
COPY package.json package-lock.json* ./
RUN if [ -f package-lock.json ]; then npm ci; else npm install; fi

# Build the SPA
COPY tsconfig.json tsconfig.node.json vite.config.ts index.html ./
COPY public ./public
COPY src ./src
RUN npm run build


FROM nginx:alpine AS runtime

# gettext provides envsubst for runtime config rendering
RUN apk add --no-cache gettext \
  && rm -rf /usr/share/nginx/html/*

# Static app assets
COPY --from=build /app/dist/ /usr/share/nginx/html/

# Nginx config and server template (envsubst'd at startup)
COPY nginx/nginx.conf /etc/nginx/nginx.conf
COPY nginx/default.conf.template /etc/nginx/templates/default.conf.template

# Entrypoint renders templates + writes /config.js from env
COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh \
  && chown -R nginx:nginx /usr/share/nginx/html /var/cache/nginx \
  && touch /tmp/nginx.pid && chown nginx:nginx /tmp/nginx.pid

USER nginx

EXPOSE 8080

ENV BACKEND_URL=http://todo-api:9090 \
    API_BASE_URL=/api

ENTRYPOINT ["/entrypoint.sh"]
