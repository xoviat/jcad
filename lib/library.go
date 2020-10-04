package lib

import (
	"bytes"
	"encoding/gob"

	"os"
	"path/filepath"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/blevesearch/bleve"
	"github.com/boltdb/bolt"
)

type ComponentType int

const (
	Resistor ComponentType = iota
	Capacitor
	Inductor
	Crystal
	IC
)

var ComponentTypes = []ComponentType{Resistor, Capacitor, Inductor, Crystal, IC}

func (s ComponentType) String() string {
	return [...]string{"resistor", "capactior", "crystal", "ic"}[s]
}

/*
	LCSC Part	First Category	Second Category	MFR.Part	Package	Solder Joint	Manufacturer	Library Type	Description	Datasheet	Price	Stock
	C25725	Resistors	Resistor Networks & Arrays	4D02WGJ0103TCE	0402_x4	8	Uniroyal Elec	Basic	Resistor Networks & Arrays 10KOhms ±5% 1/16W 0402_x4 RoHS	https://datasheet.lcsc.com/szlcsc/Uniroyal-Elec-4D02WGJ0103TCE_C25725.pdf	1-199:0.006956522,200-:0.002717391	79847
*/

func Exists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	}

	return true
}

/*
	return an encoded object as bytes
*/
func Marshal(v interface{}) ([]byte, error) {
	b := new(bytes.Buffer)
	err := gob.NewEncoder(b).Encode(v)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

/*
	return a decoded object from bytes
*/
func Unmarshal(data []byte, v interface{}) error {
	b := bytes.NewBuffer(data)
	return gob.NewDecoder(b).Decode(v)
}

type Library struct {
	root  string
	db    *bolt.DB
	index bleve.Index
}

/*
	Import a library from an excel file
*/
func (l *Library) Import(src string) error {
	f, err := excelize.OpenFile(src)
	if err != nil {
		return err
	}

	sheet := f.GetSheetList()[0]
	rows, err := f.Rows(sheet)
	if err != nil {
		return err
	}
	/*
		new plan: have a long-running indexing worker

		chindex := make(chan *LibraryComponent, 500)
		chdone := make(chan bool, 10)
		workers := 8
		for i := 0; i < workers; i++ {
			go func() {
				for {
					component := <-chindex
					if component == nil {
						break
					}

					l.index.Index(component.ID, *component)
				}
				chdone <- true
			}()
		}

	*/
	l.db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("components"))
		tx.DeleteBucket([]byte("unindexed"))
		tx.DeleteBucket([]byte("parts"))

		tx.CreateBucket([]byte("components"))
		tx.CreateBucket([]byte("unindexed"))
		tx.CreateBucket([]byte("parts"))

		return nil
	})

	chrows := make(chan []string, 100)
	go func() {
		for {
			if end := !rows.Next(); end {
				chrows <- []string{}

				return
			}

			row, err := rows.Columns()
			if err != nil {
				continue
			}

			if len(row) < 9 {
				continue
			}

			chrows <- row
		}
	}()

	i := 0
	/*
		amount per transaction
	*/
	k := 2000
	row := []string{""}
	for len(row) != 0 {
		if err := l.db.Update(func(tx *bolt.Tx) error {
			components := tx.Bucket([]byte("components"))
			unindexed := tx.Bucket([]byte("unindexed"))
			parts := tx.Bucket([]byte("parts"))
			pmap := make(map[string][]string)

			zmap := make(map[ComponentType]map[string]string)
			for _, componentType := range ComponentTypes {
				zmap[componentType] = make(map[string]string)
			}

			updateparts := func() error {
				bytes := []byte("")
				eIDs := []string{}
				for part, IDs := range pmap {
					eIDs = []string{}
					if bytes = parts.Get([]byte(part)); bytes != nil {
						Unmarshal(bytes, &eIDs)
					}
					bytes, _ = Marshal(append(eIDs, IDs...))
					err = parts.Put([]byte(part), bytes)
					if err != nil {
						return err
					}
				}

				return nil
			}

			/*
				Do it this way to save memory
			*/
			for j := 0; j < k; j++ {
				if row = <-chrows; len(row) == 0 {
					return updateparts()
				}

				component := LibraryComponent{
					ID:             row[0],
					FirstCategory:  row[1],
					SecondCategory: row[2],
					Part:           row[3],
					Package:        row[4],
					SolderJoint:    row[5],
					Manufacturer:   row[6],
					LibraryType:    row[7],
					Description:    row[8],
				}

				if _, ok := pmap[component.Part]; !ok {
					pmap[component.Part] = []string{}
				}
				pmap[component.Part] = append(pmap[component.Part], component.ID)

				bytes, err := Marshal(component)
				if err != nil {
					return err
				}

				err = components.Put([]byte(component.ID), bytes)
				if err != nil {
					return err
				}

				/*
					ids are removed from unindexed once they are indexed
				*/
				err = unindexed.Put([]byte(component.ID), []byte(""))
				if err != nil {
					return err
				}

				i++
			}

			return updateparts()
		}); err != nil {
			return err
		}
	}

	return nil
}

/*
	Create or open library from root
*/
func NewLibrary(root string) (*Library, error) {
	db, err := bolt.Open(filepath.Join(root, "JCAD.db"), 0777, nil)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists([]byte("components"))
		tx.CreateBucketIfNotExists([]byte("unindexed"))
		tx.CreateBucketIfNotExists([]byte("parts"))

		return nil
	})

	var index bleve.Index
	ipath := filepath.Join(root, "JCAD.index")
	if Exists(ipath) {
		index, err = bleve.Open(ipath)
	} else {
		index, err = bleve.New(ipath, bleve.NewIndexMapping())
	}

	return &Library{
		root:  root,
		db:    db,
		index: index,
	}, nil
}

type LibraryComponent struct {
	ID             string
	FirstCategory  string
	SecondCategory string
	Part           string
	Package        string
	SolderJoint    string
	Manufacturer   string
	LibraryType    string
	Description    string
}

/*
	Find library components, given a search string
*/
func (l *Library) Find(description string) []*LibraryComponent {
	query := bleve.NewMatchQuery(description)
	query.SetField("Description")

	result, err := l.index.Search(bleve.NewSearchRequest(query))
	if err != nil {
		return []*LibraryComponent{}
	}

	components := []*LibraryComponent{}
	for _, hit := range result.Hits {
		_ = hit
		components = append(components, &LibraryComponent{})
	}

	return []*LibraryComponent{}
}

/*
	Find the best suitable component, given the comment and package

	Prefer a basic part, if available
	Require the package (footprint) to match

	Return nil if no part foundl
*/
func (l *Library) FindMatching(prefix, comment, pkg string) *LibraryComponent {
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

	components := []*LibraryComponent{}
	part := comment

	l.db.View(func(tx *bolt.Tx) error {
		bcomponents := tx.Bucket([]byte("components"))
		bparts := tx.Bucket([]byte("parts"))

		IDs := []string{}
		if bytes := bparts.Get([]byte(part)); bytes != nil {
			Unmarshal(bytes, &IDs)
		}

		if len(IDs) == 0 {
			return nil
		}

		for _, ID := range IDs {
			component := LibraryComponent{}
			if bytes := bcomponents.Get([]byte(ID)); bytes != nil {
				Unmarshal(bytes, &component)
			}

			components = append(components, &component)
		}

		return nil
	})

	if len(components) == 0 {
		return nil
	}

	/*
		TODO: check package
	*/

	return components[0]
}
