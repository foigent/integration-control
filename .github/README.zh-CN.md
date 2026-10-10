# Foigent integration-control

[English](README.md) · [贡献指南（英文）](../CONTRIBUTING.md) · [许可证](../LICENSE)

integration-control 是独立于具体产品的离线 Go 工具与合同库，用于检查任务清单、
产品确认和授权材料的身份、规范化字节、摘要及签名。UniRouter 和 X-Insight
各自在经过审阅的路径中使用这些事实。

它只处理本地文件，不连接产品、不持有私钥、不签署授权、不操作数据库或部署服务。
校验通过本身不授予执行产品操作的权限。

## 开始使用

仓库公开后，安装 Go 1.25.12 并执行：

```sh
git clone https://github.com/foigent/integration-control.git
cd integration-control
GOTOOLCHAIN=local go test ./... -count=1
GOTOOLCHAIN=local go run ./cmd/foigent-integration-control schemas check --dir contracts/v1
```

本地测试使用合成样例，无需模型密钥、外部服务或付费调用。
使用自己的材料前，请阅读[合同说明](../docs/CONTRACTS.md)。

项目以英文作为第一语言，中文作为第二语言；两种语言的提问和贡献均欢迎。
参与方式与正式分发校验的边界见[贡献指南](../CONTRIBUTING.md)。
行为投诉使用[行为准则](../CODE_OF_CONDUCT.md)中的私密联系方式；
漏洞请按[安全报告说明](../SECURITY.md)处理。

项目自有内容采用与 UniRouter 相同的 Apache-2.0，第三方材料保留原许可。
N1 `v0.1.0-candidate.1` 的历史源码与五项发布资产保持原有身份；
许可随附方式见[许可说明](../docs/LICENSING.md)。本分支仍处于开源准备阶段，
不代表仓库已公开、匿名下载已验证或新的产品运行行为已发布。
