package lib

import (
	"fmt"
	"testing"
)

func TestSelectBaseComponentList(t *testing.T) {
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

	components, err := jlc.SelectBaseComponentList()

	for _, component := range components {
		fmt.Println(component.CID())
	}

	_, _ = components, err
}
