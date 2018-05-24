package asset

import (
	"testing"
	"net/http"
	"bytes"
	"io/ioutil"
	"encoding/json"
)

const checkMark = "\u2713"
const ballotX = "\u2717"

type BaseParam struct {
	Rc  int64 `json:"rc"`
	Msg string `json:"msg"`
}
type DataParam struct {
	Day    string `json:"day"`
	Month  string `json:"month"`
	Casual string `json:"casual"`
}

type Response struct {
	B *BaseParam `json:"base"`
	D *DataParam `json:"data"`
}

func TestWithdrawQuotaHandler_Handle(t *testing.T) {
	var url = "http://localhost:8080/asset/v1/withdrawal/quota/query"

	testParams := make([]map[string]string, 5)
	//level0
	testParams[0] = map[string]string{"uid": "20000006", "daily": "0.00000000", "monthly": "0.00000000"}
	//level1
	testParams[1] = map[string]string{"uid": "135021967", "daily": "0.00000000", "monthly": "0.00000000"}
	//level2
	testParams[2] = map[string]string{"uid": "153657784", "daily": "200.00000000", "monthly": "200.00000000"}
	//level3
	testParams[3] = map[string]string{"uid": "200100034", "daily": "200.00000000", "monthly": "200.00000000"}
	//level4
	testParams[4] = map[string]string{"uid": "199406315", "daily": "200.00000000", "monthly": "200.00000000"}

	t.Log("Given the need to test query withdrawal quota different params.")
	{
		statusCode := 200
		for _, params := range testParams {

			t.Logf("\tWhen checking \"%s\" for status code \"%d\"", url, statusCode)
			{
				post := "{\"uid\":\"" + params["uid"] + "\"}"
				var jsonStr = []byte(post)
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
					var dat = Response{}
					if err := json.Unmarshal(body, &dat); err == nil {
						if dat.D.Day == params["daily"] && dat.D.Month == params["monthly"] {
							t.Logf("\t\tShould quota day: %s month: %s. %v", dat.D.Day, dat.D.Month, checkMark)
						} else {
							t.Errorf("\t\tShould quota day: %s month: %s %v, but response day: %s month: %s", params["daily"], params["monthly"], ballotX, dat.D.Day, dat.D.Month)
						}
					} else {
						panic(err)
					}
				} else {
					t.Errorf("\t\tShould receive a \"%d\" status. %v %v", statusCode, ballotX, resp.StatusCode)
				}
			}

		}

	}
}
