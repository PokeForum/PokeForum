---
trigger: always_on
---

## 任务须知
- 项目依赖库：Go+Gin+EntORM
- 项目依赖中间件：PgSQL+Redis
- 参数验证使用 github.com/go-playground/validator/v10 库实现
- 请求响应体代码需要在Schema目录下创建
- 接口需要添加符合 gin-swagger 要求的注释, 方便生成API文档
- API接口请遵守RESTAPI风格
- 项目数据表设计严禁使用外键关联, 一切相关逻辑需要在应用层使用逻辑查询