#!/bin/bash

# 进入项目根目录（根据实际情况调整）
cd /usr/local/go_project/w-book

# 查找并删除所有 .log 文件
find app -type f -name "*.log" -exec rm -v {} \;

echo "清理完成"