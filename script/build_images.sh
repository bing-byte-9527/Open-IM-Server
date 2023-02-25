#!/usr/bin/env bash
image=openim/open_im_server:v1.0.5
rm Open-IM-Server -rf
git clone https://github.com/bing-byte-9527/Open-IM-Server.git --recursive
cd Open-IM-Server
git checkout tuoyun
cd cmd/Open-IM-SDK-Core/
git checkout tuoyun
cd ../../
docker build -t  $image . -f deploy.Dockerfile
docker push $image