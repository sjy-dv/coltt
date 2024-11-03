FROM golang:1.23 AS builder

ARG RELEASE_TAG=v2.16.0
ARG ARCH=amd64
ARG USEARCH_VERSION=2.16.0


RUN apt-get update && apt-get install -y wget dpkg

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN wget https://github.com/unum-cloud/usearch/releases/download/${RELEASE_TAG}/usearch_linux_${ARCH}_${USEARCH_VERSION}.deb \
    && dpkg -i usearch_linux_${ARCH}_${USEARCH_VERSION}.deb || apt-get install -f -y \
    && ldconfig \
    && ls -l /usr/local/lib/libusearch_c.so

RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o bin/sample main.go

# usearch required GLIBC >= 2.32
FROM debian:bookworm-slim

COPY --from=builder /usr/local/lib/ /usr/local/lib/
COPY --from=builder /app/bin/sample /bin/sample

RUN apt-get update && apt-get install -y libgcc1 \
    && rm -rf /var/lib/apt/lists/*

ENV LD_LIBRARY_PATH=/usr/local/lib

RUN chmod +x /bin/sample

ENTRYPOINT ["/bin/sample"]
