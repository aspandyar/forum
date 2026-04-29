package request

import (
	"net/http"
	"strconv"
	"strings"
)

func PathParts(r *http.Request) []string {
	return strings.Split(r.URL.Path, "/")
}

func PathInt(parts []string, idx int) (int, error) {
	return strconv.Atoi(parts[idx])
}
