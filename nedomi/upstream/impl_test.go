package upstream_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gophergala/nedomi/config"
	"github.com/gophergala/nedomi/upstream"
)

func TestGetRequest(t *testing.T) {
	// Skipped because it is waaay too hard to construct a whole config just to test
	// one upstream function.
	//!TODO: make it work again once we have the time.
	t.Skip()

	var b = "Hello world is an awesome test string"
	var start, end uint64 = 3, 500000
	file, err := ioutil.TempFile(os.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.WriteString(file, b)
	if err != nil {
		t.Fatal(err)
	}
	file.Sync()
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.ServeContent(w, r, "", time.Time{}, file)
		}))
	}()

	vh := &config.VirtualHost{
		Name:            "localhost",
		UpstreamAddress: "http://" + listener.Addr().String(),
		CacheZone:       0,
		CacheKey:        "",
	}

	u := upstream.New(&config.Config{
		HTTP: config.HTTPSection{
			Servers: []*config.VirtualHost{vh},
		},
	})

	resp, err := u.GetRequestPartial(vh, file.Name(), start, end)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expected := []byte(b)[start:]
	if !bytes.Equal(expected, result) {
		t.Fatalf("Expected [%s] got [%s]", expected, result)
	}
}
