#!/bin/bash

VERSION="v0.0.0"

set -e
CWD=`pwd`
OUT="build/Linux64"

if ! [ -x "$(command -v git)" ]; then
	echo "Missing utilitiy: unzip"
	exit 1
fi

echo "removing existing builds"
rm -rf $OUT

mkdir -p $OUT/download
mkdir -p $OUT/services/gnatsd
mkdir -p $OUT/services/influx
mkdir -p $OUT/services/liftbridge
mkdir -p $OUT/n3dispatcher

echo "Downloading NATS ..."
NATS="nats-server-v2.1.0-linux-amd64"
curl \
    -L https://github.com/nats-io/nats-server/releases/download/v2.1.0/$NATS.zip \
    --output $OUT/download/$NATS.zip
cd $OUT/download
unzip $NATS.zip && cd $NATS
cp nats-server ../../services/gnatsd
cd $CWD

echo "Downloading Influx ..."
INFLUX="influxdb-1.7.8-static_linux_amd64"
cd $OUT/download
wget https://dl.influxdata.com/influxdb/releases/$INFLUX.tar.gz && tar xfz $INFLUX.tar.gz
cd ./influxdb*/
cp influx influxd ../../services/influx
cd ../../services/influx && ./influxd &
cd $CWD

echo "Downloading LiftBridge ..."
LIFTBRIDGE=""
cd $OUT/download
LB_REPO="liftbridge-binary"
git clone https://github.com/nsip/$LB_REPO && cd ./$LB_REPO
tar -xzf liftbridge-0.0.1_linux_amd64.tar.gz && cp liftbridge ../../services/liftbridge
cd $CWD

rm -rf $OUT/download

# XXX Make sure these are all "pull" to master
go get -u github.com/cdutwhu/go-util
go get -u github.com/cdutwhu/go-wrappers
go get -u github.com/cdutwhu/go-gjxy

# XXX Move lifgtbride downlaod down here
# XXX Repeat for each of the three
# XXX Zi files

GOOS=linux
GOARCH=amd64
LDFLAGS="-s -w"

echo "Creating N3 binaries @ n3node ..."
cd ./app/n3node
go get
cd $CWD
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o $OUT/n3node ./app/n3node
cd $CWD

echo "Creating N3 binaries @ n3cli ..."
cd ./app/n3cli
go get
cd $CWD
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o $OUT/n3cli ./app/n3cli
cd $CWD

echo "Creating N3 binaries @ n3dispatcher ..."
cd ./app/n3dispatcher
go get
cd $CWD
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o $OUT/n3dispatcher/n3dispatcher ./app/n3dispatcher
cd $CWD


echo "ZIP Linux64"
cd $CWD/build/Linux64
zip -qr ../n3-transport-Linux64-$VERSION.zip *

#echo "ZIP Win64"
#cd $CWD/build/Win64
#zip -qr ../n3-transport-Win64-$VERSION.zip *
#
#echo "ZIP Mac"
##cd $CWD/build/Mac
#zip -qr ../n3-transport-Mac-$VERSION.zip *

echo "Successful, head into $OUT and run n3node. enjoy ... :)"