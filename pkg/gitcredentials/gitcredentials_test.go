package gitcredentials

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUser_WorkingDirResolvesIncludeIf(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	projectDir := t.TempDir()
	require.NoError(t, exec.Command("git", "init", projectDir).Run())

	projectConfigPath := filepath.Join(tmpHome, "project.gitconfig")
	require.NoError(t, os.WriteFile(projectConfigPath, []byte(`[user]
	name = Project User
	email = project@example.com
`), 0o644))

	globalConfigPath := filepath.Join(tmpHome, ".gitconfig")
	require.NoError(t, os.WriteFile(globalConfigPath, []byte(fmt.Sprintf(`[user]
	name = Global User
	email = global@example.com
[includeIf "gitdir:%s/"]
	path = %s
`, projectDir, projectConfigPath)), 0o644))

	user, err := GetUser("", projectDir)
	require.NoError(t, err)
	assert.Equal(t, "Project User", user.Name)
	assert.Equal(t, "project@example.com", user.Email)
}

func TestGetUser_EmptyWorkingDirUsesGlobal(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	globalConfigPath := filepath.Join(tmpHome, ".gitconfig")
	require.NoError(t, os.WriteFile(globalConfigPath, []byte(`[user]
	name = Global User
	email = global@example.com
`), 0o644))

	user, err := GetUser("", "")
	require.NoError(t, err)
	assert.Equal(t, "Global User", user.Name)
	assert.Equal(t, "global@example.com", user.Email)
}

func TestGetUser_WorkingDirWithNoMatchingIncludeIfUsesGlobal(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")

	otherDir := t.TempDir()

	projectConfigPath := filepath.Join(tmpHome, "project.gitconfig")
	require.NoError(t, os.WriteFile(projectConfigPath, []byte(`[user]
	name = Project User
	email = project@example.com
`), 0o644))

	globalConfigPath := filepath.Join(tmpHome, ".gitconfig")
	require.NoError(t, os.WriteFile(globalConfigPath, []byte(fmt.Sprintf(`[user]
	name = Global User
	email = global@example.com
[includeIf "gitdir:%s/"]
	path = %s
`, otherDir, projectConfigPath)), 0o644))

	// workingDir does not match the includeIf pattern — global values should come through
	projectDir := t.TempDir()
	require.NoError(t, exec.Command("git", "init", projectDir).Run())

	user, err := GetUser("", projectDir)
	require.NoError(t, err)
	assert.Equal(t, "Global User", user.Name)
	assert.Equal(t, "global@example.com", user.Email)
}
