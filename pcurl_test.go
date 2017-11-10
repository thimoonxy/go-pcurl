package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func Test_Getres(t *testing.T) {
	//vars
	lock := make(chan bool, 0)
	MaxIdleConnections := 20
	RequestTimeout := 5
	Clientvar = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: MaxIdleConnections,
			DisableCompression:  true, // client will compress it by default
		},
		Timeout: time.Duration(RequestTimeout) * time.Second,
	}
	call := func(i int) {
		res := getres(Clientvar, "http://mirrors.163.com/centos/7/isos/x86_64/", -1, -1)
		_, err := io.Copy(ioutil.Discard, res.Body)
		if err != nil {
			t.Error(err)
		} else {
			t.Log("#", i, " done @", time.Now().Format("15:04:05.99"))
		}
		res.Body.Close()
	}
	// 1st batch
	for i := 0; i < 2; i++ {
		go func(i int) {
			call(i)
			lock <- true
		}(i)

	}
	for i := 0; i < 2; i++ {
		<-lock
	}

	time.Sleep(1e9)

	// 2nd batch
	for i := 2; i < 4; i++ {
		go func(i int) {
			call(i)
			lock <- true
		}(i)
	}
	for i := 0; i < 2; i++ {
		<-lock
	}
}
