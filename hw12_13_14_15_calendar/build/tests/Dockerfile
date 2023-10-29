FROM golang:1.21

ENV CODE_DIR /go/src/
ENV TEST_DIR /go/src/internal/tests

COPY ./build/bin/wait-for-it.sh /bin/wait-for-it.sh
RUN chmod +x /bin/wait-for-it.sh

WORKDIR ${CODE_DIR}
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . ${CODE_DIR}

WORKDIR ${TEST_DIR}

CMD go test -tag integration -race .