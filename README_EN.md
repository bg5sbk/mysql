Introduction
============

A simple Go MySQL client library based on libmysqlclient (the offical C library).

NOTE: some code fork from `youtube/vitess` project.

How to install
==============

If `mysql_config` command already installed on your machine. Just use this shell script to install the library.

```shell
CGO_CFLAGS=`mysql_config --cflags` \
CGO_LDFLAGS=`mysql_config --libs` \
go get github.com/funny/mysql
```

Or you can set `CGO_CFLAGS` and `CGO_LDFLAGS` environment variables by manual.

Example:

```shell
CGO_CFLAGS="-I/usr/local/Cellar/mysql/5.6.15/include/mysql" \
CGO_LD_FLAGS="-L/usr/local/Cellar/mysql/5.6.15/lib -lmysqlclient" \
go get github.com/funny/mysql
```

How to test
===========

The unit test use some environment variable to override connection parameter.

```
TEST_MYSQL_HOST - MySQL server host name or IP address. Default 127.0.0.1
TEST_MYSQL_PORT - MySQL server port. Default 3306
TEST_MYSQL_DBNAME - Database name for unit test. Default oursql_test
TEST_MYSQL_UNAME - The user name. Default root
TEST_MYSQL_PASS - The password.
```

Example:

```shell
TEST_MYSQL_PASS="password" \
CGO_CFLAGS=`mysql_config --cflags` \
CGO_LDFLAGS=`mysql_config --libs` \
go test -v
```