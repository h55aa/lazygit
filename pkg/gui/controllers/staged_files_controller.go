package controllers

import (
	"errors"
	"unicode/utf8"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/filetree"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/jesseduffield/lazygit/pkg/utils"
	"github.com/samber/lo"
)

type StagedFilesController struct {
	baseController
	*ListControllerTrait[*filetree.FileNode]
	c *ControllerCommon
}

var _ types.IController = &StagedFilesController{}

func NewStagedFilesController(
	c *ControllerCommon,
) *StagedFilesController {
	return &StagedFilesController{
		baseController: baseController{},
		ListControllerTrait: NewListControllerTrait(
			c,
			c.Contexts().StagedFiles,
			c.Contexts().StagedFiles.GetSelected,
			c.Contexts().StagedFiles.GetSelectedItems,
		),
		c: c,
	}
}

func (self *StagedFilesController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		{
			Key:               opts.GetKey(opts.Config.Universal.Select),
			Handler:           self.handleSelect,
			GetDisabledReason: self.require(self.itemsSelected()),
			Description:       self.c.Tr.Stage,
			Tooltip:           self.c.Tr.StageTooltip,
			DisplayOnScreen:   true,
		},
		{
			Key:               opts.GetKey(opts.Config.Universal.Remove),
			Handler:           self.withItems(self.press),
			GetDisabledReason: self.require(self.itemsSelected()),
			Description:       self.c.Tr.Stage,
			Tooltip:           self.c.Tr.StageTooltip,
		},
		{
			Key:         opts.GetKey(opts.Config.Files.ToggleTreeView),
			Handler:     self.toggleTreeView,
			Description: self.c.Tr.ToggleTreeView,
			Tooltip:     self.c.Tr.ToggleTreeViewTooltip,
		},
		{
			Key:         opts.GetKey(opts.Config.Files.CollapseAll),
			Handler:     self.collapseAll,
			Description: self.c.Tr.CollapseAll,
		},
		{
			Key:         opts.GetKey(opts.Config.Files.ExpandAll),
			Handler:     self.expandAll,
			Description: self.c.Tr.ExpandAll,
		},
		{
			Key:               opts.GetKey(opts.Config.Universal.GoInto),
			Handler:           self.enter,
			GetDisabledReason: self.require(self.singleItemSelected()),
			Description:       self.c.Tr.FileEnter,
			Tooltip:           self.c.Tr.FileEnterTooltip,
			DisplayOnScreen:   true,
		},
	}
}

func (self *StagedFilesController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName:    self.context().GetViewName(),
			FocusedView: self.context().GetViewName(),
			Key:         gocui.MouseLeft,
			Handler:     self.onClickActionButton,
		},
	}
}

func (self *StagedFilesController) GetOnRenderToMain() func() {
	return func() {
		self.c.Helpers().Diff.WithDiffModeCheck(func() {
			// Ensure we don't leave the merge conflicts view "stuck" from a previous selection.
			self.c.Helpers().MergeConflicts.ResetMergeState()

			node := self.context().GetSelected()

			if node == nil {
				self.c.RenderToMainViews(types.RefreshMainOpts{
					Pair: self.c.MainViewPairs().Normal,
					Main: &types.ViewUpdateOpts{
						Title:    self.c.Tr.StagedChanges,
						SubTitle: self.c.Helpers().Diff.IgnoringWhitespaceSubTitle(),
						Task:     types.NewRenderStringTask(self.c.Tr.NoChangedFiles),
					},
				})
				return
			}

			split := self.c.UserConfig().Gui.SplitDiff == "always" || (node.GetHasUnstagedChanges() && node.GetHasStagedChanges())

			cmdObj := self.c.Git().WorkingTree.WorktreeFileDiffCmdObj(node, false, true)
			refreshOpts := types.RefreshMainOpts{
				Pair: self.c.MainViewPairs().Normal,
				Main: &types.ViewUpdateOpts{
					Task:     types.NewRunPtyTask(cmdObj.GetCmd()),
					SubTitle: self.c.Helpers().Diff.IgnoringWhitespaceSubTitle(),
					Title:    self.c.Tr.StagedChanges,
				},
			}

			if split {
				cmdObj := self.c.Git().WorkingTree.WorktreeFileDiffCmdObj(node, false, false)
				refreshOpts.Secondary = &types.ViewUpdateOpts{
					Title:    self.c.Tr.UnstagedChanges,
					SubTitle: self.c.Helpers().Diff.IgnoringWhitespaceSubTitle(),
					Task:     types.NewRunPtyTask(cmdObj.GetCmd()),
				}
			}

			self.c.RenderToMainViews(refreshOpts)
		})
	}
}

func (self *StagedFilesController) GetOnClick() func() error {
	return self.withItemGraceful(self.pressSingle)
}

func (self *StagedFilesController) handleSelect() error {
	selectedNodes, _, _ := self.context().GetSelectedItems()
	if len(selectedNodes) == 0 {
		return errors.New(self.c.Tr.NoItemSelected)
	}

	if len(selectedNodes) == 1 && selectedNodes[0] != nil && selectedNodes[0].File == nil {
		return self.handleToggleDirCollapsed()
	}

	return self.press(selectedNodes)
}

// For staged files, pressing will always unstage them
func (self *StagedFilesController) press(selectedNodes []*filetree.FileNode) error {
	if err := self.pressWithLock(selectedNodes); err != nil {
		return err
	}

	self.c.Refresh(types.RefreshOptions{Scope: []types.RefreshableView{types.FILES}, Mode: types.ASYNC})

	self.context().HandleFocus(types.OnFocusOpts{})
	return nil
}

func (self *StagedFilesController) pressWithLock(selectedNodes []*filetree.FileNode) error {
	self.c.Mutexes().RefreshingFilesMutex.Lock()
	defer self.c.Mutexes().RefreshingFilesMutex.Unlock()

	toPaths := func(nodes []*filetree.FileNode) []string {
		return lo.Map(nodes, func(node *filetree.FileNode, _ int) string {
			return node.GetPath()
		})
	}

	selectedNodes = normalisedSelectedNodes(selectedNodes)

	self.c.LogAction(self.c.Tr.Actions.UnstageFile)

	// Partition into tracked and untracked
	trackedNodes, untrackedNodes := utils.Partition(selectedNodes, func(node *filetree.FileNode) bool {
		return !node.IsFile() || node.GetIsTracked()
	})

	if len(untrackedNodes) > 0 {
		if err := self.c.Git().WorkingTree.UnstageUntrackedFiles(toPaths(untrackedNodes)); err != nil {
			return err
		}
	}

	if len(trackedNodes) > 0 {
		if err := self.c.Git().WorkingTree.UnstageTrackedFiles(toPaths(trackedNodes)); err != nil {
			return err
		}
	}

	return nil
}

func (self *StagedFilesController) pressSingle(node *filetree.FileNode) error {
	return self.press([]*filetree.FileNode{node})
}

func (self *StagedFilesController) enter() error {
	node := self.context().GetSelected()
	if node == nil || node.File != nil {
		return nil
	}

	return self.handleToggleDirCollapsed()
}

func (self *StagedFilesController) Context() types.Context {
	return self.c.Contexts().StagedFiles
}

func (self *StagedFilesController) context() *context.StagedFilesContext {
	return self.c.Contexts().StagedFiles
}

func (self *StagedFilesController) handleToggleDirCollapsed() error {
	node := self.context().GetSelected()
	if node == nil || node.File != nil {
		return nil
	}

	self.context().FileTreeViewModel.ToggleCollapsed(node.GetInternalPath())
	self.c.PostRefreshUpdate(self.context())
	return nil
}

func (self *StagedFilesController) toggleTreeView() error {
	self.context().FileTreeViewModel.ToggleShowTree()

	self.c.PostRefreshUpdate(self.context())

	return nil
}

func (self *StagedFilesController) collapseAll() error {
	self.context().FileTreeViewModel.CollapseAll()

	self.c.PostRefreshUpdate(self.context())

	return nil
}

func (self *StagedFilesController) expandAll() error {
	self.context().FileTreeViewModel.ExpandAll()

	self.c.PostRefreshUpdate(self.context())

	return nil
}

func (self *StagedFilesController) onClickActionButton(opts gocui.ViewMouseBindingOpts) error {
	modelLineIdx := self.context().ViewIndexToModelIndex(opts.Y)
	if modelLineIdx < 0 || modelLineIdx > self.context().GetList().Len()-1 {
		return gocui.ErrKeybindingNotHandled
	}

	node := self.context().Get(modelLineIdx)
	if node == nil {
		return gocui.ErrKeybindingNotHandled
	}

	filter := self.context().GetFilter()
	buttonToken := presentation.FileActionButtonToken(filter, node.GetHasUnstagedChanges(), node.GetHasStagedChanges())
	if buttonToken == "" {
		return gocui.ErrKeybindingNotHandled
	}

	line, ok := self.context().GetView().Line(opts.Y)
	if !ok {
		return gocui.ErrKeybindingNotHandled
	}

	startCol := lastIndexByRune(line, buttonToken)
	if startCol < 0 {
		return gocui.ErrKeybindingNotHandled
	}
	endCol := startCol + utf8.RuneCountInString(buttonToken)
	if opts.X < startCol || opts.X >= endCol {
		return gocui.ErrKeybindingNotHandled
	}

	self.context().GetList().SetSelection(modelLineIdx)
	self.context().HandleFocus(types.OnFocusOpts{})
	return self.press([]*filetree.FileNode{node})
}
