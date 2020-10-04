package lib

import "testing"


func TestImportComponents(t *testing.T) {
	library, err := NewLibrary("test_library") 
	if err != nil {
		t.FailNow()
	}
	
	library.Import("../test-data/JLCPCB SMT Parts Library(20201003).xlsx")
}


/*
	Test loading and saving the schematic to ensure that no data is lost
*/
func TestLoadSaveSchematic(t *testing.T) {

}
