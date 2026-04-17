FROM node:latest AS node-builder
WORKDIR /web
COPY package*.json .
RUN npm install
COPY . .
RUN npm run build

# FROM node:latest AS node-builder
# WORKDIR /web
# COPY . .
# COPY --from=node-deps /web/node_modules .
# RUN npm run build

FROM nginx:latest AS web
COPY nginx/tls_certs.sh /docker-entrypoint.d/tls_certs.sh
RUN chmod +x /docker-entrypoint.d/tls_certs.sh
COPY nginx/nginx.conf /etc/nginx
RUN mkdir -p /var/www/html/kryvea/static
ENV APP_HOME=/var/www/kryvea/html
RUN mkdir -p $APP_HOME
RUN mkdir -p $APP_HOME/static
WORKDIR $APP_HOME
COPY --from=node-builder /web/dist .
RUN chown -R nginx:nginx /var/www/kryvea
