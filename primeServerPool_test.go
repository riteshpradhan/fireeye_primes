/*
* @Author: ritesh
* @Date:   2016-12-29 14:50:52
* @Last Modified by:   Ritesh Pradhan
* @Last Modified time: 2016-12-29 18:35:16
*/

package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	// "github.com/delaemon/go-uuidv4"
)


func setup() {
	uuid_prime_map["d989e9a9-dab0-445c-ac85-274a2f3cd389"] = "[2, 3, 5]"
}

func setupAll() {
	uuid_prime_map["d989e9a9-dab0-445c-ac85-274a2f3cd389"] = "[2, 3, 5]"
	uuid_prime_map["d989e9a9-dab0-445c-ac85-274a2f3cd390"] = "[2, 3, 5, 7]"
}

func tearDownPrime() {
	os.Exit(1)
}

func TestDefaultHandler(t *testing.T) {
	rootRequest, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal("Root request error: %s", err)
	}

	cases := []struct {
		w                    *httptest.ResponseRecorder
		r                    *http.Request
		expectedResponseCode int
		expectedResponseBody []byte
	}{
		{
			w:                    httptest.NewRecorder(),
			r:                    rootRequest,
			expectedResponseCode: http.StatusOK,
			expectedResponseBody: []byte("/primes?start_num=aNum&end_num=bNum\n/result?id=[UUIDv4]\n/results\n"),
		},
	}

	for _, c := range cases {
		defaultHandler(c.w, c.r)

		if c.expectedResponseCode != c.w.Code {
			t.Errorf("Status Code didn't match:\n\t%q\n\t%q", c.expectedResponseCode, c.w.Code)
		}

		if !bytes.Equal(c.expectedResponseBody, c.w.Body.Bytes()) {
			t.Errorf("Body didn't match:\n\t%q\n\t%q", string(c.expectedResponseBody), c.w.Body.String())
		}
	}
}


func TestGetSingleResult(t *testing.T) {
	setup()
	rootRequest, err := http.NewRequest("GET", "/result?id=d989e9a9-dab0-445c-ac85-274a2f3cd389", nil)
	if err != nil {
		t.Fatal("Root request error: %s", err)
	}

	cases := []struct {
		w                    *httptest.ResponseRecorder
		r                    *http.Request
		expectedResponseCode int
		expectedResponseBody []byte
	}{
		{
			w:                    httptest.NewRecorder(),
			r:                    rootRequest,
			expectedResponseCode: http.StatusOK,
			expectedResponseBody: []byte("[2, 3, 5]\n"),
		},
		{
			w:                    httptest.NewRecorder(),
			r:                    rootRequest,
			expectedResponseCode: http.StatusNoContent ,
			expectedResponseBody: []byte("No such UUID=d989e9a9-dab0-445c-ac85-274a2f3cd389 found.\n"),
		},

	}

	for _, c := range cases {
		getSingleResult(c.w, c.r)

		fmt.Printf("status : %v, body: %v", c.w.Code, c.w.Body.String())
		if http.StatusOK != c.w.Code && http.StatusNoContent != c.w.Code {
			t.Errorf("Status Code didn't match:\n\t%q\n\t%q", c.expectedResponseCode, c.w.Code)
		}

		if c.w.Body.String() != string(c.expectedResponseBody) {
			t.Errorf("Body didn't match:\n\t%q\n\t%q", string(c.expectedResponseBody), c.w.Body.String())
		}
	}
}


func TestGetAllResults(t *testing.T) {
	setupAll()
	rootRequest, err := http.NewRequest("GET", "/results", nil)
	if err != nil {
		t.Fatal("Root request error: %s", err)
	}

	cases := []struct {
		w                    *httptest.ResponseRecorder
		r                    *http.Request
		expectedResponseCode int
		expectedResponseBody []byte
	}{
		{
			w:                    httptest.NewRecorder(),
			r:                    rootRequest,
			expectedResponseCode: http.StatusOK,
			expectedResponseBody: []byte("d989e9a9-dab0-445c-ac85-274a2f3cd389: [2, 3, 5]\nd989e9a9-dab0-445c-ac85-274a2f3cd390: [2, 3, 5, 7]\n"),
		},
	}

	for _, c := range cases {
		getAllResults(c.w, c.r)

		if http.StatusOK != c.w.Code {
			t.Errorf("Status Code didn't match:\n\t%q\n\t%q", c.expectedResponseCode, c.w.Code)
		}

		if !strings.Contains(string(c.expectedResponseBody),  strings.Split(c.w.Body.String(), "\n")[0]) {
			t.Errorf("Body didn't match:\n\t%q\n\t%q", string(c.expectedResponseBody), c.w.Body.String())
		}
	}
}



