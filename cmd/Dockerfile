FROM golang:1.15 AS build
ARG gitcommithash
RUN echo "Based on commit: $gitcommithash"
COPY . /go/src/bot
WORKDIR /go/src/bot
RUN make build hash="$gitcommithash"

# 2 шаг - создание контейнера для запуска
FROM debian:buster-slim
WORKDIR /app
COPY --from=build /go/src/bot/bin/* /app/
COPY --from=build /go/src/bot/configs/* /app/configs/
COPY --from=build /go/src/bot/scripts/wait-for-it.sh /app/wait-for-it.sh
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
RUN chmod +x /app/*
EXPOSE 8090/tcp 50105/tcp
CMD /app/bot