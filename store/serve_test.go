package store

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"testing"

	"github.com/intob/rocketkv/client"
	"github.com/intob/rocketkv/protocol"
)

/*

Test the Server using the client package to send messages.

Note:
Each test should use a separate port number, so that they can run in parallel.

*/

// Starts up a TCP server & returns a client connected to it
func getTestServerAndClient(port int, authSecret string) *client.Client {
	addr := fmt.Sprintf(":%s", strconv.Itoa(port))
	ready := make(chan bool)

	go func() {
		st := getTestStore(8, false)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			panic(err)
		}
		ready <- true
		close(ready)
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		st.ServeConn(conn, authSecret, 512)
	}()

	<-ready

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		panic(err)
	}

	return client.NewClient(conn)
}

func TestPing(t *testing.T) {
	client := getTestServerAndClient(42500, "")
	defer client.Close()

	err := client.Ping()
	if err != nil {
		panic(err)
	}

	resp := <-client.Msgs
	if resp.Status != protocol.StatusOk {
		t.FailNow()
	}
	if resp.Op != protocol.OpPong {
		t.FailNow()
	}
}

func TestUnauthorized(t *testing.T) {
	client := getTestServerAndClient(42501, "test")
	defer client.Close()

	err := client.Get("someKey")
	if err != nil {
		panic(err)
	}

	resp := <-client.Msgs
	if resp.Status != protocol.StatusUnauthorized {
		t.FailNow()
	}
}

func TestWrongSecret(t *testing.T) {
	client := getTestServerAndClient(42502, "test")
	defer client.Close()

	err := client.Auth("wrongSecret")
	if err != nil {
		panic(err)
	}

	resp := <-client.Msgs
	if resp.Status != protocol.StatusUnauthorized {
		t.FailNow()
	}
}

func TestAuthorized(t *testing.T) {
	authSecret := "test"
	client := getTestServerAndClient(42503, authSecret)
	defer client.Close()

	err := client.Auth(authSecret)
	if err != nil {
		panic(err)
	}
	resp := <-client.Msgs
	if resp.Status != protocol.StatusOk {
		t.FailNow()
	}

	err = client.Get("someRandomKey")
	if err != nil {
		panic(err)
	}

	resp = <-client.Msgs
	if resp.Status != protocol.StatusNotFound {
		t.FailNow()
	}
}

func TestServerSetAndGet(t *testing.T) {
	client := getTestServerAndClient(42504, "")
	defer client.Close()

	key := "testKey"
	value := []byte("testing")

	// set value
	err := client.Set(key, value, 0, true)
	if err != nil {
		panic(err)
	}
	// wait for response
	<-client.Msgs

	// fetch value
	err = client.Get(key)
	if err != nil {
		panic(err)
	}

	resp := <-client.Msgs

	if !bytes.Equal(resp.Value, value) {
		t.FailNow()
	}
}

func TestServerDel(t *testing.T) {
	client := getTestServerAndClient(42505, "")
	defer client.Close()

	key := "testKey"

	err := client.Del(key, true)
	if err != nil {
		panic(err)
	}

	resp := <-client.Msgs

	if resp.Status != protocol.StatusOk {
		t.FailNow()
	}
}

func TestServerList(t *testing.T) {
	client := getTestServerAndClient(42506, "")
	defer client.Close()

	keyAdded := "testing"

	err := client.Set(keyAdded, []byte("just testing"), 0, false)
	if err != nil {
		panic(err)
	}

	// list all keys beginning with "test"
	// we added "testing" to an empty dataset,
	// so we should get exactly 1 result
	err = client.List("test")
	if err != nil {
		panic(err)
	}

	gotStreamEndMsg := false
	gotKey := false
	for m := range client.Msgs {
		if m.Key == keyAdded {
			gotKey = true
			continue
		}
		if m.Status == protocol.StatusStreamEnd {
			gotStreamEndMsg = true
			break
		}
	}

	if !gotStreamEndMsg || !gotKey {
		t.FailNow()
	}
}

func TestServerListWhenEmpty(t *testing.T) {
	client := getTestServerAndClient(42507, "")
	defer client.Close()

	err := client.List("test")
	if err != nil {
		panic(err)
	}

	gotStreamEndMsg := false
	for m := range client.Msgs {
		if m.Status == protocol.StatusStreamEnd {
			gotStreamEndMsg = true
			break
		}
	}

	if !gotStreamEndMsg {
		t.FailNow()
	}
}

func TestServerCount(t *testing.T) {
	client := getTestServerAndClient(42508, "")
	defer client.Close()

	keyAdded := "testing"

	err := client.Set(keyAdded, []byte("just testing"), 0, false)
	if err != nil {
		panic(err)
	}

	// count all keys beginning with "test"
	// we added "testing" to an empty dataset,
	// so we should get exactly 1
	err = client.Count("test")
	if err != nil {
		panic(err)
	}
	resp := <-client.Msgs
	c := binary.BigEndian.Uint64(resp.Value)
	if c != uint64(1) {
		t.FailNow()
	}
}
