package controllers

import (
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type CommitButtonController struct {
	baseController
	c *ControllerCommon
}

var _ types.IController = &CommitButtonController{}

func NewCommitButtonController(c *ControllerCommon) *CommitButtonController {
	return &CommitButtonController{
		baseController: baseController{},
		c:              c,
	}
}

func (self *CommitButtonController) Context() types.Context {
	return self.c.Contexts().CommitButton
}

func (self *CommitButtonController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: self.Context().GetViewName(),
			Key:      gocui.MouseLeft,
			Handler:  self.onClick,
		},
	}
}

func (self *CommitButtonController) onClick(gocui.ViewMouseBindingOpts) error {
	// Avoid leaving focus on a non-focusable view.
	self.c.Context().Push(self.c.Contexts().CommitInput, types.OnFocusOpts{})

	message := strings.TrimSpace(self.c.Views().CommitInput.TextArea.GetContent())
	return self.c.Helpers().WorkingTree.CommitStagedWithMessage(message, false, func() error {
		view := self.c.Views().CommitInput
		view.ClearTextArea()
		view.RenderTextArea()
		return nil
	})
}

