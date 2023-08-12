REM The following environment variables can be set
REM    REGISTRY_TOEKN   - the token for the REGISTRY_USERID
@echo off
docker run -d --rm --name requester --network host ^
   -e "DATAROOTDIR=/home/requester/data" \
   -e "CONFIG=/home/requester/config.yaml" \
   -v %TEMP%/requester/data:/home/requester/data \
   tli551/requesterdt:latest