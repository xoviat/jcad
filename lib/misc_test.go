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

/*
	Test loading the schematic
*/
func TestParseSchematic(t *testing.T) {
	components := ParseSchematic("../test-data/input/STM32F4_Breakout.sch")

	_ = components
}

func TestGenerateBOM(t *testing.T) {

}
