workflow "on pull request merge, delete the branch" {
  on = "pull_request"
  resolves = ["branch cleanup"]
}

action "branch cleanup" {
  needs = "merged-filter"
  uses = "giantswarm/branch-cleanup-action@master"
  secrets = ["GITHUB_TOKEN"]
}

action "merged-filter" {
  uses = "actions/bin/filter@master"
  args = "merged true"
}
