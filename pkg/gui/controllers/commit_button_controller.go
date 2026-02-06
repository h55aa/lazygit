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

const (
	commitButtonDefaultLabel = "[ Commit ]"
	commitButtonFocusedLabel = "[*Commit*]"
)

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

func (self *CommitButtonController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		{
			Key:         opts.GetKey(opts.Config.Universal.SubmitEditorText),
			Handler:     self.submit,
			Description: self.c.Tr.Actions.Commit,
		},
		{
			Key:     opts.GetKey(opts.Config.Universal.PrevBlock),
			Handler: self.focusGenerateButton,
		},
		{
			Key:     opts.GetKey(opts.Config.Universal.NextBlock),
			Handler: self.focusPushButton,
		},
		{
			Key:     opts.GetKey(opts.Config.Universal.PrevBlockAlt),
			Handler: self.focusGenerateButton,
		},
		{
			Key:     opts.GetKey(opts.Config.Universal.NextBlockAlt),
			Handler: self.focusPushButton,
		},
		{
			Key:     opts.GetKey(opts.Config.Universal.PrevBlockAlt2),
			Handler: self.focusGenerateButton,
		},
		{
			Key:     opts.GetKey(opts.Config.Universal.NextBlockAlt2),
			Handler: self.focusPushButton,
		},
		{
			Key:     opts.GetKey(opts.Config.Universal.NextItem),
			Handler: self.focusStagedFiles,
		},
		{
			Key:     opts.GetKey(opts.Config.Universal.NextItemAlt),
			Handler: self.focusStagedFiles,
		},
	}
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

func (self *CommitButtonController) submit() error {
	return self.onClick(gocui.ViewMouseBindingOpts{})
}

func (self *CommitButtonController) GetOnFocus() func(types.OnFocusOpts) {
	return func(types.OnFocusOpts) {
		self.c.SetViewContent(self.c.Views().CommitButton, commitButtonFocusedLabel)
	}
}

func (self *CommitButtonController) GetOnFocusLost() func(types.OnFocusLostOpts) {
	return func(types.OnFocusLostOpts) {
		self.c.SetViewContent(self.c.Views().CommitButton, commitButtonDefaultLabel)
	}
}

func (self *CommitButtonController) focusGenerateButton() error {
	self.c.Context().Push(self.c.Contexts().CommitGenerateButton, types.OnFocusOpts{})
	return nil
}

func (self *CommitButtonController) focusPushButton() error {
	self.c.Context().Push(self.c.Contexts().CommitPushButton, types.OnFocusOpts{})
	return nil
}

func (self *CommitButtonController) focusStagedFiles() error {
	self.c.Context().Push(self.c.Contexts().StagedFiles, types.OnFocusOpts{})
	return nil
}

func (self *CommitButtonController) onClick(gocui.ViewMouseBindingOpts) error {
	// Return focus to the input so users can continue editing after commit.
	self.c.Context().Push(self.c.Contexts().CommitInput, types.OnFocusOpts{})

	message := strings.TrimSpace(self.c.Views().CommitInput.TextArea.GetContent())
	return self.c.Helpers().WorkingTree.CommitStagedWithMessage(message, false, func() error {
		view := self.c.Views().CommitInput
		view.ClearTextArea()
		view.RenderTextArea()
		return nil
	})
}
