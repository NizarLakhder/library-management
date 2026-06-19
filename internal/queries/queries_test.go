package queries

import "testing"

// Every query must fail cleanly (no panic, non-nil error) when handed a nil
// database handle, which is the state before the user connects.
func TestAllQueriesErrorOnNilDB(t *testing.T) {
	if len(All) == 0 {
		t.Fatal("All is empty")
	}
	for _, q := range All {
		data, err := q.Execute(nil)
		if err == nil {
			t.Errorf("query %q: expected error on nil db, got nil", q.Label)
		}
		if data != nil {
			t.Errorf("query %q: expected nil data on error, got %v", q.Label, data)
		}
	}
}

// Each query must carry consistent, non-empty UI metadata with a unique label.
func TestAllQueriesMetadata(t *testing.T) {
	seen := make(map[string]bool)
	for i, q := range All {
		if q.Label == "" {
			t.Errorf("query #%d has an empty label", i)
		}
		if seen[q.Label] {
			t.Errorf("duplicate query label %q", q.Label)
		}
		seen[q.Label] = true

		if len(q.ColumnWidths) == 0 {
			t.Errorf("query %q has no column widths", q.Label)
		}
		for j, w := range q.ColumnWidths {
			if w <= 0 {
				t.Errorf("query %q: column width #%d must be positive, got %v", q.Label, j, w)
			}
		}
		if q.Execute == nil {
			t.Errorf("query %q has a nil Execute function", q.Label)
		}
	}
}
