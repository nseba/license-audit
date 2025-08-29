FROM node:18-alpine

RUN apk add --no-cache git curl
RUN npm install -g express
RUN pip install flask requests

COPY . /app
WORKDIR /app

RUN apt-get update && apt-get install -y python3-pip
RUN yum install -y nodejs npm