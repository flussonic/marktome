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
RUN go build -v

ADD DEBIAN tmproot/DEBIAN
RUN mkdir -p tmproot/usr/bin/ && mv marktome tmproot/usr/bin/
ARG VERSION
RUN sed -i "s/VERSION/${VERSION}/" tmproot/DEBIAN/control
RUN dpkg-deb -Zgzip --build tmproot marktome_${VERSION}_all.deb




