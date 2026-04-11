# ---- Stage 1: Build ----
FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

# Pass the API URL as a build arg (populated from docker-compose environment)
ARG VITE_API_URL=http://localhost:4000
ENV VITE_API_URL=$VITE_API_URL

COPY . .
RUN npm run build

# ---- Stage 2: Serve ----
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html

# nginx config for React Router (handle client-side routes)
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
