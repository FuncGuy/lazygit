package commands

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/commands/oscommands"
)

// RenameCommit renames the topmost commit with the given name
func (c *GitCommand) RenameCommit(name string) error {
	return c.RunCommand("git commit --allow-empty --amend --only -m %s", c.GetOSCommand().Quote(name))
}

// ResetToCommit reset to commit
func (c *GitCommand) ResetToCommit(sha string, strength string, options oscommands.RunCommandOptions) error {
	return c.GetOSCommand().RunCommandWithOptions(fmt.Sprintf("git reset --%s %s", strength, sha), options)
}

func (c *GitCommand) CommitCmdStr(message string, flags string) string {
	splitMessage := strings.Split(message, "\n")
	lineArgs := ""
	for _, line := range splitMessage {
		lineArgs += fmt.Sprintf(" -m %s", c.GetOSCommand().Quote(line))
	}

	flagsStr := ""
	if flags != "" {
		flagsStr = fmt.Sprintf(" %s", flags)
	}

	return fmt.Sprintf("git commit%s%s", flagsStr, lineArgs)
}

// Get the subject of the HEAD commit
func (c *GitCommand) GetHeadCommitMessage() (string, error) {
	cmdStr := "git log -1 --pretty=%s"
	message, err := c.GetOSCommand().RunCommandWithOutput(cmdStr)
	return strings.TrimSpace(message), err
}

func (c *GitCommand) GetCommitMessage(commitSha string) (string, error) {
	cmdStr := "git rev-list --format=%B --max-count=1 " + commitSha
	messageWithHeader, err := c.GetOSCommand().RunCommandWithOutput(cmdStr)
	message := strings.Join(strings.SplitAfter(messageWithHeader, "\n")[1:], "\n")
	return strings.TrimSpace(message), err
}

func (c *GitCommand) GetCommitMessageFirstLine(sha string) (string, error) {
	return c.RunCommandWithOutput("git show --no-patch --pretty=format:%%s %s", sha)
}

// AmendHead amends HEAD with whatever is staged in your working tree
func (c *GitCommand) AmendHead() error {
	return c.GetOSCommand().RunCommand(c.AmendHeadCmdStr())
}

func (c *GitCommand) AmendHeadCmdStr() string {
	return "git commit --amend --no-edit --allow-empty"
}

func (c *GitCommand) ShowCmdStr(sha string, filterPath string) string {
	filterPathArg := ""
	if filterPath != "" {
		filterPathArg = fmt.Sprintf(" -- %s", c.GetOSCommand().Quote(filterPath))
	}
	return fmt.Sprintf("git show --submodule --color=%s --no-renames --stat -p %s %s", c.colorArg(), sha, filterPathArg)
}

// Revert reverts the selected commit by sha
func (c *GitCommand) Revert(sha string) error {
	return c.RunCommand("git revert %s", sha)
}

func (c *GitCommand) RevertMerge(sha string, parentNumber int) error {
	return c.RunCommand("git revert %s -m %d", sha, parentNumber)
}

// CherryPickCommits begins an interactive rebase with the given shas being cherry picked onto HEAD
func (c *GitCommand) CherryPickCommits(commits []*models.Commit) error {
	todo := ""
	for _, commit := range commits {
		todo = "pick " + commit.Sha + " " + commit.Name + "\n" + todo
	}

	cmd, err := c.PrepareInteractiveRebaseCommand("HEAD", todo, false)
	if err != nil {
		return err
	}

	return c.GetOSCommand().RunPreparedCommand(cmd)
}

// CreateFixupCommit creates a commit that fixes up a previous commit
func (c *GitCommand) CreateFixupCommit(sha string) error {
	return c.RunCommand("git commit --fixup=%s", sha)
}
