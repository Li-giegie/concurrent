#!/bin/bash

# 循环执行，直到 git push 成功
while true; do
    echo "====================================="
    echo "开始执行 git push origin main"
    echo "====================================="
    git push origin main
    # 获取上一条命令退出码
    ret=$?
    if [ $ret -eq 0 ]; then
        echo "✅ push 成功，退出循环"
        break
    fi
    echo "❌ push 失败，3秒后重试..."
    sleep 3
done
