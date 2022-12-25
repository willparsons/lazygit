package cherry_pick

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var CherryPick = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Cherry pick commits from the subcommits view, without conflicts",
	ExtraCmdArgs: "",
	Skip:         false,
	SetupConfig:  func(config *config.AppConfig) {},
	SetupRepo: func(shell *Shell) {
		shell.
			EmptyCommit("base").
			NewBranch("first-branch").
			NewBranch("second-branch").
			Checkout("first-branch").
			EmptyCommit("one").
			EmptyCommit("two").
			Checkout("second-branch").
			EmptyCommit("three").
			EmptyCommit("four").
			Checkout("first-branch")
	},
	Run: func(shell *Shell, input *Input, assert *Assert, keys config.KeybindingConfig) {
		input.SwitchToBranchesWindow()
		assert.CurrentViewName("localBranches")

		assert.CurrentViewLines(
			Contains("first-branch"),
			Contains("second-branch"),
			Contains("master"),
		)

		input.NextItem()

		input.Enter()

		assert.CurrentViewName("subCommits")
		assert.CurrentViewLines(
			Contains("four"),
			Contains("three"),
			Contains("base"),
		)

		// copy commits 'four' and 'three'
		input.Press(keys.Commits.CherryPickCopy)
		assert.ViewContent("information", Contains("1 commit copied"))
		input.NextItem()
		input.Press(keys.Commits.CherryPickCopy)
		assert.ViewContent("information", Contains("2 commits copied"))

		input.SwitchToCommitsWindow()
		assert.CurrentViewName("commits")

		assert.CurrentViewLines(
			Contains("two"),
			Contains("one"),
			Contains("base"),
		)

		input.Press(keys.Commits.PasteCommits)
		input.Alert(Equals("Cherry-Pick"), Contains("Are you sure you want to cherry-pick the copied commits onto this branch?"))

		assert.CurrentViewName("commits")
		assert.CurrentViewLines(
			Contains("four"),
			Contains("three"),
			Contains("two"),
			Contains("one"),
			Contains("base"),
		)

		assert.ViewContent("information", Contains("2 commits copied"))
		input.Press(keys.Universal.Return)
		assert.ViewContent("information", NotContains("commits copied"))
	},
})
