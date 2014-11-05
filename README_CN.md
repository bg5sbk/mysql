简介
====

这是一个Go语言的MySQL客户端库，它基于的是MySQL官方的C API库`libmysqlclient`。

注: 一些代码和思路是从`youtube/vitess`项目中参考和演化过来的.

如何安装
=======

如果你的机器上有`mysql_config`命令。 那么只需要通过以下Shell脚本安装：

```shell
CGO_CFLAGS=`mysql_config --cflags` \
CGO_LDFLAGS=`mysql_config --libs` \
go get github.com/funny/mysql
```

或者你可以手工设置`CGO_CFLAGS`和`CGO_LDFLAGS`这两个环境变量：

```shell
CGO_CFLAGS="-I/usr/local/Cellar/mysql/5.6.15/include/mysql" \
CGO_LD_FLAGS="-L/usr/local/Cellar/mysql/5.6.15/lib -lmysqlclient" \
go get github.com/funny/mysql
```

运行测试
=======

本项目的单元测试使用一些环境变量来设置数据库连接参数。

```
TEST_MYSQL_HOST - MySQL服务器地址。 默认值：127.0.0.1
TEST_MYSQL_PORT - MySQL服务器端口号。 默认值：3306
TEST_MYSQL_DBNAME - 单元测试用的数据库名。 默认值：oursql_test
TEST_MYSQL_UNAME - 数据库用户名。默认值：root
TEST_MYSQL_PASS - 数据库密码。
```

示例：

```shell
TEST_MYSQL_PASS="password" \
CGO_CFLAGS=`mysql_config --cflags` \
CGO_LDFLAGS=`mysql_config --libs` \
go test -v
```