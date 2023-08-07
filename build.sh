#!/bin/bash
# build and remove dangling images in the system
docker build -t tli551/requester:latest .
docker rmi -f $(docker images -f "dangling=true" -q)