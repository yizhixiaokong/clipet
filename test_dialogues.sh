#!/bin/bash

# Test dialogues for egg stage
cd /home/kong/code/go/my/clipet

# Clean slate
./clipet reset -y
echo "1\nTest" | ./clipet init > /dev/null 2>&1

echo "=== 蛋阶段对话测试 ==="
echo "在TUI中测试，蛋阶段应该显示简单声响："
echo "咔嗒... 咚... 咔... 等等"

# 进化到幼年
./clipet feed > /dev/null 2>&1
./clipet-dev timeskip --hours 25 > /dev/null 2>&1
echo
echo "=== 进化到幼年 ==="
echo "现在应该进化到幼年阶段，对话变为简单的喵喵声："
echo "喵~ 喵喵~ 喵呜~ 噜噜~ 等等"

echo "测试完成！"
