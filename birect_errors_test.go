package birect_test

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/marcuswestin/go-birect"
	"github.com/marcuswestin/go-errs"
)

type ResponseError struct {
	RetryIn int
}

func TestApplicationError(t *testing.T) {
	var err error
	server, client := setupServerClient()

	type AuthenticateParams struct{ Secret string }
	type AuthenticateResponse struct{ SessionToken string }
	server.HandleJSONReq("Authenticate", func(req *birect.JSONReq) (res interface{}, err error) {
		var params AuthenticateParams
		req.ParseParams(&params)

		if params.Secret != "foobarcat" {
			err = birect.NewError(errs.Info{"Secret": params.Secret}, "Wrong secret:", params.Secret)
			return
		}

		res = AuthenticateResponse{"asd71lgd1892d1"}
		return
	})

	var res AuthenticateResponse
	err = client.SendJSONReq("Authenticate", AuthenticateParams{"badsecret"}, &res)
	assert(t, err != nil)
	assert(t, err.Error() == "Wrong secret: badsecret")
	assert(t, res.SessionToken == "")
	err = client.SendJSONReq("Authenticate", AuthenticateParams{"foobarcat"}, &res)
	assert(t, err == nil)
	assert(t, res.SessionToken == "asd71lgd1892d1")
}

func TestDefaultErrorMessage(t *testing.T) {
	var err error
	server, client := setupServerClient()
	server.HandleJSONReq("TestDefaultErrorMessage", func(req *birect.JSONReq) (res interface{}, err error) {
		return nil, errors.New("Internal, secret error message")
	})
	err = client.SendJSONReq("TestDefaultErrorMessage", nil, nil)
	assert(t, err.Error() == birect.DefaultPublicErrorMessage)
}

// Misc utils
/////////////
var lastPort = 18000

func setupServerClient() (*birect.Server, *birect.Client) {
	lastPort += 1
	go http.ListenAndServe(fmt.Sprintf("localhost:%d", lastPort), nil)
	server := birect.UpgradeRequests(fmt.Sprintf("/birect/upgrade/%d", lastPort))
	client, err := birect.Connect(fmt.Sprintf("http://localhost:%d/birect/upgrade/%d", lastPort, lastPort))
	if err != nil {
		panic(err)
	}
	return server, client
}

func assert(t *testing.T, ok bool, msg ...interface{}) {
	if !ok {
		// t.Fatal("assert failed", msg)
		log.Panic(msg...)
	}
}
