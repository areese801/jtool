package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create the application menu
	appMenu := createAppMenu(app)

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "jtool",
		Width:     1024,
		Height:    768,
		MinWidth:  1024,
		MinHeight: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Menu:             appMenu,
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title:   "jtool",
				Message: "A JSON diff and analysis tool",
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// createAppMenu builds the native application menu bar.
// This works cross-platform: on macOS it appears in the system menu bar,
// on Windows/Linux it appears in the application window.
//
// Python comparison:
//   - macOS menus in Python would use PyObjC or a framework like rumps
//   - Wails provides a cross-platform menu API that maps to native menus
//   - The menu.Menu type is like a tree structure of menu items
func createAppMenu(app *App) *menu.Menu {
	appMenu := menu.NewMenu()

	// App menu (jtool menu with Quit, etc.)
	appMenu.Append(menu.AppMenu())

	// Edit menu (with standard copy/paste)
	appMenu.Append(menu.EditMenu())

	// Help menu
	helpMenu := appMenu.AddSubmenu("Help")
	helpMenu.AddText("Settings...", nil, func(_ *menu.CallbackData) {
		app.ShowSettingsTab()
	})
	helpMenu.AddSeparator()
	helpMenu.AddText("Report a Bug...", nil, func(_ *menu.CallbackData) {
		// Open GitHub Issues with bug label pre-set
		runtime.BrowserOpenURL(app.ctx, "https://github.com/areese801/jtool/issues/new?labels=bug")
	})

	return appMenu
}
