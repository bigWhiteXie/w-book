#!/bin/bash

# 定义要清理的端口
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PORTS=(20800 20810 20850 20860 20830 20840 20821 20871 20872)

# 清理指定端口的进程
echo "Cleaning up processes..."
for port in "${PORTS[@]}"; do
    pid=$(lsof -ti:$port)
    if [ ! -z "$pid" ]; then
        echo "Killing process on port $port (PID: $pid)"
        kill -9 $pid
    fi
done