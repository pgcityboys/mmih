package utils;

import "os"

func EnvWithDefaults(key, def string) string {
	res, ok := os.LookupEnv(key)
	if !ok {
		res = def
	}
	return res
}
