package data

type ScQuery struct {
	ScAddress string   `json:"scAddress"`
	FuncName  string   `json:"funcName"`
	Args      []string `json:"args"`
}

type ScIntResult struct {
	Data struct {
		Data string `json:"data"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}

type ScQueryResult struct {
	Data struct {
		Data struct {
			ReturnData [][]byte `json:"ReturnData"`
		} `json:"data"`
	} `json:"data"`
	Error string `json:"error"`
	Code  string `json:"code"`
}
