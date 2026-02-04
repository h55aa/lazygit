package controllers

import (
	"github.com/jesseduffield/lazygit/pkg/gui/context"
	"github.com/jesseduffield/lazygit/pkg/gui/filetree"
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
			Handler:           self.withItems(self.press),
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
	}
}

func (self *StagedFilesController) GetOnClick() func() error {
	return self.withItemGraceful(self.pressSingle)
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

func (self *StagedFilesController) Context() types.Context {
	return self.c.Contexts().StagedFiles
}

func (self *StagedFilesController) context() *context.StagedFilesContext {
	return self.c.Contexts().StagedFiles
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
