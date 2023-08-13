package yaml

import (
	"gopkg.in/yaml.v3"
	"strings"
	"time"
)

const dateFormat = "02/01/2006"

type DateType struct {
	time.Time
}

func (dt *DateType) MarshalYAML() (interface{}, error) {
	return dt.Time.Format(dateFormat), nil
}

func (dt *DateType) UnmarshalYAML(value *yaml.Node) error {
	var stringIn string
	err := value.Decode(&stringIn)
	if err != nil {
		return nil
	}
	pt, err := time.Parse(dateFormat, strings.TrimSpace(stringIn))
	if err != nil {
		return err
	}
	dt.Time = pt
	return nil
}
