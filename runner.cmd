REM The script to start the data requester application
@echo off
docker run -d --rm --name runner --network host ^
   -e "DATAROOTDIR=/home/requester/data" \
   -e "CONFIG=/home/requester/config.yaml" \
   -v %TEMP%/requester/data:/home/requester/data \
   tli551/requesterdt:latest