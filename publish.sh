#!/bin/bash
# git_auto_commit.sh

# 获取当前时间作为 commit message，格式：2026-10-03 15:30:22
TIME_MSG=$(date '+%Y-%m-%d %H:%M:%S')

echo "===== Git自动提交脚本 ====="
echo "Commit 信息：$TIME_MSG"

# 检查是否有改动
git status --porcelain
if [[ -z $(git status --porcelain) ]]; then
    echo "✅ 没有文件改动，无需提交"
    exit 0
fi

echo "📝 存在改动，开始 add ..."
git add .

echo "📌 执行 commit"
git commit -m "$TIME_MSG"

echo "🚀 推送到远程 origin"
git push origin

# 检查push结果
if [ $? -eq 0 ]; then
    echo "✅ 提交推送成功！"
else
    echo "❌ push失败，请检查网络/冲突"
    exit 1
fi
