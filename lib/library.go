package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/blevesearch/bleve"
	"github.com/boltdb/bolt"
)

var (
	COMPONENTS_BKT     = []byte("components")             // Contains all of the JLCPCB components
	CATEGORIES_BKT     = []byte("categories")             // Associates the categories with all contained within
	PACKAGES_BKT       = []byte("packages")               // Contains a list of eagle packages
	SYMBOLS_BKT        = []byte("symbols")                // Contains a list of eagle symbols
	COMPONENTS_ASC_BKT = []byte("component-associations") // Associates a BoardComponent Key with a LibraryComponent
	PACKAGE_ASC_BKT    = []byte("package-associations")   // Associates a KiCad package with a JLCPCB package
)

var (
	re1 *regexp.Regexp = regexp.MustCompile("[^a-zA-Z]+")
	re2 *regexp.Regexp = regexp.MustCompile(`[0-9\.]+(nF|pF|uF)`)
	re3 *regexp.Regexp = regexp.MustCompile(`[0-9\.]+(k|MOhms|KOhms|Ohms)`)
	re4 *regexp.Regexp = regexp.MustCompile(`[0-9\.]+(uH|mH)`)

	iprefixes = map[string]string{ // Associates a component value with a list of actual components
		"R": "index-resistors",
		"C": "index-capacitors",
		"L": "index-inductors",
	}
)

type Library struct {
	root  string
	db    *bolt.DB
	index bleve.Index
}

/*
Import a library from an excel or csv file
*/
func (l *Library) Import(rows <-chan []string) error {
	fromID := func(ID string) int {
		i, err := strconv.Atoi(strings.TrimPrefix(ID, "C"))
		if err != nil {
			return 0
		}

		return i
	}

	l.db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket(COMPONENTS_BKT)
		tx.DeleteBucket(CATEGORIES_BKT)
		for _, iprefix := range iprefixes {
			tx.DeleteBucket([]byte(iprefix))
		}

		tx.CreateBucket(COMPONENTS_BKT)
		tx.CreateBucket(CATEGORIES_BKT)
		for _, iprefix := range iprefixes {
			tx.CreateBucket([]byte(iprefix))
		}

		return nil
	})

	i := 0
	/*
		amount per transaction
	*/
	k := 10000
	row := []string{""}
	ok := true
	categories := make(map[string][]int)
	indexes := make(map[string]map[string][]int)
	indexes["R"] = make(map[string][]int)
	indexes["C"] = make(map[string][]int)
	indexes["L"] = make(map[string][]int)

	for len(row) != 0 {
		if err := l.db.Update(func(tx *bolt.Tx) error {
			components := tx.Bucket(COMPONENTS_BKT)

			/*
				Do it this way to save memory
			*/
			for j := 0; j < k; j++ {
				if row, ok = <-rows; !ok {
					return nil
				}

				component := LibraryComponent{
					ID:             fromID(row[0]),
					FirstCategory:  row[1],
					SecondCategory: row[2],
					Part:           row[3],
					Package:        row[4],
					SolderJoint:    row[5],
					Manufacturer:   row[6],
					LibraryType:    row[7],
					Description:    row[8],
				}

				//				fmt.Printf(
				//					"%1.0d, %s, %s, %s\n",
				//					component.ID, component.FirstCategory, component.SecondCategory, component.Part,
				//				)

				for _, each := range []string{component.FirstCategory, component.SecondCategory} {
					if _, ok := categories[each]; !ok {
						categories[each] = []int{}
					}
					categories[each] = append(categories[each], component.ID)
				}

				if component.LibraryType == "Basic" && component.Prefix() != "" &&
					component.Value() != "" {

					if _, ok := indexes[component.Prefix()][component.Value()]; !ok {
						indexes[component.Prefix()][component.Value()] = []int{}
					}
					indexes[component.Prefix()][component.Value()] = append(
						indexes[component.Prefix()][component.Value()], component.ID,
					)

					//					if component.Prefix() == "R" {
					//						fmt.Printf(
					//							"%s + %s: %1.0d\n", component.Prefix(), component.Value(), component.ID,
					//						)
					//					}
				}

				bytes, err := Marshal(component)
				if err != nil {
					return err
				}

				err = components.Put([]byte(component.CID()), bytes)
				if err != nil {
					return err
				}

				i++
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return l.db.Update(func(tx *bolt.Tx) error {
		bcategories := tx.Bucket(CATEGORIES_BKT)
		for category, components := range categories {
			bytes, err := Marshal(components)
			if err != nil {
				return err
			}

			// fmt.Println(category)

			err = bcategories.Put([]byte(category), bytes)
			if err != nil {
				return err
			}
		}
		for prefix, bname := range iprefixes {
			bucket := tx.Bucket([]byte(bname))
			for value, components := range indexes[prefix] {
				bytes, err := Marshal(components)
				if err != nil {
					return err
				}

				err = bucket.Put([]byte(value), bytes)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

/*
export assocations to an excel file
*/
func (l *Library) ExportAssociations() <-chan []string {
	rows := make(chan []string, 100)
	go func() {
		defer close(rows)
		l.db.View(func(tx *bolt.Tx) error {
			bassociations := tx.Bucket(COMPONENTS_ASC_BKT)
			cur := bassociations.Cursor()

			for key, val := cur.First(); key != nil; key, val = cur.Next() {
				rows <- []string{string(key), string(val)}
			}

			return nil
		})
	}()

	return rows
}

/*
import assocations from an excel file
*/
func (l *Library) ImportAssocations(rows <-chan []string) error {
	// Skip processing footprint assocations for now

	//	l.db.Update(func(tx *bolt.Tx) error {
	//		bassociations := tx.Bucket(COMPONENTS_ASC_BKT)
	//		bfootprints := tx.Bucket(PACKAGE_ASC_BKT)
	//
	//		if lcomponent == nil {
	//			return bassociations.Delete(bcomponent.Key())
	//		}
	//
	//		err := bassociations.Put(bcomponent.Key(), []byte(lcomponent.CID()))
	//		if err != nil {
	//			return err
	//		}
	//
	//		return bfootprints.Put([]byte(bcomponent.Package), []byte(lcomponent.Package))
	//	})

	return nil
}

func NewDefaultLibrary() (*Library, error) {
	path := filepath.Join(GetLocalAppData(), "jcad")
	os.MkdirAll(path, 0777)

	return NewLibrary(path)
}

/*
Create or open library from root
*/
func NewLibrary(root string) (*Library, error) {
	db, err := bolt.Open(filepath.Join(root, "jcad.db"), 0777, nil)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists(COMPONENTS_BKT)
		tx.CreateBucketIfNotExists(CATEGORIES_BKT)
		tx.CreateBucketIfNotExists(PACKAGES_BKT)
		tx.CreateBucketIfNotExists(SYMBOLS_BKT)
		tx.CreateBucketIfNotExists(COMPONENTS_ASC_BKT)
		tx.CreateBucketIfNotExists(PACKAGE_ASC_BKT)
		for _, iprefix := range iprefixes {
			tx.CreateBucketIfNotExists([]byte(iprefix))
		}

		return nil
	})

	var index bleve.Index
	ipath := filepath.Join(root, "jcad.index")
	if Exists(ipath) {
		index, err = bleve.Open(ipath)
	} else {
		index, err = bleve.New(ipath, bleve.NewIndexMapping())
	}
	if err != nil {
		return nil, err
	}

	return &Library{
		root:  root,
		db:    db,
		index: index,
	}, nil
}

/*
referred to as 'component'
*/
type LibraryComponent struct {
	ID             int
	FirstCategory  string
	SecondCategory string
	Part           string
	Package        string
	SolderJoint    string
	Manufacturer   string
	LibraryType    string
	Description    string
	Rotation       float64
}

func (lc *LibraryComponent) CID() string {
	return fmt.Sprintf("C%1.1d", lc.ID)
}

func (lc *LibraryComponent) Prefix() string {
	switch lc.FirstCategory {
	case "Capacitors":
		return "C"
	case "Resistors":
		return "R"
	case "Inductors & Chokes & Transformers":
		return "L"
	}

	return ""
}

/*
Attempt to determine the value from the description
*/
func (lc *LibraryComponent) Value() string {
	switch lc.FirstCategory {
	case "Capacitors":
		return NormalizeValue(re2.FindString(lc.Description))
	case "Resistors":
		return NormalizeValue(re3.FindString(lc.Description))
	case "Inductors & Chokes & Transformers":
		return NormalizeValue(re4.FindString(lc.Description))
	}

	return ""
}

/*
Determine whether it is possible to place the component using the SMT process
*/
func (l *Library) CanAssemble(bcomponent *BoardComponent) bool {
	switch re1.ReplaceAllString(bcomponent.Designator, "") {
	case "J":
		return false
	case "H":
		return false
	case "G":
		return false
	case "JP":
		return false
	case "DRA":
		return false
	case "DS":
		return false
	case "SW":
		return false
	}

	return true
}

func (l *Library) SetRotation(component *LibraryComponent, rotation float64) {
	err := l.db.Update(func(tx *bolt.Tx) error {
		bcomponents := tx.Bucket(COMPONENTS_BKT)
		component.Rotation = rotation

		bytes, err := Marshal(component)
		if err != nil {
			return err
		}

		err = bcomponents.Put([]byte(component.CID()), bytes)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		fmt.Printf("error in set-rotation: %s\n", err)
	}
}

/*
Find the best suitable library componentx, given the board components

Prefer a basic part, if available
Require the package (footprint) to match

Return nil if no part found
*/
func (l *Library) FindMatching(bcomponent *BoardComponent) []*LibraryComponent {
	/*
		This method is not trivial! The comment may refer to a part number,
		a resistor value, such as 2k2, or a capacitor value. A list of possible
		combinations for the parameters is given below:

		U	AMS1117-3.3		SOT-223-3_TabPin2,25.225001
		U	STM32F405RGT6	LQFP-64_10x10mm_P0.5mm
		F	500mA			Fuse_0603_1608Metric
		FB	100 @ 100 MHz	L_0805_2012Metric
		C	100nf			C_0402_1005Metric
		R	220				R_0402_1005Metric
		R	2k2				R_0603_1608Metric

		The desired results are given below:

		Power Management ICs				AMS1117-3.3		SOT-223					Positive Fixed 1.3V @ 800mA 15V 3.3V 1A
		Embedded Processors & Controllers	STM32F405RGT6	LQFP-64_10.0x10.0x0.5P	STMicroelectronics
		N/A
	*/

	iname, ok := iprefixes[bcomponent.Prefix()]
	if !ok {
		return []*LibraryComponent{}
	}

	// todo: filter further using package associations
	components := []*LibraryComponent{}
	pkg := ""
	l.db.View(func(tx *bolt.Tx) error {
		bcomponents := tx.Bucket(COMPONENTS_BKT)
		bindex := tx.Bucket([]byte(iname))
		bpackages := tx.Bucket(PACKAGE_ASC_BKT)

		IDs := []int{}
		if bytes := bindex.Get([]byte(bcomponent.Value())); bytes != nil {
			Unmarshal(bytes, &IDs)
		}

		if bytes := bpackages.Get([]byte(bcomponent.Package)); bytes != nil {
			pkg = string(bytes)
		}

		components = make([]*LibraryComponent, len(IDs))
		for i, ID := range IDs {
			if bytes := bcomponents.Get(
				[]byte((&LibraryComponent{ID: ID}).CID()),
			); bytes != nil {
				Unmarshal(bytes, &components[i])
			}
		}

		return nil
	})

	if pkg == "" {
		return components
	}

	i := 0
	for _, component := range components {
		if component.Package != pkg {
			continue
		}

		components[i] = component
		i++
	}
	components = components[:i]

	return components
}

func (l *Library) Exact(id string) *LibraryComponent {
	if id == "C0" {
		return &LibraryComponent{}
	}

	component := LibraryComponent{}

	l.db.View(func(tx *bolt.Tx) error {
		bcomponents := tx.Bucket(COMPONENTS_BKT)
		if bytes := bcomponents.Get([]byte(id)); bytes != nil {
			Unmarshal(bytes, &component)
		}

		return nil
	})

	return &component
}

func (l *Library) FindAssociated(bcomponent *BoardComponent) *LibraryComponent {
	if !l.CanAssemble(bcomponent) {
		return &LibraryComponent{}
	}

	component := LibraryComponent{}
	skip := false

	l.db.View(func(tx *bolt.Tx) error {
		bassociations := tx.Bucket(COMPONENTS_ASC_BKT)
		bcomponents := tx.Bucket(COMPONENTS_BKT)

		// fmt.Printf("FindAssociated: %s\n", bcomponent.Key())
		ID := ""
		if bytes := bassociations.Get(bcomponent.Key()); bytes != nil {
			ID = string(bytes)
		}

		skip = ID == "C0"

		if bytes := bcomponents.Get([]byte(ID)); bytes != nil {
			Unmarshal(bytes, &component)
		}

		return nil
	})

	if component.ID == 0 && !skip {
		return nil
	}

	return &component
}

func (l *Library) FindInCategory(category string) []*LibraryComponent {
	components := []*LibraryComponent{}

	l.db.View(func(tx *bolt.Tx) error {
		bcategories := tx.Bucket(CATEGORIES_BKT)
		bcomponents := tx.Bucket(COMPONENTS_BKT)

		IDs := []int{}
		if bytes := bcategories.Get([]byte(category)); bytes != nil {
			Unmarshal(bytes, &IDs)
		}

		for _, ID := range IDs {
			component := LibraryComponent{}
			if bytes := bcomponents.Get([]byte((&LibraryComponent{ID: ID}).CID())); bytes != nil {
				Unmarshal(bytes, &component)
			}

			components = append(components, &component)
		}

		return nil
	})

	return components
}

func (l *Library) Associate(bcomponent *BoardComponent, lcomponent *LibraryComponent) {
	// fmt.Printf("associating %s with %s\n", string(bcomponent.Key()), lcomponent.CID())
	l.db.Update(func(tx *bolt.Tx) error {
		bassociations := tx.Bucket(COMPONENTS_ASC_BKT)
		bfootprints := tx.Bucket(PACKAGE_ASC_BKT)

		if lcomponent == nil {
			return bassociations.Delete(bcomponent.Key())
		}

		err := bassociations.Put(bcomponent.Key(), []byte(lcomponent.CID()))
		if err != nil {
			return err
		}

		return bfootprints.Put([]byte(bcomponent.Package), []byte(lcomponent.Package))
	})

}
