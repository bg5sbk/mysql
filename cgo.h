#ifndef CGO_H
#define CGO_H

#include <stdlib.h>
#include <mysql.h>

typedef enum my_mode {
	MY_MODE_NONE,
	MY_MODE_TABLE,
	MY_MODE_READER
} MY_MODE;

// This API provides convenient C wrapper functions for mysql client.

// !!! Call this before everything else !!!
extern void my_library_init(void);

// Create a connection. You must call my_close even if my_open fails.
extern int my_open(
	MYSQL         *mysql,
	const char    *host,
	const char    *user,
	const char    *passwd,
	const char    *db,
	unsigned int  port,
	const char    *unix_socket,
	const char    *csname,
	unsigned long client_flag
);

extern void my_close(MYSQL *mysql);

// Pass-through to mysql
extern unsigned long my_thread_id(MYSQL *mysql);
extern unsigned int my_errno(MYSQL *mysql);
extern const char *my_error(MYSQL *mysql);

typedef struct my_res_meta {
	unsigned int num_fields;
	MYSQL_FIELD  *fields;
} MY_RES_META;

typedef struct my_res {
	MYSQL        *mysql;
	my_ulonglong affected_rows;
	my_ulonglong insert_id;
	MY_RES_META meta;
	MYSQL_RES    *result;
} MY_RES;

typedef struct my_row {
	int           has_error;
	MYSQL_ROW     mysql_row;
	unsigned long *lengths;
	my_bool       *is_nulls;
} MY_ROW;

// stream!=0 uses streaming (use_result). Otherwise it prefetches (store_result).
extern int my_query(MYSQL *mysql, MY_RES *res, const char *sql_str, unsigned long sql_len, MY_MODE mode);

// Iterate on this function until mysql_row==NULL or has_error!=0.
extern MY_ROW my_fetch_next(MY_RES *res);

// If my_query has results, you must call this before the next invocation.
extern void my_close_result(MY_RES *res);

typedef struct my_stmt {
	MYSQL_STMT    *s;
	unsigned long param_count;
	MY_RES_META  meta;
	my_bool       meta_init;
	char          **row_cache;
	size_t        *row_cache_len;
	MYSQL_BIND    *outputs;
	unsigned long *output_lengths;
} MY_STMT;

typedef struct my_stmt_res {
	MYSQL        *mysql;
	MY_STMT     *stmt;
	my_ulonglong affected_rows;
	my_ulonglong insert_id;
} MY_STMT_RES;


extern int my_prepare(MY_STMT *stmt, MYSQL *mysql, const char *sql_str, unsigned long sql_len);

extern int my_stmt_errno(MY_STMT *stmt);

extern const char *my_stmt_error(MY_STMT *stmt);

extern int my_stmt_execute(MY_STMT *stmt, MYSQL_BIND *binds, MY_STMT_RES *res, MY_MODE mode);

extern void my_stmt_close(MY_STMT *stmt);

extern MY_ROW my_stmt_fetch_next(MY_STMT_RES *res);

extern void my_stmt_close_result(MY_STMT_RES *res);

#endif