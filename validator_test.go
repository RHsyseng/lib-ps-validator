package lib_ps_validator

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var (
	Output_Empty = WebData{
		ResultOK:  "",
		ResultKO:  "",
		ResultCon: "",
	}
	Output_Valid = WebData{
		ResultOK:  "127.0.0.1",
		ResultKO:  "",
		ResultCon: "",
	}
	json_malformed      = []byte(`aaaa`)
	json_empty          = []byte(`{}`)
	json_authEMpty      = []byte(`{"auths"}`)
	json_KO_WITHOUTAUTH = []byte(`{
	"auths": {
		"localhost": {
			"auth": "mockedFAKE"
		}
	  }
     }
    `)
	json_OK = []byte(`{
	"auths": {
		"localhost:8443": {
			"auth": "aG9sYTpwYXNz"
		}
	 }
}
`)
)

func TestValidate_jsonMalformed(t *testing.T) {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	Validate(json_malformed)

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	assert.Contains(t, string(out), "error")
}

func TestValidate_jsonEmpty(t *testing.T) {

	actual := Validate(json_empty)
	assert.Equal(t, Output_Empty.ResultOK, actual.ResultOK)
	assert.Equal(t, Output_Empty.ResultKO, actual.ResultKO)
	assert.Equal(t, Output_Empty.ResultCon, actual.ResultCon)
}

func TestValidate_AuthEmpty(t *testing.T) {

	actual := Validate(json_authEMpty)
	assert.Equal(t, Output_Empty.ResultOK, actual.ResultOK)
	assert.Equal(t, Output_Empty.ResultKO, actual.ResultKO)
	assert.Equal(t, Output_Empty.ResultCon, actual.ResultCon)
}

func TestValidate_JsonKO_WITHOUT(t *testing.T) {

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	mux.HandleFunc("/someotherroute/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	httptest.NewTLSServer(mux)

	actual := Validate(json_KO_WITHOUTAUTH)
	assert.Equal(t, Output_Empty.ResultOK, actual.ResultOK)

}

func TestValidate_JsonOK_FirstReq_AUTH(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	mux.HandleFunc("/someotherroute/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})

	ts := httptest.NewTLSServer(mux)
	conn, _ := net.Dial("tcp", ts.Listener.Addr().String())
	defer conn.Close()

	json_OK2 := []byte(`{
	"auths": {
		"` + ts.URL[len("https://"):] + `": {
			"auth": "aG9sYTpwYXNz"
		}
	 }
}
`)

	actual := Validate(json_OK2)
	assert.Contains(t, actual.ResultOK, Output_Valid.ResultOK)

}

func TestValidate_JsonOK_SecondReq_V2(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	ts := httptest.NewTLSServer(mux)
	conn, _ := net.Dial("tcp", ts.Listener.Addr().String())
	defer conn.Close()

	json_OK2 := []byte(`{
	"auths": {
		"` + ts.URL[len("https://"):] + `": {
			"auth": "aG9sYTpwYXNz"
		}
	 }
}
`)
	actual := Validate(json_OK2)
	assert.Contains(t, actual.ResultOK, Output_Valid.ResultOK)
}

func TestValidate_JsonKO_tokenUnauthorized_FirstReq(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
	})
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
	})

	ts := httptest.NewTLSServer(mux)
	conn, _ := net.Dial("tcp", ts.Listener.Addr().String())
	defer conn.Close()

	json_OK2 := []byte(`{
	"auths": {
		"` + ts.URL[len("https://"):] + `": {
			"auth": "aG9sYTpwYXNz"
		}
	 }
}
`)
	actual := Validate(json_OK2)
	assert.Contains(t, actual.ResultKO, Output_Valid.ResultKO)

}

func TestValidate_JsonKO_tokenUnauthorized_SecondReq(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
	})

	ts := httptest.NewTLSServer(mux)
	conn, _ := net.Dial("tcp", ts.Listener.Addr().String())
	defer conn.Close()

	json_OK2 := []byte(`{
	"auths": {
		"` + ts.URL[len("https://"):] + `": {
			"auth": "aG9sYTpwYXNz"
		}
	 }
}
`)
	actual := Validate(json_OK2)
	assert.Contains(t, actual.ResultKO, Output_Valid.ResultKO)

}

func TestValidate_JsonKO_ConnIssue_FirstReq(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
	})
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
	})

	ts := httptest.NewTLSServer(mux)
	conn, _ := net.Dial("tcp", ts.Listener.Addr().String())
	defer conn.Close()

	json_OK2 := []byte(`{
	"auths": {
		"` + ts.URL[len("https://xxxxxx"):] + `": {
			"auth": "aG9sYTpwYXNz"
		}
	 }
}
`)
	actual := Validate(json_OK2)
	assert.Contains(t, actual.ResultCon, Output_Valid.ResultCon)

}

func TestValidate_JsonKO_ConnIssue_SecReq(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/auth", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	mux.HandleFunc("/v2/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	})

	ts := httptest.NewTLSServer(mux)
	conn, _ := net.Dial("tcp", ts.Listener.Addr().String())
	defer conn.Close()

	json_OK2 := []byte(`{
	"auths": {
		"` + ts.URL[len("https://"):] + `": {
			"auth": "aG9sYTpwYXNz"
		}
	 }
}
`)
	actual := Validate(json_OK2)
	assert.Contains(t, actual.ResultCon, Output_Valid.ResultCon)

}
