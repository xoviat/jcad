package lib

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

type JLC struct {
	lock *sync.Mutex
}

type JLCLibraryComponent struct {
	CID string `json:"componentCode"`
	LibraryComponent
}

func NewJLC() *JLC {
	return &JLC{
		lock: &sync.Mutex{},
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

func (jlc *JLC) makeRequest(request jlcRequest, response interface{}) error {
	jlc.lock.Lock()
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
		"https://jlcpcb.com/api/overseas-pcb-order/v1/shoppingCart/smtGood/"+request.Method(),
		bufio.NewReader(rd),
	)

	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
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

func (jlc *JLC) SelectComponentList(keyword string) (map[int64]*LibraryComponent, error) {
	request := jlcSelectComponentListRequest{
		CurrentPage:  1,
		PageSize:     25,
		SearchSource: "search",
		Keyword:      &keyword,
	}

	response := jlcSelectComponentListResponse{}
	jlc.makeRequest(request, &response)

	components := make(map[int64]*LibraryComponent)
	for _, component := range response.Data.ComponentPageInfo.List {
		component.ID = FromCID(component.CID)
		components[component.ID] = &component.LibraryComponent
	}

	return components, nil
}

func (jlc *JLC) Exact(cid string) *LibraryComponent {
	components, err := jlc.SelectComponentList(cid)
	if err != nil {
		return &LibraryComponent{ID: FromCID(cid), Description: "No description available"}
	}

	component, ok := components[FromCID(cid)]
	if !ok {
		return &LibraryComponent{ID: FromCID(cid), Description: "No description available"}
	}

	return component
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
