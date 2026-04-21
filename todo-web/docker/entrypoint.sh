#!/bin/sh
set -eu

# Defaults
: "${BACKEND_URL:=http://todo-api:9090}"
: "${API_BASE_URL:=/api}"

export BACKEND_URL API_BASE_URL

# Render nginx config from template with BACKEND_URL substituted
envsubst '${BACKEND_URL}' \
  < /etc/nginx/templates/default.conf.template \
  > /etc/nginx/conf.d/default.conf

# Generate runtime config.js consumed by the SPA
cat > /usr/share/nginx/html/config.js <<EOF
window.__TODOOO_CONFIG__ = {
  apiBaseUrl: "${API_BASE_URL}"
};
EOF

# Ensure nginx can write temp/pid as non-root
mkdir -p /tmp/client_body /tmp/proxy /tmp/fastcgi /tmp/uwsgi /tmp/scgi

exec nginx -g 'daemon off;'
