#!/usr/bin/env bash
# 设置一个变量，用于存储 Docker 镜像的名称和版本号
image=openim/open_im_server:v1.0.5
# 删除 Open-IM-Server 文件夹，并强制删除其中所有内容
rm Open-IM-Server -rf
# 从远程 Git 仓库中克隆 Open-IM-Server 代码库，--recursive 选项用于同时克隆子模块
git clone https://github.com/bing-byte-9527/Open-IM-Server.git --recursive
# 切换到 Open-IM-Server 代码库的 tuoyun 分支
cd Open-IM-Server
git checkout main
# 切换到 Open-IM-SDK-Core 子目录，该子目录包含了需要编译的代码
cd cmd/Open-IM-SDK-Core/
git checkout main
# 切换回 Open-IM-Server 代码库的根目录
cd ../../
# 构建 Docker 镜像，-t 选项用于指定镜像名称和版本号，-f 选项用于指定 Dockerfile 的位置
docker build -t  $image . -f deploy.Dockerfile
# 推送 Docker 镜像到 Docker Hub 仓库，该命令将会上传构建好的 Docker 镜像到远程 Docker 仓库
docker push $image
