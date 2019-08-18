package webanalyzer

import (
	"testing"
)

func TestSimple(t *testing.T) {
	symbolTab := map[string]bool{
		"1":     true,
		"2":     false,
		"3":     true,
		"4":     false,
		"name1": true,
		"name2": false,
		"name3": true,
		"name4": false,
	}

	conds := map[string]bool{
		"1":                               true,
		"2":                               false,
		"name1":                           true,
		"name2":                           false,
		"((((name1))))":                   true,
		"name1 and name2":                 false,
		"name1 and not name2":             true,
		"name1 or name2":                  true,
		"name2 or name1 and name2":        false,
		"name1 and not (name1 and name2)": true,
		"(name1 or name2) and (name3 and (1 or 2))": true,
	}

	p := Parser{}
	for k, v := range conds {
		r, err := p.Parse(k, symbolTab)
		if err != nil {
			t.Error("unknown error", err)
		}

		if r != v {
			t.Errorf("unexpect result: %s = %#v", k, v)
		}
	}
}

func TestInvalid(t *testing.T) {
	symbolTab := map[string]bool{
		"include space": false,
		"2":             false,
		"name1":         true,
		"name2":         false,
	}

	conds := []string{
		"include space",
		"name1 name2",
		"name1 or",
		"()",
		"and name1",
		"not_exists_name",
		"name1 or not_exists_name",
		"name1 and not",
		"(name1 and name2",
	}

	p := Parser{}
	for _, v := range conds {
		r, err := p.Parse(v, symbolTab)
		if err == nil {
			t.Error("expect error, got nil")
		}
		if r != false {
			t.Errorf("r should be false got %v, error: %v", r, err)
			t.Error(v)
		}
	}
}
