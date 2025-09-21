FROM node:24-alpine3.22 AS build

ARG VITE_API_BASE_URL
ARG VITE_WS_BASE_URL

WORKDIR /app

COPY package.json package-lock.json ./

RUN npm install

COPY . .

RUN echo "VITE_API_BASE_URL=${VITE_API_BASE_URL}" > .env \
    && echo "VITE_WS_BASE_URL=${VITE_WS_BASE_URL}" >> .env

RUN npm run build

FROM nginx:1.29.0-alpine

COPY --from=build /app/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]