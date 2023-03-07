package lib

import (
	"testing"
)

func TestSelectComponentList(t *testing.T) {
	jlc := NewJLC()

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

	req := jlcSelectComponentListRequest{
		ComponentLibraryType: "base",
		CurrentPage:          2,
		PageSize:             25,
		SearchSource:         "search",
	}

	jlc.selectComponentList(&req)
}
