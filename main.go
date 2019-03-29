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

	"github.com/gonutz/w32"
	"github.com/gonutz/wui"
)

func main() {
	if len(os.Args) != 2 {
		wui.MessageBoxError("Error", "Expecting the commit message file as argument")
		return
	}

	msgPath := os.Args[1]

	font, _ := wui.NewFont(wui.FontDesc{Name: "Tahoma", Height: -13})
	fixedFont, _ := wui.NewFont(wui.FontDesc{Name: "Consolas", Height: -15})
	w := wui.NewDialogWindow()
	w.SetFont(font)
	w.SetTitle("Enter Commit Message")
	w.SetClientSize(800, 600)

	writeLine := func(text string, y int) {
		l := wui.NewLabel()
		w.Add(l)
		l.SetText(text)
		l.SetBounds(0, y, w.ClientWidth(), 20)
		l.SetCenterAlign()
	}
	writeLine("Press CTRL+ENTER to commit", 20)
	writeLine("Press ESC to abort commit", 40)
	writeLine("Press CTRL+F to format the title and message", 60)

	titleCap := wui.NewLabel()
	w.Add(titleCap)
	titleCap.SetBounds(0, 100, 50, 25)
	titleCap.SetRightAlign()
	titleCap.SetText("Title")

	title := wui.NewEditLine()
	w.Add(title)
	title.SetBounds(60, 100, w.ClientWidth()-120, 25)
	title.SetFont(fixedFont)
	title.SetCharacterLimit(72)

	titleLength := wui.NewLabel()
	w.Add(titleLength)
	titleLength.SetBounds(w.ClientWidth()-50, 100, 40, 25)
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

	text := wui.NewTextEdit()
	w.Add(text)
	text.SetBounds(10, 130, w.ClientWidth()-20, w.ClientHeight()-140)
	text.SetFont(fixedFont)

	w.SetShortcut(wui.ShortcutKeys{Mod: wui.ModControl, Rune: 'F'}, func() {
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
	})
	w.SetShortcut(wui.ShortcutKeys{Key: w32.VK_ESCAPE}, w.Close)
	w.SetShortcut(wui.ShortcutKeys{Mod: wui.ModControl, Key: w32.VK_RETURN}, func() {
		output := title.Text()
		if len(text.Text()) > 0 {
			output += "\r\n\r\n" + text.Text()
		}
		if err := ioutil.WriteFile(msgPath, []byte(output), 0666); err != nil {
			wui.MessageBoxError("Error", "Failed to save commit message: "+err.Error())
		}
		// after a commit we want the contents to be empty when we open commited
		// the next time
		text.SetText("")
		title.SetText("")
		w.Close()
	})
	toggleFocus := func() {
		if title.HasFocus() {
			text.Focus()
		} else {
			title.Focus()
		}
	}
	w.SetShortcut(wui.ShortcutKeys{Key: w32.VK_TAB}, toggleFocus)
	w.SetShortcut(wui.ShortcutKeys{Mod: wui.ModShift, Key: w32.VK_TAB}, toggleFocus)

	sessionPath := filepath.Join(os.Getenv("APPDATA"), "commited_last")
	w.SetOnShow(func() {
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
