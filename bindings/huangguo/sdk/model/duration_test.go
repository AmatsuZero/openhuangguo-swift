package model

import (
	"encoding/json"
	"testing"
)

func TestDetailAcceptsNumericDurationStringWithoutDroppingIdentity(t *testing.T) {
	for _, raw := range []string{`207`, `"207"`} {
		var detail Detail
		err := json.Unmarshal([]byte(`{"id":12,"duration":`+raw+`,"episodes":[{"ep_num":2,"duration":9}]}`), &detail)
		if err != nil {
			t.Fatal(err)
		}
		if detail.Duration != 207 || detail.ID != 12 || detail.Episodes[0].EpNum != 2 {
			t.Fatalf("fields lost: %+v", detail)
		}
		encoded, err := json.Marshal(detail)
		if err != nil {
			t.Fatal(err)
		}
		var decoded Detail
		if err := json.Unmarshal(encoded, &decoded); err != nil || decoded.Duration != 207 {
			t.Fatalf("numeric roundtrip: %s %v", encoded, err)
		}
	}
}

func TestDetailRejectsInvalidDurationInsteadOfHidingServerErrors(t *testing.T) {
	for _, raw := range []string{`"unknown"`, `true`, `1.5`, `"99999999999999999999999"`} {
		var detail Detail
		if json.Unmarshal([]byte(`{"id":12,"duration":`+raw+`}`), &detail) == nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
