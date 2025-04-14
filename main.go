package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	// Pobierz listę sesji tmux
	cmd := exec.Command("tmux", "ls")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Brak aktywnych sesji tmux.")
		return
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var sessionNames []string
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) > 0 {
			sessionNames = append(sessionNames, parts[0])
		}
	}

	if len(sessionNames) == 0 {
		fmt.Println("Brak aktywnych sesji tmux.")
		return
	}

	app := tview.NewApplication()
	list := tview.NewList().ShowSecondaryText(false)

	selected := map[int]bool{}

	// Dodaj sesje do listy
	for i, name := range sessionNames {
		index := i
		list.AddItem("☐ "+name, "", 0, func() {
			selected[index] = !selected[index]
			prefix := "☑"
			if !selected[index] {
				prefix = "☐"
			}
			list.SetItemText(index, prefix+" "+sessionNames[index], "")
		})
	}

	// Obsługa j/k i innych klawiszy
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			current := list.GetCurrentItem()
			if current < len(sessionNames)-1 {
				list.SetCurrentItem(current + 1)
			}
			return nil
		case 'k':
			current := list.GetCurrentItem()
			if current > 0 {
				list.SetCurrentItem(current - 1)
			}
			return nil
		case ' ':
			index := list.GetCurrentItem()
			selected[index] = !selected[index]
			prefix := "☑"
			if !selected[index] {
				prefix = "☐"
			}
			list.SetItemText(index, prefix+" "+sessionNames[index], "")
			return nil
		case 0: // Ctrl + coś – ignoruj
			return event
		}
		if event.Key() == tcell.KeyEnter {
			app.Stop()
			return nil
		}
		return event
	})

	// Layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().
			SetText("Wybierz sesje do usunięcia (spacja = zaznacz, j/k lub strzałki = poruszanie, Enter = usuń)").
			SetTextColor(tcell.ColorYellow), 1, 0, false).
		AddItem(list, 0, 1, true)

	// Start aplikacji
	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}

	// Po ENTER
	for i, isSelected := range selected {
		if isSelected {
			name := sessionNames[i]
			fmt.Printf("Zabijam sesję: %s\n", name)
			exec.Command("tmux", "kill-session", "-t", name).Run()
		}
	}
}
