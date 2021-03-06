package tango

import (
	"testing"
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync"
)

type PoolAction struct {
}

func (PoolAction) Get() string {
	return "pool"
}

func TestPool(t *testing.T) {
	o := Classic()
	o.Get("/", new(PoolAction))
	
	var wg sync.WaitGroup
	// default pool size is 800
	for i:=0; i< 1000; i++ {
		wg.Add(1)
		go func(){
			buff := bytes.NewBufferString("")
			recorder := httptest.NewRecorder()
			recorder.Body = buff
			req, err := http.NewRequest("GET", "http://localhost:8000/", nil)
			if err != nil {
				t.Error(err)
			}

			o.ServeHTTP(recorder, req)
			expect(t, recorder.Code, http.StatusOK)
			refute(t, len(buff.String()), 0)
			expect(t, buff.String(), "pool")
			wg.Done()
		}()
	}
	wg.Wait()
}
