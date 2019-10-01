#!/bin/bash

# delete liftbridge cache
rm -rf ./build/services/liftbridge/services/
rm -f ./build/services/influx/influxlog.txt

# delete all binary files
find . -type f -executable -exec sh -c "file -i '{}' | grep -q 'x-executable; charset=binary'" \; -print | xargs rm -f