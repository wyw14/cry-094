# Bug 是什么

同一枚刷新令牌在并发轮换时可以被消费两次，并产生两套成功的后继凭据。

# 如何触发

用启动门闩让两个请求同时读取同一枚有效刷新令牌，再同时放行两个轮换操作。

# 错误信息

`same refresh token produced 2 successful rotations`
