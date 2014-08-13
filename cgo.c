#include "cgo.h"
#include <stdio.h>
#include <string.h>

// All functions must call mysql_thread_init before calling mysql. This is
// because the go runtime controls thread creation, and we don't control
// which thread these functions will be called from.

// this macro produces a compilation-time check for a condition
// if the condition is different than zero, this will abort
// if the condition is zero, this won't generate any code
// (this was imported from Linux kernel source tree)
#define BUILD_BUG_ON(condition) ((void)sizeof(char[1 - 2*!!(condition)]))

void my_library_init(void) {
	// we depend on linking with the 64 bits version of the MySQL library:
	// the go code depends on mysql_fetch_lengths() returning 64 bits unsigned.
	BUILD_BUG_ON(sizeof(unsigned long) - 8);
	mysql_library_init(0, 0, 0);
}

int my_open(
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

void my_close(MYSQL *mysql) {
	if (mysql) {
		mysql_thread_init();
		mysql_close(mysql);
		mysql = NULL;
	}
}

unsigned long my_thread_id(MYSQL *mysql) {
	mysql_thread_init();
	return mysql_thread_id(mysql);
}

unsigned int my_errno(MYSQL *mysql) {
	mysql_thread_init();
	return mysql_errno(mysql);
}

const char *my_error(MYSQL *mysql) {
	mysql_thread_init();
	return mysql_error(mysql);
}

int my_query(MYSQL *mysql, MY_RES *res, const char *sql_str, unsigned long sql_len, MY_MODE mode) {
	mysql_thread_init();

	if (mysql_real_query(mysql, sql_str, sql_len) != 0) {
		return 1;
	}

	if (mode != MY_MODE_NONE) {
		if (mode == MY_MODE_TABLE) {
			res->result = mysql_store_result(mysql);
		} else {
			res->result = mysql_use_result(mysql);
		}

		if (res->result == NULL) {
			return 1;
		}

		res->meta.num_fields = mysql_num_fields(res->result);
		res->meta.fields =  mysql_fetch_fields(res->result);
	}

	res->mysql = mysql;
	res->affected_rows = mysql_affected_rows(mysql);
	res->insert_id = mysql_insert_id(mysql);

	return 0;
}

MY_ROW my_fetch_next(MY_RES *res) {
	MY_ROW row = {0, 0, 0};

	if(res->meta.num_fields == 0) {
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

void my_close_result(MY_RES *res) {
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

int my_prepare(MY_STMT *stmt, MYSQL *mysql, const char *sql_str, unsigned long sql_len) {
	mysql_thread_init();

	stmt->s = mysql_stmt_init(mysql);

	if (mysql_stmt_prepare(stmt->s, sql_str, sql_len) != 0) {
		return 1;
	}

	stmt->param_count = mysql_stmt_param_count(stmt->s);

	return 0;
}

int my_stmt_errno(MY_STMT *stmt) {
	mysql_thread_init();
	return mysql_stmt_errno(stmt->s);
}

const char *my_stmt_error(MY_STMT *stmt) {
  mysql_thread_init();
  return mysql_stmt_error(stmt->s);
}

int my_stmt_execute(MY_STMT *stmt, MYSQL_BIND *binds, MY_STMT_RES *res, MY_MODE mode) {
	mysql_thread_init();

	if (binds != NULL) {
		if (mysql_stmt_bind_param(stmt->s, binds) != 0) {
			return 1;
		}
	}

	if (mode != MY_MODE_NONE && stmt->meta_init == 0) {
		stmt->meta_init = 1;

		MYSQL_RES *meta = mysql_stmt_result_metadata(stmt->s);
		if (meta == NULL) {
			return 1;
		}
		stmt->meta.num_fields = mysql_num_fields(meta);
		stmt->meta.fields =  mysql_fetch_fields(meta);
		mysql_free_result(meta);

		my_bool enable = 1;
		if (mysql_stmt_attr_set(stmt->s, STMT_ATTR_UPDATE_MAX_LENGTH, (void*)&enable) != 0) {
			return 1;
		}
	}

	if (mysql_stmt_execute(stmt->s) != 0) {
		return 1;
	}

	if (mode != MY_MODE_NONE) {
		if (mode == MY_MODE_TABLE) {
			if (mysql_stmt_store_result(stmt->s) != 0) {
				return 1;
			}
		}

		if (stmt->row_cache == NULL) {
			stmt->row_cache = calloc(sizeof(char*), stmt->meta.num_fields);
			stmt->row_cache_len = calloc(sizeof(size_t), stmt->meta.num_fields);
		}

		if (stmt->outputs == NULL) {
			stmt->outputs = calloc(sizeof(MYSQL_BIND), stmt->meta.num_fields);
			stmt->output_lengths = calloc(sizeof(unsigned long), stmt->meta.num_fields);
		} else {
			memset(stmt->outputs, 0, sizeof(MYSQL_BIND) * stmt->meta.num_fields);
			memset(stmt->output_lengths, 0, sizeof(unsigned long) * stmt->meta.num_fields);
		}

		for (int i = 0; i < stmt->meta.num_fields; i ++) {
			size_t size = 0;

			switch (stmt->meta.fields[i].type) {
				case MYSQL_TYPE_TINY:      size = sizeof(signed char); break;   // TINYINT
				case MYSQL_TYPE_SHORT:     size = sizeof(short int); break;     // SMALLINT
				case MYSQL_TYPE_INT24:     size = sizeof(int); break;           // MEDIUMINT
				case MYSQL_TYPE_LONG:      size = sizeof(int); break;           // INT
				case MYSQL_TYPE_LONGLONG:  size = sizeof(long long int); break; // BIGINT
				case MYSQL_TYPE_FLOAT:     size = sizeof(float); break;         // FLOAT
				case MYSQL_TYPE_DOUBLE:    size = sizeof(double); break;        // DOUBLE
				case MYSQL_TYPE_YEAR:      size = sizeof(short int); break;     // YEAR
				case MYSQL_TYPE_TIME:      size = sizeof(MYSQL_TIME); break;    // TIME
				case MYSQL_TYPE_DATE:      size = sizeof(MYSQL_TIME); break;    // DATE
				case MYSQL_TYPE_DATETIME:  size = sizeof(MYSQL_TIME); break;    // DATETIME
				case MYSQL_TYPE_TIMESTAMP: size = sizeof(MYSQL_TIME); break;    // TIMESTAMP
				case MYSQL_TYPE_DECIMAL:     // DECIMAL
				case MYSQL_TYPE_NEWDATE:     // MYSQL_TYPE_NEWDATE
				case MYSQL_TYPE_NEWDECIMAL:  // DECIMAL
				case MYSQL_TYPE_STRING:      // CHAR, BINARY
				case MYSQL_TYPE_VAR_STRING:  // VARCHAR, VARBINARY
				case MYSQL_TYPE_TINY_BLOB:   // TINYBLOB, TINYTEXT
				case MYSQL_TYPE_BLOB:        // BLOB, TEXT
				case MYSQL_TYPE_MEDIUM_BLOB: // MEDIUMBLOB, MEDIUMTEXT
				case MYSQL_TYPE_LONG_BLOB:   // LONGBLOB, LONGTEXT
				case MYSQL_TYPE_BIT:         // BIT
				default: break;
			}

			if (size != 0 && stmt->row_cache[i] == NULL) {
				stmt->row_cache[i] = malloc(size);
				stmt->row_cache_len[i] = size;
			}

			stmt->outputs[i].buffer = size != 0 ? stmt->row_cache[i] : NULL;
			stmt->outputs[i].buffer_length = size;
			stmt->outputs[i].buffer_type = stmt->meta.fields[i].type;
			stmt->outputs[i].length = &(stmt->output_lengths[i]);
		}

		if (mysql_stmt_bind_result(stmt->s, stmt->outputs) != 0) {
			return 1;
		}
	}

	res->stmt = stmt;
	res->affected_rows = mysql_stmt_affected_rows(stmt->s);
	res->insert_id = mysql_stmt_insert_id(stmt->s);

	return 0;
}

void my_stmt_close(MY_STMT *stmt) {
	mysql_thread_init();

	if (stmt->row_cache != NULL) {
		for (int i = 0; i < stmt->meta.num_fields; i ++) {
			if (stmt->row_cache[i] != NULL) {
				free(stmt->row_cache[i]);
			}
		}
		free(stmt->row_cache);
		free(stmt->row_cache_len);
	}

	if (stmt->outputs != NULL) {
		free(stmt->outputs);
		free(stmt->output_lengths);
	}

	mysql_stmt_close(stmt->s);
}

MY_ROW my_stmt_fetch_next(MY_STMT_RES *res) {
	MY_STMT *stmt = res->stmt;

	MY_ROW row = {0, 0, 0};

	if (stmt->meta.num_fields == 0) {
		return row;
	}

	mysql_thread_init();

	int t = mysql_stmt_fetch(stmt->s);
	if (t != 0 && t != MYSQL_DATA_TRUNCATED) {
		if (t != MYSQL_NO_DATA) {
			row.has_error = 1;
		}
		return row;
	}

	for (int i = 0; i < stmt->meta.num_fields; i ++) {
		if (stmt->output_lengths[i] == 0) {
			continue;
		}

		switch (stmt->meta.fields[i].type) {
			case MYSQL_TYPE_DECIMAL:     // DECIMAL
			case MYSQL_TYPE_NEWDATE:     // MYSQL_TYPE_NEWDATE
			case MYSQL_TYPE_NEWDECIMAL:  // DECIMAL
			case MYSQL_TYPE_STRING:      // CHAR, BINARY
			case MYSQL_TYPE_VAR_STRING:  // VARCHAR, VARBINARY
			case MYSQL_TYPE_TINY_BLOB:   // TINYBLOB, TINYTEXT
			case MYSQL_TYPE_BLOB:        // BLOB, TEXT
			case MYSQL_TYPE_MEDIUM_BLOB: // MEDIUMBLOB, MEDIUMTEXT
			case MYSQL_TYPE_LONG_BLOB:   // LONGBLOB, LONGTEXT
			case MYSQL_TYPE_BIT:         // BIT
				if (stmt->row_cache[i] == NULL || stmt->row_cache_len[i] < stmt->output_lengths[i]) {
					if (stmt->row_cache[i] != NULL) {
						free(stmt->row_cache[i]);
					}
					stmt->row_cache[i] = malloc(stmt->output_lengths[i]);
					stmt->row_cache_len[i] = stmt->output_lengths[i];
				}

				stmt->outputs[i].buffer = stmt->row_cache[i];
				stmt->outputs[i].buffer_length = stmt->output_lengths[i];

				if (mysql_stmt_fetch_column(stmt->s, &stmt->outputs[i], i, 0) != 0) {
					row.has_error = 1;
					return row;
				}
				break;
			default:
				break;
		}
	}

	row.mysql_row = (MYSQL_ROW)stmt->row_cache;
	row.lengths = stmt->output_lengths;

	return row;
}

void my_stmt_close_result(MY_STMT_RES *res) {
	mysql_thread_init();
	mysql_stmt_free_result(res->stmt->s);
}
