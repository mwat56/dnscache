/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package main

import (
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/mwat56/dnscache"
	"github.com/rivo/tview"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

// `connectToExistingInstance()` connects to an existing instance in remote
// control mode
//
// Parameters:
//   - `aConfig`: Configuration for the remote instance.
func connectToExistingInstance(aConfig tConfiguration) {
	// Create application
	theApp := tview.NewApplication()

	// Try to connect to the existing instance
	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", aConfig.Port+1))
	if nil != err {
		// Failed to connect, show error and exit
		fmt.Printf("Failed to connect to existing instance: %v\n", err)
		os.Exit(1)
	}

	// Initialize application state structure with remote resolver
	state := &tAppState{
		app:      theApp,
		pages:    tview.NewPages(),
		resolver: dnscache.New(0),
		statusBar: tview.NewTextView().
			SetTextColor(colourText).
			SetText("Connected to remote instance"),
		remoteConn: conn,
	}

	// Run application
	err = theApp.SetRoot(createMainLayout(state), true).
		EnableMouse(true).
		EnablePaste(true).
		Run()
	if nil != err {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
} // connectToExistingInstance()

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

// `main()` runs the application.
func main() {
	// Parse command line arguments
	cmdLineConf := parseCmdLineArgs(os.Args[1:])

	// Load configuration file
	config, err := loadConfiguration(cmdLineConf.ConfigPathName)
	if nil != err {
		// Use default values if config file doesn't exist
		config = tConfiguration{
			DataDir:         os.TempDir(),
			CacheSize:       1024,
			RefreshInterval: 5,
			TTL:             60,
		}
	}

	// Use command line address if provided, otherwise use config address
	if "" != cmdLineConf.Address {
		config.Address = cmdLineConf.Address
	}
	// Use command line port if provided
	if 0 != cmdLineConf.Port {
		config.Port = cmdLineConf.Port
	}

	// Check for existing instance
	if isInstanceRunning() {
		if cmdLineConf.ConsoleMode {
			// Connect to existing instance in remote control mode
			connectToExistingInstance(config)
			return
		} else {
			fmt.Println("An instance is already running. Use --console to connect to it.")
			os.Exit(1)
		}
	}

	// Run as daemon if requested (Linux only)
	if cmdLineConf.DaemonMode && ("linux" == runtime.GOOS) {
		if err := runAsDaemon(); nil != err {
			fmt.Printf("Failed to start daemon: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Create myResolver with configuration
	myResolver := dnscache.NewWithOptions(dnscache.TResolverOptions{
		DNSservers:      config.DNSServers,
		DataDir:         config.DataDir,
		CacheSize:       config.CacheSize,
		RefreshInterval: config.RefreshInterval,
		TTL:             config.TTL,
	})

	// Start DNS server if not in console mode
	if !cmdLineConf.ConsoleMode {
		if err := startDNSserver(myResolver, config.Address, config.Port, config.Forwarder); nil != err {
			fmt.Printf("Failed to start DNS server: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Console mode - create and run the UI
	theApp := tview.NewApplication()

	// Initialize application state structure
	state := &tAppState{
		app:      theApp,
		pages:    tview.NewPages(),
		resolver: myResolver,
		statusBar: tview.NewTextView().
			SetTextColor(colourText).
			SetText("Ready"),
	}

	// Run application
	err = theApp.SetRoot(createMainLayout(state), true).
		EnableMouse(true).
		EnablePaste(true).
		Run()
	if nil != err {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
} // main()

/* _EoF_ */
