package job

import (
	"encoding/json"
	"time"
)

// metadata contains the data parsed from "started.json" or "finished.json".
type metadata struct {
	time   time.Time
	result string
}

// UnmarshalJSON implements json.Unmarshal for metadata. The purpose of this
// method is to parse a time.Time out of an epoch timestamp.
func (m *metadata) UnmarshalJSON(src []byte) error {
	var data struct {
		Timestamp int64
		Result    string
	}
	err := json.Unmarshal(src, &data)
	if err != nil {
		return err
	}

	t := time.Unix(data.Timestamp, 0)

	*m = metadata{
		time:   t.In(time.UTC),
		result: data.Result,
	}

	return nil
}

type templateParams struct {
	Name       string
	Parameters []struct {
		Name  string
		Value string
	}
}

// UnmarshalJSON implements json.Unmarshal for templateinstance. The purpose is to get template
// params like cluster type
func (t *templateParams) UnmarshalJSON(src []byte) error {
	var data struct {
		Items []struct {
			Metadata struct {
				Name string
			}
			Spec struct {
				Template struct {
					Parameters []struct {
						Name  string
						Value string
					}
				}
			}
		}
	}
	err := json.Unmarshal(src, &data)
	if err != nil {
		return err
	}

	*t = templateParams{
		Name:       data.Items[0].Metadata.Name,
		Parameters: data.Items[0].Spec.Template.Parameters,
	}

	return nil
}
