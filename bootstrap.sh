#!/bin/bash -e

# build an docker image named nodecontainer
docker build -t nodecontainer . 

# here be rootfs
mkdir rootfs

# create an container and export it into rootfs directory
docker export $(docker create nodecontainer) | tar -C rootfs -xvf -