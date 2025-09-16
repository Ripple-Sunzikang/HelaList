#!/bin/bash
# Redis集群清理脚本

BASE_DIR="/opt/redis-cluster"
LOG_DIR="/var/log/redis-cluster"
DATA_DIR="/var/lib/redis-cluster"

REDIS_PORTS=(7001 7002 7003 7004 7005 7006)
SENTINEL_PORTS=(27001 27002 27003)

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

echo "开始清理Redis集群..."

# 停止所有Redis和Sentinel进程
log "停止Redis和Sentinel进程..."
for port in "${REDIS_PORTS[@]}" "${SENTINEL_PORTS[@]}"; do
    log "停止端口 $port 上的Redis进程"
    pkill -f "redis.*:$port" 2>/dev/null || true
    pkill -f "redis.*$port" 2>/dev/null || true
done

# 等待进程完全停止
sleep 3

# 强制杀死残留进程
pkill -f "redis-server" 2>/dev/null || true
pkill -f "redis-sentinel" 2>/dev/null || true

# 清理目录
log "清理数据目录..."
sudo rm -rf $BASE_DIR $LOG_DIR $DATA_DIR

# 检查是否还有Redis进程
remaining_processes=$(ps aux | grep redis | grep -v grep | wc -l)
if [ $remaining_processes -eq 0 ]; then
    log "所有Redis进程已清理完毕"
else
    warn "仍有 $remaining_processes 个Redis进程运行"
    ps aux | grep redis | grep -v grep
fi

log "Redis集群清理完成！"
