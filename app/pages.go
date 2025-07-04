/*
Copyright © 2025  M.Watermann, 10247 Berlin, Germany

	    All rights reserved
	EMail : <support@mwat.de>
*/
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mwat56/dnscache"
	"github.com/rivo/tview"
)

//lint:file-ignore ST1017 - I prefer Yoda conditions

const (
	// UI colours:
	colourBackground = tcell.ColorTeal
	// colourError      = tcell.ColorRed
	colourHeader    = tcell.ColorYellow
	colourHighlight = tcell.ColorGreen
	colourText      = tcell.ColorWhite
)

type (
	// `tAppState` is the application state.
	tAppState struct {
		app           *tview.Application
		pages         *tview.Pages
		resolver      *dnscache.TResolver
		statusBar     *tview.TextView
		configChanged bool
		remoteConn    net.Conn // Connection to remote instance (if any)
	}

	// `tCacheEntry` represents a DNS cache entry with hostname and IPs
	tCacheEntry struct {
		Hostname string
		IPs      []net.IP
	}
)

// ---------------------------------------------------------------------------
// Helper functions:

// `getCacheEntries()` returns all entries in the DNS cache.
//
// Parameters:
//   - `aState`: The application state.
//
// Returns:
//   - `[]CacheEntry`: List of cache entries.
//   - `error`: Error if any occurred.
func getCacheEntries(aState *tAppState) ([]tCacheEntry, error) {
	if (nil == aState) || (nil == aState.resolver) {
		return nil, fmt.Errorf("app or resolver not initialised")
	}

	entries := []tCacheEntry{}

	// Create a context with timeout for the operation
	ctx, cancel := context.WithTimeout(context.Background(), time.Second<<2)
	defer cancel()

	// Get all hostnames from the cache
	for hostname := range aState.resolver.ICacheList.Range(ctx) {
		// Get IPs for this hostname
		ips, ok := aState.resolver.ICacheList.IPs(ctx, hostname)
		if !ok || (0 == len(ips)) {
			continue
		}

		// We don't have direct access to expiration time in the interface
		// For now, we'll use a placeholder expiration time
		// In a real implementation, we would need to extend the ICacheList interface
		// expiresAt := time.Now().Add(aState.resolver.ttl)

		entries = append(entries, tCacheEntry{
			Hostname: hostname,
			IPs:      ips,
			// ExpiresAt: expiresAt,
		})
	}

	return entries, nil
} // getCacheEntries()

// ---------------------------------------------------------------------------
// Page related functions:

// `showAllowListPage()` displays the allow lists
//
// Parameters:
//   - `aState`: The application state.
func showAllowListPage(aState *tAppState) {
	// Create form for managing allow lists
	form := tview.NewForm()
	// Store the box for potential styling/layout adjustments
	formBox := form.Box.SetTitle(" Allow Lists ").
		SetBackgroundColor(colourBackground).
		SetTitleColor(colourHeader).
		SetBorder(true)

	somethingChanged := false
	changeFunc := func(_ string) {
		somethingChanged = true
	}

	// Add input field for new allow list
	form.AddInputField("Add Allow List (File Path):", "", 50, nil, changeFunc)

	// Add buttons
	form.AddButton("Add", func() {
		if !somethingChanged {
			return
		}
		// Get the input value
		filePath := form.GetFormItem(0).(*tview.InputField).GetText()
		if "" != filePath {
			// TODO: Add allow list using the resolver
			// This requires a method to add an allow list at runtime

			if err := aState.resolver.LoadAllowlist(filePath); nil != err {
				aState.statusBar.SetText(fmt.Sprintf("Error loading allowlist: %v", err))
			} else {
				aState.statusBar.SetText("Allowlist added successfully")
				// Refresh the table
				// TODO: Implement refresh logic
			}
		}
	})

	// Create table for existing allow lists
	table := tview.NewTable().SetBorders(true)
	// Store the box for potential styling/layout adjustments
	tableBox := table.Box.SetTitle(" Current Allow Lists ").
		SetBackgroundColor(colourBackground).
		SetTitleColor(colourHeader).
		SetBorder(true)

	// Add headers
	table.SetCell(0, 0, tview.NewTableCell("Source").
		SetTextColor(colourHighlight).SetSelectable(false))
	table.SetCell(0, 1, tview.NewTableCell("Entries").
		SetTextColor(colourHighlight).SetSelectable(false))

	// TODO: Populate with actual allow lists
	// This requires a method to get current allow lists

	// Use the box objects for layout adjustments
	// Adjust padding or other box properties
	formBox.SetBorderPadding(1, 1, 2, 2)
	tableBox.SetBorderPadding(0, 0, 1, 1)

	// Create layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 7, 0, true).
		AddItem(table, 0, 1, false)

	// Add page
	aState.pages.AddPage("allow", flex, true, true)
	aState.app.SetFocus(form)

	// Update status
	aState.statusBar.SetText("Managing allow lists")
} // showAllowListPage()

// `showBlockListPage()` displays the block lists
//
// Parameters:
//   - `aState`: The application state.
func showBlockListPage(aState *tAppState) {
	// Create form for managing block lists
	form := tview.NewForm()
	formBox := form.Box.SetTitle(" Block Lists ").
		SetBackgroundColor(colourBackground).
		SetTitleColor(colourHeader).
		SetBorder(true)

	somethingChanged := false
	changeFunc := func(_ string) {
		somethingChanged = true
	}

	// Add input field for new block list
	form.AddInputField("Add Block List (URL):", "", 50, nil, changeFunc)

	// Create table for existing block lists
	table := tview.NewTable().SetBorders(true)
	tableBox := table.Box.SetTitle(" Current Block Lists ").
		SetBackgroundColor(colourBackground).
		SetTitleColor(colourHeader).
		SetBorder(true)

	// Add headers
	table.SetCell(0, 0, tview.NewTableCell("Source").SetTextColor(colourHighlight).SetSelectable(false))
	table.SetCell(0, 1, tview.NewTableCell("Entries").SetTextColor(colourHighlight).SetSelectable(false))

	// Function to populate the table with block lists
	populateTable := func() {
		// Clear existing rows (except header)
		// First get the row count
		if rowCount := table.GetRowCount(); 1 < rowCount {
			// Clear the table and re-add the header
			table.Clear()

			// Re-add headers
			table.SetCell(0, 0, tview.NewTableCell("Source").SetTextColor(colourHighlight).SetSelectable(false))
			table.SetCell(0, 1, tview.NewTableCell("Entries").SetTextColor(colourHighlight).SetSelectable(false))
		}

		// TODO: Get actual block lists
		// For now, we'll add a placeholder row
		table.SetCell(1, 0, tview.NewTableCell("(No block lists available)"))
		table.SetCell(1, 1, tview.NewTableCell(""))
	}

	// Initial population
	populateTable()

	// Add buttons
	form.AddButton("Add", func() {
		if !somethingChanged {
			return
		}
		// Get the input value
		url := form.GetFormItem(0).(*tview.InputField).GetText()
		if "" != url {
			// Add block list
			if err := aState.resolver.LoadBlocklists([]string{url}); nil != err {
				aState.statusBar.SetText(fmt.Sprintf("Error loading blocklist: %v", err))
			} else {
				aState.statusBar.SetText("Blocklist added successfully")
				// Clear the input field
				form.GetFormItem(0).(*tview.InputField).SetText("")
				// Refresh the table
				populateTable()
			}
		}
	})

	// Use the box objects for styling
	formBox.SetBorderPadding(1, 1, 2, 2)
	tableBox.SetBorderPadding(0, 0, 1, 1)

	// Create layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(form, 7, 0, true).
		AddItem(table, 0, 1, false)

	// Add page
	aState.pages.AddPage("block", flex, true, true)
	aState.app.SetFocus(form)

	// Update status
	aState.statusBar.SetText("Managing block lists")
} // showBlockListPage()

// `showCachePage()` displays the DNS cache entries.
//
// Parameters:
//   - `aState`: The application state.
func showCachePage(aState *tAppState) {
	// Create table for cache entries
	table := tview.NewTable().SetBorders(true)
	tableBox := table.Box.SetTitle(" DNS Cache Entries ").
		SetBackgroundColor(colourBackground).
		SetTitleColor(colourHeader).
		SetBorder(true)

	// Add headers
	table.SetCell(0, 0, tview.NewTableCell("Hostname").SetTextColor(colourHighlight).SetSelectable(false))
	table.SetCell(0, 1, tview.NewTableCell("IP Addresses").SetTextColor(colourHighlight).SetSelectable(false))
	// table.SetCell(0, 2, tview.NewTableCell("Expires").SetTextColor(colourHighlight).SetSelectable(false))

	// Function to populate the table with cache entries
	populateTable := func() {
		// Clear existing rows (except header)
		// First get the row count
		if rowCount := table.GetRowCount(); 1 < rowCount {
			// If there are rows beyond the header
			// clear the table and re-add the header
			table.Clear()

			// Re-add headers
			table.SetCell(0, 0, tview.NewTableCell("Hostname").SetTextColor(colourHighlight).SetSelectable(false))
			table.SetCell(0, 1, tview.NewTableCell("IP Addresses").SetTextColor(colourHighlight).SetSelectable(false))
		}

		// Get cache entries
		entries, err := getCacheEntries(aState)
		if nil != err {
			aState.statusBar.SetText(fmt.Sprintf("Error getting cache entries: %v", err))
			return
		}

		// Add entries to table
		for i, entry := range entries {
			row := i + 1 // +1 for header row

			// Format IP addresses as comma-separated string
			ipStrings := make([]string, len(entry.IPs))
			for idx, ip := range entry.IPs {
				ipStrings[idx] = ip.String()
			}
			ipText := strings.Join(ipStrings, ", ")

			// Format expiration time
			// expiresText := entry.ExpiresAt.Format("2006-01-02 15:04:05")

			// Add cells
			table.SetCell(row, 0, tview.NewTableCell(entry.Hostname))
			table.SetCell(row, 1, tview.NewTableCell(ipText))
			// table.SetCell(row, 2, tview.NewTableCell(expiresText))
		}

		aState.statusBar.SetText(fmt.Sprintf("Loaded %d cache entries", len(entries)))
	}

	// Initial population
	populateTable()

	// Set up key handling for the table
	table.SetInputCapture(func(aEvent *tcell.EventKey) *tcell.EventKey {
		switch aEvent.Key() {
		case tcell.KeyEscape:
			aState.pages.SwitchToPage("default")
			return nil

		case tcell.KeyDelete:
			// Delete the selected entry
			row, _ := table.GetSelection()
			if (0 < row) && (table.GetRowCount() > row) {
				hostname := table.GetCell(row, 0).Text

				// Create context for the operation
				ctx, cancel := context.WithTimeout(context.Background(), time.Second<<2)
				defer cancel()

				// Delete from cache
				aState.resolver.ICacheList.Delete(ctx, hostname)

				// Refresh table
				populateTable()

				aState.statusBar.SetText(fmt.Sprintf("Deleted entry for %s", hostname))
			}
			return nil

		case tcell.KeyRune:
			if ('r' == aEvent.Rune()) || ('R' == aEvent.Rune()) {
				// Refresh the table
				populateTable()
				return nil
			}
		}
		return aEvent
	})

	// Add controls
	controls := tview.NewTextView().
		SetText("Enter: View Details | Delete: Remove Entry | R: Refresh | ESC: Back").
		SetTextColor(colourText)

	// Use the box for styling
	tableBox.SetBorderPadding(0, 0, 1, 1)

	// Create layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(controls, 1, 0, false)

	// Add page
	aState.pages.AddPage("cache", flex, true, true)
	aState.app.SetFocus(table)

	// Update status
	aState.statusBar.SetText("Viewing DNS cache entries")
} // showCachePage()

// `showConfigPage()` displays the configuration
//
// Parameters:
//   - `aState`: The application state.
func showConfigPage(aState *tAppState) {
	// Try to load existing configuration
	config, err := loadConfiguration(gConfigFile)
	if nil != err {
		// Use default values if config file doesn't exist
		config = tConfiguration{
			DataDir:         os.TempDir(),
			CacheSize:       1024,
			Port:            53,
			RefreshInterval: 5,
			TTL:             60,
		}
	}
	somethingChanged := false
	changeFunc := func(_ string) {
		somethingChanged = true
	}

	// Create form for configuration
	form := tview.NewForm()
	formBox := form.SetTitle(" Configuration ").
		SetBackgroundColor(colourBackground).
		SetTitleColor(colourHeader).
		SetBorder(true)

	// Add configuration fields
	form.AddInputField("Data Directory:", config.DataDir, 50, nil, changeFunc)
	form.AddInputField("Cache Size:", strconv.Itoa(config.CacheSize), 10, nil, changeFunc)
	form.AddInputField("IP Address (for daemon mode, empty for all interfaces):", config.Address, 50, nil, changeFunc)
	form.AddInputField("Port (for daemon mode):", strconv.Itoa(int(config.Port)), 10, nil, changeFunc)
	form.AddInputField("DNS Forwarder (for non-A/AAAA requests):", config.Forwarder, 50, nil, changeFunc)
	form.AddInputField("Refresh Interval (minutes):", strconv.Itoa(int(config.RefreshInterval)), 10, nil, changeFunc)
	form.AddInputField("TTL (minutes):", strconv.Itoa(int(config.TTL)), 10, nil, changeFunc)

	// Add buttons
	form.AddButton("Save", func() {
		if !somethingChanged {
			return
		}

		// Get values from form
		dataDir := form.GetFormItem(0).(*tview.InputField).GetText()
		cacheSize, _ := strconv.Atoi(form.GetFormItem(1).(*tview.InputField).GetText())
		if 0 >= cacheSize {
			cacheSize = 1 << 10
		}
		address := form.GetFormItem(2).(*tview.InputField).GetText()
		port, _ := strconv.Atoi(form.GetFormItem(3).(*tview.InputField).GetText())
		if (0 >= port) || (65535 < port) {
			port = 53
		}
		forwarder := form.GetFormItem(4).(*tview.InputField).GetText()
		refreshInterval, _ := strconv.Atoi(form.GetFormItem(5).(*tview.InputField).GetText())
		if (0 > refreshInterval) || (255 < refreshInterval) {
			refreshInterval = 0
		}
		ttl, _ := strconv.Atoi(form.GetFormItem(6).(*tview.InputField).GetText())
		if 0 > ttl {
			ttl = 0
		} else if 255 < ttl {
			ttl = 255
		}

		// Update configuration
		newConfig := tConfiguration{
			Address:         address,
			DataDir:         dataDir,
			CacheSize:       cacheSize,
			Forwarder:       forwarder,
			Port:            port,
			RefreshInterval: uint8(refreshInterval), //#nosec G115
			TTL:             uint8(ttl),             //#nosec G115
		}

		// Save configuration
		if err := saveConfiguration(newConfig, gConfigFile); nil != err {
			aState.statusBar.SetText(fmt.Sprintf("Error saving configuration: %v", err))
			return
		}

		// TODO: Save configuration
		// This requires methods to update configuration at runtime

		aState.configChanged = true
		aState.statusBar.SetText("Configuration saved. Restart required for changes to take effect.")
	}) // Save button

	form.AddButton("Cancel", func() {
		if !somethingChanged {
			return
		}
		aState.pages.SwitchToPage("default")
	})

	// Use the box for styling
	formBox.SetBorderPadding(1, 1, 2, 2)

	// Add page
	aState.pages.AddPage("config", form, true, true)
	aState.app.SetFocus(form)

	// Update status
	aState.statusBar.SetText("Editing configuration")
} // showConfigPage()

// `showMetricsPage()` displays the metrics
//
// Parameters:
//   - `aState`: The application state.
func showMetricsPage(aState *tAppState) {
	// Create text view for metrics
	metrics := tview.NewTextView().
		SetDynamicColors(true)
	_ = metrics.Box.SetTitle(" DNS Cache Metrics ")
	_ = metrics.Box.SetBackgroundColor(colourBackground)
	_ = metrics.Box.SetTitleColor(colourHeader)
	_ = metrics.Box.SetBorder(true)

	// Get current metrics
	metrics.SetText(aState.resolver.Metrics().String())

	// Add refresh button
	refresh := tview.NewButton("Refresh").SetSelectedFunc(func() {
		// Update metrics
		metrics.SetText(aState.resolver.Metrics().String())
	})

	// Create layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(metrics, 0, 1, true).
		AddItem(refresh, 1, 0, false)

	// Add page
	aState.pages.AddPage("metrics", flex, true, true)
	aState.app.SetFocus(metrics)

	// Update status
	aState.statusBar.SetText("Viewing metrics")
} // showMetricsPage()

/* _EoF_ */
