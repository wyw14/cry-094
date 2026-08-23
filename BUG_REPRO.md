# Bug 是什么

包含隐式回边的依赖图没有阻断编排，也没有稳定保留完整的循环证据链。

# 如何触发

创建 prepare、deploy、verify 三个节点，加入 prepare 到 deploy、deploy 到 verify 以及 verify 到 prepare 的隐式依赖，再生成分析结果。

# 错误信息

`cyclic dependency graph must not be orchestratable`
