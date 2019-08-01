# build.sh


# creates a Mac build of the n3 components

# 
# note for windows build, this go get update is
# required on mac to pull in necessary dependencies
# 
# GOOS=windows go get -u github.com/spf13/cobra
# 

set -e
CWD=`pwd`


echo "removing existing builds"
rm -rf build

echo "Creating n3 binaries"

GOOS=darwin
GOARCH=amd64
LDFLAGS="-s -w"

GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o build/n3cli ./app/n3cli
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o build/n3dispatcher/n3dispatcher ./app/n3dispatcher
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o build/n3node ./app/n3node

echo "Creating service binaries"

GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o build/services/gnatsd/gnatsd ../../nats-io/gnatsd
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o build/services/liftbridge/liftbridge ../../liftbridge-io/liftbridge

echo "Libraries"
cd n3crypto; go get; cd ..
cd n3config; go get; cd ..
cd n3dispatcher; go get; cd ..
cd n3liftbridge; go get; cd ..

echo "Building influx"
mkdir ./build/services/influx
cd $HOME/go/src/github.com/influxdata/influxdb
go install ./...
# include command shell for inspecting db
cp $HOME/go/bin/influx $HOME/go/src/github.com/nsip/n3-transport/build/services/influx
# main db daemon
cp $HOME/go/bin/influxd $HOME/go/src/github.com/nsip/n3-transport/build/services/influx
# include our modified db config file
cp $HOME/go/src/github.com/nsip/n3-transport/app/test/scripts/influxdb.conf $HOME/go/src/github.com/nsip/n3-transport/build/services/influx

echo "infux built"

echo "build complete"
