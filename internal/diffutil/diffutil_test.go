package diffutil

import (
	"strings"
	"testing"
)

func TestParseModifiedGoFile(t *testing.T) {
	diff := `diff --git a/internal/cmd/status.go b/internal/cmd/status.go
index abc1234..def5678 100644
--- a/internal/cmd/status.go
+++ b/internal/cmd/status.go
@@ -10,6 +10,7 @@ import (
 	"fmt"
+	"strings"
 )
@@ -25,7 +26,10 @@ func runStatus(cmd *cobra.Command, args []string) error {
-	target := "."
+	target, err := getTarget(args)
+	if err != nil {
+		return err
+	}
`

	summary, err := Parse(strings.NewReader(diff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summary.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(summary.Files))
	}

	f := summary.Files[0]
	if f.ChangeType != "modified" {
		t.Errorf("expected change type 'modified', got %q", f.ChangeType)
	}
	if f.Added != 5 {
		t.Errorf("expected 5 added lines, got %d", f.Added)
	}
	if f.Removed != 1 {
		t.Errorf("expected 1 removed line, got %d", f.Removed)
	}
	if !f.Imports {
		t.Error("expected imports to be true")
	}
	if len(f.Functions) == 0 || f.Functions[0] != "runStatus" {
		t.Errorf("expected function 'runStatus', got %v", f.Functions)
	}
	if summary.TotalAdded != 5 {
		t.Errorf("expected total added 5, got %d", summary.TotalAdded)
	}
}

func TestParseNewFile(t *testing.T) {
	diff := `diff --git a/internal/newpkg/new.go b/internal/newpkg/new.go
new file mode 100644
index 0000000..abc1234
--- /dev/null
+++ b/internal/newpkg/new.go
@@ -0,0 +1,10 @@
+package newpkg
+
+type Config struct {
+	Name string
+}
+
+func New() *Config {
+	return &Config{}
+}
`

	summary, err := Parse(strings.NewReader(diff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f := summary.Files[0]
	if f.ChangeType != "added" {
		t.Errorf("expected change type 'added', got %q", f.ChangeType)
	}
	if f.Added != 9 {
		t.Errorf("expected 9 added lines, got %d", f.Added)
	}
	if len(f.Types) == 0 || f.Types[0] != "Config" {
		t.Errorf("expected type 'Config', got %v", f.Types)
	}
	if len(f.Functions) == 0 || f.Functions[0] != "New" {
		t.Errorf("expected function 'New', got %v", f.Functions)
	}
}

func TestParseDeletedFile(t *testing.T) {
	diff := `diff --git a/old.go b/old.go
deleted file mode 100644
index abc1234..0000000
--- a/old.go
+++ /dev/null
@@ -1,5 +0,0 @@
-package old
-
-func Deprecated() {
-	// nothing
-}
`

	summary, err := Parse(strings.NewReader(diff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f := summary.Files[0]
	if f.ChangeType != "deleted" {
		t.Errorf("expected change type 'deleted', got %q", f.ChangeType)
	}
	if f.Removed != 5 {
		t.Errorf("expected 5 removed lines, got %d", f.Removed)
	}
}

func TestParseRenamedFile(t *testing.T) {
	diff := `diff --git a/old_name.go b/new_name.go
similarity index 95%
rename from old_name.go
rename to new_name.go
index abc1234..def5678 100644
--- a/old_name.go
+++ b/new_name.go
@@ -1,3 +1,3 @@ package pkg
-func OldFunc() {}
+func NewFunc() {}
`

	summary, err := Parse(strings.NewReader(diff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f := summary.Files[0]
	if f.ChangeType != "renamed" {
		t.Errorf("expected change type 'renamed', got %q", f.ChangeType)
	}
	if f.OldName != "old_name.go" || f.NewName != "new_name.go" {
		t.Errorf("expected old_name.go -> new_name.go, got %s -> %s", f.OldName, f.NewName)
	}
}

func TestParseBinaryFile(t *testing.T) {
	diff := `diff --git a/image.png b/image.png
index abc1234..def5678 100644
Binary files a/image.png and b/image.png differ
`

	summary, err := Parse(strings.NewReader(diff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f := summary.Files[0]
	if f.ChangeType != "binary" {
		t.Errorf("expected change type 'binary', got %q", f.ChangeType)
	}
	if !f.Binary {
		t.Error("expected binary to be true")
	}
}

func TestParseMultipleFiles(t *testing.T) {
	diff := `diff --git a/a.go b/a.go
index abc..def 100644
--- a/a.go
+++ b/a.go
@@ -1,3 +1,4 @@ func Foo() {
+	// comment
diff --git a/b.txt b/b.txt
index abc..def 100644
--- a/b.txt
+++ b/b.txt
@@ -1,2 +1,3 @@
+new line
-old line
`

	summary, err := Parse(strings.NewReader(diff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(summary.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(summary.Files))
	}
	if summary.TotalAdded != 2 {
		t.Errorf("expected total added 2, got %d", summary.TotalAdded)
	}
	if summary.TotalRemoved != 1 {
		t.Errorf("expected total removed 1, got %d", summary.TotalRemoved)
	}
}

func TestParseNonGoFile(t *testing.T) {
	diff := `diff --git a/README.md b/README.md
index abc..def 100644
--- a/README.md
+++ b/README.md
@@ -1,3 +1,5 @@
 # Title
+
+New paragraph.
`

	summary, err := Parse(strings.NewReader(diff))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	f := summary.Files[0]
	if f.ChangeType != "modified" {
		t.Errorf("expected 'modified', got %q", f.ChangeType)
	}
	if len(f.Functions) != 0 {
		t.Errorf("expected no functions for non-Go file, got %v", f.Functions)
	}
}

func TestParseEmptyDiff(t *testing.T) {
	summary, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(summary.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(summary.Files))
	}
}

func TestDedup(t *testing.T) {
	result := dedup([]string{"a", "b", "a", "c", "b"})
	if len(result) != 3 {
		t.Errorf("expected 3 unique items, got %d", len(result))
	}
}
