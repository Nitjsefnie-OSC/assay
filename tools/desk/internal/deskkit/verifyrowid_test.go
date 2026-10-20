package deskkit

import (
	"encoding/json"
	"testing"
)

// regression: #1731
func TestVerifyRowIDCompatibilityAndRefusal(t *testing.T) {
	for _, raw := range []string{`1`, `"6a"`, `"6A"`, `"6"`} {
		var id VerifyRowID
		if err := json.Unmarshal([]byte(raw), &id); err != nil {
			t.Fatalf("%s: %v", raw, err)
		}
		encoded, err := json.Marshal(id)
		if err != nil {
			t.Fatal(err)
		}
		var again VerifyRowID
		if err = json.Unmarshal(encoded, &again); err != nil || again != id {
			t.Fatalf("identity changed: %s -> %s", raw, encoded)
		}
	}
	for _, raw := range []string{`0`, `-1`, `1.5`, `1e2`, `null`, `true`, `""`, `"06"`, `"6aa"`, `" 6a"`, `"6/1"`, `"<6>"`, `{}`} {
		var id VerifyRowID
		if json.Unmarshal([]byte(raw), &id) == nil {
			t.Errorf("accepted malformed identity %s", raw)
		}
	}
	for _, raw := range []string{
		`{"rows":[6,"6a"],"rows_total":2,"rows_passed":0,"row_results":[{"row":6,"result":"could-not-run"},{"row":"6a","result":"could-not-run"}]}`,
		`{"rows":[6],"rows_total":1,"rows_passed":0,"row_results":[{"row":6,"result":"could-not-run"}]}`,
	} {
		if _, err := ParseJoinInput([]byte(raw)); err != nil {
			t.Errorf("compatible input %s: %v", raw, err)
		}
	}
	for _, raw := range []string{`{"rows":[6,"6"]}`, `{"rows":[0]}`, `{"rows":["6aa"]}`, `{"rows_total":2,"row_results":[{"row":6,"result":"pass"},{"row":"6","result":"pass"}]}`} {
		if _, err := ParseJoinInput([]byte(raw)); err == nil {
			t.Errorf("writer accepted ambiguous identity %s", raw)
		}
	}
}

// regression: #1731
func TestVerifyRowIDWriterRejectsHeldPassOverlap(t *testing.T) {
	raw := []byte(`{"rows":["6a"],"rows_total":2,"rows_passed":1,"row_results":[{"row":6,"result":"could-not-run"},{"row":"6a","result":"pass"}]}`)
	if _, err := ParseJoinInput(raw); err == nil {
		t.Fatal("writer accepted held row also claiming PASS")
	}
}
