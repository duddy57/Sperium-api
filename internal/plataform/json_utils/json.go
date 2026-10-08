package json_utils

import (
	"encoding/json"
	"net/http"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 4*1024)
		return &b
	},
}

func WriteJSON(w http.ResponseWriter, code int, body any) {
	buf := bufferPool.Get().(*[]byte)
	defer func() {
		*buf = (*buf)[:0]
		bufferPool.Put(buf)
	}()

	data, err := json.Marshal(body)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "encoding failed")
		return
	}
	*buf = append(*buf, data...)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(*buf)
}
func WriteError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
