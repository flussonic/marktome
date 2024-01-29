FROM golang:1.21

ARG GO_VERSION
ENV GO_VERSION=${GO_VERSION}

RUN apt update && DEBIAN_FRONTEND=noninteractive apt install -y wget
RUN mkdir -p /etc/apt/trusted.gpg.d && \
  wget -q -O /etc/apt/trusted.gpg.d/flussonic.gpg http://apt.flussonic.com/binary/gpg.key && \
  echo "deb http://apt.flussonic.com/repo tools/" > /etc/apt/sources.list.d/tools.list && \
  echo "deb http://apt.flussonic.com/repo master/" >> /etc/apt/sources.list.d/flussonic.list

RUN apt-get update && \
  DEBIAN_FRONTEND=noninteractive apt-get install \
    -y \
    --no-install-recommends \
    --no-install-suggests autodeb python3


WORKDIR /usr/src/marktome
COPY go.mod go.sum ./
RUN go mod download && go mod verify

ADD main.go ./
ADD md2json md2json
ADD multiarch.sh tmproot/usr/bin/marktome

ARG VERSION

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -ldflags="-X 'main.Version=${VERSION}'" -o tmproot/usr/bin/x86_64-linux-gnu/marktome main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 \
  go build -ldflags="-X 'main.Version=${VERSION}'" -o tmproot/usr/bin/aarch64-linux-gnu/marktome main.go



ADD DEBIAN tmproot/DEBIAN
RUN sed -i "s/VERSION/${VERSION}/" tmproot/DEBIAN/control
RUN dpkg-deb -Zgzip --build tmproot marktome_${VERSION}_all.deb




