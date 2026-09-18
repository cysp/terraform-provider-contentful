package provider_test

import (
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

type appActionResponseFault struct {
	t            *testing.T
	next         http.Handler
	method       string
	truncate     bool
	fail         atomic.Bool
	requestMutex sync.Mutex
	actionID     string
	methods      []string
}

func (fault *appActionResponseFault) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	record := httptest.NewRecorder()
	fault.next.ServeHTTP(record, r)

	if r.Method != http.MethodGet {
		fault.requestMutex.Lock()

		fault.methods = append(fault.methods, r.Method)
		if r.Method == http.MethodPost && record.Code == http.StatusCreated {
			fault.actionID = appActionRecoveryResponseID(fault.t, record.Body.Bytes())
		}
		fault.requestMutex.Unlock()
	}

	if r.Method == fault.method && fault.fail.Swap(false) {
		assert.Less(fault.t, record.Code, 300, "mutation must commit before losing its response")

		if !fault.truncate {
			panic(http.ErrAbortHandler)
		}

		maps.Copy(w.Header(), record.Header())
		w.Header().Set("Content-Length", strconv.Itoa(record.Body.Len()))
		w.WriteHeader(record.Code)
		_, _ = io.WriteString(w, record.Body.String()[:record.Body.Len()/2])

		return
	}

	maps.Copy(w.Header(), record.Header())
	w.WriteHeader(record.Code)
	_, _ = w.Write(record.Body.Bytes())
}
