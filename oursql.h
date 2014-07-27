#ifndef OUR_SQL_H
#define OUR_SQL_H

#include <stdlib.h>
#include <mysql.h>

typedef enum our_mode {
	OUR_MODE_NON,
	OUR_MODE_SET,
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

typedef struct our_res {
	MYSQL        *mysql;
	my_ulonglong affected_rows;
	my_ulonglong insert_id;
	unsigned int num_fields;
	MYSQL_FIELD  *fields;
	MYSQL_RES    *result;
} OUR_RES;

typedef struct our_row {
	int           has_error;
	MYSQL_ROW     mysql_row;
	unsigned long *lengths;
} OUR_ROW;

// stream!=0 uses streaming (use_result). Otherwise it prefetches (store_result).
extern int our_query(MYSQL *mysql, OUR_RES *res, const char *sql_str, unsigned long sql_len, OUR_MODE mode);

// Iterate on this function until mysql_row==NULL or has_error!=0.
extern OUR_ROW our_fetch_next(OUR_RES *res);

// If our_query has results, you must call this before the next invocation.
extern void our_close_result(OUR_RES *res);

#endif