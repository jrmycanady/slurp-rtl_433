#!/bin/bash

# validating a version was provided
if [ "$#" -eq 0 ]; then
    echo "no version value provided"
    exit
fi

BUILD="$(git rev-parse --verify --short HEAD)"

# update version of exectuable for build
sed -i 's/dev/'"$1"'/g' version.go

# general build for linux
# generate binary
env GOOS=linux GOARCH=amd64 go build -o ./install/usr/bin/slurp-rtl_433

# generate rpm
cd ./install
fpm -s dir -t rpm -n slurp-rtl_433 --config-files ./etc/slurp-rtl_433/config.toml -v $1 ./
mv ./slurp-rtl_433* ../builds/

# generate deb
fpm -s dir -t deb -n slurp-rtl_433 --config-files ./etc/slurp-rtl_433/config.toml -v $1 ./
mv ./slurp-rtl-433* ../builds/

# revert version
cd ..
sed -i 's/'"$1"'/dev/g' version.go
