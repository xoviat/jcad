package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/blevesearch/bleve"
	"github.com/boltdb/bolt"
)

var (
	re1 *regexp.Regexp = regexp.MustCompile("[^a-zA-Z]+")
)

type Library struct {
	root  string
	db    *bolt.DB
	index bleve.Index
}

/*
	Indexes the library. This function may take a long time.
*/
func (l *Library) Index() error {
	// l.index.Index(component.ID, *component)

	return nil
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

	l.db.Update(func(tx *bolt.Tx) error {
		tx.DeleteBucket([]byte("components"))
		tx.DeleteBucket([]byte("unindexed"))

		tx.CreateBucket([]byte("components"))
		tx.CreateBucket([]byte("unindexed"))

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

			/*
				Do it this way to save memory
			*/
			for j := 0; j < k; j++ {
				if row = <-chrows; len(row) == 0 {
					return nil
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

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func NewDefaultLibrary() (*Library, error) {
	path := filepath.Join(GetLocalAppData(), "JCAD")
	os.MkdirAll(path, 0777)

	return NewLibrary(path)
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
		tx.CreateBucketIfNotExists([]byte("associations"))
		tx.CreateBucketIfNotExists([]byte("packages"))

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
	Rotation       float64
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
	}

	return true
}

func (l *Library) SetRotation(ID string, rotation float64) {
	component := LibraryComponent{}
	err := l.db.Update(func(tx *bolt.Tx) error {
		bcomponents := tx.Bucket([]byte("components"))

		if bytes := bcomponents.Get([]byte(ID)); bytes != nil {
			Unmarshal(bytes, &component)
		}

		component.Rotation = rotation

		bytes, err := Marshal(component)
		if err != nil {
			return err
		}

		err = bcomponents.Put([]byte(component.ID), bytes)
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

	return []*LibraryComponent{}
}

func (l *Library) Exact(id string) *LibraryComponent {
	component := LibraryComponent{}

	l.db.View(func(tx *bolt.Tx) error {
		bcomponents := tx.Bucket([]byte("components"))
		if bytes := bcomponents.Get([]byte(id)); bytes != nil {
			Unmarshal(bytes, &component)
		}

		return nil
	})

	return &component
}

func (l *Library) FindAssociated(bcomponent *BoardComponent) *LibraryComponent {
	component := LibraryComponent{}

	l.db.View(func(tx *bolt.Tx) error {
		bassociations := tx.Bucket([]byte("associations"))
		bcomponents := tx.Bucket([]byte("components"))

		ID := ""
		key := bcKey(bcomponent)
		if bytes := bassociations.Get(key); bytes != nil {
			ID = string(bytes)
		}

		if bytes := bcomponents.Get([]byte(ID)); bytes != nil {
			Unmarshal(bytes, &component)
		}

		return nil
	})

	if component.ID == "" {
		return nil
	}

	return &component
}

func (l *Library) Associate(bcomponent *BoardComponent, lcomponent *LibraryComponent) {
	l.db.Update(func(tx *bolt.Tx) error {
		bassociations := tx.Bucket([]byte("associations"))
		bfootprints := tx.Bucket([]byte("packages"))

		key := bcKey(bcomponent)

		err := bassociations.Put(key, []byte(lcomponent.ID))
		if err != nil {
			return err
		}

		return bfootprints.Put([]byte(bcomponent.Package), []byte(lcomponent.Package))
	})

}
