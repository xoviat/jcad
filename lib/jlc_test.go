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

	// categories := make(map[string]struct{})
	components, err := jlc.SelectBaseComponentList()
	//	for component := range components {
	//		categories[component.Category] = struct{}{}
	//	}
	//
	//	for category := range categories {
	//		fmt.Println(category)
	//	}

	//	for component := range components {
	//		_ = component
	//	}

	//	for component := range components {
	//		key := component.BasicKey()
	//		name := component.CID()
	//		category := component.Description
	//		pkg := component.Package
	//
	//		fmt.Printf("%s : %s : %s : %s\n", name, category, pkg, key)
	//	}

	_, _ = components, err
}
func TestSelectComponentList(t *testing.T) {
	jlc := NewJLC()

	components, err := jlc.SelectComponentList("6TPE330MAP")
	for _, component := range components {
		fmt.Println(component.Description)
	}

	_, _ = components, err
}
