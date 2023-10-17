# Собираем в гошке
FROM golang:1.21 as build

ENV BIN_FILE /opt/calendar/calendar-app
ENV MIGRATE_FILE /opt/calendar/migrate
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

ARG LDFLAGS
RUN CGO_ENABLED=0 GOOS=linux go build \
        -ldflags "$LDFLAGS" \
        -o ${BIN_FILE} cmd/calendar/* &&\
    CGO_ENABLED=0 GOOS=linux go build \
        -o ${MIGRATE_FILE} cmd/migrate/*

# На выходе тонкий образ
FROM alpine:3.9

LABEL ORGANIZATION=""
LABEL SERVICE="calendar"
LABEL MAINTAINERS="inenagl@hotmail.com"

ENV BIN_FILE "/opt/calendar/calendar-app"
ENV MIGRATE_FILE /opt/calendar/migrate
COPY --from=build ${BIN_FILE} ${BIN_FILE}
COPY --from=build ${MIGRATE_FILE} ${MIGRATE_FILE}

WORKDIR /go/m
COPY --from=build /go/src/migrations ./migrations

ENV CONFIG_FILE /etc/calendar/config.yaml
COPY ./configs/config.yaml ${CONFIG_FILE}

EXPOSE 8081
EXPOSE 8082

CMD ${MIGRATE_FILE} up migrations && \
${BIN_FILE} -config ${CONFIG_FILE}
