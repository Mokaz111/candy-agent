#!/bin/bash

CURDIR=$(cd $(dirname $0); pwd)
OUTPUT_DIR=$CURDIR/output
BIN_DIR=$OUTPUT_DIR/bin
CONF_DIR=$OUTPUT_DIR/conf
LOG_DIR=$OUTPUT_DIR/log
SCRIPT_DIR=$OUTPUT_DIR/script

# 创建输出目录
mkdir -p $BIN_DIR
mkdir -p $CONF_DIR
mkdir -p $LOG_DIR
mkdir -p $SCRIPT_DIR

# 设置环境变量
export GO_ENV=dev
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct

# 编译
echo "Building candy-agent..."
go build -o $BIN_DIR/candy-agent

# 复制配置文件
echo "Copying configuration files..."
cp -r $CURDIR/conf/* $CONF_DIR/

# 复制启动脚本
echo "Copying bootstrap script..."
cp $CURDIR/script/bootstrap.sh $SCRIPT_DIR/

# 设置执行权限
chmod +x $BIN_DIR/candy-agent
chmod +x $SCRIPT_DIR/bootstrap.sh

echo "Build completed. Output directory: $OUTPUT_DIR"
echo "Run the following command to start the service:"
echo "cd $OUTPUT_DIR && ./script/bootstrap.sh"