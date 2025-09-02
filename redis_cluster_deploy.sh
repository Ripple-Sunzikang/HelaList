#!/bin/bash
# Redis集群+哨兵部署脚本

set -e  # 遇到错误立即退出

# 配置变量
REDIS_VERSION="7.0"
BASE_DIR="/opt/redis-cluster"
LOG_DIR="/var/log/redis-cluster"
DATA_DIR="/var/lib/redis-cluster"

# 服务器配置 (本地模拟集群，使用localhost)
SERVERS=("127.0.0.1")
REDIS_PORTS=(7001 7002 7003 7004 7005 7006)
SENTINEL_PORTS=(27001 27002 27003)

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] $1${NC}"
}

warn() {
    echo -e "${YELLOW}[WARNING] $1${NC}"
}

error() {
    echo -e "${RED}[ERROR] $1${NC}"
    exit 1
}

# 检查Redis是否已安装
check_redis() {
    if ! command -v redis-server &> /dev/null; then
        error "Redis未安装，请先安装Redis"
    fi
    log "Redis已安装: $(redis-server --version)"
}

# 清理旧的Redis进程和数据
cleanup() {
    log "清理旧的Redis进程和数据..."
    
    # 停止所有Redis进程
    for port in "${REDIS_PORTS[@]}" "${SENTINEL_PORTS[@]}"; do
        pkill -f "redis.*:$port" 2>/dev/null || true
    done
    
    # 清理数据目录
    sudo rm -rf $BASE_DIR $LOG_DIR $DATA_DIR
    sudo mkdir -p $BASE_DIR $LOG_DIR $DATA_DIR
    sudo chown -R $USER:$USER $BASE_DIR $LOG_DIR $DATA_DIR
}

# 生成Redis配置文件
generate_redis_config() {
    local port=$1
    local config_file="$BASE_DIR/redis-$port.conf"
    
    cat > $config_file << EOF
# Redis $port 配置
port $port
bind 127.0.0.1
protected-mode no

# 集群配置
cluster-enabled yes
cluster-config-file $DATA_DIR/nodes-$port.conf
cluster-node-timeout 15000
cluster-require-full-coverage no

# 内存配置
maxmemory 100mb
maxmemory-policy allkeys-lru

# 持久化配置
dir $DATA_DIR
appendonly yes
appendfsync everysec
save 900 1
save 300 10
save 60 10000

# 日志配置
loglevel notice
logfile $LOG_DIR/redis-$port.log
daemonize yes
pidfile $DATA_DIR/redis-$port.pid

# 网络配置
tcp-keepalive 300
timeout 0
EOF
    
    log "生成Redis配置文件: $config_file"
}

# 生成Sentinel配置文件
generate_sentinel_config() {
    local port=$1
    local config_file="$BASE_DIR/sentinel-$port.conf"
    
    cat > $config_file << EOF
# Sentinel $port 配置
port $port
bind 127.0.0.1
protected-mode no

# 监控配置 - 等集群创建后再配置
# sentinel monitor master1 127.0.0.1 7001 2
# sentinel monitor master2 127.0.0.1 7003 2  
# sentinel monitor master3 127.0.0.1 7005 2

# 故障检测配置
sentinel down-after-milliseconds master1 5000
sentinel down-after-milliseconds master2 5000
sentinel down-after-milliseconds master3 5000

# 故障转移配置
sentinel parallel-syncs master1 1
sentinel parallel-syncs master2 1
sentinel parallel-syncs master3 1

sentinel failover-timeout master1 60000
sentinel failover-timeout master2 60000
sentinel failover-timeout master3 60000

# 日志配置
logfile $LOG_DIR/sentinel-$port.log
daemonize yes
pidfile $DATA_DIR/sentinel-$port.pid
EOF
    
    log "生成Sentinel配置文件: $config_file"
}

# 启动Redis实例
start_redis_instances() {
    log "启动Redis实例..."
    
    for port in "${REDIS_PORTS[@]}"; do
        generate_redis_config $port
        
        log "启动Redis实例: $port"
        redis-server $BASE_DIR/redis-$port.conf
        
        # 等待启动完成
        sleep 2
        
        # 检查是否启动成功
        if redis-cli -p $port ping | grep -q PONG; then
            log "Redis $port 启动成功"
        else
            error "Redis $port 启动失败"
        fi
    done
}

# 创建Redis集群
create_cluster() {
    log "创建Redis集群..."
    
    # 构建集群节点列表
    local nodes=""
    for port in "${REDIS_PORTS[@]}"; do
        nodes="$nodes 127.0.0.1:$port"
    done
    
    log "集群节点: $nodes"
    
    # 创建集群 (3主3从)
    echo "yes" | redis-cli --cluster create $nodes --cluster-replicas 1
    
    # 验证集群状态
    sleep 3
    redis-cli -p 7001 cluster info
    redis-cli -p 7001 cluster nodes
}

# 配置Sentinel监控
setup_sentinel() {
    log "配置Sentinel监控..."
    
    # 获取主节点信息
    local masters=($(redis-cli -p 7001 cluster nodes | grep master | awk '{print $2}' | cut -d: -f2))
    
    log "检测到的主节点端口: ${masters[@]}"
    
    # 为每个Sentinel生成完整配置
    for i in "${!SENTINEL_PORTS[@]}"; do
        local port=${SENTINEL_PORTS[$i]}
        local config_file="$BASE_DIR/sentinel-$port.conf"
        
        # 重新生成包含监控配置的Sentinel配置
        cat > $config_file << EOF
port $port
bind 127.0.0.1
protected-mode no

# 监控主节点
sentinel monitor master1 127.0.0.1 ${masters[0]} 2
sentinel monitor master2 127.0.0.1 ${masters[1]} 2
sentinel monitor master3 127.0.0.1 ${masters[2]} 2

# 故障检测配置
sentinel down-after-milliseconds master1 5000
sentinel down-after-milliseconds master2 5000
sentinel down-after-milliseconds master3 5000

# 故障转移配置
sentinel parallel-syncs master1 1
sentinel parallel-syncs master2 1
sentinel parallel-syncs master3 1

sentinel failover-timeout master1 60000
sentinel failover-timeout master2 60000
sentinel failover-timeout master3 60000

# 日志配置
logfile $LOG_DIR/sentinel-$port.log
daemonize yes
pidfile $DATA_DIR/sentinel-$port.pid
EOF
        
        log "启动Sentinel: $port"
        redis-sentinel $config_file
        
        sleep 2
        
        # 检查Sentinel是否启动成功
        if redis-cli -p $port ping | grep -q PONG; then
            log "Sentinel $port 启动成功"
        else
            warn "Sentinel $port 启动失败，继续..."
        fi
    done
}

# 验证部署
verify_deployment() {
    log "验证部署状态..."
    
    echo ""
    echo "=== Redis集群状态 ==="
    redis-cli -p 7001 cluster info | grep -E "cluster_state|cluster_slots_assigned|cluster_known_nodes"
    
    echo ""
    echo "=== 集群节点信息 ==="
    redis-cli -p 7001 cluster nodes
    
    echo ""
    echo "=== Sentinel状态 ==="
    for port in "${SENTINEL_PORTS[@]}"; do
        echo "Sentinel $port:"
        redis-cli -p $port sentinel masters 2>/dev/null | head -20 || echo "  未响应"
        echo ""
    done
    
    echo ""
    echo "=== 进程状态 ==="
    ps aux | grep redis | grep -v grep
}

# 生成测试脚本
generate_test_script() {
    cat > $BASE_DIR/test_cluster.sh << 'EOF'
#!/bin/bash
# Redis集群测试脚本

echo "=== 测试Redis集群 ==="

# 测试写入数据
echo "1. 测试数据写入..."
for i in {1..10}; do
    redis-cli -c -p 7001 set "test:key:$i" "value:$i"
done

# 测试读取数据
echo "2. 测试数据读取..."
for i in {1..10}; do
    value=$(redis-cli -c -p 7001 get "test:key:$i")
    echo "test:key:$i = $value"
done

# 测试集群信息
echo "3. 集群信息..."
redis-cli -p 7001 cluster info

echo "4. 测试故障转移..."
echo "请手动停止一个主节点，然后观察Sentinel的反应"
echo "例如: pkill -f 'redis.*:7001'"

EOF
    chmod +x $BASE_DIR/test_cluster.sh
    log "测试脚本已生成: $BASE_DIR/test_cluster.sh"
}

# 主函数
main() {
    log "开始部署Redis集群+哨兵..."
    
    check_redis
    cleanup
    start_redis_instances
    create_cluster
    setup_sentinel
    verify_deployment
    generate_test_script
    
    echo ""
    log "Redis集群+哨兵部署完成！"
    echo ""
    echo "管理命令:"
    echo "  查看集群状态: redis-cli -p 7001 cluster info"
    echo "  查看节点信息: redis-cli -p 7001 cluster nodes"
    echo "  查看Sentinel: redis-cli -p 27001 sentinel masters"
    echo "  运行测试: $BASE_DIR/test_cluster.sh"
    echo ""
    echo "配置文件目录: $BASE_DIR"
    echo "日志目录: $LOG_DIR"
    echo "数据目录: $DATA_DIR"
}

# 执行主函数
main "$@"
