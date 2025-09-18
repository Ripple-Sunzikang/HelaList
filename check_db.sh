#!/bin/bash

# 设置数据库连接信息
export PGPASSWORD=suzuki
DB_HOST=localhost
DB_PORT=5432
DB_USER=suzuki
DB_NAME=hela

echo "======================================"
echo "HelaList 数据库结构分析"
echo "======================================"
echo ""

echo "1. 数据库连接测试..."
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version();" > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ 数据库连接成功"
else
    echo "❌ 数据库连接失败"
    exit 1
fi
echo ""

echo "2. 查看所有表..."
echo "======================================"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    schemaname,
    tablename,
    tableowner
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY tablename;
"
echo ""

echo "3. 查看所有表的详细信息..."
echo "======================================"

# 获取所有表名
TABLES=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "
SELECT tablename 
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY tablename;
" | tr -d ' ')

for table in $TABLES; do
    if [ ! -z "$table" ]; then
        echo ""
        echo "📋 表: $table"
        echo "--------------------------------------"
        
        # 查看表结构
        psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
        SELECT 
            column_name,
            data_type,
            is_nullable,
            column_default,
            character_maximum_length
        FROM information_schema.columns 
        WHERE table_name = '$table' 
        AND table_schema = 'public'
        ORDER BY ordinal_position;
        "
        
        # 查看表数据量
        COUNT=$(psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM $table;" | tr -d ' ')
        echo "📊 数据量: $COUNT 行"
        
        # 如果数据量不大，显示前几行样例数据
        if [ "$COUNT" -le 100 ] && [ "$COUNT" -gt 0 ]; then
            echo ""
            echo "🔍 样例数据 (前5行):"
            psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT * FROM $table LIMIT 5;"
        fi
        
        echo ""
    fi
done

echo ""
echo "4. 查看索引信息..."
echo "======================================"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes 
WHERE schemaname = 'public'
ORDER BY tablename, indexname;
"

echo ""
echo "5. 查看外键关系..."
echo "======================================"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT
    tc.table_name, 
    kcu.column_name, 
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name 
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
      AND tc.table_schema = kcu.table_schema
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
      AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY' 
AND tc.table_schema = 'public';
"

echo ""
echo "6. 数据库大小信息..."
echo "======================================"
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
SELECT 
    pg_database.datname,
    pg_size_pretty(pg_database_size(pg_database.datname)) AS size
FROM pg_database
WHERE datname = '$DB_NAME';
"

echo ""
echo "======================================"
echo "数据库分析完成！"
echo "======================================"