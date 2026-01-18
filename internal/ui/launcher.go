package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func LaunchApp() {
	// 1. Create Application
	myApp := app.New()
	myWindow := myApp.NewWindow("Data Coupler")
	myWindow.Resize(fyne.NewSize(500, 400))

	// 2. UI Components
	// A. Header
	title := widget.NewLabel("Data Coupler v0.1")
	title.TextStyle = fyne.TextStyle{Bold: true}

	// B. File Selection (Visuals only for now)
	lblFile := widget.NewLabel("Input CSV:")
	entryFile := widget.NewEntry()
	entryFile.SetPlaceHolder("No file selected...")
	entryFile.Disable() // Read-only, user must use the button

	btnSelectFile := widget.NewButton("Open File...", func() {
		// Logic to come in Step 2
		entryFile.SetText("C:\\Fake\\Path\\test_data.csv")
	})

	// C. Profile Selection
	lblProfile := widget.NewLabel("Select Profile:")
	comboProfile := widget.NewSelect([]string{"Profile A", "Profile B (Mock)"}, func(value string) {
		// Logic to come in Step 2
	})
	comboProfile.PlaceHolder = "Choose a mapping..."

	// D. Action Button
	btnConvert := widget.NewButton("Run Conversion", func() {
		// Logic to come in Step 2
	})
	btnConvert.Importance = widget.HighImportance // Makes it blue/prominent

	// 3. Layout (Vertical Box)
	// We stack them one after another
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		lblFile,
		container.NewBorder(nil, nil, nil, btnSelectFile, entryFile), // Puts button to the right of entry
		lblProfile,
		comboProfile,
		widget.NewSeparator(),
		btnConvert,
	)

	// 4. Run
	myWindow.SetContent(content)
	myWindow.ShowAndRun()
}


