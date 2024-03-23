FROM postgres:latest as psql
FROM golang:latest
WORKDIR /app
COPY . .
EXPOSE 1234
CMD ["make", "run"]