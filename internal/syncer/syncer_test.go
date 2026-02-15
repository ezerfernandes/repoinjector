package syncer

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initBareCloneEnv creates a bare repo with an initial commit and a clone.
func initBareCloneEnv(t *testing.T) (bareDir, cloneDir string) {
	t.Helper()

	bareDir = filepath.Join(t.TempDir(), "bare.git")
	run(t, "", "git", "init", "--bare", bareDir)

	cloneDir = filepath.Join(t.TempDir(), "clone")
	run(t, "", "git", "clone", bareDir, cloneDir)

	writeFile(t, filepath.Join(cloneDir, "README.md"), "init")
	run(t, cloneDir, "git", "add", ".")
	run(t, cloneDir, "git", "commit", "-m", "initial commit")
	run(t, cloneDir, "git", "push")

	return
}

// pushCommitFromSecondClone pushes a new commit to bare from a second clone,
// making the original clone behind by 1.
func pushCommitFromSecondClone(t *testing.T, bareDir string) {
	t.Helper()
	clone2 := filepath.Join(t.TempDir(), "clone2")
	run(t, "", "git", "clone", bareDir, clone2)
	writeFile(t, filepath.Join(clone2, "new.txt"), "from clone2")
	run(t, clone2, "git", "add", ".")
	run(t, clone2, "git", "commit", "-m", "upstream commit")
	run(t, clone2, "git", "push")
}

func run(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, out)
	}
	return string(out)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestCheckStatusCurrent(t *testing.T) {
	_, cloneDir := initBareCloneEnv(t)

	s := CheckStatus(cloneDir, true)
	if s.State != StateCurrent {
		t.Errorf("expected current, got %s: %s", s.State, s.Detail)
	}
	if s.Branch == "" {
		t.Error("expected non-empty branch")
	}
}

func TestCheckStatusBehind(t *testing.T) {
	bareDir, cloneDir := initBareCloneEnv(t)
	pushCommitFromSecondClone(t, bareDir)

	// Use noFetch=false so it fetches the new commit
	s := CheckStatus(cloneDir, false)
	if s.State != StateBehind {
		t.Errorf("expected behind, got %s: %s", s.State, s.Detail)
	}
	if s.Behind != 1 {
		t.Errorf("expected 1 behind, got %d", s.Behind)
	}
}

func TestCheckStatusDirty(t *testing.T) {
	_, cloneDir := initBareCloneEnv(t)
	writeFile(t, filepath.Join(cloneDir, "dirty.txt"), "uncommitted")

	s := CheckStatus(cloneDir, true)
	if s.State != StateDirty {
		t.Errorf("expected dirty, got %s: %s", s.State, s.Detail)
	}
	if !s.Dirty {
		t.Error("expected Dirty=true")
	}
}

func TestCheckStatusDiverged(t *testing.T) {
	bareDir, cloneDir := initBareCloneEnv(t)

	// Push from second clone
	pushCommitFromSecondClone(t, bareDir)

	// Make a local commit too
	writeFile(t, filepath.Join(cloneDir, "local.txt"), "local change")
	run(t, cloneDir, "git", "add", ".")
	run(t, cloneDir, "git", "commit", "-m", "local commit")

	// Fetch to see the divergence
	run(t, cloneDir, "git", "fetch")

	s := CheckStatus(cloneDir, true) // noFetch since we already fetched
	if s.State != StateDiverged {
		t.Errorf("expected diverged, got %s: %s", s.State, s.Detail)
	}
	if s.Ahead != 1 || s.Behind != 1 {
		t.Errorf("expected 1/1, got %d/%d", s.Ahead, s.Behind)
	}
}

func TestCheckStatusNoUpstream(t *testing.T) {
	repo := t.TempDir()
	run(t, "", "git", "init", repo)
	writeFile(t, filepath.Join(repo, "file.txt"), "content")
	run(t, repo, "git", "add", ".")
	run(t, repo, "git", "commit", "-m", "init")

	s := CheckStatus(repo, true)
	if s.State != StateNoUpstream {
		t.Errorf("expected no-upstream, got %s: %s", s.State, s.Detail)
	}
}

func TestCheckStatusAhead(t *testing.T) {
	_, cloneDir := initBareCloneEnv(t)

	// Make a local commit without pushing
	writeFile(t, filepath.Join(cloneDir, "local.txt"), "local")
	run(t, cloneDir, "git", "add", ".")
	run(t, cloneDir, "git", "commit", "-m", "local commit")

	s := CheckStatus(cloneDir, true)
	if s.State != StateAhead {
		t.Errorf("expected ahead, got %s: %s", s.State, s.Detail)
	}
	if s.Ahead != 1 {
		t.Errorf("expected 1 ahead, got %d", s.Ahead)
	}
}

func TestSyncRepoPulls(t *testing.T) {
	bareDir, cloneDir := initBareCloneEnv(t)
	pushCommitFromSecondClone(t, bareDir)

	r := SyncRepo(cloneDir, SyncOptions{NoFetch: false})
	if r.Action != "pulled" {
		t.Errorf("expected pulled, got %s: %s", r.Action, r.PostDetail)
	}
	if r.State != StatePulled {
		t.Errorf("expected state pulled, got %s", r.State)
	}
}

func TestSyncRepoDryRun(t *testing.T) {
	bareDir, cloneDir := initBareCloneEnv(t)
	pushCommitFromSecondClone(t, bareDir)

	r := SyncRepo(cloneDir, SyncOptions{DryRun: true, NoFetch: false})
	if r.Action != "dry-run" {
		t.Errorf("expected dry-run, got %s: %s", r.Action, r.PostDetail)
	}
}

func TestSyncRepoSkipsDirty(t *testing.T) {
	bareDir, cloneDir := initBareCloneEnv(t)
	pushCommitFromSecondClone(t, bareDir)
	writeFile(t, filepath.Join(cloneDir, "dirty.txt"), "uncommitted")

	r := SyncRepo(cloneDir, SyncOptions{NoFetch: false})
	if r.Action != "skipped" {
		t.Errorf("expected skipped, got %s: %s", r.Action, r.PostDetail)
	}
}

func TestSyncRepoAutoStash(t *testing.T) {
	bareDir, cloneDir := initBareCloneEnv(t)
	pushCommitFromSecondClone(t, bareDir)
	writeFile(t, filepath.Join(cloneDir, "dirty.txt"), "uncommitted")

	r := SyncRepo(cloneDir, SyncOptions{AutoStash: true, NoFetch: false})
	if r.Action != "pulled" {
		t.Errorf("expected pulled, got %s: %s", r.Action, r.PostDetail)
	}
}

func TestSyncRepoSkipsCurrent(t *testing.T) {
	_, cloneDir := initBareCloneEnv(t)

	r := SyncRepo(cloneDir, SyncOptions{NoFetch: true})
	if r.Action != "skipped" {
		t.Errorf("expected skipped, got %s: %s", r.Action, r.PostDetail)
	}
}

func TestSyncRepoSkipsDiverged(t *testing.T) {
	bareDir, cloneDir := initBareCloneEnv(t)
	pushCommitFromSecondClone(t, bareDir)

	writeFile(t, filepath.Join(cloneDir, "local.txt"), "local")
	run(t, cloneDir, "git", "add", ".")
	run(t, cloneDir, "git", "commit", "-m", "local")
	run(t, cloneDir, "git", "fetch")

	r := SyncRepo(cloneDir, SyncOptions{NoFetch: true})
	if r.Action != "skipped" {
		t.Errorf("expected skipped, got %s: %s", r.Action, r.PostDetail)
	}
}

func TestSyncAllParallel(t *testing.T) {
	parent := t.TempDir()

	// Create 3 repos, each behind by 1
	var repos []string
	for _, name := range []string{"repo-a", "repo-b", "repo-c"} {
		bareDir := filepath.Join(t.TempDir(), name+".git")
		run(t, "", "git", "init", "--bare", bareDir)

		cloneDir := filepath.Join(parent, name)
		run(t, "", "git", "clone", bareDir, cloneDir)

		writeFile(t, filepath.Join(cloneDir, "README.md"), "init")
		run(t, cloneDir, "git", "add", ".")
		run(t, cloneDir, "git", "commit", "-m", "init")
		run(t, cloneDir, "git", "push")

		pushCommitFromSecondClone(t, bareDir)
		repos = append(repos, cloneDir)
	}

	results, summary := SyncAll(repos, SyncOptions{Jobs: 2, NoFetch: false})

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if summary.Total != 3 {
		t.Errorf("expected total=3, got %d", summary.Total)
	}
	if summary.Pulled != 3 {
		t.Errorf("expected pulled=3, got %d", summary.Pulled)
	}
	for i, r := range results {
		if r.Action != "pulled" {
			t.Errorf("repo %d: expected pulled, got %s: %s", i, r.Action, r.PostDetail)
		}
	}
}

func TestStatusAll(t *testing.T) {
	parent := t.TempDir()
	var repos []string

	for _, name := range []string{"repo-a", "repo-b"} {
		bareDir := filepath.Join(t.TempDir(), name+".git")
		run(t, "", "git", "init", "--bare", bareDir)

		cloneDir := filepath.Join(parent, name)
		run(t, "", "git", "clone", bareDir, cloneDir)

		writeFile(t, filepath.Join(cloneDir, "README.md"), "init")
		run(t, cloneDir, "git", "add", ".")
		run(t, cloneDir, "git", "commit", "-m", "init")
		run(t, cloneDir, "git", "push")
		repos = append(repos, cloneDir)
	}

	statuses := StatusAll(repos, true, 1)
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
	for i, s := range statuses {
		if s.State != StateCurrent {
			t.Errorf("repo %d: expected current, got %s: %s", i, s.State, s.Detail)
		}
	}
}
