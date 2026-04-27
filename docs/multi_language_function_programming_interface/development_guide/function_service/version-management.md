# 版本管理

openYuanrong 支持发布多个版本的函数，实现软件开发生命周期中的持续集成和发布，确保函数的稳定性和可靠性。版本相当于函数的快照，包括函数代码及函数配置。当您发布版本时，openYuanrong 会为函数生成快照，并自动分配一个版本号与其关联，以供后续使用。

函数注册时，会生成 latest 版本，对函数的更新也是在 latest 版本上。您可以在测试稳定后发布函数版本，用稳定的版本来服务线上请求，并且可以继续在 latest 版本上开发测试。

在未发布任何版本前，latest 版本是您拥有的唯一函数版本，latest 版本不能被删除。版本发布后，已发布的版本不可更改。且版本号单调递增，不会被重复使用。

相关 REST API：

- [发布函数](../../api/function_service/release_function.md)：可发布一个函数版本。 
- [查询函数的指定版本](../../api/function_service/get_function_by_version.md)：查询函数指定版本的信息。
- [查询函数的所有版本](../../api/function_service/get_function_by_name.md)：查询函数所有版本的信息。
- [删除函数](../../api/function_service/delete_function.md)：可删除一个函数版本。
