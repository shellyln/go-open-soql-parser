package types

import (
	"encoding/json"
	"time"
)

func unmarshalSoqlFieldInfoValue(b json.RawMessage, typ SoqlFieldInfoType) (interface{}, error) {
	switch typ {
	case SoqlFieldInfo_Literal_Null:
		return nil, nil
	case SoqlFieldInfo_Literal_Int:
		{
			var v int64
			if err := json.Unmarshal(b, &v); err != nil {
				return nil, err
			}
			return v, nil
		}
	case SoqlFieldInfo_Literal_Float:
		{
			var v float64
			if err := json.Unmarshal(b, &v); err != nil {
				return nil, err
			}
			return v, nil
		}
	case SoqlFieldInfo_Literal_Bool:
		{
			var v bool
			if err := json.Unmarshal(b, &v); err != nil {
				return nil, err
			}
			return v, nil
		}
	case SoqlFieldInfo_Literal_String:
		{
			var v string
			if err := json.Unmarshal(b, &v); err != nil {
				return nil, err
			}
			return v, nil
		}
	case SoqlFieldInfo_Literal_Blob:
		{
			var v []byte
			if err := json.Unmarshal(b, &v); err != nil {
				return nil, err
			}
			return v, nil
		}
	case SoqlFieldInfo_Literal_Date,
		SoqlFieldInfo_Literal_DateTime,
		SoqlFieldInfo_Literal_Time:
		{
			var v time.Time
			if err := json.Unmarshal(b, &v); err != nil {
				return nil, err
			}
			return v, nil
		}
	case SoqlFieldInfo_Literal_DateTimeRange:
		{
			var v SoqlTimeRange
			if err := json.Unmarshal(b, &v); err != nil {
				return nil, err
			}
			return v, nil
		}
	case SoqlFieldInfo_Literal_List:
		{
			var v []SoqlListItem
			if err := json.Unmarshal(b, &v); err != nil {
				return nil, err
			}
			return v, nil
		}
	case SoqlFieldInfo_DateTimeLiteralName:
		{
			var v SoqlDateTimeLiteralName
			if err := json.Unmarshal(b, &v); err != nil {
				return nil, err
			}
			return v, nil
		}
	default:
		return nil, nil
	}
}

func unmarshalSoqlListItemValue(b json.RawMessage, typ SoqlFieldInfoType) (interface{}, error) {
	switch typ {
	case SoqlFieldInfo_ParameterizedValue,
		SoqlFieldInfo_DateTimeLiteralName:
		{
			var v string
			if err := json.Unmarshal(b, &v); err != nil {
				return nil, err
			}
			return v, nil
		}
	default:
		return unmarshalSoqlFieldInfoValue(b, typ)
	}
}
