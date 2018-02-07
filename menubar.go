package main

import "github.com/murlokswarm/app"

// MenuBar is the component that define the menu bar.
type MenuBar struct{}

// Render returns return the HTML describing the menu bar.
func (m *MenuBar) Render() string {
	return `
<menu>
	<menu label="app">
		<menuitem label="Close" selector="performClose:" shortcut="meta+w" />
		<menuitem label="Quit" selector="terminate:" shortcut="meta+q" /> 
	</menu>
	<menu label="Play">
		<menuitem label="Toggle pause/play" onclick="TogglePause" shortcut="meta+p" />
		<menuitem label="Next" onclick="Next" shortcut="meta+n" /> 
		<menuitem label="Back" onclick="Back" shortcut="meta+b" /> 
	</menu>
</menu>
	`
}

func (m *MenuBar) Next() {
	playlist.Next()
}

func (m *MenuBar) Back() {
	playlist.Back()
}

func (m *MenuBar) TogglePause() {
	playlist.TogglePause()
}

// /!\ Register the component. Required to use the component into a context.
func init() {
	app.RegisterComponent(&MenuBar{})
}
