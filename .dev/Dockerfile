FROM golang:1.20

ENV GOPROXY=https://goproxy.cn,direct

RUN apt-get update && apt-get install -y sudo curl gnupg

RUN curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -

RUN apt-get install -y nodejs

RUN npm config set registry https://registry.npmmirror.com

RUN npm install -g npm

COPY .dev/conf /opt/1panel/conf