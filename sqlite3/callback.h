#ifndef CALLBACK_H
#define CALLBACK_H
// user_data is the fourth argument passed to sqlite3_exec
typedef int (*exec_callback)(void *user_data, int count, char** data, char** names);

extern int goExecCallback(void* callbackFn, int columnCount, char** columnData, char** columnNames);
extern int goExecDataCallback(void* callbackFn, int columnCount, char** columnData, char** columnNames);

extern int cgoExecCallback(void *user_data, int count, char** data, char** names);
extern int cgoExecDataCallback(void *user_data, int count, char** data, char** names);
#endif