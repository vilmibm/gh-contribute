# `gh contribute` GitHub CLI extension

A [gh](https://github.com/cli/cli) extension for finding issues to help with in a GitHub repository.

This extension suggests a random issue in a given repository to work on.

## Installation
```
gh extension install vilmibm/gh-contribute
```

## Usage
```
gh contibute [flags]
```

### Flags
-h, --help          help for `contribute`
-R, --repo string   Repository to contribuet to

### Details

The extension either uses the current repository or a repository suppled via `-R` like `gh` itself.

Suggested issue is selected based on the following criteria:

- labelled `help wanted` or `good first issue`
- not labelled `blocked`
- opened within the past year
- do not have a PR associated with them

## Author

vilmibm <vilmibm@github.com>
