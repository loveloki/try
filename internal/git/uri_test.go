package git

import (
	"os"
	"path/filepath"
	"testing"
)

func checkParse(t *testing.T, uri, wantUser, wantRepo, wantHost string) {
	t.Helper()
	got := ParseGitURI(uri)
	if wantUser == "" && wantRepo == "" && wantHost == "" {
		if got != nil {
			t.Errorf("ParseGitURI(%q) = %+v, want nil", uri, got)
		}
		return
	}
	if got == nil {
		t.Fatalf("ParseGitURI(%q) = nil, want {User:%q Repo:%q Host:%q}", uri, wantUser, wantRepo, wantHost)
	}
	if got.User != wantUser || got.Repo != wantRepo || got.Host != wantHost {
		t.Errorf("ParseGitURI(%q) = {User:%q Repo:%q Host:%q}, want {User:%q Repo:%q Host:%q}",
			uri, got.User, got.Repo, got.Host, wantUser, wantRepo, wantHost)
	}
}

func TestParseGitURI(t *testing.T) {
	tests := []struct {
		name                         string
		uri                          string
		wantUser, wantRepo, wantHost string
	}{
		{"https github", "https://github.com/tobi/try", "tobi", "try", "github.com"},
		{"https github .git", "https://github.com/tobi/try.git", "tobi", "try", "github.com"},
		{"https gitlab", "https://gitlab.com/team/project", "team", "project", "gitlab.com"},
		{"https self-hosted", "https://git.company.com/org/repo", "org", "repo", "git.company.com"},
		{"http", "http://github.com/user/repo", "user", "repo", "github.com"},
		{"ssh github", "git@github.com:tobi/try", "tobi", "try", "github.com"},
		{"ssh github .git", "git@github.com:tobi/try.git", "tobi", "try", "github.com"},
		{"ssh gitlab", "git@gitlab.com:team/project", "team", "project", "gitlab.com"},
		{"ssh self-hosted", "git@git.company.com:org/repo", "org", "repo", "git.company.com"},
		{"invalid url", "not-a-url", "", "", ""},
		{"empty", "", "", "", ""},
		{"bare path", "/tmp/repo", "", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkParse(t, tt.uri, tt.wantUser, tt.wantRepo, tt.wantHost)
		})
	}
}

func checkIsGitURI(t *testing.T, arg string, want bool) {
	t.Helper()
	got := IsGitURI(arg)
	if got != want {
		t.Errorf("IsGitURI(%q) = %v, want %v", arg, got, want)
	}
}

func TestIsGitURI(t *testing.T) {
	tests := []struct {
		arg  string
		want bool
	}{
		{"https://github.com/user/repo", true},
		{"http://github.com/user/repo", true},
		{"git@github.com:user/repo", true},
		{"something.github.com/path", true},
		{"gitlab.com/team/project", true},
		{"repo.git", true},
		{"plain-text", false},
		{"", false},
		{"/tmp/local-dir", false},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			checkIsGitURI(t, tt.arg, tt.want)
		})
	}
}

func checkDirName(t *testing.T, uri, customName, dateSuffix, want string) {
	t.Helper()
	got := generateCloneDirNameWithDate(uri, customName, dateSuffix)
	if got != want {
		t.Errorf("generateCloneDirNameWithDate(%q, %q, %q) = %q, want %q", uri, customName, dateSuffix, got, want)
	}
}

func TestGenerateCloneDirName(t *testing.T) {
	date := "2025-08-17"
	tests := []struct {
		name       string
		uri        string
		customName string
		want       string
	}{
		{"auto naming", "https://github.com/tobi/try.git", "", "tobi-try-" + date},
		{"custom name", "https://github.com/tobi/try.git", "my-fork", "my-fork"},
		{"ssh auto naming", "git@github.com:user/repo", "", "user-repo-" + date},
		{"invalid uri no custom", "not-a-url", "", ""},
		{"invalid uri with custom", "not-a-url", "name", "name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkDirName(t, tt.uri, tt.customName, date, tt.want)
		})
	}
}

func checkUnique(t *testing.T, existing []string, base, suffix, want string) {
	t.Helper()
	tmpDir := t.TempDir()
	for _, name := range existing {
		if err := os.MkdirAll(filepath.Join(tmpDir, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	got := ResolveUniqueName(tmpDir, base, suffix)
	if got != want {
		t.Errorf("ResolveUniqueName(base=%q, suffix=%q, existing=%v) = %q, want %q", base, suffix, existing, got, want)
	}
}

func TestResolveUniqueName(t *testing.T) {
	tests := []struct {
		name     string
		existing []string
		base     string
		suffix   string
		want     string
	}{
		{"no conflict", nil, "foo", "2025-08-17", "foo"},
		{"one conflict appends -2", []string{"foo-2025-08-17"}, "foo", "2025-08-17", "foo-2"},
		{"two conflicts appends -3", []string{"foo-2025-08-17", "foo-2-2025-08-17"}, "foo", "2025-08-17", "foo-3"},
		{"trailing number increments", []string{"v2-2025-08-17"}, "v2", "2025-08-17", "v3"},
		{"trailing number chain", []string{"v2-2025-08-17", "v3-2025-08-17"}, "v2", "2025-08-17", "v4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkUnique(t, tt.existing, tt.base, tt.suffix, tt.want)
		})
	}
}
