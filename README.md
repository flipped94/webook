## 登录实现

登录本身分为两部分

- 实现登录功能 (/users/login接口)
- 登录态校验 (Cookie, Session, JWT)

### 登录态实现

```sequence
participant 前端 as A
participant JWT登录校验 as B
participant 登录接口 as C
participant 其他接口 as D
A->>C: 登录
C->>C: 生成token
C->>A: 返回token

A->>B: token
B->>B: 校验token通过

B->>D:
D->>D:
D->>A:
```



过程：

- 在登录接口中，登录成功后生成 JWT token。
  - 在 JWT token 中写入数据。
  - 把 JWT token 通过 HTTP Response Header `x-jwt-token` 返回。
- HTTP请求时携带 JWT token。
- 登录校验 Gin middleware。
  - 读取 JWT token。
  - 验证 JWT token 是否合法。

### 验证码登录

除了登录功能，修改密码、危险操作二次验证等业务都需要用到验证码。

所以，手机验证码应该是一个独立的功能。

- 如果是模块，那么它是一个独立的模块。
- 如果是微服务，那么它是一个独立的微服务。

手机验证码要通过短信来发送， 那么短信也会被别的业务使用，另外可能换短信供应商，比如腾讯云和阿里云。

综合考虑，以下两点：

- 不同业务都要用短信功能和验证码功能
- 可能换供应商

![短信功能设计](./resources/png/sms.png)