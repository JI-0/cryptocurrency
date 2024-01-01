#!/bin/sh

cd randomx
mkdir lib
git clone https://github.com/tevador/RandomX RandomX
# git checkout "102f8ac"
cd RandomX
mkdir build
cd build
cmake -DARCH=native ..
make && cp librandomx.a ../../lib/