#!/bin/bash

# 定义要清理的端口
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# 进入app目录
cd "$SCRIPT_DIR/app"

# 遍历所有工程
for dir in */; do
    # 移除目录名末尾的斜杠
    service_name=${dir%/}
    
    # 检查是否存在cmd目录
    if [ -d "$dir/cmd" ]; then
        echo "Starting $service_name..."
        
        # 查找cmd目录下的go文件
        cmd_file=$(find "$dir/cmd" -name "*.go" | head -n 1)
        
        if [ ! -z "$cmd_file" ]; then
            # 在后台运行服务并将输出重定向到日志文件
            rm -f "$dir/$service_name.log"
            nohup go run $cmd_file > "$dir/$service_name.log" 2>&1 &
            echo "$service_name started, log file: $dir/$service_name.log"
        fi
    fi
done

echo "All services started!"