package lib

import "testing"


func TestFindMatchingComponent(t *testing.T) {
	library, err := NewLibrary("test_library") 
	if err != nil {
		t.FailNow()
	}
	
	if err := library.Import("../test-data/JLCPCB SMT Parts Library(20201003).xlsx"); err != nil {
		t.Fail()
	}

	component := library.FindMatching("TPS5420DR", "")
	if component == nil {
		t.Fail()
	}

}

func TestGenerateBOM(t *testing.T) {
	
}


/*
	Test loading and saving the schematic to ensure that no data is lost
*/
func TestLoadSaveSchematic(t *testing.T) {

}
