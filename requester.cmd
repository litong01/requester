REM The following environment variables can be set
REM    REGISTRY_TOEKN   - the token for the REGISTRY_USERID
@echo off
docker run -it --rm --name requester --network host ^
   -e "REGISTRY_TOKEN=%REGISTRY_TOKEN%" ^
   -v %TEMP%/requester:/home/work/requester ^
   tli551/requesterdt:latest time requester %*
