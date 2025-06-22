/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package main

import (
	"fmt"
	"os"

	"github.com/mwat56/dnscache"
	"github.com/rivo/tview"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

// `createMainLayout()` creates the main UI layout.
//
// Parameters:
//   - `aState`: The application state.
//
// Returns:
//   - `tview.Primitive`: The main UI layout.
func createMainLayout(aState *tAppState) tview.Primitive {
	// Create menu
	menu := tview.NewList().
		AddItem("DNS Cache", "View and manage DNS cache entries", 'c',
			func() {
				showCachePage(aState)
			}).
		AddItem("Allow Lists", "View and manage allow lists", 'a',
			func() {
				showAllowListPage(aState)
			}).
		AddItem("Block Lists", "View and manage block lists", 'b',
			func() {
				showBlockListPage(aState)
			}).
		AddItem("Metrics", "View DNS cache and resolver metrics", 'm',
			func() {
				showMetricsPage(aState)
			}).
		AddItem("Configuration", "Edit configuration", 's',
			func() {
				showConfigPage(aState)
			}).
		AddItem("Quit", "Exit the application", 'q',
			func() {
				aState.app.Stop()
			})

	menu.SetBorderPadding(1, 1, 2, 2)
	menu.SetTitle(" DNS Cache Manager ")
	menu.SetTitleColor(colourHeader)
	menu.SetBorder(true)

	// Create main layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(menu, 30, 1, true).
			AddItem(aState.pages, 0, 3, false),
			0, 1, true).
		AddItem(aState.statusBar, 1, 1, false)

	// Add default page
	aState.pages.AddPage("default", tview.NewBox(), true, true)

	return flex
} // createMainLayout()

// `main()` runs the application
func main() {
	// Create theResolver with default options
	// theResolver := dnscache.New(0) // 5-minute refresh interval

	// Create application
	theApp := tview.NewApplication()

	// // Create thePages for navigation
	// thePages := tview.NewPages()

	// // Create status bar
	// theStatus := tview.NewTextView().
	// 	SetTextColor(colourText).
	// 	SetText("Ready")

	// Initialize application state structure
	state := &tAppState{
		app:      theApp,
		pages:    tview.NewPages(),
		resolver: dnscache.New(0),
		statusBar: tview.NewTextView().
			SetTextColor(colourText).
			SetText("Ready"),
	}

	// // Create main layout
	// layout := createMainLayout(state)

	// Run application
	err := theApp.SetRoot(createMainLayout(state), true).
		EnableMouse(true).
		Run()
	if nil != err {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
} // main()

/* _EoF_ */
