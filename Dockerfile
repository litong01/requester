FROM node:current-alpine3.18
LABEL maintainer="litong01"

RUN npm install -g newman
RUN mkdir -p /home/requester

WORKDIR /home/requester
CMD /bin/sh