package utils;

import "os"

func EnvWithDefaults(key, def string) string {
	res, ok := os.LookupEnv(key)
	if !ok {
		res = def
	}
	return res
}

func FindInSlice(key string, slice []string) int {
	for i, e := range slice {
		if key == e {
			return i
		}
	}
	return -1;
}

func RemoveFromSlice(i int, s []string) []string{
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
