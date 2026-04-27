**这是什么类型的PR？**

必须是以下之一：

/kind fix (缺陷修复)
/kind feat (新增功能)
/kind perf (性能提升)
/kind style (代码可读性提升)
/kind chore (非关键代码修改)
/kind revert (代码回退)
/kind refactor (代码重构)
/kind ci (CI配置和脚本修改)
/kind test (添加或修改测试用例)
/kind docs (添加或修改项目文档)
/kind build (编译或依赖包的修改)

----

**这个PR是做什么的/我们为什么需要它**



----

**此PR修复了哪些问题**:

Fixes #

----

**PR对程序接口进行了哪些修改？**



----

**Commit提交信息说明**：

**MR标题格式**：`fix[function_proxy]：xxxxx`

- `fix` 表示PR类型（可以是 fix/feat/perf/style/chore/revert/refactor/ci/test/docs/build 等）
- `function_proxy` 表示修改的模块名称（根据实际修改的模块填写）
- `xxxxx` 为具体的修改描述

**Commit信息格式**：

- Commit信息开头必须包含PR类型前缀，格式：`fix: xxxxx`
- Commit信息需要包含详细的修改说明，描述具体修改内容和原因
- Commit信息最后必须包含 Signed-off-by，通过 `git commit -s` 命令生成

**示例**：
```
fix: 修复InstanceProxy序列化反序列化时is_async属性无法还原的问题

- 在constants.py中添加IS_ASYNC常量
- 在serialization_方法中保存is_async属性
- 在deserialization_方法中恢复is_async属性，确保序列化后再反序列化能正确还原is_async值

Signed-off-by: YourName <your.email@example.com>
```

----

**Self-checklist**:（**请自检，在[ ]内打上x，我们将检视你的完成情况，否则会导致pr无法合入**）

+ - [ ] **设计**：PR对应的方案是否已经经过Maintainer评审，方案检视意见是否均已答复并完成方案修改
+ - [ ] **测试**：PR中的代码是否已有UT/ST测试用例进行充分的覆盖，新增测试用例是否随本PR一并上库或已经上库
+ - [ ] **验证**：PR描述信息中是否已包含对该PR对应的Feature、Refactor、Bugfix的预期目标达成情况的详细验证结果描述
+ - [ ] **接口**：是否涉及对外接口变更，相应变更已得到接口评审组织的通过，API对应的注释信息已经刷新正确
+ - [ ] **文档**：是否涉及官网文档修改，如果涉及请及时提交资料到Doc仓


----

