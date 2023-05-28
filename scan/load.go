package scan

import (
	"encoding/json"
	"os"

	"github.com/joomcode/errorx"
)

var (
	scr       *ScansRecord
	scansPath = "./scans.json"
)

func Load() error {
	f, err := os.Open(scansPath)
	if os.IsNotExist(err) {
		f, err = os.Create(scansPath)
		if err != nil {
			return errorx.Decorate(err, "unable to create file `%s`, after not detected", scansPath)
		}
	} else if err != nil {
		return errorx.Decorate(err, "unable to open file `%s`", scansPath)
	}

	st, err := f.Stat()
	if err != nil {
		return errorx.Decorate(err, "unable to stat file `%s`", scansPath)
	}

	if st.Size() > 0 {
		jd := json.NewDecoder(f)
		return jd.Decode(&scr)
	} else {
		scr = &ScansRecord{
			Scans: map[string]ScanRecords{},
			Diffs: map[string]DiffRecords{},
		}
	}

	return nil
}

func (s *ScansRecord) Flush() error {
	f, err := os.OpenFile(scansPath, os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return errorx.Decorate(err, "unable to open file '%s' for 'flush'", scansPath)
	}

	je := json.NewEncoder(f)
	return je.Encode(*scr)
}
