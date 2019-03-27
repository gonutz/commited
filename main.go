package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"time"
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
	w := wui.NewDialogWindow()
	w.SetFont(font)
	w.SetTitle("Enter Commit Message")
	w.SetClientSize(800, 600)

	ok := wui.NewLabel()
	w.Add(ok)
	ok.SetText("Press CTRL+ENTER to commit")
	ok.SetBounds(0, 0, w.ClientWidth(), 20)
	ok.SetCenterAlign()

	esc := wui.NewLabel()
	w.Add(esc)
	esc.SetText("Press ESC to abort commit")
	esc.SetBounds(0, 20, w.ClientWidth(), 20)
	esc.SetCenterAlign()

	title := wui.NewEditLine()
	w.Add(title)
	title.SetBounds(10, 40, w.ClientWidth()-70, 25)

	titleLength := wui.NewLabel()
	w.Add(titleLength)
	titleLength.SetBounds(w.ClientWidth()-50, 40, 40, 25)
	go func() {
		for {
			n := strconv.Itoa(utf8.RuneCountInString(title.Text()))
			if titleLength.Text() != n {
				titleLength.SetText(n)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	text := wui.NewTextEdit()
	w.Add(text)
	text.SetBounds(10, 70, w.ClientWidth()-20, w.ClientHeight()-80)

	w.SetShortcut(wui.ShortcutKeys{Key: w32.VK_ESCAPE}, w.Close)
	w.SetShortcut(wui.ShortcutKeys{Mod: wui.ModControl, Key: w32.VK_RETURN}, func() {
		output := title.Text()
		if len(text.Text()) > 0 {
			output += "\r\n\r\n" + text.Text()
		}
		if err := ioutil.WriteFile(msgPath, []byte(output), 0666); err != nil {
			wui.MessageBoxError("Error", "Failed to save commit message: "+err.Error())
		}
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
	w.SetOnShow(func() {
		title.Focus()
	})
	w.Show()
}
