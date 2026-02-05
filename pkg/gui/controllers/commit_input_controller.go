package controllers

import (
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type CommitInputController struct {
	baseController
	c *ControllerCommon
}

var _ types.IController = &CommitInputController{}

func NewCommitInputController(c *ControllerCommon) *CommitInputController {
	return &CommitInputController{
		baseController: baseController{},
		c:              c,
	}
}

func (self *CommitInputController) Context() types.Context {
	return self.c.Contexts().CommitInput
}

func (self *CommitInputController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		{
			Key:         opts.GetKey(opts.Config.Universal.SubmitEditorText),
			Handler:     self.confirm,
			Description: self.c.Tr.Actions.Commit,
		},
	}
}

func (self *CommitInputController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: self.Context().GetViewName(),
			Key:      gocui.MouseLeft,
			Handler:  self.onClick,
		},
	}
}

func (self *CommitInputController) onClick(gocui.ViewMouseBindingOpts) error {
	self.c.Context().Push(self.c.Contexts().CommitInput, types.OnFocusOpts{})
	return nil
}

func (self *CommitInputController) confirm() error {
	// The default keybinding for this action is "<enter>", which means that we
	// also get here when pasting multi-line text that contains newlines. In
	// that case we don't want to commit; instead we just insert a space so that
	// the pasted message stays roughly readable on one line.
	if self.c.GocuiGui().IsPasting && self.c.UserConfig().Keybinding.Universal.SubmitEditorText == "<enter>" {
		view := self.c.Views().CommitInput
		view.Editor.Edit(view, gocui.KeySpace, ' ', 0)
		return nil
	}

	message := strings.TrimSpace(self.c.Views().CommitInput.TextArea.GetContent())
	return self.c.Helpers().WorkingTree.CommitStagedWithMessage(message, false, func() error {
		view := self.c.Views().CommitInput
		view.ClearTextArea()
		view.RenderTextArea()
		return nil
	})
}

