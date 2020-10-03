package lib

import (
	"bytes"
	"encoding/gob"
	"os"
	"path/filepath"
	"strconv"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/blevesearch/bleve"
	"github.com/boltdb/bolt"
)

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

	scomponents := make(chan *LibraryComponent, 100)
	go func() {
		for {
			component := <-scomponents
			if component == nil {
				break
			}

			l.index.Index(component.ID, *component)
		}
	}()

	f, err := excelize.OpenFile(src)
	if err != nil {
		return err
	}

	cv := func(row int, col string) string {
		v, _ := f.GetCellValue("", col+strconv.Itoa(row+2))

		return v
	}

	i := 0
	k := 500
	end := false
	for {
		/*
			500 per transaction
		*/
		l.db.Update(func(tx *bolt.Tx) error {
			components := tx.Bucket([]byte("components"))
			/*
				Do it this way to save memory
			*/
			for j := 0; j < k; j++ {
				component := LibraryComponent{
					ID:             cv(i, "A"),
					FirstCategory:  cv(i, "B"),
					SecondCategory: cv(i, "C"),
					MFRPart:        cv(i, "D"),
					Package:        cv(i, "E"),
					SolderJoint:    cv(i, "F"),
					Manufacturer:   cv(i, "G"),
					LibraryType:    cv(i, "H"),
					Description:    cv(i, "I"),
				}

				if component.ID == "" {
					end = true

					return nil
				}

				bytes, err := Marshal(component)
				if err != nil {
					return err
				}

				err = components.Put([]byte(component.ID), bytes)
				if err != nil {
					return err
				}

				scomponents <- &component
				i++
			}

			return nil
		})

		if end {
			break
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
	MFRPart        string
	Package        string
	SolderJoint    string
	Manufacturer   string
	LibraryType    string
	Description    string
}

/*
	Find library components, given a search string
*/
func (l *Library) Find(text string, typ string) []*LibraryComponent {
	query := bleve.NewMatchQuery(text)
	query.SetField("Title")

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
