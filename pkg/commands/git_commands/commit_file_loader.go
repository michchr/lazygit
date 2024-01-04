package git_commands

import (
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/commands/oscommands"
	"github.com/samber/lo"
)

type CommitFileLoader struct {
	*GitCommon
	cmd oscommands.ICmdObjBuilder
}

func NewCommitFileLoader(gitCommon *GitCommon, cmd oscommands.ICmdObjBuilder) *CommitFileLoader {
	return &CommitFileLoader{
		GitCommon: gitCommon,
		cmd:       cmd,
	}
}

// GetFilesInDiff get the specified commit files
func (self *CommitFileLoader) GetFilesInDiff(from string, to string, reverse bool) ([]*models.CommitFile, error) {
	cmdArgs := NewGitCmd("diff").
		Arg("--submodule").
		Arg("--no-ext-diff").
		Arg("--name-status").
		Arg("-z").
		Arg("--no-renames").
		ArgIf(reverse, "-R").
		Arg(from).
		Arg(to).
		ToArgv()

	filenames, err := self.cmd.New(cmdArgs).DontLog().RunWithOutput()
	if err != nil {
		return nil, err
	}

	return self.GetCommitFilesFromFilenames(filenames), nil
}

// filenames string is something like "MM\x00file1\x00MU\x00file2\x00AA\x00file3\x00"
// so we need to split it by the null character and then map each status-name pair to a commit file
func (self *CommitFileLoader) GetCommitFilesFromFilenames(filenames string) []*models.CommitFile {
	lines := strings.Split(strings.TrimRight(filenames, "\x00"), "\x00")
	if len(lines) == 1 {
		return []*models.CommitFile{}
	}

	worktreePath := self.GitCommon.repoPaths.WorktreePath() + "/"

	// typical result looks like 'A my_file' meaning my_file was added
	return lo.Map(lo.Chunk(lines, 2), func(chunk []string, _ int) *models.CommitFile {
		return &models.CommitFile{
			ChangeStatus: chunk[0],
			Name:         worktreePath + chunk[1],
		}
	})
}
