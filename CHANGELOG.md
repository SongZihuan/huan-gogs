# Changelog

All notable changes to Huan Gogs are documented in this file.

Notable changes to the upstream version of Gogs are documented in the file [Gogs-CHANGELOG.md](Gogs-CHANGELOG.md).

## 1.0.0

### Gogs

更新到 Gogs 本版：v0.14.0+dev (main) [#4acaaac8](https://github.com/SongZihuan/gogs/commit/4acaaac85aca427771030ab2e9a1465e9517ba1d)

### Added

- 新增了用户创建仓库的模式（管理员不受限制创建、管理员的仓库不受限制创建）
- 探索功能添加了隐私保护（非管理员不显示用户列表）
- 用户邮箱区分：主邮箱、公开邮箱、系统自动分配的虚假邮箱（例如：cffebd19-fb54-4fd4-8880-3327fe768dee@fake.localhost）。 其中，公开邮箱用于个人主页公开展示，主邮箱用于绑定git、接收通知等非公共行为。

### Fixed

- 修复了自定义邮箱模板的问题。
- 为主页添加了`/home`路由，从而可以设定`/`路由（即首页）是`/home`还是`/explore`。
- 对外迁移仓库等待时间可能过长，导致HTTP无响应，因此设置的5s内返回，并且在仓库也没设置了等待加载提升。

### Develop

- 优化了`SMTP`的`TLS`握手过程。`TCP`建连时就开始尝试`TLS`握手。
- 整合优化了`Token`的除了，使用JWT+签名机制（非必要修改，我做出此修改的原因是我使用原`Token`无法激活用户，后发现是激活函数遗漏设置`User`的`IsActive`值导致）。
- 对`User->Email`的绑定：原情况下、用户注册的第一个`Email`是不会添加到`UserEmail`绑定表中，而是在用户切换主邮箱时再保存。而现在除了主邮箱、还有公开邮箱，因此不得不修改为：用户注册时就要把邮箱写入`UserEmail`表绑定。

### Note

- 处于我对编写测试样例的不熟悉，我对大部分新增功能没做自动化测试。
- 处于我对系统api还未深入了解，因此api的相关修改还未完善。
