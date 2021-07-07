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

func Validate(input []byte) WebData {

	var (
		payload          Payload
		resultOKArray    []byte
		resultKOArray    []byte
		resultKOConArray []byte
	)

	err := json.Unmarshal(input, &payload)
	if err != nil {
		fmt.Println(ERROR, err)
	}

	resultKOConArray = []byte("")
	resultKOArray = []byte("")
	resultOKArray = []byte("")

	for k, v := range payload.Auths {

		sDec, _ := b64.StdEncoding.DecodeString(v.Auth)
		auth := string(sDec)

		err, res := loginToRegistry(k, auth)
		if err != nil || res == RES_CONERROR {
			resultKOConArray = append(resultKOConArray, k+"\n"...)
		} else if res == RES_VALID {
			resultOKArray = append(resultOKArray, k+"\n"...)
		} else if res == RES_EXPIRED {
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

	resp, err2 := http.DefaultClient.Do(req)

	if err2 != nil {
		return err2, RES_CONERROR
	} else if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		return nil, RES_VALID
	} else if resp.StatusCode == http.StatusNotFound {
		return nil, RES_CONERROR
	} else if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, RES_EXPIRED
	} else {
		return err, RES_CONERROR
	}

	defer resp.Body.Close()
	return nil, ""

}
