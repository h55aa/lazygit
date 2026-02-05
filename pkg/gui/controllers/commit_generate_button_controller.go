package controllers

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
)

type CommitGenerateButtonController struct {
	baseController
	c *ControllerCommon
}

var _ types.IController = &CommitGenerateButtonController{}

func NewCommitGenerateButtonController(c *ControllerCommon) *CommitGenerateButtonController {
	return &CommitGenerateButtonController{
		baseController: baseController{},
		c:              c,
	}
}

func (self *CommitGenerateButtonController) Context() types.Context {
	return self.c.Contexts().CommitGenerateButton
}

func (self *CommitGenerateButtonController) GetMouseKeybindings(opts types.KeybindingsOpts) []*gocui.ViewMouseBinding {
	return []*gocui.ViewMouseBinding{
		{
			ViewName: self.Context().GetViewName(),
			Key:      gocui.MouseLeft,
			Handler:  self.onClick,
		},
	}
}

func (self *CommitGenerateButtonController) onClick(gocui.ViewMouseBindingOpts) error {
	// Focus the input immediately so the user sees where the message will land.
	self.c.Context().Push(self.c.Contexts().CommitInput, types.OnFocusOpts{})

	self.c.WithWaitingStatus("AI: generating commit message", func(gocui.Task) error {
		message, err := self.generateCommitMessage()
		if err != nil {
			self.c.OnUIThread(func() error {
				self.c.ErrorToast(fmt.Sprintf("AI: %v", err))
				return nil
			})
			return nil
		}

		message = sanitizeCommitMessage(message)
		if message == "" {
			self.c.OnUIThread(func() error {
				self.c.ErrorToast("AI: empty commit message")
				return nil
			})
			return nil
		}

		self.c.OnUIThread(func() error {
			view := self.c.Views().CommitInput
			view.ClearTextArea()
			view.TextArea.TypeString(message)
			view.RenderTextArea()

			self.c.Context().Push(self.c.Contexts().CommitInput, types.OnFocusOpts{})
			return nil
		})

		return nil
	})

	return nil
}

func (self *CommitGenerateButtonController) generateCommitMessage() (string, error) {
	repoPath := self.c.Git().RepoPaths.RepoPath()
	zeemuxPath, err := resolveBundledZeeMuxPath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command(zeemuxPath, "llm", "commit-msg", "--repo", repoPath)
	cmd.Dir = repoPath

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return "", fmt.Errorf("zeemux llm commit-msg failed: %s", msg)
		}
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

func sanitizeCommitMessage(message string) string {
	clean := strings.ReplaceAll(message, "\r", " ")
	clean = strings.ReplaceAll(clean, "\n", " ")
	clean = strings.TrimSpace(clean)
	if len(clean) > 200 {
		clean = clean[:200]
	}
	return clean
}

func resolveBundledZeeMuxPath() (string, error) {
	if exePath, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exePath), "zeemux")
		if isExecutableFile(candidate) {
			return candidate, nil
		}
	}

	if pathLookup, err := exec.LookPath("zeemux"); err == nil {
		return pathLookup, nil
	}

	return "", fmt.Errorf("zeemux not found (expected `zeemux` next to lazygit or in PATH)")
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if !info.Mode().IsRegular() {
		return false
	}
	return info.Mode()&0o111 != 0
}
