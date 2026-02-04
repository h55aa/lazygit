package context

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/commands/models"
	"github.com/jesseduffield/lazygit/pkg/gui/filetree"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation"
	"github.com/jesseduffield/lazygit/pkg/gui/presentation/icons"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/samber/lo"
)

type StagedFilesContext struct {
	*filetree.FileTreeViewModel
	*ListContextTrait
	*SearchTrait
}

var (
	_ types.IListContext       = (*StagedFilesContext)(nil)
	_ types.ISearchableContext = (*StagedFilesContext)(nil)
)

func NewStagedFilesContext(c *ContextCommon) *StagedFilesContext {
	viewModel := filetree.NewFileTreeViewModel(
		func() []*models.File { return c.Model().Files },
		c.Common,
		c.UserConfig().Gui.ShowFileTree,
	)

	// Set default filter to Staged (use SetInitialFilter to avoid tree rebuild before model is ready)
	viewModel.SetInitialFilter(filetree.DisplayStaged)

	getDisplayStrings := func(_ int, _ int) [][]string {
		showFileIcons := icons.IsIconEnabled() && c.UserConfig().Gui.ShowFileIcons
		showNumstat := c.UserConfig().Gui.ShowNumstatInFilesView
		lines := presentation.RenderFileTree(viewModel, c.Model().Submodules, showFileIcons, showNumstat, &c.UserConfig().Gui.CustomIcons, c.UserConfig().Gui.ShowRootItemInFileTree)
		return lo.Map(lines, func(line string, _ int) []string {
			return []string{line}
		})
	}

	ctx := &StagedFilesContext{
		SearchTrait:       NewSearchTrait(c),
		FileTreeViewModel: viewModel,
		ListContextTrait: &ListContextTrait{
			Context: NewSimpleContext(NewBaseContext(NewBaseContextOpts{
				View:       c.Views().StagedFiles,
				WindowName: "stagedFiles",
				Key:        STAGED_FILES_CONTEXT_KEY,
				Kind:       types.SIDE_CONTEXT,
				Focusable:  true,
			})),
			ListRenderer: ListRenderer{
				list:              viewModel,
				getDisplayStrings: getDisplayStrings,
			},
			c: c,
		},
	}

	ctx.GetView().SetRenderSearchStatus(ctx.SearchTrait.RenderSearchStatus)
	ctx.GetView().SetOnSelectItem(ctx.OnSearchSelect)

	return ctx
}

func (self *StagedFilesContext) ModelSearchResults(searchStr string, caseSensitive bool) []gocui.SearchPosition {
	return nil
}
