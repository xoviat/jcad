package lib

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
)

type JLC struct {
	client *http.Client
}

func NewJLC() *JLC {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic("failed to create cookiejar for some reason")
	}

	return &JLC{client: &http.Client{Jar: jar}}
}

type jlcSelectComponentListRequest struct {
	ComponentAttributes    []string `json:"componentAttributes"`
	ComponentBrand         string   `json:"componentBrand"`
	ComponentLibraryType   string   `json:"componentLibraryType"`
	ComponentSpecification *string  `json:"componentSpecification"`
	CurrentPage            int      `json:"currentPage"`
	FirstSortId            string   `json:"firstSortId"`
	FirstSortName          string   `json:"firstSortName"`
	FirstSortNameNew       string   `json:"firstSortNameNew"`
	Keyword                *string  `json:"keyword"`
	PageSize               int      `json:"pageSize"`
	SearchSource           string   `json:"searchSource"`
	SecondSortName         string   `json:"secondSortName"`
	StockFlag              *string  `json:"stockFlag"`
	StockSort              *string  `json:"stockSort"`
}

type jlcSelectComponentListResponse struct {
	Code int `json:"code"`
	Data struct {
		ComponentPageInfo struct {
			EndRow          int                 `json:"endRow"`
			HasNextPage     bool                `json:"hasNextPage"`
			HasPreviousPage bool                `json:"hasPreviousPage"`
			IsFirstPage     bool                `json:"isFirstPage"`
			IsLastPage      bool                `json:"isLastPage"`
			List            []*LibraryComponent `json:"list"`
		} `json:"componentPageInfo"`
	} `json:"data"`
}

func (jlc *JLC) selectComponentList(request *jlcSelectComponentListRequest) ([]*LibraryComponent, error) {
	rd, wr := io.Pipe()

	go func() {
		enc := json.NewEncoder(wr)
		enc.Encode(request)
		wr.Close()
	}()

	req, err := http.NewRequest(
		"POST",
		"https://jlcpcb.com/shoppingCart/smtGood/selectSmtComponentList",
		bufio.NewReader(rd),
	)

	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return []*LibraryComponent{}, err
	}

	resp, err := jlc.client.Do(req)
	if err != nil {
		return []*LibraryComponent{}, err
	}

	var response jlcSelectComponentListResponse
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&response)

	_ = resp

	return []*LibraryComponent{}, err

	/*
		componentAttributes	:  []
		componentBrand	: 	null
		componentLibraryType	: 	"base"
		componentSpecification	: 	null
		currentPage	: 	2
		firstSortId	: 	""
		firstSortName	: 	""
		firstSortNameNew	: 	""
		keyword	: 	null
		pageSize	: 	25
		searchSource	: 	"search"
		secondSortName	: 	""
		stockFlag	: 	null
		stockSort	: 	null
	*/
}

func (jlc *JLC) getComponentDetail() {
	/* https://jlcpcb.com/shoppingCart/smtGood/getComponentDetail?componentLcscId=1899 */
}
