#!/bin/sh

cd randomx
mkdir lib
git clone https://github.com/tevador/RandomX RandomX
# git checkout "102f8ac"
mv RandomX/src/configuration.h RandomX/src/configuration.h.old
cp configuration.h RandomX/src/configuration.h
cd RandomX
mkdir build
cd build
cmake -DARCH=native ..
make && cp librandomx.a ../../lib/
cd ..
mv src/configuration.h.old src/configuration.h