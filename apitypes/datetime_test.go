package microjson

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDate(t *testing.T) {

	val := Date{
		Year:  2016,
		Month: 12,
		Day:   30,
	}

	b, err := json.Marshal(val)
	if err != nil {
		t.Fatal(err)
	}

	str := strings.TrimSpace(string(b))
	if str != `"2016-12-30"` {
		t.Errorf("got %s", str)
	}

	v2 := Date{}
	if err := json.Unmarshal([]byte(`"2016-12-30"`), &v2); err != nil {
		t.Fatal(err)
	}
	if v2.Year != 2016 || v2.Month != 12 || v2.Day != 30 {
		t.Errorf("Wrong date: %#v", v2)
	}
}
