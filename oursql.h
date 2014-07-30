#ifndef OUR_SQL_H
#define OUR_SQL_H

#include <stdlib.h>
#include <mysql.h>

typedef enum our_mode {
	OUR_MODE_NONE,
	OUR_MODE_TABLE,
	OUR_MODE_READER
} OUR_MODE;

// This API provides convenient C wrapper functions for mysql client.

// !!! Call this before everything else !!!
extern void our_library_init(void);

// Create a connection. You must call our_close even if our_connect fails.
extern int our_connect(
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

extern void our_close(MYSQL *mysql);

// Pass-through to mysql
extern unsigned long our_thread_id(MYSQL *mysql);
extern unsigned int our_errno(MYSQL *mysql);
extern const char *our_error(MYSQL *mysql);

typedef struct our_res_meta {
	unsigned int num_fields;
	MYSQL_FIELD  *fields;
} OUR_RES_META;

typedef struct our_res {
	MYSQL        *mysql;
	my_ulonglong affected_rows;
	my_ulonglong insert_id;
	OUR_RES_META meta;
	MYSQL_RES    *result;
} OUR_RES;

typedef struct our_row {
	int           has_error;
	MYSQL_ROW     mysql_row;
	unsigned long *lengths;
	my_bool       *is_nulls;
} OUR_ROW;

// stream!=0 uses streaming (use_result). Otherwise it prefetches (store_result).
extern int our_query(MYSQL *mysql, OUR_RES *res, const char *sql_str, unsigned long sql_len, OUR_MODE mode);

// Iterate on this function until mysql_row==NULL or has_error!=0.
extern OUR_ROW our_fetch_next(OUR_RES *res);

// If our_query has results, you must call this before the next invocation.
extern void our_close_result(OUR_RES *res);

typedef struct our_stmt {
	MYSQL_STMT    *s;
	unsigned long param_count;
	OUR_RES_META  meta;
	my_bool       meta_init;
	char          **row_cache;
	size_t        *row_cache_len;
	MYSQL_BIND    *outputs;
	unsigned long *output_lengths;
} OUR_STMT;

typedef struct our_stmt_res {
	MYSQL        *mysql;
	OUR_STMT     *stmt;
	my_ulonglong affected_rows;
	my_ulonglong insert_id;
} OUR_STMT_RES;


extern int our_prepare(OUR_STMT *stmt, MYSQL *mysql, const char *sql_str, unsigned long sql_len);

extern int our_stmt_errno(OUR_STMT *stmt);

extern const char *our_stmt_error(OUR_STMT *stmt);

extern int our_stmt_execute(OUR_STMT *stmt, MYSQL_BIND *binds, OUR_STMT_RES *res, OUR_MODE mode);

extern void our_stmt_close(OUR_STMT *stmt);

extern OUR_ROW our_stmt_fetch_next(OUR_STMT_RES *res);

extern void our_stmt_close_result(OUR_STMT_RES *res);

#endif