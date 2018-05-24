package asset

import (
	"testing"
	"net/http"
	"bytes"
	"io/ioutil"
	"encoding/json"
	"utils"
)

const checkMark = "\u2713"
const ballotX = "\u2717"

type BaseParam struct {
	Rc  int64 `json:"rc"`
	Msg string `json:"msg"`
}
type DataParam struct {
	Day    int64 `json:"day"`
	Month  int64 `json:"month"`
	Casual int64 `json:"casual"`
}

type Response struct {
	B *BaseParam `json:"base"`
	D *DataParam `json:"data"`
}

func TestWithdrawQuotaHandler_Handle(t *testing.T) {
	var url = "http://localhost:8080/asset/v1/withdrawal/quota/query"

	testParams := make([]map[string]int64, 5)
	//level0
	testParams[0] = map[string]int64{"uid": 20000006, "daily": 0, "monthly": 0}
	//level1
	testParams[1] = map[string]int64{"uid": 135021967, "daily": 0, "monthly": 0}
	//level2
	testParams[2] = map[string]int64{"uid": 153657784, "daily": 200, "monthly": 200}
	//level3
	testParams[3] = map[string]int64{"uid": 200100034, "daily": 200, "monthly": 200}
	//level4
	testParams[4] = map[string]int64{"uid": 199406315, "daily": 200, "monthly": 200}

	t.Log("Given the need to test query withdrawal quota different params.")
	{
		statusCode := 200
		for _, params := range testParams {

			t.Logf("\tWhen checking \"%s\" for status code \"%d\"", url, statusCode)
			{
				post := "{\"uid\":\"" + utils.Int642Str(params["uid"]) + "\"}"
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
							t.Logf("\t\tShould quota day: %d month: %d. %v", dat.D.Day, dat.D.Month, checkMark)
						} else {
							t.Errorf("\t\tShould quota day: %d month: %d %v, but response day: %d month: %d", params["daily"], params["monthly"], ballotX, dat.D.Day, dat.D.Month)
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