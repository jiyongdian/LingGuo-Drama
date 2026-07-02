FROM node:20-alpine AS builder

WORKDIR /src

COPY package*.json ./
RUN npm ci

COPY . .

ARG VITE_API_URL=
ARG VITE_API_URL_PREFIX=/admin/v1
ENV VITE_API_URL=${VITE_API_URL}
ENV VITE_API_URL_PREFIX=${VITE_API_URL_PREFIX}

RUN npm run build

FROM nginx:1.27-alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /src/dist /usr/share/nginx/html

EXPOSE 80
