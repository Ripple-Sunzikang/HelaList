# 目录结构概述

## 文件总结构

```
HelaList/
├── api/              # API定义
├── build/            # 构建
├── cmd/              # 命令行实现
├── configs/          # 配置文件
├── docs/             # 项目文档
├── internal/         # 后端核心代码
├── pkg/              # 可被外部项目引用的公共库代码
├── test/             # 集成测试
├── vue/              # 前端实现
├── ARCHITECTURE.md   # 目录结构概述
├── go.mod            # Go 模块依赖文件
├── go.sum            # Go 模块校验和
└── README.md         # 项目介绍
```

## internal 后端核心代码

```
internal/
├── handler/          # (表示层)Gin的HTTP处理器
│   ├── middleware/   # 中间件(如认证、日志)
│   └── router.go     # 注册所有路由
│
├── service/          # (业务逻辑层)核心业务逻辑
│
├── repository/       # (数据访问层)数据库操作
│   ├── user_repository.go  # 封装对用户集合的 MongoDB 操作
│   └── file_repository.go  # 封装对文件/目录元数据集合的操作
│
├── model/            # (模型层)定义数据结构(DO，DTO，VO)
│   ├── file.go       # 文件/目录的数据结构
│   ├── request.go    # 请求体结构
│   └── response.go   # 响应体结构
│   └── user.go       # 用户的数据结构
│
└── pkg/              # 项目内部共享的工具包
```

## vue 前端核心代码

```
vue/
├── public/
├── src/
│   ├── api/          # 调用后端API的封装(例如axios)
│   ├── assets/       # 静态资源
│   ├── components/   # 可复用的Vue组件
│   ├── router/       # 路由配置
│   ├── stores/       # 状态管理(如Pinia或Vuex)
│   ├── views/        # 页面级组件
│   └── main.js       # Vue 应用入口
├── .env.development  # 开发环境变量
├── .env.production   # 生产环境变量
├── package.json
└── vite.config.js    # 或 vue.config.js
```

## docs 文档

```
docs
├── database          # 数据库文档
└── rup               # RUP开发过程文档
    ├── 1_inception   # 构思阶段
    ├── 2_elaboration # 细化阶段
    ├── 3_construction  # 构建阶段
    └── 4_transition  # 移交阶段
```