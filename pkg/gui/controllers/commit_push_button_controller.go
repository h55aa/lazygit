package controllers

import (
	"fmt"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type CommitPushButtonController struct {
	baseController
	c          *ControllerCommon
	handlePush func() error
}

const (
	commitPushButtonDefaultLabel = "[ Push ]"
	commitPushButtonFocusedLabel = "[*Push*]"
)

var _ types.IController = &CommitPushButtonController{}

func NewCommitPushButtonController(c *ControllerCommon, handlePush func() error) *CommitPushButtonController {
	return &CommitPushButtonController{
		baseController: baseController{},
		c:              c,
		handlePush:     handlePush,
	}
}

func (self *CommitPushButtonController) Context() types.Context {
	return self.c.Contexts().CommitPushButton
}

func (self *CommitPushButtonController) GetKeybindings(opts types.KeybindingsOpts) []*types.Binding {
	return []*types.Binding{
		{
			Key:         opts.GetKey(opts.Config.Universal.SubmitEditorText),
			Handler:     self.submit,
			Description: self.c.Tr.Actions.Push,
		},
		{
			Key:     opts.GetKey(opts.Config.Universal.PrevBlock),
			Handler: self.focusCommitButton,
		},
		{
			Key:     opts.GetKey(opts.Config.Universal.NextBlock),
			Handler: self.focusStagedFiles,
		},
		{
			Key:     opts.GetKey(opts.Config.Universal.PrevBlockAlt),
			Handler: self.focusCommitButton,
		},
		{
			Key:     opts.GetKey(opts.Config.Universal.NextBlockAlt),
			Handler: self.focusStagedFiles,
		},
		{
			Key:     opts.GetKey(opts.Config.Universal.PrevBlockAlt2),
			Handler: self.focusCommitButton,
		},
		{
			Key:     opts.GetKey(opts.Config.Universal.NextBlockAlt2),
			Handler: self.focusStagedFiles,
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

func (self *CommitPushButtonController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: self.Context().GetViewName(),
			Key:      gocui.MouseLeft,
			Handler:  self.onClick,
		},
	}
}

func (self *CommitPushButtonController) submit() error {
	return self.onClick(gocui.ViewMouseBindingOpts{})
}

func (self *CommitPushButtonController) GetOnFocus() func(types.OnFocusOpts) {
	return func(types.OnFocusOpts) {
		self.c.SetViewContent(self.c.Views().CommitPushButton, commitPushButtonFocusedLabel)
	}
}

func (self *CommitPushButtonController) GetOnFocusLost() func(types.OnFocusLostOpts) {
	return func(types.OnFocusLostOpts) {
		self.c.SetViewContent(self.c.Views().CommitPushButton, commitPushButtonDefaultLabel)
	}
}

func (self *CommitPushButtonController) focusCommitButton() error {
	self.c.Context().Push(self.c.Contexts().CommitButton, types.OnFocusOpts{})
	return nil
}

func (self *CommitPushButtonController) focusStagedFiles() error {
	self.c.Context().Push(self.c.Contexts().StagedFiles, types.OnFocusOpts{})
	return nil
}

func (self *CommitPushButtonController) onClick(gocui.ViewMouseBindingOpts) error {
	if self.handlePush == nil {
		return fmt.Errorf("push action unavailable")
	}

	return self.handlePush()
}
