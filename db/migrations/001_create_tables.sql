-- Generated migration: create core tables (metas, storages, users, objects)
-- Assumptions:
--  - uuid columns use uuid_generate_v4() if available; code uses uuid v7, but we use uuid as type
--  - timestamps stored as timestamptz where appropriate

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- metas
CREATE TABLE IF NOT EXISTS metas (
    id uuid PRIMARY KEY,
    path text UNIQUE NOT NULL,
    password text,
    p_sub boolean DEFAULT false,
    write boolean DEFAULT false,
    w_sub boolean DEFAULT false,
    hide text,
    h_sub boolean DEFAULT false,
    readme text,
    r_sub boolean DEFAULT false,
    header text,
    header_sub boolean DEFAULT false
);

-- storages
CREATE TABLE IF NOT EXISTS storages (
    id uuid PRIMARY KEY,
    mount_path text UNIQUE NOT NULL,
    "order" integer DEFAULT 0,
    driver text,
    cache_expiration integer,
    status text,
    addition text,
    remark text,
    modified_time timestamptz,
    disabled boolean DEFAULT false,
    order_by text,
    order_direction text,
    extract_folder text
);

-- users
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    username varchar(50) UNIQUE NOT NULL,
    email varchar(100) UNIQUE NOT NULL,
    password_hash text NOT NULL,
    salt text UNIQUE NOT NULL,
    password text,
    base_path text,
    identity integer NOT NULL,
    disabled boolean NOT NULL DEFAULT false,
    password_ts bigint
);

-- objects
CREATE TABLE IF NOT EXISTS objects (
    id uuid PRIMARY KEY,
    path text,
    name text,
    size bigint,
    modified_time timestamptz,
    created_time timestamptz,
    is_folder boolean
);

-- indexes (additional)
CREATE INDEX IF NOT EXISTS idx_objects_path ON objects(path);
CREATE INDEX IF NOT EXISTS idx_storages_mount_path ON storages(mount_path);
CREATE INDEX IF NOT EXISTS idx_metas_path ON metas(path);

-- test data (参考 test/ 脚本)
-- 注意: 密码哈希应通过应用层生成；这里插入占位 hash/salt 以便测试查询。
INSERT INTO users (id, username, email, password_hash, salt, base_path, identity, disabled, password_ts)
VALUES
    (gen_random_uuid(), 'suzuki', 'suzuki@example.local', 'placeholder_hash', 'placeholder_salt_suzuki', '/home/suzuki', 0, false, extract(epoch from now())::bigint),
    (gen_random_uuid(), 'admin', 'admin@example.local', 'placeholder_hash', 'placeholder_salt_admin', '/home/admin', 1, false, extract(epoch from now())::bigint),
    (gen_random_uuid(), 'user1', 'user1@example.local', 'placeholder_hash', 'placeholder_salt_user1', '/home/user1', 0, false, extract(epoch from now())::bigint)
ON CONFLICT (username) DO NOTHING;

INSERT INTO storages (id, mount_path, "order", driver, cache_expiration, status, addition, remark, modified_time, disabled)
VALUES (
    gen_random_uuid(),
    '/localwebdav',
    1,
    'webdav',
    3600,
    'work',
    '{"address":"https://dav.jianguoyun.com/dav","username":"1063046101@qq.com","password":"ae3yjvwhptqjpsh8","tls_insecure_skip_verify":false}',
    '挂载本地 WebDAV 服务',
    now(),
    false
)
ON CONFLICT (mount_path) DO NOTHING;

INSERT INTO metas (id, path, password, p_sub, write, w_sub, hide, h_sub, readme, r_sub, header, header_sub)
VALUES (
    gen_random_uuid(),
    '/localwebdav',
    NULL,
    false,
    true,
    false,
    NULL,
    false,
    '测试 meta',
    false,
    NULL,
    false
)
ON CONFLICT (path) DO NOTHING;
