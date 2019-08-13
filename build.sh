# build.sh


# creates a Mac build of the n3 components

#
# note for windows build, this go get update is
# required on mac to pull in necessary dependencies
#
# GOOS=windows go get -u github.com/spf13/cobra
#
# OS:
#   V1 - Just mac
#   V2 - WIP, build for current platform
#   V3 - Cross build all platforms

set -e
CWD=`pwd`

echo "removing existing builds"
rm -r build

mkdir build
mkdir build/download
mkdir build/services

echo "Downloading Influxdb"
curl \
	https://dl.influxdata.com/influxdb/releases/influxdb-1.7.7_darwin_amd64.tar.gz \
	-o build/download/influxdb.tar.gz
cd build/download
tar xvzf influxdb.tar.gz
cp */usr/bin/influx ../services/
cp */usr/bin/influxd ../services/
cp */etc/influxdb/influxdb.conf ../services/
cd $CWD

exit 1

echo "Creating n3 binaries"

GOOS=darwin
GOARCH=amd64
LDFLAGS="-s -w"

GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o build/n3cli ./app/n3cli
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o build/n3dispatcher/n3dispatcher ./app/n3dispatcher
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o build/n3node ./app/n3node

#echo "Creating service binaries"
#
#GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o build/services/gnatsd/gnatsd ../../nats-io/gnatsd
#GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o build/services/liftbridge/liftbridge ../../liftbridge-io/liftbridge
#
#echo "Building influx"
#mkdir ./build/services/influx
#cd $GOPATH/src/github.com/influxdata/influxdb
#go install ./...
## include command shell for inspecting db
#cp $GOPATH/bin/influx $GOPATH/src/github.com/nsip/n3-transport/build/services/influx
## main db daemon
#cp $GOPATH/bin/influxd $GOPATH/src/github.com/nsip/n3-transport/build/services/influx
## include our modified db config file
#cp $GOPATH/src/github.com/nsip/n3-transport/app/test/scripts/influxdb.conf $GOPATH/src/github.com/nsip/n3-transport/build/services/influx
#
#echo "infux built"

echo "build complete"
