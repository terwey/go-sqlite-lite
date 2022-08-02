#include "callback.h"

int cgoExecCallback(void *user_data, int count, char** data, char** names) {
	return goExecCallback(user_data, count, data, names);
}

int cgoExecDataCallback(void *user_data, int count, char** data, char** names) {
	return goExecDataCallback(user_data, count, data, names);
}