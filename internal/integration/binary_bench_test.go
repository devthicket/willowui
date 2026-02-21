package integration

import (
	"testing"

	ui "github.com/devthicket/willowui"
)

// complexXML is a realistic, deeply nested template with bindings, events,
// directives, and many component types to stress the parser/decoder.
const complexXML = `
<Panel layout="vbox" spacing="8">
	<Label bind:text="appTitle" />
	<Panel layout="hbox" spacing="12" size="600,40">
		<Button text="Home" on:click="goHome" />
		<Button text="Settings" on:click="goSettings" />
		<Button text="Profile" on:click="goProfile" />
		<Button bind:text="navLabel" on:click="toggleNav" />
	</Panel>
	<Panel layout="vbox" spacing="4" size="600,400" background="#1a1a2eff">
		<Label text="Dashboard" />
		<Panel layout="hbox" spacing="8">
			<Panel layout="vbox" spacing="4" size="280,300" background="#222244ff">
				<Label text="Stats" />
				<ProgressBar size="260,20" />
				<Label bind:text="stat1" />
				<Label bind:text="stat2" />
				<Label bind:text="stat3" />
				<Slider size="260,24" />
				<Panel ui:show="showDetails" layout="vbox" spacing="4">
					<Label text="Detail A" />
					<Label text="Detail B" />
					<Label bind:text="detailC" />
					<Panel layout="hbox" spacing="4">
						<Button text="Action 1" on:click="action1" />
						<Button text="Action 2" on:click="action2" />
					</Panel>
				</Panel>
				<Toggle />
				<Checkbox text="Enable feature" />
				<Checkbox text="Auto-save" />
			</Panel>
			<Panel layout="vbox" spacing="4" size="280,300" background="#223344ff">
				<Label text="Items" />
				<Panel layout="vbox" spacing="2">
					<Label text="Item 1 - Sword of Testing" />
					<Label text="Item 2 - Shield of CI" />
					<Label text="Item 3 - Helm of Coverage" />
					<Label text="Item 4 - Boots of Speed" />
					<Label text="Item 5 - Ring of Power" />
					<Label text="Item 6 - Amulet of Wisdom" />
					<Label text="Item 7 - Cloak of Shadows" />
					<Label text="Item 8 - Gauntlets of Might" />
				</Panel>
				<Panel ui:if="hasInventory" layout="vbox" spacing="4">
					<Label text="Inventory" />
				</Panel>
				<Panel layout="hbox" spacing="4">
					<Button text="Sort" on:click="sortItems" />
					<Button text="Filter" on:click="filterItems" />
					<Button text="Clear" on:click="clearItems" />
				</Panel>
			</Panel>
		</Panel>
		<Panel layout="hbox" spacing="8">
			<Button text="Save" on:click="save" />
			<Button text="Load" on:click="load" />
			<Button bind:text="statusLabel" />
			<Label bind:text="footerText" />
		</Panel>
	</Panel>
	<Panel layout="hbox" spacing="16" size="600,30">
		<Label text="v1.0.0" />
		<Label bind:text="connectionStatus" />
		<Label bind:text="fpsCounter" />
	</Panel>
</Panel>
`

var (
	benchXMLBytes []byte
	benchBinBytes []byte
)

func init() {
	benchXMLBytes = []byte(complexXML)
	ir, err := ui.CompileXML(benchXMLBytes)
	if err != nil {
		panic("bench init: " + err.Error())
	}
	benchBinBytes, err = ui.EncodeIR(ir)
	if err != nil {
		panic("bench init: " + err.Error())
	}
}

func BenchmarkCompileXML(b *testing.B) {
	for b.Loop() {
		_, err := ui.CompileXML(benchXMLBytes)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeIR(b *testing.B) {
	for b.Loop() {
		_, err := ui.DecodeIR(benchBinBytes)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncodeIR(b *testing.B) {
	ir, _ := ui.CompileXML(benchXMLBytes)
	b.ResetTimer()
	for b.Loop() {
		_, err := ui.EncodeIR(ir)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRegisterXML(b *testing.B) {
	for b.Loop() {
		reg := ui.NewTemplateRegistry()
		if err := reg.RegisterXML("bench", benchXMLBytes); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRegisterBinary(b *testing.B) {
	for b.Loop() {
		reg := ui.NewTemplateRegistry()
		if err := reg.RegisterBinary("bench", benchBinBytes); err != nil {
			b.Fatal(err)
		}
	}
}
