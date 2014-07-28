#include "oursql.h"
#include <stdio.h>

// All functions must call mysql_thread_init before calling mysql. This is
// because the go runtime controls thread creation, and we don't control
// which thread these functions will be called from.

// this macro produces a compilation-time check for a condition
// if the condition is different than zero, this will abort
// if the condition is zero, this won't generate any code
// (this was imported from Linux kernel source tree)
#define BUILD_BUG_ON(condition) ((void)sizeof(char[1 - 2*!!(condition)]))

void our_library_init(void) {
	// we depend on linking with the 64 bits version of the MySQL library:
	// the go code depends on mysql_fetch_lengths() returning 64 bits unsigned.
	BUILD_BUG_ON(sizeof(unsigned long) - 8);
	mysql_library_init(0, 0, 0);
}

int our_connect(
    MYSQL         *mysql,
    const char    *host,
    const char    *user,
    const char    *passwd,
    const char    *db,
    unsigned int  port,
    const char    *unix_socket,
    const char    *csname,
    unsigned long client_flag
) {
	mysql_thread_init();

	mysql_init(mysql);

	if (!mysql_real_connect(mysql, host, user, passwd, db, port, unix_socket, client_flag)) {
		return 1;
	}

	if (!mysql_set_character_set(mysql, csname)) {
		return 1;
	}

	return 0;
}

void our_close(MYSQL *mysql) {
	if (mysql) {
		mysql_thread_init();
		mysql_close(mysql);
		mysql = NULL;
	}
}

unsigned long our_thread_id(MYSQL *mysql) {
	mysql_thread_init();
	return mysql_thread_id(mysql);
}

unsigned int our_errno(MYSQL *mysql) {
	mysql_thread_init();
	return mysql_errno(mysql);
}

const char *our_error(MYSQL *mysql) {
	mysql_thread_init();
	return mysql_error(mysql);
}

int our_query(MYSQL *mysql, OUR_RES *res, const char *sql_str, unsigned long sql_len, OUR_MODE mode) {
	mysql_thread_init();

	if (mysql_real_query(mysql, sql_str, sql_len) != 0) {
		return 1;
	}

	if (mode != OUR_MODE_NON) {
		if (mode == OUR_MODE_TABLE) {
			res->result = mysql_store_result(mysql);
		} else {
			res->result = mysql_use_result(mysql);
		}

		if (res->result == NULL) {
			return 1;
		}

		res->num_fields = mysql_num_fields(res->result);
		res->fields =  mysql_fetch_fields(res->result);
	}

	res->mysql = mysql;
	res->affected_rows = mysql_affected_rows(mysql);
	res->insert_id = mysql_insert_id(mysql);

	return 0;
}

OUR_ROW our_fetch_next(OUR_RES *res) {
	OUR_ROW row = {0, 0, 0};

	if(res->num_fields == 0) {
		return row;
	}

	mysql_thread_init();

	row.mysql_row = mysql_fetch_row(res->result);
	if (!row.mysql_row) {
		if(mysql_errno(res->mysql)) {
			row.has_error = 1;
			return row;
		}
	} else {
		row.lengths = mysql_fetch_lengths(res->result);
	}

	return row;
}

void our_close_result(OUR_RES *res) {
	MYSQL_RES *result;

	mysql_thread_init();

	if (res->result) {
		mysql_free_result(res->result);
	}

	// Ignore subsequent results if any. We only
	// return the first set of results for now.
	while (mysql_next_result(res->mysql) == 0) {
		result = mysql_store_result(res->mysql);
		if (result) {
			// while(mysql_fetch_row(result)) {
			// }
			mysql_free_result(result);
		}
	}
}

int our_prepare(OUR_STMT *stmt, MYSQL *mysql, const char *sql_str, unsigned long sql_len) {
	mysql_thread_init();

	stmt->s = mysql_stmt_init(mysql);

	if (mysql_stmt_prepare(stmt->s, sql_str, sql_len) != 0) {
		return 1;
	}

	stmt->param_count = mysql_stmt_param_count(stmt->s);

	return 0;
}

int our_stmt_errno(OUR_STMT *stmt) {
	mysql_thread_init();
	return mysql_stmt_errno(stmt->s);
}

const char *our_stmt_error(OUR_STMT *stmt) {
  mysql_thread_init();
  return mysql_stmt_error(stmt->s);
}

int our_stmt_execute(OUR_STMT *stmt, MYSQL_BIND *binds) {
	mysql_thread_init();

	if (mysql_stmt_bind_param(stmt->s, binds) != 0) {
		return 1;
	}

	if (mysql_stmt_execute(stmt->s) != 0) {
		return 1;
	}

	stmt->affected_rows = mysql_stmt_affected_rows(stmt->s);
	stmt->insert_id = mysql_stmt_insert_id(stmt->s);

	return 0;
}

void our_stmt_close(OUR_STMT *stmt) {
	mysql_thread_init();
	mysql_stmt_close(stmt->s);
}
