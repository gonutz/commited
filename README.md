![screenshot.png](https://raw.githubusercontent.com/gonutz/commited/master/screenshot.png)

# commited

This is a Windows editor to use for git commit messages. To use it as the
default editor, put the executable somewhere in your PATH and tell git about it:

	git config --global core.editor "commited"

Here are seven rules for git commit messages from

	https://chris.beams.io/posts/git-commit

1. Separate subject from body with a blank line
2. Limit the subject line to 50 characters
3. Capitalize the subject line
4. Do not end the subject line with a period
5. Use the imperative mood in the subject line
6. Wrap the body at 72 characters
7. Use the body to explain what and why vs. how

This editor helps you follow these rules:

- title and message are separate input fields, the commit message will have a
  blank line separating them (rule 1)
- the title input field shows the title's length and displays a "!" for titles
  longer then 50 characters (rule 2)
- Ctrl+F formats the title to start with a capital letter and removes trailing
  periods (rules 3 and 4)
- Ctrl+F wraps the message text to a maximum of 72 characters per line (rule 6)

Rules 5 and 7 are about the message's content so this is left to the author :-)

# build

If you have [Go](https://go.dev/) installed, just run

	go install github.com/gonutz/commited@latest

to download, build and install the latest verison.
