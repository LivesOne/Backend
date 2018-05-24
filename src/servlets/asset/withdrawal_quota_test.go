package asset

import (
	"testing"
	"net/http"
	"bytes"
	"io/ioutil"
	"fmt"
)

const checkMark = "\u2713"
const ballotX = "\u2717"

func TestWithdrawQuotaHandler_Handle(t *testing.T) {
	var url = "http://localhost:8080/asset/v1/withdrawal/quota/query"
	post := "{\"uid\":\"133902136\"}"
	var jsonStr = []byte(post)

	t.Log("Given the need to test query withdrawal quota different params.")
	{
		statusCode := 200
		t.Logf("\tWhen checking \"%s\" for status code \"%d\"", url, statusCode)
		{
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			t.Log("\t\tShould be able to make the Get call.", checkMark)

			defer resp.Body.Close()
			if resp.StatusCode == statusCode {
				t.Logf("\t\tShould receive a \"%d\" status. %v", statusCode, checkMark)
				body, _ := ioutil.ReadAll(resp.Body)
				fmt.Println(string(body))
			} else {
				t.Errorf("\t\tShould receive a \"%d\" status. %v %v", statusCode, ballotX, resp.StatusCode)
			}
		}

	}
}
