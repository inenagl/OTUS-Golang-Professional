# Собираем в гошке
FROM golang:1.21 as build

ENV BIN_FILE /opt/calendar/calendar-sender
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

# Кэшируем слои с модулями
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY ../.. ${CODE_DIR}

ARG LDFLAGS
RUN CGO_ENABLED=0 GOOS=linux go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/calendar_sender/*

# На выходе тонкий образ
FROM alpine:3.9

LABEL ORGANIZATION=""
LABEL SERVICE="calendar_sender"
LABEL MAINTAINERS="inenagl@hotmail.com"

ENV BIN_FILE "/opt/calendar/calendar-sender"
COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENV CONFIG_FILE /etc/calendar/config.yaml
COPY ./configs/config_sender.yaml ${CONFIG_FILE}

CMD ${BIN_FILE} -config ${CONFIG_FILE}
