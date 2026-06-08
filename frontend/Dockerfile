# Build
FROM node:22-alpine AS builder

WORKDIR /app

ARG VITE_API_URL=http://localhost:8080/api/v1
ARG VITE_WS_URL=ws://localhost:8080/ws/connect

ENV VITE_API_URL=$VITE_API_URL
ENV VITE_WS_URL=$VITE_WS_URL

COPY package.json ./
RUN npm install

COPY . .
RUN npm run build

# Runtime
FROM nginx:1.27-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

HEALTHCHECK --interval=10s --timeout=5s --retries=3 \
  CMD wget -q -O /dev/null http://localhost:80/ || exit 1

CMD ["nginx", "-g", "daemon off;"]
