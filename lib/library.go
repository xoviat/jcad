package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/boltdb/bolt"
)

var (
	COMPONENTS_BKT     = []byte("components")             // Contains all of the JLCPCB components
	COMPONENTS_ASC_BKT = []byte("component-associations") // Associates a BoardComponent Key with a LibraryComponent
	PACKAGE_ASC_BKT    = []byte("package-associations")   // Associates a KiCad package with a JLCPCB package
)

var (
	re1 *regexp.Regexp = regexp.MustCompile("[^a-zA-Z]+")
	re2 *regexp.Regexp = regexp.MustCompile(`[0-9\.]+(pF|nF|uF|mF)`)
	re3 *regexp.Regexp = regexp.MustCompile(`[0-9\.]+(mΩ|Ω|kΩ|MΩ)`)
	re4 *regexp.Regexp = regexp.MustCompile(`[0-9\.]+(nH|uH|mH)`)
)

var (
	BASIC_CAT_MAP = map[string][]string{
		"R": {"Chip Resistor - Surface Mount"},
		"C": {"Multilayer Ceramic Capacitors MLCC - SMD/SMT", "Tantalum Capacitors"},
		"L": {"Inductors (SMD)"},
	}
	BASIC_FP_MAP = map[string]string{
		"0402": "_0402_1005Metric",
		"0603": "_0603_1608Metric",
		"0805": "_0805_2012Metric",
		"1206": "_1206_3216Metric",
	}

	BASIC_CAT_MAP_COMP map[string]string
)

func init() {
	BASIC_CAT_MAP_COMP = make(map[string]string)

	for prefix, categories := range BASIC_CAT_MAP {
		for _, category := range categories {
			BASIC_CAT_MAP_COMP[category] = prefix
		}
	}
}

type Library struct {
	root string
	db   *bolt.DB
}

/*
Import all of the basic parts into the library
*/
func (l *Library) ImportBasic(components <-chan *LibraryComponent, errs <-chan error) error {
	l.db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket(COMPONENTS_BKT)
		tx.CreateBucket(COMPONENTS_BKT)

		return nil
	})

	return l.db.Update(func(tx *bolt.Tx) error {
		bcomponents := tx.Bucket(COMPONENTS_BKT)
		bassociations := tx.Bucket(COMPONENTS_ASC_BKT)

		for component := range components {
			bytes, err := Marshal(component)
			if err != nil {
				return err
			}

			if key := component.BasicKey(); key != "" {
				err = bassociations.Put([]byte(key), []byte(component.CID()))
			}
			if err != nil {
				return err
			}

			err = bcomponents.Put([]byte(component.CID()), bytes)
			if err != nil {
				return err
			}
		}

		select {
		case err := <-errs:
			return err
		default:
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

	return l.db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket(COMPONENTS_ASC_BKT)
		bassociations, err := tx.CreateBucket(COMPONENTS_ASC_BKT)
		if err != nil {
			return err
		}

		for row := range rows {
			bassociations.Put([]byte(row[0]), []byte(row[1]))
		}

		return nil
	})
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
		tx.CreateBucketIfNotExists(COMPONENTS_ASC_BKT)
		tx.CreateBucketIfNotExists(PACKAGE_ASC_BKT)

		return nil
	})

	return &Library{
		root: root,
		db:   db,
	}, nil
}

/*
referred to as 'library component'
*/
type LibraryComponent struct {
	ID           int64
	Category     string `json:"componentTypeEn"`
	Part         string `json:"componentModelEn"`
	Package      string `json:"componentSpecificationEn"`
	Manufacturer string `json:"componentBrandEn"`
	Description  string `json:"describe"`
	Basic        bool
}

func (lc LibraryComponent) CID() string {
	return fmt.Sprintf("C%1.1d", lc.ID)
}

/*
compute the key for a basic resistor, capacitor or inductor
*/
func (lc LibraryComponent) BasicKey() string {
	if !lc.Basic {
		return ""
	}

	prefix := lc.Prefix()
	if prefix == "" {
		return ""
	}

	pkg, _ := BASIC_FP_MAP[lc.Package]
	if pkg == "" {
		return ""
	}

	return BoardComponent{
		Designator: prefix,
		Comment:    lc.Value(),
		Package:    prefix + pkg,
	}.StringKey()
}

func (lc LibraryComponent) Prefix() string {
	prefix, _ := BASIC_CAT_MAP_COMP[lc.Category]

	return prefix
}

/*
Attempt to determine the value from the description
*/
func (lc LibraryComponent) Value() string {
	switch lc.Prefix() {
	case "C":
		return NormalizeValue(re2.FindString(lc.Description))
	case "R":
		return NormalizeValue(re3.FindString(lc.Description))
	case "L":
		return NormalizeValue(re4.FindString(lc.Description))
	}

	return ""
}

func (l *Library) Exact(cid string) *LibraryComponent {
	if cid == "C0" {
		return &LibraryComponent{}
	}

	component := LibraryComponent{}
	l.db.View(func(tx *bolt.Tx) error {
		bcomponents := tx.Bucket(COMPONENTS_BKT)
		if bytes := bcomponents.Get([]byte(cid)); bytes != nil {
			Unmarshal(bytes, &component)
		}

		return nil
	})

	if component.ID != 0 {
		return &component
	}

	return &LibraryComponent{ID: FromCID(cid)}
}

/*
returns:
  - nil if no associated component
  - LibraryComponent{ID:0} if this component is to be skipped
  - LibraryComponent{ID:int64} if an associated component is found
*/
func (l *Library) FindAssociated(bcomponent *BoardComponent) *LibraryComponent {
	if !bcomponent.CanAssemble() {
		return &LibraryComponent{}
	}

	component := LibraryComponent{}
	cid := ""

	l.db.View(func(tx *bolt.Tx) error {
		bassociations := tx.Bucket(COMPONENTS_ASC_BKT)
		bcomponents := tx.Bucket(COMPONENTS_BKT)

		// fmt.Printf("FindAssociated: %s\n", bcomponent.Key())
		if bytes := bassociations.Get(bcomponent.Key()); bytes != nil {
			cid = string(bytes)
		}

		if bytes := bcomponents.Get([]byte(cid)); bytes != nil {
			Unmarshal(bytes, &component)
		}

		return nil
	})

	if component.ID == 0 && cid != "C0" && cid != "" {
		return &LibraryComponent{ID: FromCID(cid)}
	} else if component.ID == 0 && cid != "C0" {
		return nil
	}

	return &component
}

func (l *Library) Associate(bcomponent *BoardComponent, lcomponent *LibraryComponent) {
	// fmt.Printf("associating %s with %s\n", string(bcomponent.Key()), lcomponent.CID())
	l.db.Update(func(tx *bolt.Tx) error {
		bassociations := tx.Bucket(COMPONENTS_ASC_BKT)
		bfootprints := tx.Bucket(PACKAGE_ASC_BKT)
		bcomponents := tx.Bucket(COMPONENTS_BKT)

		bytes, err := Marshal(lcomponent)
		if err != nil {
			return err
		}

		err = bcomponents.Put([]byte(lcomponent.CID()), bytes)
		if err != nil {
			return err
		}

		if lcomponent == nil {
			return bassociations.Delete(bcomponent.Key())
		}

		err = bassociations.Put(bcomponent.Key(), []byte(lcomponent.CID()))
		if err != nil {
			return err
		}

		return bfootprints.Put([]byte(bcomponent.Package), []byte(lcomponent.Package))
	})

}
