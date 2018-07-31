// Code generated by "enumer -type=MaterialBehavior -json"; DO NOT EDIT.

package mapping

import (
	"encoding/json"
	"fmt"
)

const _MaterialBehaviorName = "ScaleNoneScaleWidthScaleHeightScaleAll"

var _MaterialBehaviorIndex = [...]uint8{0, 9, 19, 30, 38}

func (i MaterialBehavior) String() string {
	if i < 0 || i >= MaterialBehavior(len(_MaterialBehaviorIndex)-1) {
		return fmt.Sprintf("MaterialBehavior(%d)", i)
	}
	return _MaterialBehaviorName[_MaterialBehaviorIndex[i]:_MaterialBehaviorIndex[i+1]]
}

var _MaterialBehaviorValues = []MaterialBehavior{0, 1, 2, 3}

var _MaterialBehaviorNameToValueMap = map[string]MaterialBehavior{
	_MaterialBehaviorName[0:9]:   0,
	_MaterialBehaviorName[9:19]:  1,
	_MaterialBehaviorName[19:30]: 2,
	_MaterialBehaviorName[30:38]: 3,
}

// MaterialBehaviorString retrieves an enum value from the enum constants string name.
// Throws an error if the param is not part of the enum.
func MaterialBehaviorString(s string) (MaterialBehavior, error) {
	if val, ok := _MaterialBehaviorNameToValueMap[s]; ok {
		return val, nil
	}
	return 0, fmt.Errorf("%s does not belong to MaterialBehavior values", s)
}

// MaterialBehaviorValues returns all values of the enum
func MaterialBehaviorValues() []MaterialBehavior {
	return _MaterialBehaviorValues
}

// IsAMaterialBehavior returns "true" if the value is listed in the enum definition. "false" otherwise
func (i MaterialBehavior) IsAMaterialBehavior() bool {
	for _, v := range _MaterialBehaviorValues {
		if i == v {
			return true
		}
	}
	return false
}

// MarshalJSON implements the json.Marshaler interface for MaterialBehavior
func (i MaterialBehavior) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for MaterialBehavior
func (i *MaterialBehavior) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("MaterialBehavior should be a string, got %s", data)
	}

	var err error
	*i, err = MaterialBehaviorString(s)
	return err
}
