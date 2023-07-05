/*
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
*/

package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gonutz/wui/v2"
)

func main() {
	if len(os.Args) != 2 {
		wui.MessageBoxError("Error", "Expecting the commit message file as argument")
		return
	}

	msgPath := os.Args[1]
	sessionPath := filepath.Join(os.Getenv("APPDATA"), "commited_last")

	var msg string
	if data, err := ioutil.ReadFile(msgPath); err == nil {
		msg = string(data)
	}

	msg = strings.Replace(msg, "\r", "", -1)
	lines := strings.Split(msg, "\n")

	commentStart := "#" // The default comment character.

	// Find the comment start character by looking at the existing message.
	// The last five lines will contain an explanation of how to edit Git
	// messages. They all start with the comment character.
	var starts []string
	for _, line := range lines {
		if line != "" {
			starts = append(starts, start(line))
		}
	}
	if len(starts) > 5 && allTheSame(starts[len(starts)-5:]) {
		commentStart = starts[len(starts)-1]
	}

	var nonCommentLines []string
	for _, line := range lines {
		if !strings.HasPrefix(line, commentStart) {
			nonCommentLines = append(nonCommentLines, line)
		}
	}

	for len(nonCommentLines) > 0 && nonCommentLines[0] == "" {
		nonCommentLines = nonCommentLines[1:]
	}

	// Change the lineH and font heights to scale the UI.
	const lineH = 25
	font, _ := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -17})
	fixedFont, _ := wui.NewFont(wui.FontDesc{Name: "Consolas", Height: -19})
	w := wui.NewWindow()
	w.SetHasMinButton(false)
	w.SetHasMaxButton(false)
	w.SetResizable(false)
	w.SetFont(font)
	w.SetTitle("Enter Commit Message")
	w.SetInnerSize(800, 600)

	lineY := 0
	writeLine := func(text string) {
		lineY += lineH
		l := wui.NewLabel()
		w.Add(l)
		l.SetText(text)
		l.SetBounds(0, lineY, w.InnerWidth(), lineH)
		l.SetAlignment(wui.AlignCenter)
	}
	writeLine("Press CTRL+ENTER to commit")
	writeLine("Press ESC to abort commit")
	writeLine("Press CTRL+F to format the title and message")

	y := lineY + 2*lineH
	titleCap := wui.NewLabel()
	w.Add(titleCap)
	titleCap.SetBounds(0, y, 50, lineH+5)
	titleCap.SetAlignment(wui.AlignRight)
	titleCap.SetText("Title")

	title := wui.NewEditLine()
	w.Add(title)
	title.SetBounds(60, y, w.InnerWidth()-120, lineH+5)
	title.SetFont(fixedFont)
	title.SetCharacterLimit(72)

	titleLength := wui.NewLabel()
	w.Add(titleLength)
	titleLength.SetBounds(w.InnerWidth()-50, y, 40, lineH+5)
	title.SetOnTextChange(func() {
		n := utf8.RuneCountInString(title.Text())
		text := strconv.Itoa(n)
		if n > 50 {
			text += " !"
		}
		if titleLength.Text() != text {
			titleLength.SetText(text)
		}
	})

	y += lineH + 10
	text := wui.NewTextEdit()
	w.Add(text)
	text.SetBounds(10, y, w.InnerWidth()-20, w.InnerHeight()-y-10)
	text.SetFont(fixedFont)

	format := func() {
		newTitle := title.Text()
		newTitle = strings.TrimSpace(newTitle)
		for {
			// trim all spaces and periods, this might take several iterations,
			// e.g. if the title is "blah . . ."
			before := newTitle
			newTitle = strings.TrimSpace(newTitle)
			newTitle = strings.TrimSuffix(newTitle, ".")
			if newTitle == before {
				break
			}
		}
		newTitle = capitalize(newTitle)
		title.SetText(newTitle)

		var newText string
		clean := text.Text()
		clean = strings.TrimSpace(clean)
		clean = strings.Replace(clean, "\t", "    ", -1)
		lines := strings.Split(clean, "\r\n")
		for _, line := range lines {
			indent := indentation(line)
			a, b := splitLine(line, indent)
			newText += a + "\r\n"
			for b != "" {
				a, b = splitLine(b, indent)
				newText += a + "\r\n"
			}
		}
		text.SetText(strings.TrimSuffix(newText, "\r\n"))
	}
	w.SetShortcut(format, wui.KeyControl, wui.KeyF)
	abort := func() {
		ioutil.WriteFile(msgPath, nil, 0666)
		w.Close()
	}
	w.SetShortcut(abort, wui.KeyEscape)
	commit := func() {
		output := title.Text()
		if len(text.Text()) > 0 {
			output += "\r\n\r\n" + text.Text()
		}

		// warn if a line will be interpreted as a comment
		commentAt := ""
		if strings.HasPrefix(output, commentStart) {
			commentAt = "Your title"
		} else if strings.Contains(output, "\n"+commentStart) {
			commentAt = "One or more lines of your commit mesage"
		}
		if commentAt != "" {
			if !wui.MessageBoxYesNo(
				"Comments Found",
				commentAt+" starts with a '"+commentStart+"' character which is interpreted by git commit as a comment.\r\n\r\nMake sure you use commit option\r\n    --cleanup=whitespace\r\nto allow lines to start with '"+commentStart+"' in your message.\r\n\r\nDo you really want to commit now?",
			) {
				return
			}
		}

		if err := ioutil.WriteFile(msgPath, []byte(output), 0666); err != nil {
			wui.MessageBoxError("Error", "Failed to save commit message: "+err.Error())
		}
		// after a commit we want the contents to be empty when we open commited
		// the next time
		text.SetText("")
		title.SetText("")
		w.Close()
	}
	w.SetShortcut(commit, wui.KeyControl, wui.KeyReturn)

	readLastSession := func() {
		// restore the last contents for title and message text
		data, err := ioutil.ReadFile(sessionPath)
		if err == nil {
			titleText := string(data)
			msgText := ""
			if i := bytes.Index(data, []byte{0}); i != -1 {
				titleText = string(data[:i])
				msgText = string(data[i+1:])
			}
			title.SetText(titleText)
			text.SetText(msgText)
		}
	}

	copyMessageToGui := func() {
		titleText := nonCommentLines[0]
		msgText := strings.TrimSpace(strings.Join(nonCommentLines[1:], "\r\n"))
		title.SetText(titleText)
		text.SetText(msgText)
	}

	w.SetOnShow(func() {
		if len(nonCommentLines) == 0 {
			readLastSession()
		} else {
			copyMessageToGui()
		}
		title.Focus()
	})
	w.SetOnClose(func() {
		// remember the contents for the next time we open commited
		ioutil.WriteFile(
			sessionPath,
			append(append([]byte(title.Text()), 0), []byte(text.Text())...),
			0666,
		)
	})
	w.Show()
}

// split line returns the two parts of a line if it is too long, or the line and
// the empty string if the line is already short enough. The given indent is
// used as a prefix for all new lines.
func splitLine(s, indent string) (a, b string) {
	const maxLineLen = 72

	if len(s) <= maxLineLen {
		return s, ""
	}

	i := strings.LastIndex(s[:maxLineLen], " ")
	if i == -1 {
		return s, "" // do not split, leave it a long line
	}
	return s[:i], indent + s[i+1:]
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}

func indentation(s string) string {
	spaces := 0
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			break
		}
		spaces++
	}
	return strings.Repeat(" ", spaces)
}

func start(s string) string {
	for _, r := range s {
		return string(r)
	}
	return ""
}

func allTheSame(s []string) bool {
	for i := 1; i < len(s); i++ {
		if s[i-1] != s[i] {
			return false
		}
	}
	return true
}
