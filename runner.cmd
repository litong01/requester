REM The script to start the data requester application
@echo off
docker run -d --rm --name runner ^
   -v %TEMP%/data:/home/requester/data ^
   tli551/requesterdt:latest

docker run -dit --rm --name drill ^
   -p 8047:8047 -p 31010:31010 ^
   -v %TEMP%/data:/tmp/data ^
   --platform=linux/amd64 apache/drill:master-openjdk-17