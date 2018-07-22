#!/bin/bash

# validating a version was provided
if [ "$#" -eq 0 ]; then
    echo "no version value provided"
    exit
fi

# update version of exectuable for build
sed -i "" 's/dev/'"$1"'/g' version.go

# general build for linux
# generate binary
evn GOOS=linux GOARCH=amd64 go build -o ./install/usr/bin

# generate rpm
cd ./install
fpm -s dir -t rpm -n slurp-rtl_433 -v $1 ./
mv slurp-rtl_433.* ../builds/

# generate deb
fpm -s dir -t deb -n slurp-rtl_433 -v $1 ./
mv slurp-rtl_433.* ../builds/

# revert version
sed -i "" 's/'"$1"'/dev/g' version.go