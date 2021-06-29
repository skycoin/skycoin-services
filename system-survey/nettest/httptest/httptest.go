package httptest

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type Input struct {
	Addr string `json:"addr"`
}

type Result struct {
	Error        error `json:"error"`
	RequestID    int   `json:"request_id"`
	RandSent     int   `json:"rand_sent"`
	RandReceived int   `json:"rand_received"`
	Success      bool  `json:"success"`
}

func Run(inp *Input) Result {
	res := Result{}

	r := mux.NewRouter()
	r.HandleFunc("/ping", pingHandler).Methods(http.MethodPost)

	srv := &http.Server{
		Addr:    inp.Addr,
		Handler: r,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()
	time.Sleep(time.Millisecond * 200)

	res.RequestID, res.RandSent = rand.Int(), rand.Int()
	b, _ := json.Marshal(map[string]int{"id": res.RequestID, "rand": res.RandSent})
	req, err := http.NewRequest(http.MethodPost, "http://"+inp.Addr+"/ping", bytes.NewBuffer(b))
	if err != nil {
		res.Error = err
		return res
	}

	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		res.Error = err
		return res
	}

	result := make(map[string]int)
	defer response.Body.Close()
	_ = json.NewDecoder(response.Body).Decode(&result)
	res.RandReceived = result["rand"]

	if res.RandReceived != res.RandSent+1 {
		return res
	}

	res.Success = true

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	_ = srv.Shutdown(ctx)

	return res
}

// pingHandler handles ping request.
func pingHandler(rw http.ResponseWriter, r *http.Request) {
	var req struct {
		ID   int `json:"id"`
		Rand int `json:"rand"`
	}

	defer r.Body.Close()
	_ = json.NewDecoder(r.Body).Decode(&req)

	req.Rand++

	rw.Header().Set("Content-Type", "application/json")

	b, _ := json.Marshal(&req)
	_, _ = rw.Write(b)
}
