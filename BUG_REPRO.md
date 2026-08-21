# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA；不要在当前修复结果源码上期待重新出现修复前失败。生成系统使用的可信验证补丁和完整验证日志仅在本地留存，不提交到结果分支。

## 问题现象

财务对账员导入一组运费结算借贷行，第二行的外部编号早已入账，接口按预期返回冲突；但刷新账本发现第一条新借记已经落下，余额少了四十，整批重试又会被这条残留拦住。另一组编号全新的借贷行可以正常平衡。请先不要修改代码，查清失败批次为什么不是全有或全无，说明对账转换、逐行入账和错误返回之间的状态时序与影响范围。

## 含 Bug 版本

- 仓库：VanceMichael/go-label-21
- 仓库地址：https://github.com/VanceMichael/go-label-21.git
- parent SHA：d637f218a7826e21193f6c53588a39be90d5c13d

## 复现步骤

```bash
git clone -- https://github.com/VanceMichael/go-label-21.git bug-repro
cd bug-repro
git checkout --detach d637f218a7826e21193f6c53588a39be90d5c13d
go test ./internal/reconciliation -run ^TestFailedSettlementImportDoesNotLeavePartialLedgerEntries$ -count=1
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/reconciliation -run ^TestFailedSettlementImportDoesNotLeavePartialLedgerEntries$ -count=1
--- FAIL: TestFailedSettlementImportDoesNotLeavePartialLedgerEntries (0.00s)
    settlement_test.go:28: failed settlement changed ledger: entries=[{ID:existing-credit TenantID:tenant-1 ShipmentID:shipment-old Currency:CNY Debit:0 Credit:100 Memo: PostedAt:2026-08-21 06:27:03.236168882 +0000 UTC} {ID:new-debit TenantID:tenant-1 ShipmentID:shipment-1 Currency:CNY Debit:40 Credit:0 Memo: PostedAt:2026-08-21 07:27:03.236168882 +0000 UTC}] balance=60
FAIL
FAIL	github.com/VanceMichael/go-base-airbridge/internal/reconciliation	0.037s
FAIL

```

stderr：

```text
(empty)
```

### linux/arm64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/reconciliation -run ^TestFailedSettlementImportDoesNotLeavePartialLedgerEntries$ -count=1
--- FAIL: TestFailedSettlementImportDoesNotLeavePartialLedgerEntries (0.00s)
    settlement_test.go:28: failed settlement changed ledger: entries=[{ID:existing-credit TenantID:tenant-1 ShipmentID:shipment-old Currency:CNY Debit:0 Credit:100 Memo: PostedAt:2026-08-21 06:27:40.879391594 +0000 UTC} {ID:new-debit TenantID:tenant-1 ShipmentID:shipment-1 Currency:CNY Debit:40 Credit:0 Memo: PostedAt:2026-08-21 07:27:40.879391594 +0000 UTC}] balance=60
FAIL
FAIL	github.com/VanceMichael/go-base-airbridge/internal/reconciliation	0.002s
FAIL

```

stderr：

```text
(empty)
```

## 通过条件

根因结论必须同时定位 internal/reconciliation/reconcile.go 的 ApplySettlement 和 internal/finance/ledger.go 的 Ledger.PostBatch/Ledger.Post，说明预先转换为何看不到既有账本编号、逐行加锁提交怎样让第一条借记在第二条冲突前永久生效，以及错误包装保留身份却不提供回滚的区别；还需解释残留记录为何阻断整批重试，并以全新借贷行最终余额为零界定正常路径。使用 TestFailedSettlementImportDoesNotLeavePartialLedgerEntries 的红测复核，目标仓库生产代码、测试和配置保持零改动，不得实施修复或只把问题归结为重复编号。
