package types

import (
	"encoding/json"
)

func (t SoqlFieldInfoType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

func (t *SoqlFieldInfoType) UnmarshalJSON(b []byte) error {
	s := string(b)
	for i := SoqlFieldInfoType(0); i < SoqlFieldInfo_EndOfConstDefinitions_; i++ {
		if `"`+i.String()+`"` == s {
			*t = i
			break
		}
	}
	return nil
}

func (t *SoqlFieldInfo) UnmarshalJSON(b []byte) error {
	t2 := &soqlFieldInfo_unmarshal{}
	if err := json.Unmarshal(b, t2); err != nil {
		return err
	}

	t.Type = t2.Type
	t.Name = t2.Name
	t.AliasName = t2.AliasName
	t.Parameters = t2.Parameters
	t.SubQuery = t2.SubQuery
	t.NotSelected = t2.NotSelected
	t.Aggregated = t2.Aggregated
	t.Hints = t2.Hints
	t.ColumnId = t2.ColumnId
	t.ColIndex = t2.ColIndex
	t.ViewId = t2.ViewId
	t.Key = t2.Key

	if v, err := unmarshalSoqlFieldInfoValue(t2.Value, t2.Type); err != nil {
		return err
	} else {
		t.Value = v
	}

	return nil
}

func (t *SoqlListItem) UnmarshalJSON(b []byte) error {
	t2 := &soqlFieldInfo_unmarshal{}
	if err := json.Unmarshal(b, t2); err != nil {
		return err
	}

	t.Type = t2.Type

	if v, err := unmarshalSoqlListItemValue(t2.Value, t2.Type); err != nil {
		return err
	} else {
		t.Value = v
	}

	return nil
}

func (t SoqlConditionOpcode) MarshalJSON() ([]byte, error) {
	return []byte(`"` + t.String() + `"`), nil
}

func (t *SoqlConditionOpcode) UnmarshalJSON(b []byte) error {
	s := string(b)
	for i := SoqlConditionOpcode(0); i < soqlConditionOpcode_EndOfConstDefinitions_; i++ {
		if `"`+i.String()+`"` == s {
			*t = i
			break
		}
	}
	return nil
}

func (t *SoqlCondition) MarshalJSON() ([]byte, error) {
	if t.Opcode == SoqlConditionOpcode_FieldInfo {
		t2 := soqlCondition_marshalAll{
			Opcode: t.Opcode,
			Value:  t.Value,
		}
		return json.Marshal(t2)
	} else {
		t2 := soqlCondition_marshalOp{
			Opcode: t.Opcode,
		}
		return json.Marshal(t2)
	}
}
