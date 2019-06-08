workflow "on pull request merge, delete the branch" {
  on = "closed"
  resolves = ["branch cleanup"]
}

action "branch cleanup" {
  uses = "giantswarm/branch-cleanup-action@master"
  secrets = ["GITHUB_TOKEN"]
}
