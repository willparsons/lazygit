package components

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jesseduffield/lazygit/pkg/gui/types"
	integrationTypes "github.com/jesseduffield/lazygit/pkg/integration/types"
)

// through this struct we assert on the state of the lazygit gui

type Assert struct {
	gui integrationTypes.GuiDriver
}

func NewAssert(gui integrationTypes.GuiDriver) *Assert {
	return &Assert{gui: gui}
}

func Contains(target string) *matcher {
	return NewMatcher(
		fmt.Sprintf("contains '%s'", target),
		func(value string) (bool, string) {
			return strings.Contains(value, target), fmt.Sprintf("Expected '%s' to be found in '%s'", target, value)
		},
	)
}

func NotContains(target string) *matcher {
	return NewMatcher(
		fmt.Sprintf("does not contain '%s'", target),
		func(value string) (bool, string) {
			return !strings.Contains(value, target), fmt.Sprintf("Expected '%s' to NOT be found in '%s'", target, value)
		},
	)
}

func MatchesRegexp(target string) *matcher {
	return NewMatcher(
		fmt.Sprintf("matches regular expression '%s'", target),
		func(value string) (bool, string) {
			matched, err := regexp.MatchString(target, value)
			if err != nil {
				return false, fmt.Sprintf("Unexpected error parsing regular expression '%s': %s", target, err.Error())
			}
			return matched, fmt.Sprintf("Expected '%s' to match regular expression '%s'", value, target)
		},
	)
}

func Equals(target string) *matcher {
	return NewMatcher(
		fmt.Sprintf("equals '%s'", target),
		func(value string) (bool, string) {
			return target == value, fmt.Sprintf("Expected '%s' to equal '%s'", value, target)
		},
	)
}

func (self *Assert) WorkingTreeFileCount(expectedCount int) {
	self.assertWithRetries(func() (bool, string) {
		actualCount := len(self.gui.Model().Files)

		return actualCount == expectedCount, fmt.Sprintf(
			"Expected %d changed working tree files, but got %d",
			expectedCount, actualCount,
		)
	})
}

func (self *Assert) CommitCount(expectedCount int) {
	self.assertWithRetries(func() (bool, string) {
		actualCount := len(self.gui.Model().Commits)

		return actualCount == expectedCount, fmt.Sprintf(
			"Expected %d commits present, but got %d",
			expectedCount, actualCount,
		)
	})
}

func (self *Assert) StashCount(expectedCount int) {
	self.assertWithRetries(func() (bool, string) {
		actualCount := len(self.gui.Model().StashEntries)

		return actualCount == expectedCount, fmt.Sprintf(
			"Expected %d stash entries, but got %d",
			expectedCount, actualCount,
		)
	})
}

func (self *Assert) AtLeastOneCommit() {
	self.assertWithRetries(func() (bool, string) {
		actualCount := len(self.gui.Model().Commits)

		return actualCount > 0, "Expected at least one commit present"
	})
}

func (self *Assert) HeadCommitMessage(matcher *matcher) {
	self.assertWithRetries(func() (bool, string) {
		return len(self.gui.Model().Commits) > 0, "Expected at least one commit to be present"
	})

	self.matchString(matcher, "Unexpected commit message.",
		func() string {
			return self.gui.Model().Commits[0].Name
		},
	)
}

func (self *Assert) CurrentViewName(expectedViewName string) {
	self.assertWithRetries(func() (bool, string) {
		actual := self.gui.CurrentContext().GetView().Name()
		return actual == expectedViewName, fmt.Sprintf("Expected current view name to be '%s', but got '%s'", expectedViewName, actual)
	})
}

func (self *Assert) CurrentWindowName(expectedWindowName string) {
	self.assertWithRetries(func() (bool, string) {
		actual := self.gui.CurrentContext().GetView().Name()
		return actual == expectedWindowName, fmt.Sprintf("Expected current window name to be '%s', but got '%s'", expectedWindowName, actual)
	})
}

func (self *Assert) CurrentBranchName(expectedViewName string) {
	self.assertWithRetries(func() (bool, string) {
		actual := self.gui.CheckedOutRef().Name
		return actual == expectedViewName, fmt.Sprintf("Expected current branch name to be '%s', but got '%s'", expectedViewName, actual)
	})
}

func (self *Assert) InListContext() {
	self.assertWithRetries(func() (bool, string) {
		currentContext := self.gui.CurrentContext()
		_, ok := currentContext.(types.IListContext)
		return ok, fmt.Sprintf("Expected current context to be a list context, but got %s", currentContext.GetKey())
	})
}

func (self *Assert) SelectedLine(matcher *matcher) {
	self.matchString(matcher, "Unexpected selected line.",
		func() string {
			return self.gui.CurrentContext().GetView().SelectedLine()
		},
	)
}

func (self *Assert) SelectedLineIdx(expected int) {
	self.assertWithRetries(func() (bool, string) {
		actual := self.gui.CurrentContext().GetView().SelectedLineIdx()
		return expected == actual, fmt.Sprintf("Expected selected line index to be %d, got %d", expected, actual)
	})
}

func (self *Assert) InPrompt() {
	self.assertWithRetries(func() (bool, string) {
		currentView := self.gui.CurrentContext().GetView()
		return currentView.Name() == "confirmation" && currentView.Editable, "Expected prompt popup to be focused"
	})
}

func (self *Assert) InConfirm() {
	self.assertWithRetries(func() (bool, string) {
		currentView := self.gui.CurrentContext().GetView()
		return currentView.Name() == "confirmation" && !currentView.Editable, "Expected confirmation popup to be focused"
	})
}

func (self *Assert) InAlert() {
	// basically the same thing as a confirmation popup with the current implementation
	self.assertWithRetries(func() (bool, string) {
		currentView := self.gui.CurrentContext().GetView()
		return currentView.Name() == "confirmation" && !currentView.Editable, "Expected alert popup to be focused"
	})
}

func (self *Assert) InCommitMessagePanel() {
	self.assertWithRetries(func() (bool, string) {
		currentView := self.gui.CurrentContext().GetView()
		return currentView.Name() == "commitMessage", "Expected commit message panel to be focused"
	})
}

func (self *Assert) InMenu() {
	self.assertWithRetries(func() (bool, string) {
		return self.gui.CurrentContext().GetView().Name() == "menu", "Expected popup menu to be focused"
	})
}

func (self *Assert) NotInPopup() {
	self.assertWithRetries(func() (bool, string) {
		currentViewName := self.gui.CurrentContext().GetView().Name()
		return currentViewName != "menu" && currentViewName != "confirmation" && currentViewName != "commitMessage", "Expected popup not to be focused"
	})
}

func (self *Assert) CurrentViewTitle(matcher *matcher) {
	self.matchString(matcher, "Unexpected current view title.",
		func() string {
			return self.gui.CurrentContext().GetView().Title
		},
	)
}

func (self *Assert) ViewContent(viewName string, matcher *matcher) {
	self.matchString(matcher, fmt.Sprintf("Unexpected content in view '%s'.", viewName),
		func() string {
			return self.gui.View(viewName).Buffer()
		},
	)
}

// asserts that the given view has lines matching the given matchers.
func (self *Assert) ViewLines(viewName string, matchers ...*matcher) {
	self.assertWithRetries(func() (bool, string) {
		lines := self.gui.View(viewName).BufferLines()
		return len(lines) == len(matchers), fmt.Sprintf("unexpected number of lines in view. Expected %d, got %d", len(matchers), len(lines))
	})

	for i, matcher := range matchers {
		self.matchString(matcher, fmt.Sprintf("Unexpected content in view '%s'.", viewName),
			func() string {
				return self.gui.View(viewName).BufferLines()[i]
			},
		)
	}
}

func (self *Assert) CurrentViewLines(matchers ...*matcher) {
	self.ViewLines(self.gui.CurrentContext().GetView().Name(), matchers...)
}

// asserts that the given view has lines matching the given matchers. So if three matchers
// are passed, we only check the first three lines of the view.
func (self *Assert) ViewTopLines(viewName string, matchers ...*matcher) {
	self.assertWithRetries(func() (bool, string) {
		lines := self.gui.View(viewName).BufferLines()
		return len(lines) >= len(matchers), fmt.Sprintf("unexpected number of lines in view. Expected at least %d, got %d", len(matchers), len(lines))
	})

	for i, matcher := range matchers {
		self.matchString(matcher, fmt.Sprintf("Unexpected content in view '%s'.", viewName),
			func() string {
				return self.gui.View(viewName).BufferLines()[i]
			},
		)
	}
}

func (self *Assert) CurrentViewTopLines(matchers ...*matcher) {
	self.ViewTopLines(self.gui.CurrentContext().GetView().Name(), matchers...)
}

func (self *Assert) CurrentViewContent(matcher *matcher) {
	self.matchString(matcher, "Unexpected content in current view.",
		func() string {
			return self.gui.CurrentContext().GetView().Buffer()
		},
	)
}

func (self *Assert) MainViewContent(matcher *matcher) {
	self.matchString(matcher, "Unexpected main view content.",
		func() string {
			return self.gui.MainView().Buffer()
		},
	)
}

func (self *Assert) SecondaryViewContent(matcher *matcher) {
	self.matchString(matcher, "Unexpected secondary view title.",
		func() string {
			return self.gui.SecondaryView().Buffer()
		},
	)
}

func (self *Assert) matchString(matcher *matcher, context string, getValue func() string) {
	self.assertWithRetries(func() (bool, string) {
		value := getValue()
		return matcher.context(context).test(value)
	})
}

func (self *Assert) assertWithRetries(test func() (bool, string)) {
	waitTimes := []int{0, 1, 1, 1, 1, 1, 5, 10, 20, 40, 100, 200, 500, 1000, 2000, 4000}

	var message string
	for _, waitTime := range waitTimes {
		time.Sleep(time.Duration(waitTime) * time.Millisecond)

		var ok bool
		ok, message = test()
		if ok {
			return
		}
	}

	self.Fail(message)
}

// for when you just want to fail the test yourself
func (self *Assert) Fail(message string) {
	self.gui.Fail(message)
}

// This does _not_ check the files panel, it actually checks the filesystem
func (self *Assert) FileSystemPathPresent(path string) {
	self.assertWithRetries(func() (bool, string) {
		_, err := os.Stat(path)
		return err == nil, fmt.Sprintf("Expected path '%s' to exist, but it does not", path)
	})
}

// This does _not_ check the files panel, it actually checks the filesystem
func (self *Assert) FileSystemPathNotPresent(path string) {
	self.assertWithRetries(func() (bool, string) {
		_, err := os.Stat(path)
		return os.IsNotExist(err), fmt.Sprintf("Expected path '%s' to not exist, but it does", path)
	})
}
