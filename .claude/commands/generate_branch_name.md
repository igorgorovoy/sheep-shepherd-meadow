# Generate Git Branch Name

Based on the `Instructions` below, take the `Variables` follow the `Run` section to generate a concise Git branch name following the specified format. Then follow the `Report` section to report the results of your work.

## Variables

issue_class: $1
issue: $2

## Instructions

- Generate a branch name in the format: `<issue_class>-issue-<issue_number>-<concise_name>`
- The `<concise_name>` should be:
  - 3-6 words maximum
  - All lowercase
  - Words separated by hyphens
  - Descriptive of the main task/feature
  - No special characters except hyphens
- Examples:
  - `feat-issue-123-add-user-auth`
  - `bug-issue-456-fix-login-error`
  - `chore-issue-789-update-dependencies`
  - `test-issue-323-fix-failing-tests`
- Extract the issue number, title, and body from the issue JSON

## Run

Generate the branch name based on the instructions above.
Do NOT create or checkout any branches - just generate the name.

## Report

Return ONLY the generated branch name (no other text)