package lib

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"
)

type JLC struct {
	client *http.Client
	lock   *sync.Mutex
}

type JLCLibraryComponent struct {
	CID string `json:"componentCode"`
	LibraryComponent
}

func NewJLC() *JLC {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic("failed to create cookiejar for some reason")
	}

	return &JLC{
		client: &http.Client{Jar: jar},
		lock:   &sync.Mutex{},
	}
}

type jlcRequest interface {
	Method() string
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

func (r jlcSelectComponentListRequest) Method() string { return "selectSmtComponentList" }

type jlcSelectComponentListResponse struct {
	Code int `json:"code"`
	Data struct {
		ComponentPageInfo struct {
			EndRow          int                    `json:"endRow"`
			HasNextPage     bool                   `json:"hasNextPage"`
			HasPreviousPage bool                   `json:"hasPreviousPage"`
			IsFirstPage     bool                   `json:"isFirstPage"`
			IsLastPage      bool                   `json:"isLastPage"`
			List            []*JLCLibraryComponent `json:"list"`
		} `json:"componentPageInfo"`
	} `json:"data"`
}

func (jlc *JLC) clearCookies() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic("failed to create cookiejar for some reason")
	}

	jlc.client.Jar = jar
}

func (jlc *JLC) makeRequest(request jlcRequest, response interface{}) error {
	jlc.lock.Lock()
	jlc.clearCookies()
	go func() {
		defer jlc.lock.Unlock()
		time.Sleep(1500 * time.Millisecond)
	}()

	rd, wr := io.Pipe()

	go func() {
		enc := json.NewEncoder(wr)
		enc.Encode(request)
		wr.Close()
	}()

	req, err := http.NewRequest(
		"POST",
		"https://jlcpcb.com/shoppingCart/smtGood/"+request.Method(),
		bufio.NewReader(rd),
	)

	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}

	resp, err := jlc.client.Do(req)
	if err != nil {
		return err
	}

	//	buf, err := ioutil.ReadAll(resp.Body)
	//	fmt.Println(string(buf))
	//
	//	dec := json.NewDecoder(bytes.NewReader(buf))

	dec := json.NewDecoder(resp.Body)
	return dec.Decode(&response)
}

func (jlc *JLC) SelectComponentList(keyword string) ([]*LibraryComponent, error) {
	request := jlcSelectComponentListRequest{
		CurrentPage:  1,
		PageSize:     25,
		SearchSource: "search",
		Keyword:      &keyword,
	}

	response := jlcSelectComponentListResponse{}
	jlc.makeRequest(request, &response)

	components := make([]*LibraryComponent, len(response.Data.ComponentPageInfo.List))
	for i, component := range response.Data.ComponentPageInfo.List {
		component.ID = FromCID(component.CID)
		components[i] = &component.LibraryComponent
	}

	return components, nil
}

func (jlc *JLC) SelectBaseComponentList() (<-chan *LibraryComponent, <-chan error) {
	size := 100
	components := make(chan *LibraryComponent, size)
	errs := make(chan error)

	go func() {
		defer close(components)
		defer close(errs)

		page := 1
		for {
			request := jlcSelectComponentListRequest{
				ComponentLibraryType: "base",
				CurrentPage:          page,
				PageSize:             size,
				SearchSource:         "search",
			}

			response := jlcSelectComponentListResponse{}
			jlc.makeRequest(request, &response)

			if len(response.Data.ComponentPageInfo.List) == 0 {
				return
			}

			for _, component := range response.Data.ComponentPageInfo.List {
				component.ID = FromCID(component.CID)
				component.Basic = true

				components <- &component.LibraryComponent
			}

			page++
		}
	}()

	return components, errs
}
