# build.sh


# creates a Mac build of the n3 components

set -e
CWD=`pwd`


echo "Creating n3 binaries"

GOOS=darwin
GOARCH=amd64
LDFLAGS="-s -w"


GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o build/n3cli ./app/n3cli
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o build/dispatcher/dispatcher ./app/dispatcher
GOOS="$GOOS" GOARCH="$GOARCH" go build -ldflags="$LDFLAGS" -o build/node ./app/node

