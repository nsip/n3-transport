#!/bin/bash

VERSION="v0.0.0"

set -e
CWD=`pwd`

if ! [ -x "$(command -v git)" ]; then
	echo "Missing utilitiy: unzip"
	exit 1
fi

echo "removing existing builds"
rm -rf ./build

## Directories

OUT="build/Linux64"
mkdir -p $OUT/download
mkdir -p $OUT/services/gnatsd
mkdir -p $OUT/services/influx
mkdir -p $OUT/services/liftbridge
mkdir -p $OUT/n3dispatcher

OUT="build/Win64"
mkdir -p $OUT/download
mkdir -p $OUT/services/gnatsd
mkdir -p $OUT/services/influx
mkdir -p $OUT/services/liftbridge
mkdir -p $OUT/n3dispatcher

OUT="build/Mac"
mkdir -p $OUT/download
mkdir -p $OUT/services/gnatsd
mkdir -p $OUT/services/influx
mkdir -p $OUT/services/liftbridge
mkdir -p $OUT/n3dispatcher

## NATS

OUT="build/Linux64"
echo "Downloading NATS (Linux64) ..."
NATS="nats-server-v2.1.0-linux-amd64"
curl \
    -L https://github.com/nats-io/nats-server/releases/download/v2.1.0/$NATS.zip \
    --output $OUT/download/$NATS.zip
cd $OUT/download
unzip $NATS.zip && cd $NATS
cp nats-server ../../services/gnatsd
cd $CWD

OUT="build/Win64"
echo "Downloading NATS (Win64) ..."
NATS="nats-server-v2.1.0-windows-amd64"
curl \
    -L https://github.com/nats-io/nats-server/releases/download/v2.1.0/$NATS.zip \
    --output $OUT/download/$NATS.zip
cd $OUT/download
unzip $NATS.zip && cd $NATS
cp nats-server.exe ../../services/gnatsd
cd $CWD

OUT="build/Mac"
echo "Downloading NATS (Mac) ..."
NATS="nats-server-v2.1.0-darwin-amd64"
curl \
    -L https://github.com/nats-io/nats-server/releases/download/v2.1.0/$NATS.zip \
    --output $OUT/download/$NATS.zip
cd $OUT/download
unzip $NATS.zip && cd $NATS
cp nats-server ../../services/gnatsd
cd $CWD

## Influx

OUT="build/Linux64"
echo "Downloading Influx (Linux64) ..."
INFLUX="influxdb-1.7.8-static_linux_amd64"
cd $OUT/download
wget https://dl.influxdata.com/influxdb/releases/$INFLUX.tar.gz && tar xfz $INFLUX.tar.gz
cd ./influxdb*/
cp influx influxd ../../services/influx
# cd ../../services/influx && ./influxd &
cd $CWD

OUT="build/Win64"
echo "Downloading Influx (Win64) ..."
INFLUX="influxdb-1.7.8_windows_amd64"
cd $OUT/download
wget https://dl.influxdata.com/influxdb/releases/$INFLUX.zip && unzip $INFLUX.zip
cd ./influxdb*/
cp influx.exe influxd.exe ../../services/influx
# cd ../../services/influx && ./influxd.exe &
cd $CWD

OUT="build/Mac"
echo "Downloading Influx (Mac) ..."
INFLUX="influxdb-1.7.8_darwin_amd64"
cd $OUT/download
wget https://dl.influxdata.com/influxdb/releases/$INFLUX.tar.gz && tar xfz $INFLUX.tar.gz
cd ./influxdb*/usr/bin/
cp influx influxd ../../../../services/influx
# cd ../../services/influx && ./influxd &
cd $CWD

## LiftBridge

mkdir -p ./build/download
cd ./build/download
echo "Downloading LiftBridge (Linux64, Win64 & Mac) ..."
LB_REPO="liftbridge-binary"
LB_VER="0.0.1"
git clone https://github.com/nsip/$LB_REPO && cd ./$LB_REPO
tar -xzf liftbridge-"$LB_VER"_linux_amd64.tar.gz && mv liftbridge ../../Linux64/services/liftbridge
tar -xzf liftbridge-"$LB_VER"_windows_amd64.tar.gz && mv liftbridge.exe ../../Win64/services/liftbridge
tar -xzf liftbridge-"$LB_VER"_darwin_amd64.tar.gz && mv liftbridge ../../Mac/services/liftbridge
cd $CWD

## remove temporary download directories

rm -rf ./build/download
rm -rf ./build/Linux64/download
rm -rf ./build/Win64/download
rm -rf ./build/Mac/download

# XXX Make sure these are all "pull" to master
go get -u github.com/cdutwhu/go-util
go get -u github.com/cdutwhu/go-wrappers
go get -u github.com/cdutwhu/go-gjxy

# XXX Move lifgtbride downlaod down here
# XXX Repeat for each of the three
# XXX Zi files

GOARCH=amd64
LDFLAGS="-s -w"

# go build Linux

OUT="build/Linux64"
GOOS=linux

echo "Creating N3 binaries @ n3node (Linux64) ..."
cd ./app/n3node
go get
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o $CWD/$OUT/n3node ./
cd $CWD

echo "Creating N3 binaries @ n3cli (Linux64) ..."
cd ./app/n3cli
go get
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o $CWD/$OUT/n3cli ./
cd $CWD

echo "Creating N3 binaries @ n3dispatcher (Linux64) ..."
cd ./app/n3dispatcher
go get
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o $CWD/$OUT/n3dispatcher/n3dispatcher ./
cd $CWD

# go build Windows

OUT="build/Win64"
GOOS=windows

echo "Creating N3 binaries @ n3node (Win64) ..."
cd ./app/n3node
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o $CWD/$OUT/n3node.exe ./
cd $CWD

echo "Creating N3 binaries @ n3cli (Win64) ..."
cd ./app/n3cli
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o $CWD/$OUT/n3cli.exe ./
cd $CWD

echo "Creating N3 binaries @ n3dispatcher (Win64) ..."
cd ./app/n3dispatcher
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o $CWD/$OUT/n3dispatcher/n3dispatcher.exe ./
cd $CWD

# go build Mac

OUT="build/Mac"
GOOS=darwin

echo "Creating N3 binaries @ n3node (Mac) ..."
cd ./app/n3node
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o $CWD/$OUT/n3node ./
cd $CWD

echo "Creating N3 binaries @ n3cli (Mac) ..."
cd ./app/n3cli
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o $CWD/$OUT/n3cli ./
cd $CWD

echo "Creating N3 binaries @ n3dispatcher (Mac) ..."
cd ./app/n3dispatcher
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o $CWD/$OUT/n3dispatcher/n3dispatcher ./
cd $CWD

# ZIP

echo "ZIP Linux64"
cd $CWD/build/Linux64
zip -qr ../n3-transport-Linux64-$VERSION.zip *

echo "ZIP Win64"
cd $CWD/build/Win64
zip -qr ../n3-transport-Win64-$VERSION.zip *

echo "ZIP Mac"
cd $CWD/build/Mac
zip -qr ../n3-transport-Mac-$VERSION.zip *

echo "Successful, head into build/your-os and run n3node. enjoy ... :)"
