package main

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client     *mongo.Client
	deviceColl *mongo.Collection
	ctx        = context.Background()
)

// 👇 AAPKA ASLI MONGODB LINK YAHAN HAI 👇
const mongoURI = "mongodb+srv://shailendrameghwal2007_db_user:A9B4nju1E2ZwDfRC@cluster0.ths9xk3.mongodb.net/?retryWrites=true&w=majority"

type Device struct {
	HardwareID string    `bson:"hardware_id"`
	DeviceName string    `bson:"device_name"`
	Username   string    `bson:"username"`
	Password   string    `bson:"password"`
	Status     string    `bson:"status"`
	IsAdmin    bool      `bson:"is_admin"`
	CreatedAt  time.Time `bson:"created_at"`
}

func initDB() {
	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		fmt.Println("MongoDB Connection Error: ", err)
		return
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = client.Ping(ctxTimeout, nil)
	if err != nil {
		fmt.Println("MongoDB Ping Failed: ", err)
		return
	}

	deviceColl = client.Database("SafeEraDB").Collection("devices")
	fmt.Println("Connected to MongoDB successfully!")
}

func main() {
	// OS specific check (Mac pe ye true return karega, Windows pe UAC mangega)
	if !isAdmin() {
		if err := elevate(); err != nil {
			fmt.Println("Elevation failed:", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	initDB()

	a := app.New()
	a.Settings().SetTheme(&sebTheme{})
	w := a.NewWindow("SafeEra")
	w.Resize(fyne.NewSize(450, 450))

	if deviceColl != nil {
		checkAutoLogin(a, w)
	} else {
		w.SetContent(container.NewVBox(
			widget.NewLabelWithStyle("Database Connection Failed", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabel("Please check your internet connection."),
		))
	}

	w.ShowAndRun()
}

func checkAutoLogin(a fyne.App, w fyne.Window) {
	hwid, _ := getDeviceInfo()

	var device Device
	err := deviceColl.FindOne(ctx, bson.M{"hardware_id": hwid}).Decode(&device)

	if err == nil {
		if device.Status == "rejected" {
			w.SetContent(container.NewVBox(
				widget.NewLabelWithStyle("Access Denied", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				widget.NewLabel("This device has been rejected by Admin."),
			))
			return
		} else if device.Status == "approved" {
			showMainApp(a, w, device)
			return
		} else {
			w.SetContent(container.NewVBox(
				widget.NewLabelWithStyle("Pending Approval", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				widget.NewLabel("This device is waiting for Admin approval."),
			))
			return
		}
	}

	showRegisterScreen(a, w)
}

func showRegisterScreen(a fyne.App, w fyne.Window) {
	hwid, realName := getDeviceInfo()
	
	w.SetTitle("SafeEra - Register Device")
	w.Resize(fyne.NewSize(450, 550)) 

	nameEntry := widget.NewEntry()
	nameEntry.SetText(realName)
	nameEntry.SetPlaceHolder("Device Name")

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("Choose Username")

	passEntry := widget.NewPasswordEntry()
	passEntry.SetPlaceHolder("Choose Password")

	registerBtn := widget.NewButtonWithIcon("Register", theme.ContentAddIcon(), func() {
		if nameEntry.Text == "" || userEntry.Text == "" || passEntry.Text == "" {
			dialog.ShowError(fmt.Errorf("Please fill all fields"), w)
			return
		}

		count, _ := deviceColl.CountDocuments(ctx, bson.M{"hardware_id": hwid})
		if count > 0 {
			dialog.ShowError(fmt.Errorf("This device is already registered!"), w)
			return
		}

		device := Device{
			HardwareID: hwid,
			DeviceName: nameEntry.Text,
			Username:   userEntry.Text,
			Password:   passEntry.Text,
			Status:     "pending",
			IsAdmin:    false,
			CreatedAt:  time.Now(),
		}

		deviceColl.InsertOne(ctx, device)

		dialog.ShowInformation("Registered", "Device registered!\nAdmin approve karega to access milega.", w)
		checkAutoLogin(a, w)
	})
	registerBtn.Importance = widget.HighImportance

	hwidShort := hwid
	if len(hwid) > 20 {
		hwidShort = hwid[:20] + "..."
	}

	content := container.NewVBox(
		widget.NewLabelWithStyle("Register New Device", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Device Name (Real):"),
		nameEntry,
		widget.NewLabel("Username:"),
		userEntry,
		widget.NewLabel("Password:"),
		passEntry,
		widget.NewLabel("Device ID (Hardware):"),
		widget.NewLabel(hwidShort), 
		widget.NewSeparator(),
		registerBtn,
	)

	w.SetContent(container.NewPadded(container.NewVScroll(content)))
}

func showMainApp(a fyne.App, w fyne.Window, device Device) {
	w.SetTitle("SafeEra - " + device.DeviceName)
	w.Resize(fyne.NewSize(480, 740))

	preview := widget.NewMultiLineEntry()
	preview.SetPlaceHolder("Clipboard content...")
	preview.Wrapping = fyne.TextWrapWord

	refreshBtn := widget.NewButtonWithIcon("Refresh Clipboard", theme.ViewRefreshIcon(), func() {
		text, err := readClipboard()
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		preview.SetText(text)
	})

	if text, err := readClipboard(); err == nil {
		preview.SetText(text)
	}

	delaySlider := widget.NewSlider(0, 10)
	delaySlider.Value = 3
	delayLabel := widget.NewLabel("Start delay: 3 seconds")
	delaySlider.OnChanged = func(v float64) {
		delayLabel.SetText(fmt.Sprintf("Start delay: %.0f seconds", v))
	}

	charDelaySlider := widget.NewSlider(1, 100)
	charDelaySlider.Value = 10
	charDelayLabel := widget.NewLabel("Char delay: 10 ms")
	charDelaySlider.OnChanged = func(v float64) {
		charDelayLabel.SetText(fmt.Sprintf("Char delay: %.0f ms", v))
	}

	sebOnlyCheck := widget.NewCheck("Type only in SEB window", nil)
	sebOnlyCheck.Checked = true

	statusLabel := widget.NewLabel("Ready.")
	statusLabel.Wrapping = fyne.TextWrapWord

	injectBtn := widget.NewButtonWithIcon("Start Typing", theme.MediaPlayIcon(), func() {
		text := preview.Text
		if text == "" {
			dialog.ShowInformation("Empty", "Clipboard is empty.", w)
			return
		}
		startDelay := int(delaySlider.Value)
		charDelay := int(charDelaySlider.Value)
		onlySEB := sebOnlyCheck.Checked
		statusLabel.SetText(fmt.Sprintf("Starting in %d seconds... Click on target window!", startDelay))
		go func() {
			err := injectText(text, startDelay, charDelay, onlySEB, func(msg string) {
				statusLabel.SetText(msg)
			})
			if err != nil {
				statusLabel.SetText("Error: " + err.Error())
			} else {
				statusLabel.SetText("Success!")
			}
		}()
	})
	injectBtn.Importance = widget.HighImportance

	switchBtn := widget.NewButtonWithIcon("Switch Window (Minimize SEB)", theme.NavigateNextIcon(), func() {
		hwnd := findSEBWindow()
		if hwnd == 0 {
			dialog.ShowInformation("Not Found", "SEB window not found.", w)
			return
		}
		showWindow(hwnd, SW_MINIMIZE)
		statusLabel.SetText("SEB minimized.")
	})

	restoreBtn := widget.NewButtonWithIcon("Restore SEB", theme.ViewRestoreIcon(), func() {
		hwnd := findSEBWindow()
		if hwnd != 0 {
			showWindow(hwnd, SW_RESTORE)
			setForeground(hwnd)
			statusLabel.SetText("SEB restored.")
		}
	})

	line := widget.NewSeparator()

	content := container.NewVBox(
		widget.NewLabelWithStyle(device.DeviceName, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		line,
		widget.NewLabelWithStyle("Paste Text:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		preview,
		refreshBtn,
		line,
		widget.NewLabelWithStyle("Typing Configuration:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		delayLabel,
		delaySlider,
		charDelayLabel,
		charDelaySlider,
		sebOnlyCheck,
		injectBtn,
		statusLabel,
		line,
		widget.NewLabelWithStyle("SEB Window Tools:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		switchBtn,
		restoreBtn,
	)

	w.SetContent(container.NewPadded(container.NewVScroll(content)))
}

type sebTheme struct{}

func (m *sebTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameBackground {
		return color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	}
	if name == theme.ColorNameButton {
		return color.NRGBA{R: 240, G: 240, B: 240, A: 255}
	}
	if name == theme.ColorNamePrimary {
		return color.NRGBA{R: 0, G: 90, B: 158, A: 255}
	}
	if name == theme.ColorNameError {
		return color.NRGBA{R: 211, G: 47, B: 47, A: 255}
	}
	return theme.DefaultTheme().Color(name, theme.VariantLight)
}

func (m *sebTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m *sebTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *sebTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return 12
	}
	return theme.DefaultTheme().Size(name)
}