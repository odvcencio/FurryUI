package progress

import "testing"

func TestProgressManagerLifecycle(t *testing.T) {
	pm := NewProgressManager()
	var snapshots [][]Progress

	pm.SetOnChange(func(items []Progress) {
		copied := make([]Progress, len(items))
		copy(copied, items)
		snapshots = append(snapshots, copied)
	})

	pm.Start("build-1", "Build", ProgressDeterminate, 10)
	pm.Update("build-1", 5)
	pm.Done("build-1")

	if len(snapshots) < 4 {
		t.Fatalf("expected at least 4 snapshots, got %d", len(snapshots))
	}

	started := snapshots[1]
	if len(started) != 1 {
		t.Fatalf("expected 1 progress after start, got %d", len(started))
	}
	if started[0].ID != "build-1" {
		t.Errorf("expected id build-1, got %q", started[0].ID)
	}
	if started[0].Label != "Build" {
		t.Errorf("expected label Build, got %q", started[0].Label)
	}

	updated := snapshots[2]
	if len(updated) != 1 {
		t.Fatalf("expected 1 progress after update, got %d", len(updated))
	}
	if updated[0].Current != 5 {
		t.Errorf("expected current 5, got %d", updated[0].Current)
	}
	if updated[0].Percent < 0.49 || updated[0].Percent > 0.51 {
		t.Errorf("expected percent around 0.5, got %f", updated[0].Percent)
	}

	done := snapshots[len(snapshots)-1]
	if len(done) != 0 {
		t.Fatalf("expected no progress after done, got %d", len(done))
	}
}
