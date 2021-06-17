package commands

import (
	gogit "github.com/jesseduffield/go-git/v5"
	"github.com/jesseduffield/lazygit/pkg/commands/oscommands"
	"github.com/jesseduffield/lazygit/pkg/config"
	"github.com/jesseduffield/lazygit/pkg/i18n"
	"github.com/sirupsen/logrus"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . IGit
type IGit interface {
	ICommander
	IGitConfigMgr

	Branches() IBranchesMgr
	Commits() ICommitsMgr
	Worktree() IWorktreeMgr
	Submodules() ISubmodulesMgr
	Status() IStatusMgr
	Stash() IStashMgr
	Tags() ITagsMgr
	Remotes() IRemotesMgr
	Reflog() IReflogMgr
	Sync() ISyncMgr
	Flow() IFlowMgr
	Rebasing() IRebasingMgr
	Patches() IPatchesMgr
	Diff() IDiffMgr

	GetLog() *logrus.Entry
	// TODO: ensure this works
	WithSpan(span string) IGit
	GetOS() oscommands.IOS
}

// this takes something like:
// * (HEAD detached at 264fc6f5)
//	remotes
// and returns '264fc6f5' as the second match
const CurrentBranchNameRegex = `(?m)^\*.*?([^ ]*?)\)?$`

// Git is our main git interface
type Git struct {
	*Commander
	*GitConfigMgr

	tagsMgr       *TagsMgr
	remotesMgr    *RemotesMgr
	commitsMgr    *CommitsMgr
	reflogMgr     *ReflogMgr
	branchesMgr   *BranchesMgr
	worktreeMgr   *WorktreeMgr
	submodulesMgr *SubmodulesMgr
	statusMgr     *StatusMgr
	stashMgr      *StashMgr
	syncMgr       *SyncMgr
	patchesMgr    *PatchesMgr
	flowMgr       *FlowMgr
	rebasingMgr   *RebasingMgr
	diffMgr       *DiffMgr

	log       *logrus.Entry
	os        oscommands.IOS
	repo      *gogit.Repository
	tr        *i18n.TranslationSet
	config    config.AppConfigurer
	dotGitDir string
}

func (c *Git) GetLog() *logrus.Entry {
	return c.log
}

type MgrCtx struct {
	ICommander

	config IGitConfigMgr
	repo   *gogit.Repository
	log    *logrus.Entry
	os     oscommands.IOS
	tr     *i18n.TranslationSet
}

// NewGit it runs git commands
func NewGit(log *logrus.Entry, oS *oscommands.OS, tr *i18n.TranslationSet, config config.AppConfigurer) (*Git, error) {
	repo, dotGitDir, err := getRepoInfo(tr)
	if err != nil {
		return nil, err
	}

	commander := NewCommander(oS.RunWithOutput, log, oS.GetLazygitPath(), oS.Quote)
	gitConfig := NewGitConfigMgr(
		commander,
		config.GetUserConfig(),
		config.GetUserConfigDir(),
		getGitConfigValue,
		log,
		repo,
		config.GetDebug(),
		dotGitDir,
	)

	mgrCtx := &MgrCtx{
		ICommander: commander,
		config:     gitConfig,
		repo:       repo,
		log:        log,
		os:         oS,
		tr:         tr,
	}

	tagsMgr := NewTagsMgr(mgrCtx)
	remotesMgr := NewRemotesMgr(mgrCtx)
	branchesMgr := NewBranchesMgr(mgrCtx)
	submodulesMgr := NewSubmodulesMgr(mgrCtx)
	statusMgr := NewStatusMgr(mgrCtx)
	worktreeMgr := NewWorktreeMgr(mgrCtx, branchesMgr, submodulesMgr)
	commitsMgr := NewCommitsMgr(mgrCtx, branchesMgr, statusMgr)
	stashMgr := NewStashMgr(mgrCtx, worktreeMgr)
	reflogMgr := NewReflogMgr(mgrCtx)
	syncMgr := NewSyncMgr(mgrCtx)
	flowMgr := NewFlowMgr(mgrCtx)
	rebasingMgr := NewRebasingMgr(mgrCtx, commitsMgr, worktreeMgr, statusMgr)
	diffMgr := NewDiffMgr(mgrCtx)
	patchesMgr := NewPatchesMgr(mgrCtx, commitsMgr, rebasingMgr, statusMgr, stashMgr, diffMgr)

	gitCommand := &Git{
		Commander:     commander,
		GitConfigMgr:  gitConfig,
		tagsMgr:       tagsMgr,
		remotesMgr:    remotesMgr,
		reflogMgr:     reflogMgr,
		commitsMgr:    commitsMgr,
		branchesMgr:   branchesMgr,
		worktreeMgr:   worktreeMgr,
		submodulesMgr: submodulesMgr,
		statusMgr:     statusMgr,
		stashMgr:      stashMgr,
		syncMgr:       syncMgr,
		patchesMgr:    patchesMgr,
		flowMgr:       flowMgr,
		rebasingMgr:   rebasingMgr,
		diffMgr:       diffMgr,
		log:           log,
		os:            oS,
		tr:            tr,
		repo:          repo,
		config:        config,
		dotGitDir:     dotGitDir,
	}

	return gitCommand, nil
}

func (c *Git) Commits() ICommitsMgr {
	return c.commitsMgr
}

func (c *Git) Branches() IBranchesMgr {
	return c.branchesMgr
}

func (c *Git) Worktree() IWorktreeMgr {
	return c.worktreeMgr
}

func (c *Git) Submodules() ISubmodulesMgr {
	return c.submodulesMgr
}

func (c *Git) Status() IStatusMgr {
	return c.statusMgr
}

func (c *Git) Stash() IStashMgr {
	return c.stashMgr
}

func (c *Git) Tags() ITagsMgr {
	return c.tagsMgr
}

func (c *Git) Remotes() IRemotesMgr {
	return c.remotesMgr
}

func (c *Git) Reflog() IReflogMgr {
	return c.reflogMgr
}

func (c *Git) Sync() ISyncMgr {
	return c.syncMgr
}

func (c *Git) Flow() IFlowMgr {
	return c.flowMgr
}

func (c *Git) Rebasing() IRebasingMgr {
	return c.rebasingMgr
}

func (c *Git) Patches() IPatchesMgr {
	return c.patchesMgr
}

func (c *Git) Diff() IDiffMgr {
	return c.diffMgr
}

func (c *Git) WithSpan(span string) IGit {
	// sometimes .WithSpan(span) will be called where span actually is empty, in
	// which case we don't need to log anything so we can just return early here
	// with the original struct
	if span == "" {
		return c
	}

	newGit := &Git{}
	*newGit = *c
	newGit.os = c.GetOS().WithSpan(span)

	return newGit
}

func (c *Git) GetOS() oscommands.IOS {
	return c.os
}
