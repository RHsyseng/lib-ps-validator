package lib_ps_validator

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	RES_VALID    = "valid"
	RES_EXPIRED  = "expired"
	RES_CONERROR = "conn-error"
	ERROR        = "error"
	CONN_ERROR   = "no such host"
)

var (
	payload          Payload
	resultOKArray    []byte
	resultKOArray    []byte
	resultKOConArray []byte
)

func Validate(input []byte) WebData {

	err := json.Unmarshal(input, &payload)
	if err != nil {
		fmt.Println(ERROR, err)
	}

	for k, v := range payload.Auths {
		//fmt.Println("Esto es una key: ", k)
		//fmt.Println("y esto su valor: ", v.Auth)
		//fmt.Println("y esto su valor: ", v.Auth)
		sDec, _ := b64.StdEncoding.DecodeString(v.Auth)
		auth := string(sDec)
		err, res := loginToRegistry(k, auth)
		if err != nil {
			resultKOConArray = append(resultKOConArray, k+"\n"...)
		} else if err == nil && res == RES_VALID {
			resultOKArray = append(resultOKArray, k+"\n"...)
		} else if err == nil && res == RES_EXPIRED {
			resultKOArray = append(resultKOArray, k+"\n"...)
		}
	}
	return WebData{input, string(resultOKArray), string(resultKOArray), string(resultKOConArray)}
}

func loginToRegistry(url, auth string) (error, string) {
	req, err := http.NewRequest("GET", "https://"+url+"/v2/auth", nil)
	if err != nil {
		return err, RES_CONERROR
	}
	s := strings.Split(auth, ":")
	req.SetBasicAuth(s[0], s[1])

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err, RES_CONERROR

	}

	defer resp.Body.Close()
	return nil, ""
}
