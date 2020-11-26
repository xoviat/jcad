package lib

import (
	"encoding/xml"
	"testing"
)

func TestLoadEasy(t *testing.T) {
	/*
		TODO: load template from test-data
	*/
	template := ""

	epackage := EasyPackage{}
	xml.Unmarshal([]byte(template), &epackage)

}
