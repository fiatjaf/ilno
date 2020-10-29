package lnurl

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fiatjaf/go-lnurl"
	cmap "github.com/orcaman/concurrent-map"
	"gopkg.in/antage/eventsource.v1"
	"github.com/fiatjaf/ilno/logger"
)

var userstreams = cmap.New()

func Auth(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	k1 := params.Get("k1")
	sig := params.Get("sig")
	key := params.Get("key")

	if ok, err := lnurl.VerifySignature(k1, sig, key); !ok {
		logger.Debug("failed to verify lnurl-auth signature: %s", err)
		json.NewEncoder(w).Encode(
			lnurl.ErrorResponse("signature verification failed."))
		return
	}

	ies, ok := userstreams.Get(k1)
	if !ok {
		logger.Debug("successful login but no browser session related")
	} else {
		// notify browser
		es := ies.(eventsource.EventSource)
		es.SendEventMessage(`{"sig": "`+sig+`", "key": "`+key+`"}`, "auth", "")
	}

	json.NewEncoder(w).Encode(lnurl.OkResponse())
}

func AuthStream(w http.ResponseWriter, r *http.Request) {
	var es eventsource.EventSource
	k1 := r.URL.Query().Get("k1")

	// try to fetch an existing stream
	ies, ok := userstreams.Get(k1)
	if ok {
		es = ies.(eventsource.EventSource)
	}

	if es == nil {
		es = eventsource.New(
			&eventsource.Settings{
				Timeout:        5 * time.Second,
				CloseOnTimeout: true,
				IdleTimeout:    1 * time.Minute,
			},
			func(r *http.Request) [][]byte {
				return [][]byte{
					[]byte("X-Accel-Buffering: no"),
					[]byte("Cache-Control: no-cache"),
					[]byte("Content-Type: text/event-stream"),
					[]byte("Connection: keep-alive"),
					[]byte("Access-Control-Allow-Origin: *"),
				}
			},
		)
		userstreams.Set(k1, es)
		go func() {
			for {
				time.Sleep(25 * time.Second)
				es.SendEventMessage("", "keepalive", "")
			}
		}()
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		es.SendRetryMessage(3 * time.Second)
	}()

	es.ServeHTTP(w, r)
}
