package birect_test

import (
	"fmt"
	"net"
	"net/http"

	"github.com/marcuswestin/go-birect"
)

func ExampleUpgradeRequests_server() {
	listener, err := net.Listen("tcp", ":8087")
	if err != nil {
		panic(err)
	}
	go http.Serve(listener, nil)
	server := birect.UpgradeRequests("/birect/upgrade")

	type EchoParams struct{ Text string }
	type EchoResponse struct{ Text string }
	server.HandleJSONReq("Echo", func(req *birect.JSONReq) (res interface{}, err error) {
		var par EchoParams
		req.ParseParams(&par)
		return EchoResponse{par.Text}, nil
	})
	// Output:
	//
}

func ExampleConnect_client() {
	conn, _ := birect.Connect("http://localhost:8087/birect/upgrade")

	type EchoParams struct{ Text string }
	type EchoResponse struct{ Text string }
	var par = EchoParams{"Hi!"}
	var res EchoResponse
	fmt.Println("Send:", par.Text)
	conn.SendJSONReq("Echo", &res, par)
	fmt.Println("Received:", res.Text)

	// Output:
	// Send: Hi!
	// Received: Hi!
}
