# Bug 是什么

同一工具被多次调用时，声明的版本约束会在依赖合并后丢失，旧版本主机没有被阻断。

# 如何触发

分析一份声明 jq >=2.1 且连续调用两次 jq 的脚本，目标主机只提供 jq 2.0。

# 错误信息

`incompatible repeated tool use produced status ready`
