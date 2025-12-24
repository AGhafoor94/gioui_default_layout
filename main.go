package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/xuri/excelize/v2"
)

const voodoo_api_key string = ""
const graph_base_uri string = "https://graph.microsoft.com/v1.0/"
const client_id string = ""
const client_secret string = ""

const side_bar_width_default float32 = 0.175

const (
	white        = iota
	blue         = iota
	dark_blue    = iota
	green        = iota
	dark_green   = iota
	yellow       = iota
	orange       = iota
	dark_red     = iota
	lighter_grey = iota
	light_grey   = iota
	black        = iota
)

var colours_list []color.NRGBA = []color.NRGBA{
	{R: 255, G: 255, B: 255, A: 255}, // White 			0
	{R: 0, G: 94, B: 184, A: 255},    // Blue			1
	{R: 0, G: 48, B: 135, A: 255},    // Dark Blue		2
	{R: 0, G: 150, B: 57, A: 255},    // Green			3
	{R: 0, G: 103, B: 71, A: 255},    // Dark Green		4
	{R: 255, G: 184, B: 28, A: 255},  // Warm Yellow	5
	{R: 237, G: 139, B: 0, A: 255},   // Orange			6
	{R: 138, G: 21, B: 56, A: 255},   // Dark Red		7
	{R: 240, G: 240, B: 240, A: 255}, // Lighter Grey	9
	{R: 232, G: 232, B: 232, A: 255}, // Light Grey		8
	{R: 0, G: 0, B: 0, A: 255},       // Black		   10
}
var month_data []month_data_struct = []month_data_struct{
	{
		name:            "April",
		month:           4,
		financial_month: "01",
	},
	{
		name:            "May",
		month:           5,
		financial_month: "02",
	},
	{
		name:            "June",
		month:           6,
		financial_month: "03",
	},
	{
		name:            "July",
		month:           7,
		financial_month: "04",
	},
	{
		name:            "August",
		month:           8,
		financial_month: "05",
	},
	{
		name:            "September",
		month:           9,
		financial_month: "06",
	},
	{
		name:            "October",
		month:           10,
		financial_month: "07",
	},
	{
		name:            "November",
		month:           11,
		financial_month: "08",
	},
	{
		name:            "December",
		month:           12,
		financial_month: "09",
	},
	{
		name:            "January",
		month:           1,
		financial_month: "10",
	},
	{
		name:            "February",
		month:           2,
		financial_month: "11",
	},
	{
		name:            "March",
		month:           3,
		financial_month: "12",
	},
}
var footer_image_point image.Point
var excel_columns_alphabet []string = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA", "AB", "AC", "AD", "AE", "AF", "AG", "AH", "AI", "AJ", "AK", "AL", "AM", "AN", "AO", "AP", "AQ", "AR", "AS", "AT", "AU", "AV", "AW", "AX", "AY", "AZ"}

type main_window struct {
	window         *app.Window
	heading        string
	heading_colour color.NRGBA
	string_error   string
	sub_heading    string
}
type main_data_struct struct {
	boxi_file                *excelize.File
	boxi_excel_file_location string
	boxi_excel_file_sheets   []string

	cash_allocation_file                *excelize.File
	cash_allocation_excel_file_location string
	cash_allocation_excel_file_sheets   []string

	mapping_table_file                *excelize.File
	mapping_table_excel_file_location string
	mapping_table_excel_file_sheets   []string

	bank_reconciliation_file                *excelize.File
	bank_reconciliation_excel_file_location string
	bank_reconciliation_excel_file_sheets   []string

	app_selection int8
	trusts        []trust_data_struct
	trust         string

	text_input_array []text_inputs_array_struct
	completed        bool
}
type text_inputs_array_struct struct {
	selection     int8
	text_inputted string
}
type mapped_data_object []map[string]interface{}
type month_data_struct struct {
	name            string
	month           int8
	financial_month string
}
type dropdown_struct struct {
	list           widget.List
	items          []trust_data_struct
	selected       string
	trust_selected trust_data_struct
}
type boxi_data_struct struct {
	header             []string
	data               [][]string
	header_start_index int8
	filtered_data      [][]string
	// start_index        int8
}
type boxi_main_data_struct struct {
	number_of_banks     int8
	header_start_column string
	balances_data       boxi_data_struct
	movements           boxi_data_struct
}
type cash_allocation_data_struct struct {
	opening_balance       string
	closing_balance       string
	sub_ledger_heading    []string
	sub_ledger_one_data   [][]string
	sub_ledger_one_code   string
	sub_ledger_two_data   [][]string
	sub_ledger_two_code   string
	journal_heading       []string
	journal_income        [][]string
	journal_payments      [][]string
	journal_code          string
	payments_heading      []string
	payments              [][]string
	other_credits_heading []string
	other_credits         [][]string
	other_data            [][]string
}
type mapping_table_data_struct struct {
	cash_allocation_header []string
	boxi_report_header     []string
	cash_allocation_data   [][]string
	boxi_report_data       [][]string
}
type bank_rec_calculation struct {
	sheet              string
	account_code       string
	difference         float64
	balance_slash_year string
	original_amounts   [][]string
	new_amounts        [][]string
	calculation        [][]string
	balance_list       [][]string
	balance_indexes    []int
}
type bank_rec_calculation_array struct {
	data []bank_rec_calculation
}
type data_allocated_struct struct {
	sheet              string   // Sheet to add data to
	location           string   // Cell location e.g. A4
	cell               string   // e.g. C
	found_in_data      string   // Data it found it in e.g. mapping table
	type_of_data       string   // cash allocation type e.g. journal income
	amount             float64  // amount, can be converted back to string FromFloat
	data               []string // original data
	additional_data    string
	length_of_one_item int16
	id                 int
}
type alerts_errors_not_allocated_struct struct {
	original_location string   // Data it found it in e.g. mapping table
	type_of_data      string   // cash allocation type e.g. journal income
	amount            float64  // amount, can be converted back to string FromFloat
	data              []string // original data
	additional_data   string
}
type settings_struct struct {
	StartMenu   int16               `json:"startMenu"`
	Trusts      []trust_data_struct `json:"trusts"`
	UserDetails user_details        `json:"userDetails"`
}
type trust_data_struct struct {
	Name  string `json:"name"`
	Trust string `json:"trust"`
	Id    string `json:"id"`
}
type user_details struct {
	Username string `json:"username"`
	UserCode string `json:"userCode"`
}

func main() {
	// BUILD: go build -ldflags -H=windowsgui
	go func() {
		window := new(app.Window)
		window.Option(app.Title("Technical Accounts Apps"))

		main_window := main_window{window: window}

		main_window.main_layout()

	}()
	app.Main()
}

func (window_state *main_window) main_layout() {
	var ops op.Ops
	theme := material.NewTheme()
	// var editor widget.Editor = widget.Editor{
	// 	SingleLine: false,
	// 	Submit:     true,
	// 	ReadOnly:   false,
	// }
	var send_to widget.Editor = widget.Editor{
		SingleLine: true,
		Submit:     true,
		ReadOnly:   false,
	}
	var send_from widget.Editor = widget.Editor{
		SingleLine: true,
		Submit:     true,
		ReadOnly:   false,
	}
	var message_body widget.Editor = widget.Editor{
		SingleLine: false,
		Submit:     false,
		ReadOnly:   false,
	}

	var (
		menu_button widget.Clickable

		show_control_accounts widget.Clickable
		show_bank_rec         widget.Clickable
		show_sms_page         widget.Clickable
		show_email_page       widget.Clickable
		show_boxi_page        widget.Clickable

		open_select_button                            widget.Clickable
		open_file_dialogue_boxi_widget                widget.Clickable
		open_file_dialogue_cash_allocation_widget     widget.Clickable
		open_file_dialogue_mapping_table_widget       widget.Clickable
		open_file_dialogue_bank_reconciliation_widget widget.Clickable
		open_file_dialogue_boxi                       widget.Clickable
		open_file_dialogue_month                      widget.Clickable
		start_button                                  widget.Clickable
		end_button                                    widget.Clickable

		submit_button widget.Clickable
	)
	side_button_text := []string{
		"Control Accounts",
		"Bank Reconciliation",
		"SMS",
		"Email",
		"BOXI",
	}
	side_buttons := []widget.Clickable{
		show_control_accounts,
		show_bank_rec,
		show_sms_page,
		show_email_page,
		show_boxi_page,
	}
	// right_side_buttons := []widget.Clickable{
	// 	right_button_one, right_button_two, right_button_three,
	// }
	control_accounts_widgets := []widget.Clickable{
		open_select_button, open_file_dialogue_boxi, open_file_dialogue_month,
	}
	bank_rec_widgets := []widget.Clickable{
		open_select_button, open_file_dialogue_boxi_widget, open_file_dialogue_cash_allocation_widget, open_file_dialogue_mapping_table_widget, open_file_dialogue_bank_reconciliation_widget, submit_button,
	}

	sms_widget_text_inputs := []widget.Editor{
		send_from, send_to, message_body,
	}
	email_widget_text_inputs := []widget.Editor{
		send_from, send_to, message_body,
	}
	show_options := false
	var side_bar_width float32 = side_bar_width_default
	var right_content_width float32 = (1.0 - side_bar_width)
	show_side_bar := true
	show_selection := 1
	window_state.heading = "Control Accounts"
	window_state.heading_colour = colours_list[blue]
	main_data_state := main_data_struct{}
	log_file, err := os.Create("./log-file-tech-accounts.txt")
	if err != nil {
		window_state.string_error = fmt.Sprintf("ERROR: %s", err.Error())
	}
	defer log_file.Close()
	settings_file, err := os.Open("settings.json")
	if err != nil {
		log_file.WriteString(fmt.Sprintf("[%s]: Unable to Load / Create Log File \n", time.Now().Local().Format("02/01/2006 15:04")))
		window_state.string_error = fmt.Sprintf("ERROR: %s. Updated Log file.", err.Error())
	}
	defer settings_file.Close()
	byte_value, _ := io.ReadAll(settings_file)
	var settings_data settings_struct
	json.Unmarshal(byte_value, &settings_data)
	// fmt.Println(settings_data)
	main_data_state.trust = "Select Trust"
	window_state.sub_heading = "Created by Adnan Ghafoor"
	main_data_state.completed = true
	user_home_directory, _ := os.UserHomeDir()
	// fmt.Println("USER HOME DIRECTORY: ", user_home_directory, " ", strings.Split(user_home_directory, "\\")[len(strings.Split(user_home_directory, "\\"))-1])
	settings_data.UserDetails.Username = strings.Split(user_home_directory, "\\")[len(strings.Split(user_home_directory, "\\"))-1]
	main_data_state.trusts = settings_data.Trusts
	main_data_state.app_selection = int8(settings_data.StartMenu)
	show_selection = int(main_data_state.app_selection)
	fmt.Println("SHOW SELECTION: ", show_selection, " MAIN DATA STATE APP SELECTION: ", main_data_state.app_selection)
	window_state.heading = side_button_text[show_selection-1]
	dropdown_menu_enum := widget.Enum{}
	dropdown := dropdown_struct{
		selected: "Load Excel File",
	}
	dropdown.selected = "Select Item"
	dropdown.items = settings_data.Trusts
	dropdown.list = widget.List{List: layout.List{Axis: layout.Vertical}}
	for {
		switch event := window_state.window.Event().(type) {
		case app.FrameEvent:
			// explorer_window.ListenEvents(event)
			graphical_context := app.NewContext(&ops, event)
			if menu_button.Clicked(graphical_context) {
				show_side_bar = !show_side_bar
				if side_bar_width == 0.0 {
					side_bar_width = side_bar_width_default
					right_content_width = (1.0 - side_bar_width)
				} else {
					side_bar_width = 0.0
					right_content_width = (1.0 - side_bar_width)
				}
			}
			// if side_buttons[0].Clicked(graphical_context) {
			// 	fmt.Println("Cash Allocation Page was selected")
			// 	show_selection = 1
			// 	window_state.heading_colour = colours_list[1]
			// 	window_state.heading = side_button_text[0]
			// }l
			// if main_data_state.completed {
			// 	window_state.sub_heading = "Completed"
			// }
			for i := range side_buttons {
				if side_buttons[i].Clicked(graphical_context) {
					show_selection = i + 1
					window_state.heading_colour = colours_list[i+1]
					window_state.heading = side_button_text[i]
					// Empty struct data on screen change
					main_data_state = main_data_struct{}
					main_data_state.trusts = settings_data.Trusts
					main_data_state.trust = "Select Trust"
				}
			}
			// if right_side_buttons[1].Clicked(graphical_context) {
			// 	fmt.Println("TEST CLICK")
			// }

			if start_button.Clicked(graphical_context) {
				switch show_selection {
				case 1:
					if len(main_data_state.boxi_excel_file_sheets) > 0 && len(main_data_state.cash_allocation_excel_file_sheets) > 0 {
						main_data_state.generate_data_control_accounts(log_file)
						main_data_state.completed = true
						window_state.sub_heading = "Completed"
					} else {
						set_background_rect_colour(graphical_context, footer_image_point, colours_list[dark_red])
						window_state.sub_heading = "Please Upload the Required Files"
						main_data_state.completed = true
					}

				case 2:
					if len(main_data_state.boxi_excel_file_sheets) > 0 && len(main_data_state.cash_allocation_excel_file_sheets) > 0 && len(main_data_state.mapping_table_excel_file_sheets) > 0 && len(main_data_state.bank_reconciliation_excel_file_sheets) > 0 {
						main_data_state.generate_data_bank_rec(log_file)
						main_data_state.completed = true
						window_state.sub_heading = "Completed"
					} else {
						set_background_rect_colour(graphical_context, footer_image_point, colours_list[dark_red])
						window_state.sub_heading = "Please Upload the Required Files"
						main_data_state.completed = true
					}

				case 3:
					send_sms_length := 0
					for i := range sms_widget_text_inputs {
						main_data_state.text_input_array = append(main_data_state.text_input_array, text_inputs_array_struct{
							selection:     int8(show_selection),
							text_inputted: sms_widget_text_inputs[i].Text(),
						})
					}
					for _, text := range filter_single_array(main_data_state.text_input_array, func(item text_inputs_array_struct) bool {
						return item.selection == 3
					}) {
						if len(strings.TrimSpace(text.text_inputted)) > 0 {
							send_sms_length += 1
						}
					}
					fmt.Println(main_data_state.text_input_array)
					if send_sms_length == 3 {
						status_code, response := main_data_state.send_sms_api_call()
						fmt.Println(status_code, " ", response)
						if strings.Contains(status_code, "20") {
							main_data_state.completed = true
							window_state.sub_heading = fmt.Sprintf("Credits used: %d, Credits Remaining: %d", response.Credits, response.Balance)
						}
					} else {
						set_background_rect_colour(graphical_context, footer_image_point, colours_list[dark_red])
						window_state.sub_heading = "Please Enter Text Above"
						main_data_state.completed = true
					}
				case 4:
					send_email_length := 0
					for i := range email_widget_text_inputs {
						main_data_state.text_input_array = append(main_data_state.text_input_array, text_inputs_array_struct{
							selection:     int8(show_selection),
							text_inputted: email_widget_text_inputs[i].Text(),
						})
					}
					for _, text := range filter_single_array(main_data_state.text_input_array, func(item text_inputs_array_struct) bool {
						return item.selection == 4
					}) {
						if len(strings.TrimSpace(text.text_inputted)) > 0 {
							send_email_length += 1
						}
					}
					fmt.Println(main_data_state.text_input_array)
					if send_email_length == 3 {
						access_token := main_data_state.get_access_token_graph_api_call()
						status := main_data_state.send_email_office_api_call(access_token)
						if strings.Contains(status, "20") {
							window_state.sub_heading = "Sent Email"
							main_data_state.completed = true
						}
					} else {
						set_background_rect_colour(graphical_context, footer_image_point, colours_list[dark_red])
						window_state.sub_heading = "Please Enter Text Above"
						main_data_state.completed = true
					}
				}
			}
			if end_button.Clicked(graphical_context) {
				settings_data.StartMenu = int16(show_selection)
				settings_json_marshaled, _ := json.Marshal(settings_data)

				os.WriteFile("settings.json", settings_json_marshaled, fs.ModeExclusive)
				os.Exit(0)
			}

			// if bank_rec_widgets[len(bank_rec_widgets)-1].Clicked(graphical_context) {
			// 	if !strings.EqualFold(editor.Text(), "") {
			// 		main_data_state.text_input_array = append(main_data_state.text_input_array, text_inputs_array_struct{
			// 			selection:     int8(show_selection),
			// 			text_inputted: editor.Text(),
			// 		})
			// 	}
			// }
			if control_accounts_widgets[0].Clicked(graphical_context) || bank_rec_widgets[0].Clicked(graphical_context) {
				show_options = !show_options
			}
			if bank_rec_widgets[1].Clicked(graphical_context) {
				cmd := exec.Command("powershell", "-Command", `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.OpenFileDialog; $f.ShowDialog() | Out-Null; $f.FileName`)

				output, err := cmd.Output()

				if err != nil {
					// main_data_state.information = "Please select xlsx file"
					log_file.WriteString(fmt.Sprintf("[%s]: User didn't select Excel file.\n", time.Now().Local().Format("02/01/2006 15:04")))
				}
				// main_data_state.information = "Loading BOXI Data"
				main_data_state.boxi_excel_file_location = strings.TrimSpace(string(output))
				go func() {
					boxi_report, err := excelize.OpenFile(main_data_state.boxi_excel_file_location)
					if err != nil {
						log_file.WriteString(fmt.Sprintf("[%s]: BOXI File Failed: %s \n", time.Now().Local().Format("02/01/2006 15:04"), main_data_state.boxi_excel_file_location))
						// boxi_report, _ = excelize.OpenFile(main_data_state.boxi_excel_file_location)
					}
					main_data_state.boxi_file = boxi_report
					main_data_state.boxi_excel_file_sheets = boxi_report.GetSheetList()
					fmt.Println("139:- Sheets in this file: ", main_data_state.boxi_excel_file_sheets)
					log_file.WriteString(fmt.Sprintf("[%s]: BOXI File Added: %s \n", time.Now().Local().Format("02/01/2006 15:04"), main_data_state.boxi_excel_file_location))
					// main_data_state.information = "Loaded BOXI Report"
				}()
			}
			if bank_rec_widgets[2].Clicked(graphical_context) {
				cmd := exec.Command("powershell", "-Command", `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.OpenFileDialog; $f.ShowDialog() | Out-Null; $f.FileName`)
				output, err := cmd.Output()

				if err != nil {
					// main_data_state.information = "Please select xlsx file"
					log_file.WriteString(fmt.Sprintf("[%s]: User didn't select Excel file.\n", time.Now().Local().Format("02/01/2006 15:04")))
				}
				// main_data_state.information = "Loading Cash Allocation Data"
				main_data_state.cash_allocation_excel_file_location = strings.TrimSpace(string(output))
				go func() {
					cash_allocation, err := excelize.OpenFile(main_data_state.cash_allocation_excel_file_location)
					if err != nil {
						log_file.WriteString(fmt.Sprintf("[%s]: Cash Allocation File Failed: %s. Error: %s \n", time.Now().Local().Format("02/01/2006 15:04"), main_data_state.cash_allocation_excel_file_location, err))
						cash_allocation, _ = excelize.OpenFile(main_data_state.cash_allocation_excel_file_location)
					}
					main_data_state.cash_allocation_file = cash_allocation
					main_data_state.cash_allocation_excel_file_sheets = cash_allocation.GetSheetList()
					fmt.Println("Sheets in this file: ", main_data_state.cash_allocation_excel_file_sheets)

					log_file.WriteString(fmt.Sprintf("[%s]: Cash Allocation File Added: %s \n", time.Now().Local().Format("02/01/2006 15:04"), main_data_state.cash_allocation_excel_file_location))
					// main_data_state.information = "Loaded Cash Allocation"
				}()
			}
			if bank_rec_widgets[3].Clicked(graphical_context) {
				cmd := exec.Command("powershell", "-Command", `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.OpenFileDialog; $f.ShowDialog() | Out-Null; $f.FileName`)
				output, err := cmd.Output()

				if err != nil {
					// main_data_state.information = "Please select xlsx file"
					log_file.WriteString(fmt.Sprintf("[%s]: User didn't select Excel file.\n", time.Now().Local().Format("02/01/2006 15:04")))
				}
				// main_data_state.information = "Loading Mapping Table Data"
				main_data_state.mapping_table_excel_file_location = strings.TrimSpace(string(output))
				go func() {
					mapping_table, err := excelize.OpenFile(main_data_state.mapping_table_excel_file_location)
					if err != nil {
						log_file.WriteString(fmt.Sprintf("[%s]: Mapping Table File Failed: %s. Error: %s \n", time.Now().Local().Format("02/01/2006 15:04"), main_data_state.mapping_table_excel_file_location, err))
						mapping_table, _ = excelize.OpenFile(main_data_state.mapping_table_excel_file_location)
					}
					main_data_state.mapping_table_file = mapping_table
					main_data_state.mapping_table_excel_file_sheets = mapping_table.GetSheetList()
					fmt.Println("Sheets in this file: ", main_data_state.mapping_table_excel_file_sheets)

					log_file.WriteString(fmt.Sprintf("[%s]: Mapping Table File Added: %s \n", time.Now().Local().Format("02/01/2006 15:04"), main_data_state.mapping_table_excel_file_location))
					// main_data_state.information = "Loaded Mapping Table"
				}()
			}
			if bank_rec_widgets[4].Clicked(graphical_context) {
				cmd := exec.Command("powershell", "-Command", `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.OpenFileDialog; $f.ShowDialog() | Out-Null; $f.FileName`)
				output, err := cmd.Output()

				if err != nil {
					// main_data_state.information = "Please select xlsx file"
					log_file.WriteString(fmt.Sprintf("[%s]: User didn't select Excel file.\n", time.Now().Local().Format("02/01/2006 15:04")))
				}
				// main_data_state.information = "Loading Bank Rec Data"
				main_data_state.bank_reconciliation_excel_file_location = strings.TrimSpace(string(output))
				go func() {
					bank_reconciliation, err := excelize.OpenFile(main_data_state.bank_reconciliation_excel_file_location)
					if err != nil {
						log_file.WriteString(fmt.Sprintf("[%s]: Bank Rec File Failed: %s. Error: %s \n", time.Now().Local().Format("02/01/2006 15:04"), main_data_state.bank_reconciliation_excel_file_location, err))
						bank_reconciliation, _ = excelize.OpenFile(main_data_state.bank_reconciliation_excel_file_location)
					}
					main_data_state.bank_reconciliation_file = bank_reconciliation
					main_data_state.bank_reconciliation_excel_file_sheets = bank_reconciliation.GetSheetList()
					fmt.Println("Sheets in this file: ", main_data_state.bank_reconciliation_excel_file_sheets)

					log_file.WriteString(fmt.Sprintf("[%s]: Bank Rec File Added: %s \n", time.Now().Local().Format("02/01/2006 15:04"), main_data_state.bank_reconciliation_excel_file_location))
					// main_data_state.information = "Loaded Bank Rec File"
				}()
			}
			if control_accounts_widgets[1].Clicked(graphical_context) {
				cmd := exec.Command("powershell", "-Command", `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.OpenFileDialog; $f.ShowDialog() | Out-Null; $f.FileName`)

				output, err := cmd.Output()

				if err != nil {
					// main_data_state.information = "Please select xlsx file"
					log_file.WriteString(fmt.Sprintf("[%s]: User didn't select Excel file.\n", time.Now().Local().Format("02/01/2006 15:04")))
				}
				// main_data_state.information = "Loading BOXI Data"
				main_data_state.boxi_excel_file_location = strings.TrimSpace(string(output))
				go func() {
					boxi_report, err := excelize.OpenFile(main_data_state.boxi_excel_file_location)
					if err != nil {
						log_file.WriteString(fmt.Sprintf("[%s]: BOXI File Failed: %s \n", time.Now().Local().Format("02/01/2006 15:04"), main_data_state.boxi_file))
						// boxi_report, _ = excelize.OpenFile(main_data_state.boxi_file)
					}
					main_data_state.boxi_file = boxi_report
					main_data_state.boxi_excel_file_sheets = boxi_report.GetSheetList()
					fmt.Println("139:- Sheets in this file: ", main_data_state.boxi_excel_file_sheets)
					log_file.WriteString(fmt.Sprintf("[%s]: BOXI File Added: %s \n", time.Now().Local().Format("02/01/2006 15:04"), main_data_state.boxi_file))
					// main_data_state.information = "Loaded BOXI Report"
				}()
			}
			if control_accounts_widgets[2].Clicked(graphical_context) {
				cmd := exec.Command("powershell", "-Command", `Add-Type -AssemblyName System.Windows.Forms; $f = New-Object System.Windows.Forms.OpenFileDialog; $f.ShowDialog() | Out-Null; $f.FileName`)
				output, err := cmd.Output()

				if err != nil {
					// main_data_state.information = "Please select xlsx file"
					log_file.WriteString(fmt.Sprintf("[%s]: User didn't select Excel file.\n", time.Now().Local().Format("02/01/2006 15:04")))
				}
				// main_data_state.information = "Loading Month Data"
				main_data_state.cash_allocation_excel_file_location = strings.TrimSpace(string(output))
				go func() {
					month_file, err := excelize.OpenFile(main_data_state.cash_allocation_excel_file_location)
					if err != nil {
						log_file.WriteString(fmt.Sprintf("[%s]: Month File Failed: %s. Error: %s \n", time.Now().Local().Format("02/01/2006 15:04"), main_data_state.cash_allocation_file, err))
						month_file, _ = excelize.OpenFile(main_data_state.cash_allocation_excel_file_location)
					}
					main_data_state.cash_allocation_file = month_file
					main_data_state.cash_allocation_excel_file_sheets = month_file.GetSheetList()
					fmt.Println("Sheets in this file: ", main_data_state.cash_allocation_excel_file_sheets)

					log_file.WriteString(fmt.Sprintf("[%s]: Month File Added: %s \n", time.Now().Local().Format("02/01/2006 15:04"), main_data_state.cash_allocation_file))
					// main_data_state.information = "Loaded Month File"
				}()
			}
			// This flex splits the window vertically.
			layout.Flex{
				Axis: layout.Vertical,
			}.Layout(graphical_context,
				layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
					return layout.Flex{
						Axis: layout.Horizontal,
					}.Layout(graphical_context,
						layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
							menu_button_layout := material.Button(theme, &menu_button, "Menu")
							menu_button_layout.Color = colours_list[light_grey]
							set_background_rect_colour(graphical_context, material.Button(theme, &menu_button, "Menu").Layout(graphical_context).Size, colours_list[orange])
							// // paint.Fill(&ops, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
							// // Here we use the material.Clickable wrapper func to animate button clicks.
							// // return label.Layout(graphical_context)
							menu_button_layout.CornerRadius = unit.Dp(0)
							menu_button_layout.Inset = layout.Inset{
								Top:    unit.Dp(23),
								Bottom: unit.Dp(23),
								Left:   unit.Dp(15),
								Right:  unit.Dp(15),
							}
							// menu_button_layout.TextSize = unit.Sp(10) {R: 0, G: 94, B: 184, A: 255}
							menu_button_layout.Background = window_state.heading_colour
							return menu_button_layout.Layout(graphical_context)
							// return menu_button_layout.Layout(graphical_context)
							// label := material.Label(theme, unit.Sp(35), "Menu")
							// label.Alignment = text.Middle
							// label.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}

							// set_background_rect_colour(graphical_context, label.Layout(graphical_context).Size, colours_list[6])
							// // paint.Fill(&ops, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
							// // Here we use the material.Clickable wrapper func to animate button clicks.
							// // return label.Layout(graphical_context)
							// // return margins.Layout(graphical_context, label.Layout)
							// return label.Layout(graphical_context)
						}),
						layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
							// margins := layout.Inset{
							// 	Top:    unit.Dp(15),
							// 	Bottom: unit.Dp(15),
							// 	Left:   unit.Dp(0),
							// 	Right:  unit.Dp(0),
							// }
							label := material.Label(theme, unit.Sp(35), window_state.heading)
							label.Alignment = text.Middle
							label.Color = colours_list[light_grey]
							set_background_rect_colour(graphical_context, layout.Inset{
								Top:    unit.Dp(10),
								Bottom: unit.Dp(10),
							}.Layout(graphical_context, label.Layout).Size, window_state.heading_colour)

							// paint.Fill(&ops, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
							// Here we use the material.Clickable wrapper func to animate button clicks.
							// return label.Layout(graphical_context)
							// return margins.Layout(graphical_context, label.Layout)

							// return label.Layout(graphical_context)
							return layout.Inset{
								Top:    unit.Dp(10),
								Bottom: unit.Dp(10),
							}.Layout(graphical_context, label.Layout)
						}),
					)
				}),
				layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
					// This flex splits the bottom pane horizontally.
					return layout.Flex{
						Axis: layout.Horizontal,
					}.Layout(graphical_context,
						// layout.Flexed(0.25, func(graphical_context layout.Context) layout.Dimensions {
						//
						// 	return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
						// 		// Here we set the minimum constraints to zero. This allows our button to be smaller
						// 		// than the entire right-side pane in the UI. Without this change, the button is forced
						// 		// to occupy all of the space.
						// 		// graphical_context.Constraints.Min = image.Point{}
						// 		// Here we inset the button a little bit on all sides.
						// 		return layout.UniformInset(8).Layout(graphical_context,
						// 			material.Button(theme, &button_one, "Button1").Layout,
						// 		)
						// 	})

						// }),

						layout.Flexed(side_bar_width, func(graphical_context layout.Context) layout.Dimensions {

							set_background_rect_colour(graphical_context, left_side_bar(theme, graphical_context, side_buttons, side_button_text, colours_list).Size, colours_list[lighter_grey])
							return left_side_bar(theme, graphical_context, side_buttons, side_button_text, colours_list)

						}),
						layout.Flexed(right_content_width, func(graphical_context layout.Context) layout.Dimensions {

							switch show_selection {
							case 1:
								return main_data_state.control_accounts_layout(graphical_context, theme, control_accounts_widgets, &show_options, &dropdown, &dropdown_menu_enum)
							case 2:
								// , &editor
								return main_data_state.bank_rec_layout(graphical_context, theme, bank_rec_widgets, &show_options, &dropdown, &dropdown_menu_enum)
							case 3:
								return main_data_state.text_input_layouts(graphical_context, theme, sms_widget_text_inputs, []string{"Add Phone Number", "Add Sender", "Add Message"})
							case 4:
								return main_data_state.text_input_layouts(graphical_context, theme, email_widget_text_inputs, []string{"Add Email Addresses", "Add Subject", "Add Message"})
							case 5:
								return main_data_state.text_input_layouts(graphical_context, theme, email_widget_text_inputs, []string{"Add Email Addresses", "Add Subject", "Add Message"})

							// case 7:
							// 	return layout.Flex{
							// 		Axis: layout.Vertical,
							// 	}.Layout(
							// 		graphical_context,
							// 		layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
							// 			label := material.Label(theme, unit.Sp(20), "TEST STRING 3")
							// 			label.Alignment = text.Start
							// 			return label.Layout(graphical_context)
							// 		}),
							// 		layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
							// 			return right_side_layout(theme, graphical_context, right_side_buttons, colours_list)
							// 		}),
							// 		layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
							// 			return right_side_layout_one(theme, graphical_context, right_side_buttons, colours_list)
							// 		}),
							// 	)
							default:

								return layout.Flex{
									Axis: layout.Vertical,
								}.Layout(graphical_context,
									layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
										label := material.Label(theme, unit.Sp(20), "Select Button to the Left")
										label.Alignment = text.Start
										return label.Layout(graphical_context)
									}),
								)
							}
						}),
						// layout.Flexed(0.25,
						// 	func(graphical_context layout.Context) layout.Dimensions {
						// 		margins := layout.Inset{
						// 			Top:    unit.Dp(10),
						// 			Bottom: unit.Dp(10),
						// 			Right:  unit.Dp(25),
						// 			Left:   unit.Dp(25),
						// 		}
						// 		return margins.Layout(
						// 			graphical_context,
						// 			func(graphical_context layout.Context) layout.Dimensions {
						// 				button := material.Button(theme, &button_one, "Load BOXI File")
						// 				// button.Color = color.NRGBA{R: 76, G: 87, B: 96, A: 255}
						// 				button.Background = color.NRGBA{R: 190, G: 75, B: 4, A: 255}
						// 				return button.Layout(graphical_context)
						// 			},
						// 		)
						// 	},
						// ),
					)
				}),
				layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {

					if main_data_state.completed {
						// window_state.sub_heading = "Completed"
						sub_heading := material.Label(theme, unit.Sp(20), window_state.sub_heading)
						sub_heading.Color = colours_list[white]
						margins := layout.Inset{
							Top:    unit.Dp(5),
							Left:   unit.Dp(10),
							Bottom: unit.Dp(5),
							Right:  unit.Dp(0),
						}
						footer_image_point = margins.Layout(graphical_context, sub_heading.Layout).Size
						set_background_rect_colour(graphical_context, margins.Layout(graphical_context, sub_heading.Layout).Size, colours_list[green])
						return margins.Layout(graphical_context, sub_heading.Layout)
					} else {
						// set_background_rect_colour(graphical_context, material.Label(theme, unit.Sp(20), window_state.sub_heading).Layout(graphical_context).Size, colours_list[light_grey])
						return layout.Dimensions{}
					}
					// // paint.Fill(&ops, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
					// // Here we use the material.Clickable wrapper func to animate button clicks.
					// // return label.Layout(graphical_context)

					// menu_button_layout.TextSize = unit.Sp(10)
				}),
				layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {

					return layout.Flex{
						Axis:    layout.Horizontal,
						Spacing: layout.SpaceAround,
					}.Layout(graphical_context,
						layout.Flexed(0.5, func(graphical_context layout.Context) layout.Dimensions {
							// margins := layout.Inset{
							// 	Top:    unit.Dp(75),
							// 	Bottom: unit.Dp(0),
							// 	Right:  unit.Dp(25),
							// 	Left:   unit.Dp(5),
							// }
							// button := material.Button(theme, &end_button, "Close")
							// button.Background = color.NRGBA{R: 163, G: 22, B: 33, A: 255}
							// return margins.Layout(graphical_context, button.Layout)
							return layout.S.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
								// Here we set the minimum constraints to zero. This allows our button to be smaller
								// than the entire right-side pane in the UI. Without this change, the button is forced
								// to occupy all of the space.
								// graphical_context.Constraints.Min = image.Point{}
								// Here we inset the button a little bit on all sides.
								button := material.Button(theme, &end_button, "Close")
								button.Background = colours_list[dark_red]
								set_background_rect_colour(graphical_context, layout.UniformInset(5).Layout(graphical_context, button.Layout).Size, colours_list[light_grey])
								return layout.UniformInset(5).Layout(graphical_context, button.Layout)
							})
						}),
						layout.Flexed(0.5, func(graphical_context layout.Context) layout.Dimensions {
							// margins := layout.Inset{
							// 	Top:    unit.Dp(75),
							// 	Bottom: unit.Dp(0),
							// 	Right:  unit.Dp(5),
							// 	Left:   unit.Dp(25),
							// }
							// button := material.Button(theme, &start_button, "Start")
							// button.Background = color.NRGBA{R: 0, G: 143, B: 0, A: 255}
							// return margins.Layout(graphical_context, button.Layout)
							return layout.S.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
								// Here we set the minimum constraints to zero. This allows our button to be smaller
								// than the entire right-side pane in the UI. Without this change, the button is forced
								// to occupy all of the space.
								// graphical_context.Constraints.Min = image.Point{}
								// Here we inset the button a little bit on all sides.
								button := material.Button(theme, &start_button, "Start")

								button.Background = colours_list[dark_green]

								set_background_rect_colour(graphical_context, layout.UniformInset(5).Layout(graphical_context, button.Layout).Size, colours_list[light_grey])
								return layout.UniformInset(5).Layout(graphical_context, button.Layout)
							})
						}),
					)
				}),
			)
			event.Frame(graphical_context.Ops)
		case app.DestroyEvent:
			settings_data.StartMenu = int16(show_selection)
			settings_byte_array, _ := json.Marshal(settings_data)
			// json.NewEncoder(os.Stdout).Encode(&settings_data)
			os.WriteFile("settings.json", settings_byte_array, fs.ModeDevice)
			os.Exit(0)
		}
	}
}
func (window *main_data_struct) parse_boxi_data() boxi_main_data_struct {
	// get headers from below from first column
	boxi_balances_sheet_index := filter_single_arrays_return_index(&window.boxi_excel_file_sheets, func(sheet string) bool {
		return strings.Contains(strings.ToLower(sheet), "balances")
	})
	movements_sheet_index := filter_single_arrays_return_index(&window.boxi_excel_file_sheets, func(sheet string) bool {
		return strings.Contains(strings.ToLower(sheet), "movement")
	})
	boxi_data_account_balances, _ := read_single_sheet_data(window.boxi_excel_file_sheets[boxi_balances_sheet_index], window.boxi_file)
	boxi_index_account_balances_sheet := 0
	if boxi_data_account_balances[0][0] == "" {
		boxi_index_account_balances_sheet = 1
	}
	boxi_main_header := filter_multiple_arrays_return_index(&boxi_data_account_balances, func(array_item []string) bool {
		if len(array_item) > 0 {
			return strings.Contains(strings.ToLower(array_item[boxi_index_account_balances_sheet]), "centre")
		} else {
			return false
		}
	})
	boxi_temp_data := boxi_data_account_balances[boxi_main_header+1:]
	number_of_banks := filter_multiple_arrays(&boxi_temp_data, func(s []string) bool {
		return s[len(s)-1] != "0.00" && strings.Contains(s[len(s)-1], ".")
	})

	// DATA for Movements:
	boxi_data_movements, _ := read_single_sheet_data(window.boxi_excel_file_sheets[movements_sheet_index], window.boxi_file)
	boxi_index_movements_sheet := 0
	if boxi_data_movements[0][0] == "" {
		boxi_index_movements_sheet = 1
	}
	boxi_posting_header := filter_multiple_arrays_return_index(&boxi_data_movements, func(array_item []string) bool {
		if len(array_item) > 0 {
			return strings.Contains(strings.ToLower(array_item[boxi_index_movements_sheet]), "posting")
		} else {
			return false
		}
	})
	// // DATA for Movements:

	// boxi_data_first_sheet, _, _ := read_single_sheet_data(window.boxi_excel_file_sheets[0], window.boxi_file)

	var balances_data boxi_data_struct = boxi_data_struct{
		header:             boxi_data_account_balances[boxi_main_header],
		header_start_index: int8(boxi_index_account_balances_sheet),
		data:               boxi_data_account_balances[boxi_main_header+1:],
		filtered_data:      number_of_banks,
	}
	var movements_data boxi_data_struct = boxi_data_struct{
		header:             boxi_data_movements[boxi_posting_header],
		header_start_index: int8(boxi_index_movements_sheet),
		data:               boxi_data_movements[boxi_posting_header+1:],
	}
	fmt.Println("movements_data")
	fmt.Println(movements_data)
	return boxi_main_data_struct{
		header_start_column: boxi_data_account_balances[boxi_main_header][boxi_index_account_balances_sheet],
		number_of_banks:     int8(len(number_of_banks)),
		balances_data:       balances_data,
		movements:           movements_data,
	}
}
func (window *main_data_struct) parse_cash_allocation_file() cash_allocation_data_struct {
	sub_ledger_heading := []string{}
	payments_heading := []string{}
	journals_heading := []string{}
	sub_ledger_one_data := [][]string{}
	sub_ledger_two_data := [][]string{}
	sub_ledger_one_code := ""
	sub_ledger_two_code := ""
	// journals_code := ""
	payments := [][]string{}
	journals_income := [][]string{}
	journals_payments := [][]string{}
	other_credits := [][]string{}
	other_data := [][]string{}

	cash_allocation_file_data, _ := read_single_sheet_data(window.cash_allocation_excel_file_sheets[0], window.cash_allocation_file)
	// fmt.Println(cash_allocation_file_data)

	if !strings.EqualFold(strings.ToLower(cash_allocation_file_data[0][0]), "") {
		filter_trust_from_settings_file := filter_single_array(window.trusts, func(item trust_data_struct) bool {
			return strings.EqualFold(strings.ToLower(item.Trust), cash_allocation_file_data[0][0])
		})
		if len(filter_trust_from_settings_file) > 0 {
			window.trust = filter_trust_from_settings_file[0].Trust
			// window.information = filter_trust_from_settings_file[0].Trust
		}
	}

	sub_ledger_one_start := filter_multiple_arrays_return_index(&cash_allocation_file_data, func(array_item []string) bool {
		if len(array_item) > 0 {
			return strings.Contains(strings.ToLower(array_item[0]), "non-nhs")
		} else {
			return false
		}
	})

	sub_ledger_two_start := filter_multiple_arrays_return_index(&cash_allocation_file_data, func(array_item []string) bool {
		if len(array_item) > 0 {
			return strings.Contains(strings.ToLower(array_item[0]), "nhs") && !strings.Contains(strings.ToLower(array_item[0]), "non")
		} else {
			return false
		}
	})
	// fmt.Println("932: SUB LEDGER ONE START: ", sub_ledger_one_start, " SUB LEDGER TWO START: ", sub_ledger_two_start)
	journals_start := filter_multiple_arrays_return_index(&cash_allocation_file_data, func(array_item []string) bool {
		if len(array_item) > 8 {
			return strings.Contains(strings.ToLower(array_item[8]), "credit jnls to code")
		} else {
			return false
		}
	})
	other_credits_start := filter_multiple_arrays_return_index(&cash_allocation_file_data, func(array_item []string) bool {
		if len(array_item) > 8 {
			return strings.Contains(strings.ToLower(array_item[8]), "other credits -")
		} else {
			return false
		}
	})
	// get sub ledger one data
	// fmt.Println("CASH ALLOC DATA SUB 1 to SUB 2: ", cash_allocation_file_data[sub_ledger_one_start+1:sub_ledger_two_start])
	sub_ledger_one_code_index := 0
	for i := range cash_allocation_file_data[sub_ledger_one_start+1 : sub_ledger_two_start] {
		if len(cash_allocation_file_data[sub_ledger_one_start+1 : sub_ledger_two_start][i]) == 0 {
			sub_ledger_one_code_index = i + 1
			break
		}
		if cash_allocation_file_data[sub_ledger_one_start+1 : sub_ledger_two_start][i][0] == "" {
			sub_ledger_one_code_index = i
			break
		}
		sub_ledger_one_data = append(sub_ledger_one_data, cash_allocation_file_data[sub_ledger_one_start+1 : sub_ledger_two_start][i][:7])
	}
	fmt.Println("Sub ledger one data: ", sub_ledger_one_data)
	for i := range cash_allocation_file_data[(sub_ledger_one_start+1)+sub_ledger_one_code_index : sub_ledger_two_start] {
		if strings.Contains(cash_allocation_file_data[(sub_ledger_one_start+1)+sub_ledger_one_code_index : sub_ledger_two_start][i][6], ".") {
			sub_ledger_one_code = cash_allocation_file_data[(sub_ledger_one_start+1)+sub_ledger_one_code_index : sub_ledger_two_start][i][5]
			break
		}
	}
	fmt.Println("SUB LEDGER ONE CODE: ", sub_ledger_one_code)
	// fmt.Println("SUB LEDGER ONE")
	// fmt.Println(sub_ledger_heading)
	// fmt.Println(sub_ledger_one_data[1:])

	// get sub ledger two data
	sub_ledger_two_code_index := 0

	for i := range cash_allocation_file_data[sub_ledger_two_start+1:] {
		if len(cash_allocation_file_data[sub_ledger_two_start+1:]) == 0 {
			sub_ledger_two_code_index = i + 1
			break
		}
		if cash_allocation_file_data[sub_ledger_two_start+1:][i][0] == "" {
			sub_ledger_two_code_index = i
			break
		}
		sub_ledger_two_data = append(sub_ledger_two_data, cash_allocation_file_data[sub_ledger_two_start+1:][i][:7])
	}

	for i := range cash_allocation_file_data[(sub_ledger_two_start+1)+sub_ledger_two_code_index:] {
		fmt.Println(cash_allocation_file_data[(sub_ledger_two_start+1)+sub_ledger_two_code_index:][i])
		if strings.Contains(cash_allocation_file_data[(sub_ledger_two_start+1)+sub_ledger_two_code_index:][i][6], ".") {
			sub_ledger_two_code = cash_allocation_file_data[(sub_ledger_two_start+1)+sub_ledger_two_code_index:][i][5]
			break
		}
	}
	fmt.Println("SUB LEDGER TWO CODE: ", sub_ledger_two_code)
	// fmt.Println("\nSUB LEDGER TWO")
	// fmt.Println(sub_ledger_two_data[1:])

	// get payments data (subledger one index start)
	for i := range cash_allocation_file_data[sub_ledger_one_start+1 : journals_start] {
		if len(cash_allocation_file_data[sub_ledger_one_start+1 : journals_start][i]) > 7 {
			if cash_allocation_file_data[sub_ledger_one_start+1 : journals_start][i][8] == "" {
				break
			}
			payments = append(payments, cash_allocation_file_data[sub_ledger_one_start+1 : journals_start][i][8:12])
		}
	}
	// fmt.Println("\nPAYMENTS")
	// fmt.Println(payments_heading)
	// fmt.Println(payments[1:])

	// find credit jnls to code for journals

	// fmt.Println(cash_allocation_file_data[journals_start+1:])
	// journals_voucher_index := 0
	for i := range cash_allocation_file_data[journals_start+1:] {
		if len(cash_allocation_file_data[journals_start+1:][i]) > 8 {
			if cash_allocation_file_data[journals_start+1:][i][8] == "" {
				break
			}
			journals_income = append(journals_income, cash_allocation_file_data[journals_start+1:][i][8:14])
		}
	}
	journals_code := ""

	for i := range cash_allocation_file_data[other_credits_start-5 : other_credits_start+1] {
		if len(cash_allocation_file_data[other_credits_start-5 : other_credits_start+1][i]) > 9 {
			if strings.Contains(strings.ToLower(cash_allocation_file_data[other_credits_start-5 : other_credits_start+1][i][9]), "voucher") {
				if strings.EqualFold(cash_allocation_file_data[other_credits_start-5 : other_credits_start+1][i][10], "") {
					journals_code = cash_allocation_file_data[other_credits_start-5 : other_credits_start+1][i][11]
				} else {
					journals_code = cash_allocation_file_data[other_credits_start-5 : other_credits_start+1][i][10]
				}
				break
			}
			// for j := range cash_allocation_file_data[other_credits_start-5 : other_credits_start+1][i] {
			// 	fmt.Println("J: ", j, " ", cash_allocation_file_data[other_credits_start-5 : other_credits_start+1][i][j])
			// }
		}
	}
	fmt.Println("JOURNALS: ", journals_code)
	// fmt.Println("JOURNALS: ", cash_allocation_file_data[other_credits_start-5:other_credits_start+1])

	// journal income -> other credits find (Journal Voucher Number)
	// fmt.Println("\njournals_income")
	// fmt.Println(journals_heading)
	// fmt.Println(journals_income[1:])
	// fmt.Println("\njournals_payments")
	for i := range cash_allocation_file_data[journals_start+1:] {
		if len(cash_allocation_file_data[journals_start+1:][i]) > 14 {
			if cash_allocation_file_data[journals_start+1:][i][14] == "" {
				break
			}
			journals_payments = append(journals_payments, cash_allocation_file_data[journals_start+1:][i][14:])
		}
	}
	// fmt.Println(journals_payments[1:])
	for i := range cash_allocation_file_data[other_credits_start+1:] {
		if len(cash_allocation_file_data[other_credits_start+1:][i]) > 8 {
			if cash_allocation_file_data[other_credits_start+1:][i][8] == "" {
				break
			}
			other_credits = append(other_credits, cash_allocation_file_data[other_credits_start+1:][i][8:14])
		}
	}
	// fmt.Println("\nOTHER CREDITS")
	// fmt.Println(other_credits[1:])

	for i := range cash_allocation_file_data[sub_ledger_one_start+1 : journals_start-1] {
		if len(cash_allocation_file_data[sub_ledger_one_start+1 : journals_start-1][i]) > 14 {
			// if i == 15 {
			// 	break
			// }
			other_data = append(other_data, cash_allocation_file_data[sub_ledger_one_start+1 : journals_start-1][i][13:])
		}
	}
	opening_balance_string := filter_multiple_arrays(&other_data, func(s []string) bool {
		return strings.Contains(strings.ToLower(s[0]), "opening balance")
	})
	closing_balance_string := filter_multiple_arrays(&other_data, func(s []string) bool {
		return strings.Contains(strings.ToLower(s[0]), "closing balance")
	})
	sub_ledger_heading = sub_ledger_one_data[0:1][0]
	sub_ledger_heading = append(sub_ledger_heading, "Found")
	payments_heading = payments[0:1][0]
	payments_heading = append(payments_heading, "Found")
	journals_heading = journals_income[0:1][0]
	if journals_heading[len(journals_heading)-1] == "" {
		journals_heading = journals_heading[0 : len(journals_heading)-1]
	}
	if strings.ToLower(journals_heading[len(journals_heading)-1]) != "amount" {
		journals_heading = append(journals_heading, "Amount")
	}
	journals_heading = append(journals_heading, "Found")
	// closing balance set to data
	return cash_allocation_data_struct{
		opening_balance:       opening_balance_string[0][len(opening_balance_string[0])-1],
		closing_balance:       closing_balance_string[0][len(closing_balance_string[0])-2],
		sub_ledger_heading:    sub_ledger_heading,
		sub_ledger_one_data:   sub_ledger_one_data[1:],
		sub_ledger_one_code:   sub_ledger_one_code,
		sub_ledger_two_data:   sub_ledger_two_data[1:],
		sub_ledger_two_code:   sub_ledger_two_code,
		journal_heading:       journals_heading,
		journal_income:        journals_income[1:],
		journal_payments:      journals_payments[1:],
		journal_code:          journals_code,
		payments_heading:      payments_heading,
		payments:              payments[1:],
		other_credits_heading: []string{"description", "amount_one", "null_1", "null_2", "null_3", "amount_two", "null_4", "found"},
		other_credits:         other_credits[1:],
		other_data:            other_data,
	}
}
func (window *main_data_struct) parse_mapping_table() mapping_table_data_struct {
	mapping_table_boxi_report_sheet_data, _ := read_single_sheet_data(window.mapping_table_excel_file_sheets[0], window.mapping_table_file)
	// first sheet header
	// fmt.Println("HEADER")
	// fmt.Println(mapping_table_boxi_report_sheet_data[0])
	// // first sheet data without header
	// fmt.Println("DATA")
	// fmt.Println(mapping_table_boxi_report_sheet_data[1:])

	mapping_table_cash_allocation_sheet_data, _ := read_single_sheet_data(window.mapping_table_excel_file_sheets[1], window.mapping_table_file)
	// second sheet header
	// fmt.Println("HEADER")
	// fmt.Println(mapping_table_cash_allocation_sheet_data[0])
	// // second sheet data without header
	// fmt.Println("DATA")
	// fmt.Println(mapping_table_cash_allocation_sheet_data[1:])
	boxi_report_header := mapping_table_boxi_report_sheet_data[0]
	boxi_report_header = append(boxi_report_header, "Found")
	cash_allocation_header := mapping_table_cash_allocation_sheet_data[0]
	cash_allocation_header = append(cash_allocation_header, "Found")
	return mapping_table_data_struct{
		boxi_report_header:     boxi_report_header,
		boxi_report_data:       mapping_table_boxi_report_sheet_data[1:],
		cash_allocation_header: cash_allocation_header,
		cash_allocation_data:   mapping_table_cash_allocation_sheet_data[1:],
	}
}
func (window *main_data_struct) generate_data_bank_rec(log_file *os.File) {
	wait_group := &sync.WaitGroup{}
	completed_channel := make(chan bool, 1)
	wait_group.Add(1)

	// window.information = fmt.Sprintf("Running through BOXI for: %s", window.trust)
	log_file.WriteString(fmt.Sprintf("[%s]: Trust Selected: %s \n", time.Now().Local().Format("02/01/2006 15:04"), window.trust))
	// continue_app := false
	// fmt.Println(month_data)
	go func() {
		defer window.boxi_file.Close()
		defer window.cash_allocation_file.Close()
		defer window.mapping_table_file.Close()
		defer window.bank_reconciliation_file.Close()
		defer log_file.Close()

		// fmt.Println(month_data)
		// window.parse_boxi_data(window.boxi_excel_file_sheets[0])
		// var boxi_data boxi_main_data_struct = window.parse_boxi_data()
		var data_allocated []data_allocated_struct
		var alerts_and_errors []alerts_errors_not_allocated_struct
		// Boxi Data

		boxi_data := window.parse_boxi_data()
		boxi_balances := convert_to_object_data(boxi_data.balances_data.header, boxi_data.balances_data.data)
		// fmt.Println("boxi_balances")
		// fmt.Println(boxi_balances)
		// fmt.Println("boxi_data.balances_data.header")
		// fmt.Println(boxi_data.balances_data.header)
		boxi_balances_header_account_code_index := filter_single_arrays_return_index(&boxi_data.balances_data.header, func(s string) bool {
			return strings.Contains(strings.ToLower(s), "account code")
		})
		// fmt.Println("boxi_balances_header_account_code_index")
		// fmt.Println(boxi_balances_header_account_code_index)
		// fmt.Println(boxi_data.balances_data.header[boxi_balances_header_account_code_index])
		// fmt.Println(strings.ToLower(strings.ReplaceAll(boxi_data.balances_data.header[boxi_balances_header_account_code_index], " ", "_")))
		fmt.Println("BOXI HEADER: ", boxi_data.movements.header)
		boxi_data_movements := convert_to_object_data(boxi_data.movements.header, boxi_data.movements.data)
		// fmt.Println("boxi_data_movements")
		// // [ Posting Date & Time Batch Reference Number Transaction Reference Code Type Period Year Our Ref Transaction Description Amount]
		// fmt.Println(boxi_data.movements.header)
		// fmt.Println(boxi_data_movements)

		// Cash Allocation

		cash_allocation_data := window.parse_cash_allocation_file()
		sub_ledger_one_data := convert_to_object_data(cash_allocation_data.sub_ledger_heading, cash_allocation_data.sub_ledger_one_data)
		sub_ledger_one_code := cash_allocation_data.sub_ledger_one_code
		sub_ledger_two_data := convert_to_object_data(cash_allocation_data.sub_ledger_heading, cash_allocation_data.sub_ledger_two_data)
		sub_ledger_two_code := cash_allocation_data.sub_ledger_two_code

		// HAVE TO ACCOUNT FOR VAT, get previous item and add amount into there and remove vat item

		journals_income_data := convert_to_object_data(cash_allocation_data.journal_heading, cash_allocation_data.journal_income)
		journals_payment_data := convert_to_object_data(cash_allocation_data.journal_heading, cash_allocation_data.journal_payments)
		journal_code := cash_allocation_data.journal_code
		payments_data := convert_to_object_data(cash_allocation_data.payments_heading, cash_allocation_data.payments)
		other_credits_data := convert_to_object_data(cash_allocation_data.other_credits_heading, cash_allocation_data.other_credits)

		fmt.Println("CODES: ", sub_ledger_one_code, " ", sub_ledger_two_code, " ", journal_code)
		// Find total in boxi movements data

		var sub_ledger_one_total float64
		var sub_ledger_two_total float64

		for i := range sub_ledger_one_data {
			sub_ledger_one_total_temp_convert, _ := strconv.ParseFloat(sub_ledger_one_data[i]["amount"].(string), 64)
			sub_ledger_one_total += sub_ledger_one_total_temp_convert
		}
		for i := range sub_ledger_two_data {
			sub_ledger_two_total_temp_convert, _ := strconv.ParseFloat(sub_ledger_two_data[i]["amount"].(string), 64)
			sub_ledger_two_total += sub_ledger_two_total_temp_convert
		}
		fmt.Println("SUBLEDGER ONE TOTAL: ", sub_ledger_one_total, " FROM FLOAT: ", strconv.FormatFloat(sub_ledger_one_total, 'f', 2, 64))
		fmt.Println("SUBLEDGER TWO TOTAL: ", sub_ledger_two_total, " FROM FLOAT: ", strconv.FormatFloat(sub_ledger_two_total, 'f', 2, 64))
		// fmt.Println("SUB LEDGER ONE DATA: ", sub_ledger_one_data, " SUB LEDGER TWO DATA: ", sub_ledger_two_data, " JOURNAL INCOME DATA: ", journals_income_data, " JOURNAL PAYMENTS DATA: ", journals_payment_data, " PAYMENTS DATA: ", payments_data)
		// fmt.Println("cash_allocation_data.other_data")
		// fmt.Println(cash_allocation_data.other_data)

		// fmt.Println("cash_allocation_data.other_credits")
		// fmt.Println(cash_allocation_data.other_credits)
		// fmt.Println("other_credits_data")
		// fmt.Println(other_credits_data)
		// fmt.Println(boxi_data)

		// Mapping Table
		var mapping_table_data mapping_table_data_struct = window.parse_mapping_table()
		// fmt.Println(mapping_table_data)
		mapping_table_boxi_data_object := convert_to_object_data(mapping_table_data.boxi_report_header, mapping_table_data.boxi_report_data)
		mapping_table_cash_allocation_data_object := convert_to_object_data(mapping_table_data.cash_allocation_header, mapping_table_data.cash_allocation_data)
		// fmt.Println("mapping_table_boxi_data_object")
		// fmt.Println(mapping_table_data.boxi_report_header)
		// fmt.Println(mapping_table_boxi_data_object)
		// fmt.Println("mapping_table_cash_allocation_data_object")
		// fmt.Println(mapping_table_data.cash_allocation_header)
		// fmt.Println(mapping_table_cash_allocation_data_object)
		// for i := range mapping_table_cash_allocation_data_object {
		// 	fmt.Println(mapping_table_cash_allocation_data_object[i]["location_on_bank_rec"])
		// }

		/* TO DO:
		loop through number of banks
		if greater than 1, split worksheets by statement location -1 (bank rec)

		find words in worksheet: Bank Reconciliation and find index in worksheets

		*/
		// split worksheets

		// fmt.Println(window.bank_reconciliation_excel_file_sheets)

		// bank_rec_worksheets_split := filter_single_array(window.bank_reconciliation_excel_file_sheets, func(s string) bool {
		// 	return strings.Contains(strings.ToLower(s), "bank rec")
		// })
		// fmt.Println(bank_rec_worksheets_split)

		var bank_rec_calc_data bank_rec_calculation_array = find_bank_reconciliation_worksheets(window.bank_reconciliation_file, window.bank_reconciliation_excel_file_sheets)
		// fmt.Println("balances_index_locations: ", bank_rec_calc_data.data[0].balance_indexes)
		// fmt.Println("bank_rec_calc_data.data: ", bank_rec_calc_data.data)
		// fmt.Println(bank_rec_calc_data)
		// fmt.Println("CASH ALLOCATION DATA")
		// closing_balance, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(cash_allocation_data.closing_balance, ",", "")), 64)
		// closing_balance_format_float := strconv.FormatFloat(closing_balance, 'f', 2, 64)
		// fmt.Println(cash_allocation_data.closing_balance)
		// fmt.Println("closing_balance")
		// fmt.Println(closing_balance)
		// fmt.Println("closing_balance_format_float")
		// fmt.Println(closing_balance_format_float)

		opening_balance, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(cash_allocation_data.opening_balance, ",", "")), 64)

		/*
			change first balance in array to closing balance
			(second balance - first original balance) + closing balance
			(third balance - second balance) = find in boxi mapping data

			sheet            string
			difference       float64
			original_amounts [][]string
			new_amounts      [][]string
			calculation      [][]string
			balance_list     [][]string
		*/

		for i := range bank_rec_calc_data.data {
			// fmt.Println(bank_rec_calc_data.data[i])
			// fmt.Println("BOXI BALANCES LIST")

			for j := range boxi_balances {
				// fmt.Println("boxi_balances")
				// fmt.Println(boxi_balances)
				// fmt.Println("boxi_balances[i]")
				// fmt.Println(boxi_balances[j])
				// fmt.Println("boxi_balances[i][strings.ToLower(strings.ReplaceAll(boxi_data.balances_data.header[boxi_balances_header_account_code_index], \" \", \"_\"))]")
				// fmt.Println(boxi_balances[j][strings.ToLower(strings.ReplaceAll(boxi_data.balances_data.header[boxi_balances_header_account_code_index], " ", "_"))])
				// bank_rec_calc_data_filter_for_account_code := filter_single_array(bank_rec_calc_data.data, func(b []bank_rec_calculation) bool {
				// 	if b
				// })
				if boxi_balances[j][strings.ToLower(strings.ReplaceAll(boxi_data.balances_data.header[boxi_balances_header_account_code_index], " ", "_"))] == bank_rec_calc_data.data[i].account_code {
					bank_rec_calc_data.data[i].balance_slash_year = boxi_balances[j]["ytd_actuals"].(string)
					break
				}
			}
			var balance_amounts_array []float64
			var string_balamce_amounts_array []string
			// var calculated_amount float64
			for j := range bank_rec_calc_data.data[i].balance_list {
				// fmt.Println(bank_rec_calc_data.data[i].balance_list[j][len(bank_rec_calc_data.data[i].balance_list[j])-1])
				bank_rec_float_balance_amount, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(bank_rec_calc_data.data[i].balance_list[j][len(bank_rec_calc_data.data[i].balance_list[j])-1], ",", "")), 64)
				// bank_rec_format_float := strconv.FormatFloat(bank_rec_float_balance_amount, 'f', 2, 64)
				// fmt.Println("BANK REC FORMAT FLOAT: ", bank_rec_format_float)
				balance_amounts_array = append(balance_amounts_array, bank_rec_float_balance_amount)
				string_balamce_amounts_array = append(string_balamce_amounts_array, strconv.FormatFloat(bank_rec_float_balance_amount, 'f', 2, 64))
			}
			// fmt.Println("balance_amounts_array")
			// fmt.Println(balance_amounts_array)
			// find in balances based on account code, boxi_data.balances_data.header[boxi_balances_header_account_code_index], strings.ToLower(strings.ReplaceAll(boxi_data.balances_data.header[boxi_balances_header_account_code_index], " ", "_"))
			// first_total_balance_amount := (balance_amounts_array[1] - balance_amounts_array[0]) + closing_balance
			// second_total_balance_amount := (balance_amounts_array[3] - balance_amounts_array[2])

			// fmt.Println("boxi_balances_sheet_amount_from_account_code")
			// fmt.Println(bank_rec_calc_data.data[i].balance_slash_year)
			bank_rec_calc_data_balance_slash_year_float, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(bank_rec_calc_data.data[i].balance_slash_year, ",", "")), 64)
			// fmt.Println(bank_rec_calc_data_balance_slash_year_float)
			// fmt.Println("BALANCE AMOUNTS: 1: ", balance_amounts_array[0], " ", strconv.FormatFloat(balance_amounts_array[0], 'f', 2, 64), " 2: ", balance_amounts_array[1], " ", strconv.FormatFloat(balance_amounts_array[1], 'f', 2, 64), " 3: ", balance_amounts_array[2], " ", strconv.FormatFloat(balance_amounts_array[2], 'f', 2, 64), " 4: ", balance_amounts_array[3], " ", strconv.FormatFloat(balance_amounts_array[3], 'f', 2, 64), " ")

			// fmt.Println("CACLULATIONS: TOP: ", (balance_amounts_array[1]-balance_amounts_array[0])+opening_balance, " ", strconv.FormatFloat((balance_amounts_array[1]-balance_amounts_array[0])+opening_balance, 'f', 2, 64), " BOTTOM: ", (balance_amounts_array[3]-balance_amounts_array[2])+bank_rec_calc_data_balance_slash_year_float, " ", strconv.FormatFloat((balance_amounts_array[3]-balance_amounts_array[2])+bank_rec_calc_data_balance_slash_year_float, 'f', 2, 64))
			for _, data := range bank_rec_calc_data.data {
				window.bank_reconciliation_file.SetCellFloat(data.sheet, fmt.Sprintf("C%d", data.balance_indexes[0]+1), opening_balance, 2, 64)
				window.bank_reconciliation_file.SetCellFloat(data.sheet, fmt.Sprintf("C%d", data.balance_indexes[2]+1), bank_rec_calc_data_balance_slash_year_float, 2, 64)
			}
			// bank_rec_calc_data.data[i].difference = ((balance_amounts_array[3] - balance_amounts_array[2]) + bank_rec_calc_data_balance_slash_year_float) - ((balance_amounts_array[1] - balance_amounts_array[0]) + opening_balance)
			bank_rec_calc_data.data[i].difference = ((balance_amounts_array[3] - balance_amounts_array[2]) + bank_rec_calc_data_balance_slash_year_float) - balance_amounts_array[1]
			// last_total_minus_last_balance := (balance_amounts_array[3] - balance_amounts_array[2]) + bank_rec_calc_data_balance_slash_year_float
			// first_total_minus_second_balance := ((balance_amounts_array[1] - balance_amounts_array[0]) + closing_balance)
			// bank_rec_calc_data.data[i].difference = last_total_minus_last_balance - first_total_minus_second_balance
			// fmt.Println("FIRST AMOUNT: ", bank_rec_calc_data.data[i].difference, " STR CONV: ", strconv.FormatFloat(bank_rec_calc_data.data[i].difference, 'f', 2, 64), " Last Total Minus Last Balance: ", last_total_minus_last_balance, " ", strconv.FormatFloat(last_total_minus_last_balance, 'f', 2, 64), " First Total Minus Second Balance: ", first_total_minus_second_balance, " ", strconv.FormatFloat(first_total_minus_second_balance, 'f', 2, 64))
			// fmt.Println("FINAL AMOUNT: ", bank_rec_calc_data.data[i].difference, " STRING: ", strconv.FormatFloat(bank_rec_calc_data.data[i].difference, 'f', 2, 64))
			// fmt.Println("string_balamce_amounts_array")
			// fmt.Println(string_balamce_amounts_array)
		}
		// fmt.Println("DIFFERENCE AMOUNT: ", bank_rec_calc_data.data[0].difference, " ", strconv.FormatFloat(bank_rec_calc_data.data[0].difference, 'f', 2, 64))
		var boxi_total_amount_in_movements float64
		var boxi_last_index_movements int16

		for i := range boxi_data_movements {

			if strconv.FormatFloat(boxi_total_amount_in_movements, 'f', 2, 64) == strconv.FormatFloat(bank_rec_calc_data.data[0].difference, 'f', 2, 64) {
				boxi_last_index_movements = int16(i)
				break
			}
			// temp_amount := fmt.Sprintf("%f", boxi_data_movements[i]["amount"])
			temp_boxi_amount, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(boxi_data_movements[i]["amount"].(string), ",", "")), 64)
			// fmt.Println(boxi_total_amount_in_movements, ",", strconv.FormatFloat(boxi_total_amount_in_movements, 'f', 2, 64), ",", d["transaction_description"], ",", boxi_data_movements[i]["amount"], ",", boxi_data_movements[i]["amount"], ",", temp_boxi_amount, ",", strconv.FormatFloat(temp_boxi_amount, 'f', 2, 64), ",", boxi_total_amount_in_movements, ",", strconv.FormatFloat(boxi_total_amount_in_movements, 'f', 2, 64))
			// fmt.Println("D AMOUNT: ", d["amount"].(string))
			boxi_total_amount_in_movements += temp_boxi_amount
		}
		// fmt.Println("LAST INDEX: ", boxi_last_index_movements, " BOXI TOTAL LAST AMOUNT: ", strconv.FormatFloat(boxi_total_amount_in_movements, 'f', 2, 64), " DIFFERENCE: ", strconv.FormatFloat(bank_rec_calc_data.data[0].difference, 'f', 2, 64))

		// Filtered total amount slice from BOXI

		boxi_filtered_found_movements := boxi_data_movements[0:boxi_last_index_movements]

		// fmt.Println(boxi_filtered_found_movements)
		// fmt.Println("mapping_table_boxi_data_object")
		var boxi_movements_index_sub_ledger_one_total int8 = 0
		var boxi_movements_index_sub_ledger_two_total int8 = 0
		for i := range boxi_filtered_found_movements {
			if boxi_filtered_found_movements[i]["amount"] == strconv.FormatFloat(sub_ledger_one_total, 'f', 2, 64) {
				boxi_movements_index_sub_ledger_one_total = int8(i)
			}
			if boxi_filtered_found_movements[i]["amount"] == strconv.FormatFloat(sub_ledger_two_total, 'f', 2, 64) {
				boxi_movements_index_sub_ledger_two_total = int8(i)
			}
		}

		fmt.Println("1458:- BOXI TOTALS: SUB LEDGER ONE: ", boxi_movements_index_sub_ledger_one_total, " ", sub_ledger_one_total, " ", boxi_filtered_found_movements[boxi_movements_index_sub_ledger_one_total], " SUB LEDGER TWO: ", boxi_movements_index_sub_ledger_two_total, " ", sub_ledger_two_total, " ", boxi_filtered_found_movements[boxi_movements_index_sub_ledger_two_total])

		id := 0
		if boxi_movements_index_sub_ledger_one_total != 0 {
			// fmt.Println("BOXI MOVEMENTS SUB LEDGER ONE LOCATION: ", boxi_filtered_found_movements[boxi_movements_index_sub_ledger_one_total])
			for _, item := range sub_ledger_one_data {
				id += 1
				item["found"] = "BOXI Report"
				temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(item["amount"].(string)), ",", ""), 64)
				data_allocated = append(data_allocated, data_allocated_struct{
					sheet:           "BOXI Report",
					found_in_data:   "Cash Allocation , BOXI Report",
					type_of_data:    "Sub Ledger One",
					cell:            "Found Total",
					amount:          temp_amount_convert_to_float,
					additional_data: item["name"].(string),
					id:              id,
					data: []string{
						fmt.Sprintf("name: %s", item["name"].(string)),
						fmt.Sprintf("amount: %s", item["amount"].(string)),
						fmt.Sprintf("date: %s", item["date"].(string)),
						fmt.Sprintf("inv_no: %s", item["inv_no"].(string)),
						fmt.Sprintf("tr_date: %s", item["tr_date"].(string)),
						fmt.Sprintf("ac_no: %s", item["ac_no"].(string)),
					},
				})
			}
		}

		if boxi_movements_index_sub_ledger_two_total != 0 {
			// fmt.Println("BOXI MOVEMENTS SUB LEDGER TWO LOCATION: ", boxi_filtered_found_movements[boxi_movements_index_sub_ledger_two_total])
			for _, item := range sub_ledger_two_data {
				id += 1
				item["found"] = "BOXI Report"
				temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(item["amount"].(string)), ",", ""), 64)
				data_allocated = append(data_allocated, data_allocated_struct{
					sheet:           "BOXI Report",
					found_in_data:   "Cash Allocation , BOXI Report",
					type_of_data:    "Sub Ledger Two",
					cell:            "Found Total",
					amount:          temp_amount_convert_to_float,
					additional_data: item["name"].(string),
					id:              id,
					data: []string{
						fmt.Sprintf("name: %s", item["name"].(string)),
						fmt.Sprintf("amount: %s", item["amount"].(string)),
						fmt.Sprintf("date: %s", item["date"].(string)),
						fmt.Sprintf("inv_no: %s", item["inv_no"].(string)),
						fmt.Sprintf("tr_date: %s", item["tr_date"].(string)),
						fmt.Sprintf("ac_no: %s", item["ac_no"].(string)),
					},
				})
			}
		}
		// fmt.Println("MAPPING BOXI DATA TABLE: ", mapping_table_boxi_data_object)
		if len(mapping_table_boxi_data_object) > 0 {

			for i := range mapping_table_boxi_data_object {
				// fmt.Println(mapping_table_boxi_data_object[i])
				/*
					find in cash allocation file
					find in boxi file

					cash_allocation: Payment runs
					location_on_bank_rec: RBS Payments
					transaction_description: Act_R89
					Batch type: PB
				*/
				id += i
				if strings.Contains(strings.ToLower(strings.TrimSpace(mapping_table_boxi_data_object[i]["cash_allocation"].(string))), "payment runs") {
					// fmt.Println("HIT INTO PAYMENT RUNS: ", strings.ToLower(strings.ReplaceAll(mapping_table_boxi_data_object[i]["transaction_description"].(string), "_", " ")))
					for j := range payments_data {
						id += j
						if strings.Contains(strings.ToLower(strings.ReplaceAll(payments_data[j]["ref"].(string), "_", " ")), strings.ToLower(strings.ReplaceAll(mapping_table_boxi_data_object[i]["transaction_description"].(string), "_", " "))) {
							// fmt.Println("\n", mapping_table_boxi_data_object[i])
							// fmt.Println("PAYMENTS: ", cash_allocation_data.payments_heading, " ", payments_data[j]["ref"], " ", payments_data[j]["chaps"])

							// PAYMENTS DATA: chaps, cheques, payment_runs, ref

							payments_data[j]["found"] = "Mapping Table"
							mapping_table_boxi_data_object[i]["found"] = "found - payments"
							var temp_amount_convert_to_float_from_formatted float64
							var additional_data string
							if payments_data[j]["chaps"] != "NULL" {
								temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(payments_data[j]["chaps"].(string)), ",", ""), 64)
								temp_amount_convert_to_float_from_formatted = temp_amount_convert_to_float
								additional_data = "chaps"
							}
							if payments_data[j]["cheques"] != "NULL" {
								temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(payments_data[j]["cheques"].(string)), ",", ""), 64)
								temp_amount_convert_to_float_from_formatted = temp_amount_convert_to_float
								additional_data = "cheques"
							}
							if payments_data[j]["payment_runs"] != "NULL" {
								temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(payments_data[j]["payment_runs"].(string)), ",", ""), 64)
								temp_amount_convert_to_float_from_formatted = temp_amount_convert_to_float
								additional_data = "payment_runs"
							}
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           mapping_table_boxi_data_object[i]["location_on_bank_rec"].(string),
								found_in_data:   "Mapping Table , Boxi Report",
								type_of_data:    "Payments",
								amount:          temp_amount_convert_to_float_from_formatted,
								additional_data: additional_data,
								id:              id,
								data: []string{
									fmt.Sprintf("ref: %s", payments_data[j]["ref"].(string)),
									fmt.Sprintf("chaps: %s", payments_data[j]["chaps"].(string)),
									fmt.Sprintf("cheques: %s", payments_data[j]["cheques"].(string)),
									fmt.Sprintf("payment_runs: %s", payments_data[j]["payment_runs"].(string)),
								},
							})
							break
						}
					}
				}
			}
		}
		if len(mapping_table_cash_allocation_data_object) > 0 {
			for i := range mapping_table_cash_allocation_data_object {
				/*
					find in cash allocation file
					sub_ledger_one_data, sub_ledger_two_data,journals_income_data,journals_payment_data,payments_data
					map[columns:D income_or_payment:Income location_on_bank_rec:Catering Salford narrative_1:AMERICAN EXPRESS P narrative_2:AX8372476143]

					type data_allocated_struct struct {
						sheet         string   // Sheet to add data to
						location      string   // Cell location e.g. A4
						cell 		  string   // e.g. C
						found_in_data string   // Data it found it in e.g. mapping table
						type_of_data  string   // cash allocation type e.g. journal income
						amount        float64  // amount, can be converted back to string FromFloat
						data          []string // original data
					}

				*/
				id += i
				if mapping_table_cash_allocation_data_object[i]["narrative_2"] != "NULL" {

					// var found_item_in_cash_allocation_from_narrative_2 []string
					for j := range sub_ledger_one_data {
						id += j
						if strings.Contains(strings.ToLower(sub_ledger_one_data[j]["name"].(string)), strings.ToLower(mapping_table_cash_allocation_data_object[i]["narrative_2"].(string))) {
							// fmt.Println("\n", mapping_table_cash_allocation_data_object[i])
							// fmt.Println("SUBLEDGER ONE: ", cash_allocation_data.sub_ledger_heading, " ", sub_ledger_one_data[j]["name"], " ", sub_ledger_one_data[j]["amount"])

							// SUB LEDGER DATA: ac_no, amount, date, inv_no, name, tr_date

							sub_ledger_one_data[j]["found"] = "Mapping Table"
							mapping_table_cash_allocation_data_object[i]["found"] = "found - sub ledger one"
							temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(sub_ledger_one_data[j]["amount"].(string)), ",", ""), 64)
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           mapping_table_cash_allocation_data_object[i]["location_on_bank_rec"].(string),
								found_in_data:   "Mapping Table , Cash Allocation",
								type_of_data:    "Sub Ledger One",
								cell:            mapping_table_cash_allocation_data_object[i]["columns"].(string),
								amount:          temp_amount_convert_to_float,
								additional_data: sub_ledger_one_data[j]["name"].(string),
								id:              id,
								data: []string{
									fmt.Sprintf("name: %s", sub_ledger_one_data[j]["name"].(string)),
									fmt.Sprintf("amount: %s", sub_ledger_one_data[j]["amount"].(string)),
									fmt.Sprintf("date: %s", sub_ledger_one_data[j]["date"].(string)),
									fmt.Sprintf("inv_no: %s", sub_ledger_one_data[j]["inv_no"].(string)),
									fmt.Sprintf("tr_date: %s", sub_ledger_one_data[j]["tr_date"].(string)),
									fmt.Sprintf("ac_no: %s", sub_ledger_one_data[j]["ac_no"].(string)),
								},
							})
							// break
						}
					}
					for j := range sub_ledger_two_data {
						id += j
						if strings.Contains(strings.ToLower(sub_ledger_two_data[j]["name"].(string)), strings.ToLower(mapping_table_cash_allocation_data_object[i]["narrative_2"].(string))) {
							// fmt.Println("\n", mapping_table_cash_allocation_data_object[i])
							// fmt.Println("SUBLEDGER TWO: ", cash_allocation_data.sub_ledger_heading, " ", sub_ledger_two_data[j]["name"], " ", sub_ledger_two_data[j]["amount"])

							// SUB LEDGER DATA: ac_no, amount, date, inv_no, name, tr_date

							sub_ledger_two_data[j]["found"] = "Mapping Table"
							mapping_table_cash_allocation_data_object[i]["found"] = "found - sub ledger two"
							temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(sub_ledger_two_data[j]["amount"].(string)), ",", ""), 64)
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           mapping_table_cash_allocation_data_object[i]["location_on_bank_rec"].(string),
								found_in_data:   "Mapping Table , Cash Allocation",
								type_of_data:    "Sub Ledger Two",
								cell:            mapping_table_cash_allocation_data_object[i]["columns"].(string),
								amount:          temp_amount_convert_to_float,
								additional_data: sub_ledger_two_data[j]["name"].(string),
								id:              id,
								data: []string{
									fmt.Sprintf("name: %s", sub_ledger_two_data[j]["name"].(string)),
									fmt.Sprintf("amount: %s", sub_ledger_two_data[j]["amount"].(string)),
									fmt.Sprintf("date: %s", sub_ledger_two_data[j]["date"].(string)),
									fmt.Sprintf("inv_no: %s", sub_ledger_two_data[j]["inv_no"].(string)),
									fmt.Sprintf("tr_date: %s", sub_ledger_two_data[j]["tr_date"].(string)),
									fmt.Sprintf("ac_no: %s", sub_ledger_two_data[j]["ac_no"].(string)),
								},
							})
							// break
						}
					}
					for j := range journals_income_data {
						id += j
						if strings.Contains(strings.ToLower(journals_income_data[j]["description"].(string)), strings.ToLower(mapping_table_cash_allocation_data_object[i]["narrative_2"].(string))) {
							// fmt.Println("\n", mapping_table_cash_allocation_data_object[i])
							// fmt.Println("JOURNALS INCOME: ", cash_allocation_data.journal_heading, " ", journals_income_data[j]["description"], " ", journals_income_data[j]["amount"])

							// JOURNAL DATA: a.c., act.c., amount, c.c., description, job.c.

							journals_income_data[j]["found"] = "Mapping Table"
							mapping_table_cash_allocation_data_object[i]["found"] = "found - journal income"
							temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(journals_income_data[j]["amount"].(string)), ",", ""), 64)
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           mapping_table_cash_allocation_data_object[i]["location_on_bank_rec"].(string),
								found_in_data:   "Mapping Table , Cash Allocation",
								type_of_data:    "Journal Income",
								cell:            mapping_table_cash_allocation_data_object[i]["columns"].(string),
								amount:          temp_amount_convert_to_float,
								additional_data: journals_income_data[j]["description"].(string),
								id:              id,
								data: []string{
									fmt.Sprintf("description: %s", journals_income_data[j]["description"].(string)),
									fmt.Sprintf("amount: %s", journals_income_data[j]["amount"].(string)),
									fmt.Sprintf("a.c.: %s", journals_income_data[j]["a.c."].(string)),
									fmt.Sprintf("act.c.: %s", journals_income_data[j]["act.c."].(string)),
									fmt.Sprintf("c.c.: %s", journals_income_data[j]["c.c."].(string)),
									fmt.Sprintf("job.c.: %s", journals_income_data[j]["job.c."].(string)),
								},
							})
							// break
						}
					}
					for j := range journals_payment_data {
						id += j
						if strings.Contains(strings.ToLower(journals_payment_data[j]["description"].(string)), strings.ToLower(mapping_table_cash_allocation_data_object[i]["narrative_2"].(string))) {
							// fmt.Println("\n", mapping_table_cash_allocation_data_object[i])
							// fmt.Println("JOURNALS PAYMENTS: ", cash_allocation_data.journal_heading, " ", journals_payment_data[j]["description"], " ", journals_payment_data[j]["amount"])

							// JOURNAL DATA: a.c., act.c., amount, c.c., description, job.c.

							journals_payment_data[j]["found"] = "Mapping Table"
							mapping_table_cash_allocation_data_object[i]["found"] = "found - journal payments"
							temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(journals_payment_data[j]["amount"].(string)), ",", ""), 64)
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           mapping_table_cash_allocation_data_object[i]["location_on_bank_rec"].(string),
								found_in_data:   "Mapping Table , Cash Allocation",
								type_of_data:    "Journal Payment",
								cell:            mapping_table_cash_allocation_data_object[i]["columns"].(string),
								amount:          temp_amount_convert_to_float,
								additional_data: journals_payment_data[j]["description"].(string),
								id:              id,
								data: []string{
									fmt.Sprintf("description: %s", journals_payment_data[j]["description"].(string)),
									fmt.Sprintf("amount: %s", journals_payment_data[j]["amount"].(string)),
									fmt.Sprintf("a.c.: %s", journals_payment_data[j]["a.c."].(string)),
									fmt.Sprintf("act.c.: %s", journals_payment_data[j]["act.c."].(string)),
									fmt.Sprintf("c.c.: %s", journals_payment_data[j]["c.c."].(string)),
									fmt.Sprintf("job.c.: %s", journals_payment_data[j]["job.c."].(string)),
								},
							})
							// break
						}
					}
					for j := range payments_data {
						id += j
						if strings.Contains(strings.ToLower(payments_data[j]["ref"].(string)), strings.ToLower(mapping_table_cash_allocation_data_object[i]["narrative_2"].(string))) {
							// fmt.Println("\n", mapping_table_cash_allocation_data_object[i])
							// fmt.Println("PAYMENTS: ", cash_allocation_data.payments_heading, " ", payments_data[j]["ref"], " ", payments_data[j]["chaps"])

							// PAYMENTS DATA: chaps, cheques, payment_runs, ref

							payments_data[j]["found"] = "Mapping Table"
							mapping_table_cash_allocation_data_object[i]["found"] = "found - payments"
							var temp_amount_convert_to_float float64

							if payments_data[j]["chaps"].(string) != "" {
								temp_amount_convert_to_float, _ = strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(payments_data[j]["chaps"].(string)), ",", ""), 64)
							} else if payments_data[j]["cheques"].(string) != "" {
								temp_amount_convert_to_float, _ = strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(payments_data[j]["cheques"].(string)), ",", ""), 64)
							} else if payments_data[j]["payment_runs"].(string) != "" {
								temp_amount_convert_to_float, _ = strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(payments_data[j]["payment_runs"].(string)), ",", ""), 64)
							}
							fmt.Println("PAYMENTS DATA: ", payments_data[j], " ", temp_amount_convert_to_float)
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           mapping_table_cash_allocation_data_object[i]["location_on_bank_rec"].(string),
								found_in_data:   "Mapping Table , Cash Allocation",
								type_of_data:    "Payments",
								cell:            mapping_table_cash_allocation_data_object[i]["columns"].(string),
								amount:          0,
								additional_data: payments_data[j]["ref"].(string),
								id:              id,
								data: []string{
									fmt.Sprintf("ref: %s", payments_data[j]["ref"].(string)),
									fmt.Sprintf("chaps: %s", payments_data[j]["chaps"].(string)),
									fmt.Sprintf("cheques: %s", payments_data[j]["cheques"].(string)),
									fmt.Sprintf("payment_runs: %s", payments_data[j]["payment_runs"].(string)),
								},
							})
							// break
						}
					}
					for j := range other_credits_data {
						id += j
						if strings.Contains(strings.ToLower(other_credits_data[j]["description"].(string)), strings.ToLower(mapping_table_cash_allocation_data_object[i]["narrative_2"].(string))) {
							// fmt.Println("\n", mapping_table_cash_allocation_data_object[i])
							// fmt.Println("OTHER CREDITS: ", cash_allocation_data.other_credits_heading, " ", other_credits_data[j]["description"], " ", other_credits_data[j]["amount_two"])

							// OTHER CREDITS DATA:  description, amount_one, null_1, null_2, null_3, amount_two, null_4

							other_credits_data[j]["found"] = "Mapping Table"
							mapping_table_cash_allocation_data_object[i]["found"] = "found - other credits"
							temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(other_credits_data[j]["amount_two"].(string)), ",", ""), 64)
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           mapping_table_cash_allocation_data_object[i]["location_on_bank_rec"].(string),
								found_in_data:   "Mapping Table , Cash Allocation",
								type_of_data:    "Other Credits",
								cell:            mapping_table_cash_allocation_data_object[i]["columns"].(string),
								amount:          temp_amount_convert_to_float,
								additional_data: other_credits_data[j]["description"].(string),
								id:              id,
								data: []string{
									fmt.Sprintf("description: %s", other_credits_data[j]["description"].(string)),
									fmt.Sprintf("amount_one: %s", other_credits_data[j]["amount_one"].(string)),
									fmt.Sprintf("amount_two: %s", other_credits_data[j]["amount_two"].(string)),
								},
							})
							// break
						}
					}
					// fmt.Println("found_item_in_cash_allocation_from_narrative_2")
					// fmt.Println(found_item_in_cash_allocation_from_narrative_2)
				} else {
					for j := range sub_ledger_one_data {
						id += j
						if strings.Contains(strings.ToLower(sub_ledger_one_data[j]["name"].(string)), strings.ToLower(mapping_table_cash_allocation_data_object[i]["narrative_1"].(string))) {
							// fmt.Println("\n", mapping_table_cash_allocation_data_object[i])
							// fmt.Println("SUBLEDGER ONE: ", cash_allocation_data.sub_ledger_heading, " ", sub_ledger_one_data[j]["name"], " ", sub_ledger_one_data[j]["amount"])

							// SUB LEDGER DATA: ac_no, amount, date, inv_no, name, tr_date

							sub_ledger_one_data[j]["found"] = "Mapping Table"
							mapping_table_cash_allocation_data_object[i]["found"] = "found - sub ledger one"
							temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(sub_ledger_one_data[j]["amount"].(string)), ",", ""), 64)
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           mapping_table_cash_allocation_data_object[i]["location_on_bank_rec"].(string),
								found_in_data:   "Mapping Table , Cash Allocation",
								type_of_data:    "Sub Ledger One",
								cell:            mapping_table_cash_allocation_data_object[i]["columns"].(string),
								amount:          temp_amount_convert_to_float,
								additional_data: sub_ledger_one_data[j]["name"].(string),
								id:              id,
								data: []string{
									fmt.Sprintf("name: %s", sub_ledger_one_data[j]["name"].(string)),
									fmt.Sprintf("amount: %s", sub_ledger_one_data[j]["amount"].(string)),
									fmt.Sprintf("date: %s", sub_ledger_one_data[j]["date"].(string)),
									fmt.Sprintf("inv_no: %s", sub_ledger_one_data[j]["inv_no"].(string)),
									fmt.Sprintf("tr_date: %s", sub_ledger_one_data[j]["tr_date"].(string)),
									fmt.Sprintf("ac_no: %s", sub_ledger_one_data[j]["ac_no"].(string)),
								},
							})
							// break
						}
					}
					for j := range sub_ledger_two_data {
						id += j
						if strings.Contains(strings.ToLower(sub_ledger_two_data[j]["name"].(string)), strings.ToLower(mapping_table_cash_allocation_data_object[i]["narrative_1"].(string))) {
							// fmt.Println("\n", mapping_table_cash_allocation_data_object[i])
							// fmt.Println("SUBLEDGER TWO: ", cash_allocation_data.sub_ledger_heading, " ", sub_ledger_two_data[j]["name"], " ", sub_ledger_two_data[j]["amount"])

							// SUB LEDGER DATA: ac_no, amount, date, inv_no, name, tr_date

							sub_ledger_two_data[j]["found"] = "Mapping Table"
							mapping_table_cash_allocation_data_object[i]["found"] = "found - sub ledger two"
							temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(sub_ledger_two_data[j]["amount"].(string)), ",", ""), 64)
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           mapping_table_cash_allocation_data_object[i]["location_on_bank_rec"].(string),
								found_in_data:   "Mapping Table , Cash Allocation",
								type_of_data:    "Sub Ledger Two",
								cell:            mapping_table_cash_allocation_data_object[i]["columns"].(string),
								amount:          temp_amount_convert_to_float,
								additional_data: sub_ledger_two_data[j]["name"].(string),
								id:              id,
								data: []string{
									fmt.Sprintf("name: %s", sub_ledger_two_data[j]["name"].(string)),
									fmt.Sprintf("amount: %s", sub_ledger_two_data[j]["amount"].(string)),
									fmt.Sprintf("date: %s", sub_ledger_two_data[j]["date"].(string)),
									fmt.Sprintf("inv_no: %s", sub_ledger_two_data[j]["inv_no"].(string)),
									fmt.Sprintf("tr_date: %s", sub_ledger_two_data[j]["tr_date"].(string)),
									fmt.Sprintf("ac_no: %s", sub_ledger_two_data[j]["ac_no"].(string)),
								},
							})
							// break
						}
					}
					for j := range journals_income_data {
						id += j
						if strings.Contains(strings.ToLower(journals_income_data[j]["description"].(string)), strings.ToLower(mapping_table_cash_allocation_data_object[i]["narrative_1"].(string))) {
							// fmt.Println("\n", mapping_table_cash_allocation_data_object[i])
							// fmt.Println("JOURNALS INCOME: ", cash_allocation_data.journal_heading, " ", journals_income_data[j]["description"], " ", journals_income_data[j]["amount"])

							// JOURNAL DATA: a.c., act.c., amount, c.c., description, job.c.

							journals_income_data[j]["found"] = "Mapping Table"
							mapping_table_cash_allocation_data_object[i]["found"] = "found - journal income"
							temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(journals_income_data[j]["amount"].(string)), ",", ""), 64)
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           mapping_table_cash_allocation_data_object[i]["location_on_bank_rec"].(string),
								found_in_data:   "Mapping Table , Cash Allocation",
								type_of_data:    "Journal Income",
								cell:            mapping_table_cash_allocation_data_object[i]["columns"].(string),
								amount:          temp_amount_convert_to_float,
								additional_data: journals_income_data[j]["description"].(string),
								id:              id,
								data: []string{
									fmt.Sprintf("description: %s", journals_income_data[j]["description"].(string)),
									fmt.Sprintf("amount: %s", journals_income_data[j]["amount"].(string)),
									fmt.Sprintf("a.c.: %s", journals_income_data[j]["a.c."].(string)),
									fmt.Sprintf("act.c.: %s", journals_income_data[j]["act.c."].(string)),
									fmt.Sprintf("c.c.: %s", journals_income_data[j]["c.c."].(string)),
									fmt.Sprintf("job.c.: %s", journals_income_data[j]["job.c."].(string)),
								},
							})
							// break
						}
					}
					for j := range journals_payment_data {
						id += j
						if strings.Contains(strings.ToLower(journals_payment_data[j]["description"].(string)), strings.ToLower(mapping_table_cash_allocation_data_object[i]["narrative_1"].(string))) {
							// fmt.Println("\n", mapping_table_cash_allocation_data_object[i])
							// fmt.Println("JOURNALS PAYMENTS: ", cash_allocation_data.journal_heading, " ", journals_payment_data[j]["description"], " ", journals_payment_data[j]["amount"])

							// JOURNAL DATA: a.c., act.c., amount, c.c., description, job.c.

							journals_payment_data[j]["found"] = "Mapping Table"
							mapping_table_cash_allocation_data_object[i]["found"] = "found - journal payments"
							temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(journals_payment_data[j]["amount"].(string)), ",", ""), 64)
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           mapping_table_cash_allocation_data_object[i]["location_on_bank_rec"].(string),
								found_in_data:   "Mapping Table , Cash Allocation",
								type_of_data:    "Journal Payment",
								cell:            mapping_table_cash_allocation_data_object[i]["columns"].(string),
								amount:          temp_amount_convert_to_float,
								additional_data: journals_payment_data[j]["description"].(string),
								id:              id,
								data: []string{
									fmt.Sprintf("description: %s", journals_payment_data[j]["description"].(string)),
									fmt.Sprintf("amount: %s", journals_payment_data[j]["amount"].(string)),
									fmt.Sprintf("a.c.: %s", journals_payment_data[j]["a.c."].(string)),
									fmt.Sprintf("act.c.: %s", journals_payment_data[j]["act.c."].(string)),
									fmt.Sprintf("c.c.: %s", journals_payment_data[j]["c.c."].(string)),
									fmt.Sprintf("job.c.: %s", journals_payment_data[j]["job.c."].(string)),
								},
							})
							// break
						}
					}
					for j := range payments_data {
						id += j
						if strings.Contains(strings.ToLower(payments_data[j]["ref"].(string)), strings.ToLower(mapping_table_cash_allocation_data_object[i]["narrative_1"].(string))) {
							// fmt.Println("\n", mapping_table_cash_allocation_data_object[i])
							// fmt.Println("PAYMENTS: ", cash_allocation_data.payments_heading, " ", payments_data[j]["ref"], " ", payments_data[j]["chaps"])

							// PAYMENTS DATA: chaps, cheques, payment_runs, ref
							payments_data[j]["found"] = "Mapping Table"
							mapping_table_cash_allocation_data_object[i]["found"] = "found - payments"
							var temp_amount_convert_to_float float64

							if payments_data[j]["chaps"].(string) != "" {
								temp_amount_convert_to_float, _ = strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(payments_data[j]["chaps"].(string)), ",", ""), 64)
							} else if payments_data[j]["cheques"].(string) != "" {
								temp_amount_convert_to_float, _ = strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(payments_data[j]["cheques"].(string)), ",", ""), 64)
							} else if payments_data[j]["payment_runs"].(string) != "" {
								temp_amount_convert_to_float, _ = strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(payments_data[j]["payment_runs"].(string)), ",", ""), 64)
							}
							fmt.Println("PAYMENTS DATA: ", payments_data[j], " ", temp_amount_convert_to_float)
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           mapping_table_cash_allocation_data_object[i]["location_on_bank_rec"].(string),
								found_in_data:   "Mapping Table , Cash Allocation",
								type_of_data:    "Payments",
								cell:            mapping_table_cash_allocation_data_object[i]["columns"].(string),
								amount:          0,
								additional_data: payments_data[j]["ref"].(string),
								id:              id,
								data: []string{
									fmt.Sprintf("ref: %s", payments_data[j]["ref"].(string)),
									fmt.Sprintf("chaps: %s", payments_data[j]["chaps"].(string)),
									fmt.Sprintf("cheques: %s", payments_data[j]["cheques"].(string)),
									fmt.Sprintf("payment_runs: %s", payments_data[j]["payment_runs"].(string)),
								},
							})
							// break
						}
					}
					for j := range other_credits_data {
						id += j
						if strings.Contains(strings.ToLower(other_credits_data[j]["description"].(string)), strings.ToLower(mapping_table_cash_allocation_data_object[i]["narrative_1"].(string))) {
							// fmt.Println("\n", mapping_table_cash_allocation_data_object[i])
							// fmt.Println("OTHER CREDITS: ", cash_allocation_data.other_credits_heading, " ", other_credits_data[j]["description"], " ", other_credits_data[j]["amount_two"])

							// OTHER CREDITS DATA:  description, amount_one, null_1, null_2, null_3, amount_two, null_4

							other_credits_data[j]["found"] = "Mapping Table"
							mapping_table_cash_allocation_data_object[i]["found"] = "found - other credits"
							temp_amount_convert_to_float, _ := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(other_credits_data[j]["amount_two"].(string)), ",", ""), 64)
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           mapping_table_cash_allocation_data_object[i]["location_on_bank_rec"].(string),
								found_in_data:   "Mapping Table , Cash Allocation",
								type_of_data:    "Other Credits",
								cell:            mapping_table_cash_allocation_data_object[i]["columns"].(string),
								amount:          temp_amount_convert_to_float,
								additional_data: other_credits_data[j]["description"].(string),
								id:              id,
								data: []string{
									fmt.Sprintf("description: %s", other_credits_data[j]["description"].(string)),
									fmt.Sprintf("amount_one: %s", other_credits_data[j]["amount_one"].(string)),
									fmt.Sprintf("amount_two: %s", other_credits_data[j]["amount_two"].(string)),
								},
							})
							// break
						}
					}
				}
			}
		}

		// fmt.Println("DATA ALLOCATED:")
		// fmt.Println(data_allocated)
		// fmt.Println("NOT FOUND MAPPING TABLE")

		// Filter cash allocation journals, subledgers, payments, other credits

		/*
			TO-DO BANK REC:

				Filter payments to get act_r89 items, combine together and check in payments sheet / sheet they give on mapping table. Check it in there if not in there, add into gl payments section

				STAGE 1: Check BOXI and Cash Allocation
					If found, -> found

				Check cash allocation and bank rec
		*/

		// fmt.Println("window.boxi_file")
		// fmt.Println(boxi_data_movements)
		for j, item := range sub_ledger_one_data {
			if item["found"] == nil {
				temp_amount_float, _ := strconv.ParseFloat(item["amount"].(string), 64)
				found_in_loop_boxi_movements := false
				for k := range boxi_data_movements {
					get_boxi_movement_amount_into_float, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(boxi_data_movements[k]["amount"].(string), ",", "")), 64)
					if get_boxi_movement_amount_into_float == temp_amount_float {
						if strings.EqualFold(strings.ToLower(boxi_data_movements[k]["transaction_description"].(string)), strings.ToLower(item["name"].(string))) {
							// fmt.Println("2027 - GET BOXI MOVEMENT AMOUNT: ", get_boxi_movement_amount_into_float, " BOXI: ", boxi_data_movements[k]["transaction_description"], " ", boxi_data_movements[k]["amount"], " SUBLEDGER ONE ITEM: ", item["name"], " ", item["amount"])
							sub_ledger_one_data[j]["found"] = "Found in Boxi"
							found_in_loop_boxi_movements = true
							id += 1
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           "Movements",
								found_in_data:   "BOXI Report , Cash Allocation",
								type_of_data:    "Sub Ledger One",
								cell:            "",
								amount:          temp_amount_float,
								additional_data: item["name"].(string),
								id:              id,
								data: []string{
									fmt.Sprintf("name: %s", item["name"].(string)),
									fmt.Sprintf("amount: %s", item["amount"].(string)),
									fmt.Sprintf("date: %s", item["date"].(string)),
									fmt.Sprintf("inv_no: %s", item["inv_no"].(string)),
									fmt.Sprintf("tr_date: %s", item["tr_date"].(string)),
									fmt.Sprintf("ac_no: %s", item["ac_no"].(string)),
								},
							})
							break

						}
					}
				}
				if !found_in_loop_boxi_movements {
					alerts_and_errors = append(alerts_and_errors, alerts_errors_not_allocated_struct{
						original_location: "Sub Ledger 1",
						type_of_data:      "Alert",
						amount:            temp_amount_float,
						additional_data:   item["name"].(string),
						data: []string{
							fmt.Sprintf("name: %s", item["name"].(string)),
							fmt.Sprintf("amount: %s", item["amount"].(string)),
							fmt.Sprintf("date: %s", item["date"].(string)),
							fmt.Sprintf("inv_no: %s", item["inv_no"].(string)),
							fmt.Sprintf("tr_date: %s", item["tr_date"].(string)),
							fmt.Sprintf("ac_no: %s", item["ac_no"].(string)),
						},
					})
				}
			}
		}
		for j, item := range sub_ledger_two_data {
			// fmt.Println("SUB LEDGER TWO: ", item)
			if item["found"] == nil {
				temp_amount_float, _ := strconv.ParseFloat(item["amount"].(string), 64)
				found_in_loop_boxi_movements := false
				for k := range boxi_data_movements {
					get_boxi_movement_amount_into_float, _ := strconv.ParseFloat(strings.TrimSpace(strings.ReplaceAll(boxi_data_movements[k]["amount"].(string), ",", "")), 64)
					if get_boxi_movement_amount_into_float == temp_amount_float {
						if strings.EqualFold(strings.ToLower(boxi_data_movements[k]["transaction_description"].(string)), strings.ToLower(item["name"].(string))) {
							// fmt.Println("2082 - GET BOXI MOVEMENT AMOUNT: ", get_boxi_movement_amount_into_float, " BOXI: ", boxi_data_movements[k]["transaction_description"], " ", boxi_data_movements[k]["amount"], " SUBLEDGER TWO ITEM: ", item["name"], " ", item["amount"])
							sub_ledger_two_data[j]["found"] = "Found in Boxi"
							found_in_loop_boxi_movements = true
							id += 1
							data_allocated = append(data_allocated, data_allocated_struct{
								sheet:           "Movements",
								found_in_data:   "BOXI Report , Cash Allocation",
								type_of_data:    "Sub Ledger Two",
								cell:            "",
								amount:          temp_amount_float,
								additional_data: item["name"].(string),
								id:              id,
								data: []string{
									fmt.Sprintf("name: %s", item["name"].(string)),
									fmt.Sprintf("amount: %s", item["amount"].(string)),
									fmt.Sprintf("date: %s", item["date"].(string)),
									fmt.Sprintf("inv_no: %s", item["inv_no"].(string)),
									fmt.Sprintf("tr_date: %s", item["tr_date"].(string)),
									fmt.Sprintf("ac_no: %s", item["ac_no"].(string)),
								},
							})
							break

						}
					}
				}
				if !found_in_loop_boxi_movements {
					alerts_and_errors = append(alerts_and_errors, alerts_errors_not_allocated_struct{
						original_location: "Sub Ledger 2",
						type_of_data:      "Alert",
						amount:            temp_amount_float,
						additional_data:   item["name"].(string),
						data: []string{
							fmt.Sprintf("name: %s", item["name"].(string)),
							fmt.Sprintf("amount: %s", item["amount"].(string)),
							fmt.Sprintf("date: %s", item["date"].(string)),
							fmt.Sprintf("inv_no: %s", item["inv_no"].(string)),
							fmt.Sprintf("tr_date: %s", item["tr_date"].(string)),
							fmt.Sprintf("ac_no: %s", item["ac_no"].(string)),
						},
					})
				}
			}
		}

		/*
			TO-DO BANK REC:
				if journal code isn't empty, journal income + journal payments = amount.
					if journal payments = 0.00 then just journal income
				get code from boxi & check in batch reference number, if not found in batch ref num check in transaction reference code
		*/

		var find_journal_code_in_boxi_report int = 0
		for i := range boxi_data_movements {
			if strings.EqualFold(boxi_data_movements[i]["batch_reference_number"].(string), journal_code) {
				fmt.Println(boxi_data_movements[i]["batch_reference_number"])
				find_journal_code_in_boxi_report = i
			}
		}
		fmt.Println("BOXI ITEM FROM JOURNAL CODE INDEX: ", boxi_data_movements[find_journal_code_in_boxi_report])
		for _, item := range journals_income_data {
			if item["found"] == nil {
				temp_amount_float, _ := strconv.ParseFloat(item["amount"].(string), 64)
				alerts_and_errors = append(alerts_and_errors, alerts_errors_not_allocated_struct{
					original_location: "Journal Income",
					type_of_data:      "Alert",
					amount:            temp_amount_float,
					additional_data:   item["description"].(string),
					data: []string{
						fmt.Sprintf("description: %s", item["description"].(string)),
						fmt.Sprintf("amount: %s", item["amount"].(string)),
						fmt.Sprintf("a.c.: %s", item["a.c."].(string)),
						fmt.Sprintf("act.c.: %s", item["act.c."].(string)),
						fmt.Sprintf("c.c.: %s", item["c.c."].(string)),
						fmt.Sprintf("job.c.: %s", item["job.c."].(string)),
					},
				})
			}
		}
		for _, item := range journals_payment_data {
			if item["found"] == nil {
				temp_amount_float, _ := strconv.ParseFloat(item["amount"].(string), 64)
				alerts_and_errors = append(alerts_and_errors, alerts_errors_not_allocated_struct{
					original_location: "Journal Payment",
					type_of_data:      "Alert",
					amount:            temp_amount_float,
					additional_data:   item["description"].(string),
					data: []string{
						fmt.Sprintf("description: %s", item["description"].(string)),
						fmt.Sprintf("amount: %s", item["amount"].(string)),
						fmt.Sprintf("a.c.: %s", item["a.c."].(string)),
						fmt.Sprintf("act.c.: %s", item["act.c."].(string)),
						fmt.Sprintf("c.c.: %s", item["c.c."].(string)),
						fmt.Sprintf("job.c.: %s", item["job.c."].(string)),
					},
				})
			}
		}
		for _, item := range payments_data {
			if item["found"] == nil {
				var temp_amount_convert_to_float float64

				if !strings.EqualFold(item["chaps"].(string), "NULL") {
					temp_amount_convert_to_float, _ = strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(item["chaps"].(string)), ",", ""), 64)
				}
				if !strings.EqualFold(item["cheques"].(string), "NULL") {
					temp_amount_convert_to_float, _ = strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(item["cheques"].(string)), ",", ""), 64)
				}
				if !strings.EqualFold(item["payment_runs"].(string), "NULL") {
					temp_amount_convert_to_float, _ = strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(item["payment_runs"].(string)), ",", ""), 64)
				}
				alerts_and_errors = append(alerts_and_errors, alerts_errors_not_allocated_struct{
					original_location: "Payments",
					type_of_data:      "Alert",
					amount:            temp_amount_convert_to_float,
					additional_data:   item["ref"].(string),
					data: []string{
						fmt.Sprintf("ref: %s", item["ref"].(string)),
						fmt.Sprintf("chaps: %s", item["chaps"].(string)),
						fmt.Sprintf("cheques: %s", item["cheques"].(string)),
						fmt.Sprintf("payment_runs: %s", item["payment_runs"].(string)),
					},
				})
			}
		}
		for _, item := range other_credits_data {
			if item["found"] == nil {
				temp_amount_float, _ := strconv.ParseFloat(item["amount_two"].(string), 64)
				alerts_and_errors = append(alerts_and_errors, alerts_errors_not_allocated_struct{
					original_location: "Other Credits",
					type_of_data:      "Alert",
					amount:            temp_amount_float,
					additional_data:   item["description"].(string),
					data: []string{
						fmt.Sprintf("description: %s", item["description"].(string)),
						fmt.Sprintf("amount_one: %s", item["amount_one"].(string)),
						fmt.Sprintf("amount_two: %s", item["amount_two"].(string)),
					},
				})
			}
		}
		/*
			TO-DO BANK REC: Loop through not found alerts and try to allocate
				split by text? find worksheet with that text? get column that contains that number?

		*/

		/*
			TO-DO BANK REC:
				if in movements and not in cash allocation then add to alerts
		*/
		// loop through alerts

		allocated_sheet_name := "Allocated"
		alerts_sheet_name := "Alerts"
		window.bank_reconciliation_file.DeleteSheet(allocated_sheet_name)
		window.bank_reconciliation_file.DeleteSheet(alerts_sheet_name)

		for _, sheet_name := range window.bank_reconciliation_excel_file_sheets {

			find_item_based_on_sheet_name := filter_single_array(data_allocated, func(i data_allocated_struct) bool {
				return strings.EqualFold(strings.ToLower(i.sheet), strings.ToLower(sheet_name))
			})
			if len(find_item_based_on_sheet_name) > 0 {
				// fmt.Println("SHEET NAME: ", sheet_name)
				// fmt.Println(find_item_based_on_sheet_name)
				bank_rec_temp_sheet_data, _ := read_single_sheet_data(sheet_name, window.bank_reconciliation_file)

				// fmt.Println("bank_rec_temp_sheet_data[find_last_item_last_index]")
				// fmt.Println(bank_rec_temp_sheet_data[find_last_item_last_index])
				// fmt.Println(len(bank_rec_temp_sheet_data[find_last_item_last_index]))
				// fmt.Println(excel_columns_alphabet)
				// print_type(find_item_based_on_sheet_name)
				var find_last_item_last_index int16
				for i := range 300 {
					if len(bank_rec_temp_sheet_data[(len(bank_rec_temp_sheet_data)-1)-i]) > 0 {
						if len(bank_rec_temp_sheet_data[(len(bank_rec_temp_sheet_data)-1)-i][0]) != 0 {
							temp_cell_formula, _ := window.bank_reconciliation_file.GetCellFormula(sheet_name, fmt.Sprintf("G%d", int16((len(bank_rec_temp_sheet_data)-1)-i)))
							fmt.Println(bank_rec_temp_sheet_data[(len(bank_rec_temp_sheet_data)-1)-i], " CELL FORMULA: ", temp_cell_formula)
							find_last_item_last_index = int16((len(bank_rec_temp_sheet_data) - 1) - i)
							break
						}
					}
				}

				// if sheet_name == "Catering Oldham" {
				// 	// Filter for First column not empty
				// 	fmt.Println("Catering Oldham: ", bank_rec_temp_sheet_data[len(bank_rec_temp_sheet_data)-1])
				// 	fmt.Println(filter_multiple_arrays(&bank_rec_temp_sheet_data, func(s []string) bool { return !strings.EqualFold(s[0], "") }))
				// 	fmt.Println("Catering Oldham: ", bank_rec_temp_sheet_data[find_last_item_last_index], " SHEET: ", sheet_name)
				// }
				for j := range find_item_based_on_sheet_name {

					// loop backwards to find column A last item with data. +1 to insert into empty row
					find_item_based_on_sheet_name[j].length_of_one_item = int16(len(bank_rec_temp_sheet_data[find_last_item_last_index]))
					find_item_based_on_sheet_name[j].location = fmt.Sprintf("%s%d", excel_columns_alphabet[0], find_last_item_last_index+1)
					cell_index_for_column_character := filter_single_arrays_return_index(&excel_columns_alphabet, func(s string) bool { return s == strings.ToUpper(find_item_based_on_sheet_name[j].cell) })
					// fmt.Println("DATA TO ADD: ", find_item_based_on_sheet_name[j].additional_data, " CELL: ", find_item_based_on_sheet_name[j].cell, " ALPHABET INDEX: ", excel_columns_alphabet[cell_index_for_column_character], " AMOUNT: ", find_item_based_on_sheet_name[j].amount)
					length_of_items_to_create_string_array := find_item_based_on_sheet_name[j].length_of_one_item
					// fmt.Println("LENGTH: ", find_item_based_on_sheet_name[j].length_of_one_item)
					if cell_index_for_column_character > len(bank_rec_temp_sheet_data[find_last_item_last_index]) {
						length_of_items_to_create_string_array = int16(cell_index_for_column_character) + 1
					}

					// fmt.Println("LENGTH: ", length_of_items_to_create_string_array)
					temp_data_to_add_to_bank_rec := make([]string, length_of_items_to_create_string_array)
					temp_data_to_add_to_bank_rec[0] = time.Now().Local().Format("02/01/2006")
					temp_data_to_add_to_bank_rec[1] = find_item_based_on_sheet_name[j].additional_data
					temp_data_to_add_to_bank_rec[cell_index_for_column_character] = strconv.FormatFloat(find_item_based_on_sheet_name[j].amount, 'f', 2, 64)

					for k := range data_allocated {
						if data_allocated[k].id == find_item_based_on_sheet_name[j].id {
							data_allocated[k].location = fmt.Sprintf("A%d", find_last_item_last_index+1)
							break
						}
					}
					find_last_item_last_index += 1
					// window.bank_reconciliation_file.InsertRows(sheet_name, int(find_last_item_last_index+2), 1)
					window.bank_reconciliation_file.DuplicateRowTo(sheet_name, int(find_last_item_last_index), int(find_last_item_last_index+2))

					// window.bank_reconciliation_file.InsertRows(sheet_name, int(find_last_item_last_index+1), 1)
					// find formulas in cells?

					// index into
					for k := range temp_data_to_add_to_bank_rec {
						// find_last_item_last_index += 1
						// if sheet_name == "Catering Oldham" {
						// 	fmt.Println("CELL INDEX COLUMN: ", cell_index_for_column_character, " ALPHABET: ", excel_columns_alphabet[cell_index_for_column_character])
						// }
						if k == 0 || k == 1 || k == cell_index_for_column_character {
							window.bank_reconciliation_file.SetCellDefault(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[k], find_last_item_last_index+2), temp_data_to_add_to_bank_rec[k])
						} else {
							temp_cell_formula, _ := window.bank_reconciliation_file.GetCellFormula(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[k], int(find_last_item_last_index)))
							// cell_type, _ := window.bank_reconciliation_file.GetCellType(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[k], int(find_last_item_last_index+2)))
							// fmt.Println("CELL SPRINT: ", fmt.Sprintf("%s%d", excel_columns_alphabet[k], int(find_last_item_last_index+2)), "  ", temp_cell_formula, " CELL TYPE: ", cell_type)
							// fmt.Println("TEMP CELL FORMULA: ", temp_cell_formula)
							if strings.EqualFold(temp_cell_formula, "") {
								window.bank_reconciliation_file.SetCellFloat(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[k], find_last_item_last_index+2), 0.00, 2, 32)
							} else {
								window.bank_reconciliation_file.SetCellFormula(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[k], find_last_item_last_index+2), temp_cell_formula)
								// fmt.Println("CATERING: ", sheet_name, " ", temp_cell_formula, " CELL: ", excel_columns_alphabet[k])
							}
						}
					}
					// window.bank_reconciliation_file.SetSheetRow(sheet_name, fmt.Sprintf("A%d", find_last_item_last_index), temp_data_to_add_to_bank_rec)

					// fmt.Println("temp_data_to_add_to_bank_rec: ", temp_data_to_add_to_bank_rec, " Length: ", len(temp_data_to_add_to_bank_rec), " BANK TEMP REC LENGTH: ", len(bank_rec_temp_sheet_data[find_last_item_last_index]), " FIND IN ITEM LENGTH: ", find_item_based_on_sheet_name[j].length_of_one_item)
					// find_last_item_last_index += 1
				}

			}
		}

		// fmt.Println("FOUND ITEMS TO ADD WORKSHEET: ", data_allocated)
		window.bank_reconciliation_file.NewSheet(allocated_sheet_name)
		window.bank_reconciliation_file.NewSheet(alerts_sheet_name)

		allocated_headings := []string{"ID", "Sheet", "Additional Data", "Amount", "Cell", "Location", "Found in", "Type", "Original Data"}
		alert_headings := []string{"Description", "Type", "Original Location", "Amount", "", "", "", "", "Original Data"}

		cell_style, _ := window.bank_reconciliation_file.NewStyle(&excelize.Style{NumFmt: 4})

		for i, data := range data_allocated {
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("A%d", i+1), data.id)
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("B%d", i+1), data.sheet)
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("C%d", i+1), data.additional_data)
			window.bank_reconciliation_file.SetCellFloat(allocated_sheet_name, fmt.Sprintf("D%d", i+1), data.amount, 2, 64)
			window.bank_reconciliation_file.SetCellStyle(allocated_sheet_name, fmt.Sprintf("D%d", i+1), fmt.Sprintf("D%d", i+1), cell_style)
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("E%d", i+1), data.cell)
			// window.bank_reconciliation_file.SetCellValue("Allocated and Alerts", fmt.Sprintf("D%d", i+1), strconv.FormatFloat(data.amount, 'f', 2, 64))
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("F%d", i+1), data.location)
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("G%d", i+1), data.found_in_data)
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("H%d", i+1), data.type_of_data)
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("I%d", i+1), data.data)
			// window.bank_reconciliation_file.SetCellValue("Allocated and Alerts",fmt.Sprintf("I%d",i),data.)
		}
		window.bank_reconciliation_file.InsertRows("Allocated and Alerts", 1, 2)
		for i, data := range allocated_headings {
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("%s2", excel_columns_alphabet[i]), data)
		}
		window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, "A1", "Allocated Data")
		last_item_in_allocated_sheet_index := len(data_allocated) + 3
		window.bank_reconciliation_file.SetCellFormula(allocated_sheet_name, fmt.Sprintf("D%d", last_item_in_allocated_sheet_index), fmt.Sprintf("SUM(D%d:D%d)", 3, len(data_allocated)+2))
		window.bank_reconciliation_file.SetCellStyle(allocated_sheet_name, fmt.Sprintf("D%d", last_item_in_allocated_sheet_index), fmt.Sprintf("D%d", last_item_in_allocated_sheet_index), cell_style)
		last_item_in_allocated_sheet_index += 1
		window.bank_reconciliation_file.InsertRows(allocated_sheet_name, last_item_in_allocated_sheet_index, 3)
		window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("A%d", last_item_in_allocated_sheet_index+1), "Alerts and Errors")
		last_item_in_allocated_sheet_index += 2
		/*
			ALERTS
			original_location string   // Data it found it in e.g. mapping table
			type_of_data      string   // cash allocation type e.g. journal income
			amount            float64  // amount, can be converted back to string FromFloat
			data              []string // original data
			additional_data   string
		*/
		for i, item := range alerts_and_errors {
			// var colour_fill []string

			// window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("A%d", (i+1)+last_item_in_allocated_sheet_index), item.additional_data)
			// window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("B%d", (i+1)+last_item_in_allocated_sheet_index), item.type_of_data)
			// window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("C%d", (i+1)+last_item_in_allocated_sheet_index), item.original_location)
			// window.bank_reconciliation_file.SetCellFloat(allocated_sheet_name, fmt.Sprintf("D%d", (i+1)+last_item_in_allocated_sheet_index), item.amount, 2, 64)
			// window.bank_reconciliation_file.SetCellStyle(allocated_sheet_name, fmt.Sprintf("D%d", (i+1)+last_item_in_allocated_sheet_index), fmt.Sprintf("D%d", (i+1)+last_item_in_allocated_sheet_index), cell_style)
			// window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("I%d", (i+1)+last_item_in_allocated_sheet_index), item.data)
			window.bank_reconciliation_file.SetCellValue(alerts_sheet_name, fmt.Sprintf("A%d", (i+1)), item.additional_data)
			window.bank_reconciliation_file.SetCellValue(alerts_sheet_name, fmt.Sprintf("B%d", (i+1)), item.type_of_data)
			window.bank_reconciliation_file.SetCellValue(alerts_sheet_name, fmt.Sprintf("C%d", (i+1)), item.original_location)
			window.bank_reconciliation_file.SetCellFloat(alerts_sheet_name, fmt.Sprintf("D%d", (i+1)), item.amount, 2, 64)
			window.bank_reconciliation_file.SetCellStyle(alerts_sheet_name, fmt.Sprintf("D%d", (i+1)), fmt.Sprintf("D%d", (i+1)), cell_style)
			window.bank_reconciliation_file.SetCellValue(alerts_sheet_name, fmt.Sprintf("I%d", (i+1)), item.data)

			colour := "000000"
			font_colour := "FFFFFF"
			switch strings.ToLower(item.original_location) {
			case "sub ledger 1":
				colour = "D10000"
				font_colour = "FFFFFF"
			case "sub ledger 2":
				colour = "EA2A2A"
				font_colour = "FFFFFF"
			case "journal income":
				colour = "FF9900"
				font_colour = "FFFFFF"
			case "journal payment":
				colour = "FBB141"
				font_colour = "FFFFFF"
			case "payments":
				colour = "CC00CC"
				font_colour = "FFFFFF"
			case "other credits":
				colour = "660066"
				font_colour = "FFFFFF"
			}

			alert_cell_style, _ := window.bank_reconciliation_file.NewStyle(&excelize.Style{
				Fill: excelize.Fill{Type: "pattern", Color: []string{colour}, Pattern: 1},
				Font: &excelize.Font{
					Color: font_colour,
				},
			})
			window.bank_reconciliation_file.SetCellStyle(alerts_sheet_name, fmt.Sprintf("A%d", (i+1)), fmt.Sprintf("I%d", (i+1)), alert_cell_style)
			alert_cell_style_float_amount, _ := window.bank_reconciliation_file.NewStyle(&excelize.Style{
				Fill: excelize.Fill{Type: "pattern", Color: []string{colour}, Pattern: 1}, NumFmt: 4,
				Font: &excelize.Font{
					Color: font_colour,
				},
			})
			window.bank_reconciliation_file.SetCellStyle(alerts_sheet_name, fmt.Sprintf("D%d", (i+1)), fmt.Sprintf("D%d", (i+1)), alert_cell_style_float_amount)
			// window.bank_reconciliation_file.SetCellStyle("Allocated and Alerts", fmt.Sprintf("E%d", (i+1)+last_item_in_allocated_sheet_index), fmt.Sprintf("I%d", (i+1)+last_item_in_allocated_sheet_index), alert_cell_style)
		}
		window.bank_reconciliation_file.SetColWidth(allocated_sheet_name, "I", "I", 150)
		window.bank_reconciliation_file.SetColWidth(alerts_sheet_name, "I", "I", 150)
		for i, data := range alert_headings {
			window.bank_reconciliation_file.SetCellValue(alerts_sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[i], i), data)
		}

		window.bank_reconciliation_file.SetCellFormula(alerts_sheet_name, fmt.Sprintf("D%d", (len(alerts_and_errors))+1), fmt.Sprintf("SUM(D%d:D%d)", last_item_in_allocated_sheet_index+1, (len(alerts_and_errors))))

		window.bank_reconciliation_file.SetCellStyle(allocated_sheet_name, fmt.Sprintf("D%d", (len(alerts_and_errors))+1), fmt.Sprintf("D%d", (len(alerts_and_errors))+1), cell_style)
		// last_item_in_allocated_sheet_index += (len(alerts_and_errors) + 5)
		last_item_in_allocated_sheet_index += 5

		for i, item := range boxi_filtered_found_movements {
			// map[amount:-8,000.00 our_ref:FASTER PAYMENT period:7 posting_date_&_time:10/21/2025 11:03:08 AM transaction_description:Quadient Uk Ltd transaction_reference_code:1994080 type:PG year:2026]

			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("A%d", (i+1)+last_item_in_allocated_sheet_index), item["posting_date_&_time"])
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("B%d", (i+1)+last_item_in_allocated_sheet_index), item["transaction_description"])
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("C%d", (i+1)+last_item_in_allocated_sheet_index), item["transaction_reference_code"])
			boxi_movements_temp_float, _ := strconv.ParseFloat(strings.ReplaceAll(item["amount"].(string), ",", ""), 64)
			window.bank_reconciliation_file.SetCellFloat(allocated_sheet_name, fmt.Sprintf("D%d", (i+1)+last_item_in_allocated_sheet_index), boxi_movements_temp_float, 2, 64)
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("E%d", (i+1)+last_item_in_allocated_sheet_index), item["period"])
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("F%d", (i+1)+last_item_in_allocated_sheet_index), item["type"])
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("G%d", (i+1)+last_item_in_allocated_sheet_index), item["our_ref"])
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("H%d", (i+1)+last_item_in_allocated_sheet_index), item["their_ref"])
			window.bank_reconciliation_file.SetCellValue(allocated_sheet_name, fmt.Sprintf("I%d", (i+1)+last_item_in_allocated_sheet_index), item["year"])
		}

		window.bank_reconciliation_file.MoveSheet(allocated_sheet_name, window.bank_reconciliation_excel_file_sheets[0])

		window.bank_reconciliation_file.UpdateLinkedValue()
		user_home_directory, _ := os.UserHomeDir()
		// fmt.Println("USER HOME DIRECTORY: ", user_home_directory)
		window.bank_reconciliation_file.SaveAs(fmt.Sprintf("M%s - %s - Bank Reconciliation.xlsx", filter_single_array(month_data, func(s month_data_struct) bool { return s.name == time.Now().Month().String() })[0].financial_month, user_home_directory))
		// // window.cash_allocation_file.Save()

		// log_file.WriteString(fmt.Sprintf("[%s]: Saved Month File \n", time.Now().Local().Format("02/01/2006 15:04")))
		// log_file.WriteString(fmt.Sprintf("[%s]: Completed \n", time.Now().Local().Format("02/01/2006 15:04")))
		// fmt.Println("Completed")
		// window.completed = true
		completed_channel <- true
		wait_group.Done()
		window.completed = true
	}()
	wait_group.Wait()
	close(completed_channel)
	fmt.Println("COMPLETED")
}
func find_bank_reconciliation_worksheets(workbook *excelize.File, worksheets []string) bank_rec_calculation_array {
	bank_rec_indexes_of_worksheets := []int8{}
	bank_rec_balances_index_locations := []int{}
	bank_rec_sheet_names_banks := []string{}
	var bank_rec_calc bank_rec_calculation_array

	// sheets           []string
	// original_amounts []string
	// new_amounts      []string
	// calculation      []string

	// fmt.Println("worksheets")
	// fmt.Println(worksheets)
	// loop through first 10 rows
	for i, sheet := range worksheets {
		// fmt.Println(sheet)
		bank_rec_data, _ := read_single_sheet_data(sheet, workbook)
		if len(bank_rec_data) > 10 {
			for j := range 10 {
				// if len(bank_rec_data[j]) > 5 {
				// 	if strings.Contains(strings.ToLower(bank_rec_data[j][0]), "bank reconciliation") {
				// 		bank_rec_indexes_of_worksheets = append(bank_rec_indexes_of_worksheets, int16(i))
				// 		break
				// 	}
				// }
				if len(bank_rec_data[j]) > 0 {

					if strings.Contains(strings.ToLower(bank_rec_data[j][0]), "reconciliation") {
						var bank_rec_calc_data bank_rec_calculation
						// fmt.Println("bank_rec_data: ", bank_rec_data[j][0], " SHEET: ", sheet)
						bank_rec_indexes_of_worksheets = append(bank_rec_indexes_of_worksheets, int8(i))
						bank_rec_sheet_names_banks = append(bank_rec_sheet_names_banks, sheet)
						bank_rec_calc_data.sheet = sheet
						bank_rec_temp_balance_cells := filter_multiple_arrays(&bank_rec_data, func(s []string) bool {
							if len(s) > 0 {
								return strings.Contains(strings.ToLower(s[0]), "balance")
							} else {
								return false
							}
						})
						fmt.Println("bank_rec_temp_balance_cells: ", bank_rec_temp_balance_cells)
						// Get index and return that value
						bank_rec_balances_index_locations = filter_multiple_arrays_return_array_of_indexes(&bank_rec_data, func(s []string) bool {
							if len(s) > 0 {
								return strings.Contains(strings.ToLower(s[0]), "balance")
							} else {
								return false
							}
						})
						fmt.Println("BALANCES LOCATION INDEXES: ", bank_rec_balances_index_locations)

						if len(bank_rec_balances_index_locations) > 0 {
							bank_rec_calc_data.balance_indexes = bank_rec_balances_index_locations
							// fmt.Println("BALANCES: 0: ", bank_rec_data[bank_rec_balances_index_locations[0]], " BALANCES: 1: ", bank_rec_data[bank_rec_balances_index_locations[1]], " BALANCES: 2: ", bank_rec_data[bank_rec_balances_index_locations[2]], " BALANCES: 3: ", bank_rec_data[bank_rec_balances_index_locations[3]])
						}
						bank_rec_calc_data.balance_list = bank_rec_temp_balance_cells
						bank_rec_calc_data.original_amounts = bank_rec_temp_balance_cells
						// fmt.Println(bank_rec_temp_balance_cells)
						bank_rec_filter_for_account_code := filter_multiple_arrays(&bank_rec_data, func(s []string) bool {
							if len(s) > 0 {
								return strings.Contains(strings.ToLower(s[0]), "account code")
							} else {
								return false
							}
						})
						if len(bank_rec_filter_for_account_code) > 0 {
							// fmt.Println("FILTER BANK REC FOR ACCOUNT CODE: ", bank_rec_filter_for_account_code[0])
							// fmt.Println("FILTER BANK REC FOR ACCOUNT CODE: ", bank_rec_filter_for_account_code[0][0])
							bank_rec_split_account_string := strings.Split(strings.ReplaceAll(strings.ReplaceAll(strings.TrimSpace(bank_rec_filter_for_account_code[0][0]), ":", ""), " ", ","), ",")
							// fmt.Println("FILTER BANK REC FOR ACCOUNT CODE: ", bank_rec_split_account_string, " LENGTH: ", len(bank_rec_split_account_string))
							// fmt.Println("LAST ITEM: ", bank_rec_split_account_string[len(bank_rec_split_account_string)-1])
							bank_rec_calc_data.account_code = strings.TrimSpace(bank_rec_split_account_string[len(bank_rec_split_account_string)-1])
						}
						bank_rec_calc.data = append(bank_rec_calc.data, bank_rec_calc_data)

						/*
							change first balance in array to closing balance
							(second balance - first original balance) + closing balance
							(third balance - second balance) = find in boxi mapping data
						*/

						break
					}
				}
			}
		}
	}
	// fmt.Println("bank_rec_sheet_names_banks")
	fmt.Println("bank_rec_calc")
	// fmt.Println(bank_rec_calc)
	// fmt.Println(bank_rec_calc.data[0].sheet)
	return bank_rec_calc
}

// Control Accounts Functions

func (window *main_data_struct) generate_data_control_accounts(log_file *os.File) {
	wait_group := &sync.WaitGroup{}
	completed_channel := make(chan bool, 1)
	wait_group.Add(1)

	// window.information = fmt.Sprintf("Running through BOXI for: %s", window.trust)
	log_file.WriteString(fmt.Sprintf("[%s]: Trust Selected: %s \n", time.Now().Local().Format("02/01/2006 15:04"), window.trust))
	// continue_app := false
	// fmt.Println(month_data)
	go func() {
		defer window.boxi_file.Close()
		defer window.cash_allocation_file.Close()
		defer log_file.Close()
		boxi_rows, trust_sheet_name, err := read_cell_data_control_accounts(window.trust, window.boxi_file, true)
		// window.trust = trust_sheet_name
		fmt.Println("TRUST SHEET NAME: ", trust_sheet_name)
		if err != nil {
			log_file.WriteString(fmt.Sprintf("[%s]: Failed to Load BOXI Data: %s \n", time.Now().Local().Format("02/01/2006 15:04"), err))
		}
		log_file.WriteString(fmt.Sprintf("[%s]: BOXI Data Rows: %d \n", time.Now().Local().Format("02/01/2006 15:04"), len(boxi_rows)))
		header := []string{}
		start_row := 0
		boxi_first_column_start_row := ""
		// BOXI Column
		for i := range 5 {
			for j := range 10 {
				cell_val, _ := window.boxi_file.GetCellValue(trust_sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[i], j+1))
				fmt.Println("CELL:", fmt.Sprintf("%s%d", excel_columns_alphabet[i], j+1), "CELL VALUE: ", cell_val)
				if strings.Contains(cell_val, "Code") && !strings.Contains(cell_val, "Transaction") {
					start_row = j + 1
					boxi_first_column_start_row = cell_val
					// fmt.Println("Cell Value: ", cell_val)
					header = boxi_rows[j]
					break
				}
			}
			if len(header) > 0 {
				fmt.Println("HEADER LENGTH: ", len(header), " START ROW NAME: ", boxi_first_column_start_row, " HEADER: ", header)
				break
			}
		}
		fmt.Println("boxi_first_column_start_row: ", boxi_first_column_start_row, " HEADER: ", header)
		log_file.WriteString(fmt.Sprintf("[%s]: Headers: %s, LENGTH: %d", time.Now().Local().Format("02/01/2006 15:04"), header, len(header)))

		if strings.EqualFold(header[0], "") {
			fmt.Println("HEADER FIRST EMPTY")
			// header = header[1:]
		}
		// Getting index of the columns in the header list

		// cost_centre_column_column_index := slices.IndexFunc(header, func(h string) bool { return strings.Contains(strings.ToLower(h), "cost centre") })

		account_column_column_index := slices.IndexFunc(header, func(h string) bool { return strings.Contains(strings.ToLower(h), "account code") })

		job_code_column_column_index := slices.IndexFunc(header, func(h string) bool { return strings.Contains(strings.ToLower(h), "job code") })

		amount_column_index := slices.IndexFunc(header, func(h string) bool { return strings.Contains(strings.ToLower(h), "amount") })

		transaction_type_column := slices.IndexFunc(header, func(h string) bool { return strings.Contains(strings.ToLower(h), "transaction type") })

		transaction_reference_column := slices.IndexFunc(header, func(h string) bool { return strings.Contains(strings.ToLower(h), "transaction reference") })

		// period_column := slices.IndexFunc(header, func(h string) bool { return strings.Contains(strings.ToLower(h), "period") })
		fmt.Println("ACCOUNT CODE: ", account_column_column_index, " HEADER: ", header)
		log_file.WriteString(fmt.Sprintf("\n[%s]: (BOXI) - Start Row: %d, Start Row Name: %s, Last Row Index: %d, Transaction Ref: %d, Transaction Type: %d \n", time.Now().Local().Format("02/01/2006 15:04"), start_row, boxi_first_column_start_row, len(boxi_rows)-1, transaction_reference_column, transaction_type_column))

		log_file.WriteString(fmt.Sprintf("[%s]: Looping through Month File \n", time.Now().Local().Format("02/01/2006 15:04")))

		// window.information = "Running through Month File"

		window.update_actr_90_sheet(excel_columns_alphabet)
		// window.cash_allocation_file.SaveAs(fmt.Sprintf("ACTR90 M%s control accounts.xlsx", filter_single_array(month_data, func(s month_data_struct) bool { return s.name == time.Now().Month().String() })[0].financial_month))

		// continue_app := false
		// if !continue_app {
		// 	os.Exit(0)
		// }

		for _, sheet_name := range window.cash_allocation_excel_file_sheets {
			second_sheet_start_row := 0
			// reconcilling_cell := 0
			reconcilling_cell := ""
			// fmt.Println("boxi_first_column_start_row")
			// fmt.Println(boxi_first_column_start_row)
			// reconcilling_cell_letter := "C"
			for i := range 50 {
				cell_val, _ := window.cash_allocation_file.GetCellValue(sheet_name, fmt.Sprintf("A%d", i))
				for j := range 4 {
					reconcilling_cell_val, _ := window.cash_allocation_file.GetCellValue(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[j], i))
					if strings.Contains(strings.ToLower(reconcilling_cell_val), "reconciling items") {
						reconcilling_cell = fmt.Sprintf("%s%d", excel_columns_alphabet[j+1], i)
						break
					}
				}
				if cell_val == boxi_first_column_start_row {
					second_sheet_start_row = i
					break
				}
			}
			// fmt.Println(second_sheet_start_row, " ", reconcilling_cell)
			// fmt.Println(row_data_slice)
			// last_row := 0
			second_sheet_data_rows, _, _ := read_cell_data_control_accounts(sheet_name, window.cash_allocation_file, false)

			first_row_from_cash_allocation_file := slices.IndexFunc(second_sheet_data_rows, func(h []string) bool {
				if len(h) > 0 {
					// fmt.Println(h[0])
					return strings.EqualFold(strings.ToLower(h[0]), strings.ToLower(boxi_first_column_start_row))
				} else {
					return false
				}
			})
			boxi_row_data_slice := boxi_rows[start_row:]

			if first_row_from_cash_allocation_file != -1 {
				// cells_with_formulas := []int16{}
				cell_with_formula := 0
				for i := range second_sheet_data_rows {
					cell_formula, err := window.cash_allocation_file.GetCellFormula(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[amount_column_index-1], i+1))
					if err != nil {
						log_file.WriteString(fmt.Sprintf("[%s]: Error First Loop (Cell Formula): 488, error: %s\n", time.Now().Local().Format("02/01/2006 15:04"), err))
					}
					if strings.Contains(cell_formula, "SUM") || strings.Contains(cell_formula, ":") {
						cell_with_formula = int(i + 1)
						break
					}
				}
				if cell_with_formula > 0 {
					monthly_sheet_slice := second_sheet_data_rows[first_row_from_cash_allocation_file:cell_with_formula]
					first_empty_row := slices.IndexFunc(monthly_sheet_slice, func(h []string) bool { return len(h) == 0 })
					if first_empty_row != -1 {

						cell_formula_temp, err := window.cash_allocation_file.GetCellFormula(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[amount_column_index-1], cell_with_formula))
						if err != nil {
							log_file.WriteString(fmt.Sprintf("[%s]: Error First Loop (Cell Formula): 488, error: %s\n", time.Now().Local().Format("02/01/2006 15:04"), err))
						}
						window.update_cash_allocation_file_data(sheet_name, boxi_row_data_slice, monthly_sheet_slice, transaction_type_column, excel_columns_alphabet, cell_with_formula, cell_formula_temp, job_code_column_column_index, account_column_column_index, amount_column_index, second_sheet_start_row, reconcilling_cell, log_file)
					} else {
						temp_count := cell_with_formula
						cell_formula_temp, err := window.cash_allocation_file.GetCellFormula(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[amount_column_index-1], cell_with_formula))
						if err != nil {
							log_file.WriteString(fmt.Sprintf("[%s]: Error First Loop (Cell Formula): 488, error: %s\n", time.Now().Local().Format("02/01/2006 15:04"), err))
						}
						for range 3 {
							// add 3 empty rows
							window.cash_allocation_file.InsertRows(sheet_name, cell_with_formula, 1)
							// monthly_sheet_slice = slices.Insert(monthly_sheet_slice, len(monthly_sheet_slice)-1, temp_array_to_append)
							temp_count += 1
						}
						window.update_cash_allocation_file_data(sheet_name, boxi_row_data_slice, monthly_sheet_slice, transaction_type_column, excel_columns_alphabet, cell_with_formula, cell_formula_temp, job_code_column_column_index, account_column_column_index, amount_column_index, second_sheet_start_row, reconcilling_cell, log_file)

					}
				}

				window.cash_allocation_file.UpdateLinkedValue()
			}

		}
		// last_section_of_month_name := strings.Split(window.second_excel_file, "\\")[len(strings.Split(window.second_excel_file, "\\"))-1]
		// fmt.Println(last_section_of_month_name)
		window.cash_allocation_file.SaveAs(fmt.Sprintf("M%s Control Accounts.xlsx", filter_single_array(month_data, func(s month_data_struct) bool { return s.name == time.Now().Month().String() })[0].financial_month))
		// window.cash_allocation_file.Save()

		log_file.WriteString(fmt.Sprintf("[%s]: Saved Month File \n", time.Now().Local().Format("02/01/2006 15:04")))
		log_file.WriteString(fmt.Sprintf("[%s]: Completed \n", time.Now().Local().Format("02/01/2006 15:04")))
		// window.completed = true
		completed_channel <- true
		wait_group.Done()
		window.completed = true
	}()
	wait_group.Wait()
	close(completed_channel)
}
func (window *main_data_struct) update_actr_90_sheet(excel_columns_alphabet []string) {

	boxi_actr_90_sheet := slices.IndexFunc(window.boxi_excel_file_sheets, func(h string) bool { return strings.Contains(strings.ToLower(h), "act") })
	montly_file_actr_90_sheet := slices.IndexFunc(window.cash_allocation_excel_file_sheets, func(h string) bool { return strings.Contains(strings.ToLower(h), "act") })
	boxi_actr_90_sheet_name := window.boxi_file.GetSheetName(boxi_actr_90_sheet)
	monthly_actr_90_sheet_name := window.cash_allocation_file.GetSheetName(montly_file_actr_90_sheet)
	fmt.Println(monthly_actr_90_sheet_name)
	boxi_sheet_data_rows, _, _ := read_cell_data_control_accounts(boxi_actr_90_sheet_name, window.boxi_file, false)
	monthly_sheet_data_rows, _, _ := read_cell_data_control_accounts(monthly_actr_90_sheet_name, window.cash_allocation_file, false)
	boxi_actr_90_header := []string{}
	// boxi_actr_90_header_index := 0
	for i := range 5 {
		if len(boxi_sheet_data_rows[i]) > 0 {
			boxi_actr_90_header = boxi_sheet_data_rows[i]
			// boxi_actr_90_header_index = i
			break
		}
	}
	if boxi_actr_90_header[0] == "" {
		boxi_actr_90_header = boxi_actr_90_header[1:]
	}
	fmt.Println(boxi_actr_90_header)
	// monthly_actr_90_table_index := 0
	// for i := range 10 {
	// 	if len(monthly_sheet_data_rows[i]) > 0 {
	// 		if monthly_sheet_data_rows[i][0] == boxi_actr_90_header[0] {
	// 			monthly_actr_90_table_index = i
	// 			break
	// 		}
	// 	}
	// }
	monthly_actr_90_start := filter_multiple_arrays_return_index(&monthly_sheet_data_rows, func(s []string) bool {
		if len(s) > 0 {
			if s[0] == "" {
				s = s[1:]
			}
			return strings.EqualFold(strings.ToLower(s[0]), strings.ToLower(boxi_actr_90_header[0]))
		} else {
			return false
		}
	})
	fmt.Println(boxi_actr_90_header)
	fmt.Println("monthly_actr_90_start: ", monthly_actr_90_start)
	// fmt.Println("784 - MONTHLY TABLE INDEX: ", monthly_actr_90_start, " DATA: ", monthly_sheet_data_rows[monthly_actr_90_start], " FILTER RETURN INDEX: ", monthly_actr_90_start)
	// fmt.Println("787 - BOXI TABLE INDEX: ", boxi_actr_90_header_index, " DATA: ", boxi_sheet_data_rows[boxi_actr_90_header_index])
	for i, boxi_item := range boxi_sheet_data_rows {
		// boxi_actr_90_header_index = i + boxi_actr_90_header_index

		// fmt.Println(boxi_sheet_data_rows[i], " CELL: ", fmt.Sprintf("A%d", monthly_actr_90_start+i))
		// if len(boxi_sheet_data_rows[i]) > 0 {
		// 	for j := range len(boxi_actr_90_header) {

		// 		fmt.Println(fmt.Sprintf("%s%d", excel_columns_alphabet[j], monthly_actr_90_start+i), " ", boxi_sheet_data_rows[i])
		// 		window.cash_allocation_file.SetCellValue(monthly_actr_90_sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[j], monthly_actr_90_start+i), boxi_sheet_data_rows[i][j+1])
		// 	}
		// }
		if len(boxi_item) > 0 {
			if strings.EqualFold(boxi_item[0], "") {
				boxi_item = boxi_item[1:]
			}
			for j := range len(boxi_item) {
				switch j {
				case 0:
					boxi_actr_90_int_amount, _ := strconv.ParseInt(strings.ReplaceAll(boxi_item[j], ",", ""), 0, 64)
					window.cash_allocation_file.SetCellInt(monthly_actr_90_sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[j], monthly_actr_90_start+i), boxi_actr_90_int_amount)
					// boxi_actr_90_float_amount, _ := strconv.ParseFloat(strings.ReplaceAll(boxi_item[j], ",", ""), 64)
					// window.cash_allocation_file.SetCellFloat(monthly_actr_90_sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[j], monthly_actr_90_start+i), boxi_actr_90_float_amount, 2, 64)
				case 2:
					boxi_actr_90_float_amount, _ := strconv.ParseFloat(strings.ReplaceAll(boxi_item[j], ",", ""), 64)
					window.cash_allocation_file.SetCellFloat(monthly_actr_90_sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[j], monthly_actr_90_start+i), boxi_actr_90_float_amount, 2, 64)
				default:
					window.cash_allocation_file.SetCellValue(monthly_actr_90_sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[j], monthly_actr_90_start+i), boxi_item[j])
				}
				// if j == 2 {
				// 	boxi_actr_90_float_amount, _ := strconv.ParseFloat(strings.ReplaceAll(boxi_item[j], ",", ""), 64)
				// 	window.cash_allocation_file.SetCellFloat(monthly_actr_90_sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[j], monthly_actr_90_start+i), boxi_actr_90_float_amount, 2, 64)
				// } else {
				// 	window.cash_allocation_file.SetCellValue(monthly_actr_90_sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[j], monthly_actr_90_start+i), boxi_item[j])
				// }
			}
		}
		// if len(boxi_sheet_data_rows) > 0 {

		// 	// window.cash_allocation_file.SetSheetRow(monthly_actr_90_sheet_name, fmt.Sprintf("A%d", monthly_actr_90_start+i), boxi_item)
		// }
	}
	window.cash_allocation_file.SetCellValue(monthly_actr_90_sheet_name, fmt.Sprintf("A%d", monthly_actr_90_start+1), boxi_actr_90_header[0])
	window.cash_allocation_file.SetCellValue(monthly_actr_90_sheet_name, fmt.Sprintf("C%d", monthly_actr_90_start+1), boxi_actr_90_header[2])
	// monthly_sheet_data_rows, _, _ := read_cell_data(monthly_actr_90_sheet_name, window.cash_allocation_file, false)
	// fmt.Println("755:- BOXI SHEET ACTR90: ", boxi_actr_90_sheet_name, " MONTHLY ACTR90 SHEET: ", monthly_actr_90_sheet_name, " BOXI SHEET ACTR90: ", boxi_actr_90_sheet, " MONTHLY ACTR90 SHEET: ", montly_file_actr_90_sheet)
	// fmt.Println(boxi_sheet_data_rows)
}
func (window *main_data_struct) update_cash_allocation_file_data(sheet_name string, boxi_row_data_slice [][]string, monthly_sheet_slice [][]string, transaction_type_column int, excel_columns_alphabet []string, cell_with_formula int, cell_with_formula_string string, job_code_column_column_index int, account_column_column_index int, amount_column_index int, second_sheet_start_row int, reconcilling_cell string, log_file *os.File) {
	style, err := window.cash_allocation_file.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "fill", Color: []string{"FFFFFF"}, Shading: 1},
	})
	if err != nil {
		log_file.WriteString(fmt.Sprintf("[%s]: Error setting style, error: %s\n", time.Now().Local().Format("02/01/2006 15:04"), err))
	}
	// not_gr_from_filtered_array := filter_multiple_arrays(&boxi_row_data_slice, func(h []string) bool { return strings.ToLower(h[transaction_type_column]) != "gr" })

	// gr_from_filtered_array := filter_multiple_arrays(&boxi_row_data_slice, func(h []string) bool { return strings.ToLower(h[transaction_type_column]) == "gr" })
	if len(sheet_name) < 9 {
		if job_code_column_column_index != -1 {
			slice_of_boxi_data_slice := filter_multiple_arrays(&boxi_row_data_slice, func(s []string) bool {
				return strings.TrimSpace(s[account_column_column_index]) == strings.TrimSpace(sheet_name) && strings.TrimSpace(s[job_code_column_column_index]) == "!"
			})
			if len(slice_of_boxi_data_slice) > 0 {
				window.cash_allocation_file.InsertRows(sheet_name, cell_with_formula, len(slice_of_boxi_data_slice)+1)
				temp_count := 0
				for _, d := range slice_of_boxi_data_slice {
					// monthly_sheet_slice = slices.Insert(monthly_sheet_slice, len(monthly_sheet_slice)-1, temp_array_to_append)
					// fmt.Println("Data: ", d, " SHEET NAME: ", sheet_name)
					if d[0] == "" {
						d = d[1:]
					}
					for i := range len(d) {
						window.cash_allocation_file.SetCellValue(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[i], cell_with_formula+temp_count), d[i])
						window.cash_allocation_file.SetCellStyle(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[i], cell_with_formula+temp_count), fmt.Sprintf("%s%d", excel_columns_alphabet[i], (cell_with_formula+temp_count)), style)
					}
					temp_count += 1
				}

				window.update_formulas(sheet_name, cell_with_formula, excel_columns_alphabet, cell_with_formula_string, temp_count, second_sheet_start_row, reconcilling_cell)
				// window.cash_allocation_file.SetCellFormula(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[amount_column_index-1], (cell_with_formula+temp_count)+1), )
				// log_file.WriteString(fmt.Sprintf("[%s]: %s\n%s", time.Now().Local().Format("02/01/2006 15:04"), slice_of_boxi_data_slice, monthly_sheet_slice))
			}

		} else {
			slice_of_boxi_data_slice := filter_multiple_arrays(&boxi_row_data_slice, func(s []string) bool {
				return strings.TrimSpace(s[account_column_column_index]) == strings.TrimSpace(sheet_name)
			})
			if len(slice_of_boxi_data_slice) > 0 {
				window.cash_allocation_file.InsertRows(sheet_name, cell_with_formula, len(slice_of_boxi_data_slice)+1)

				temp_count := 0
				for _, d := range slice_of_boxi_data_slice {
					// monthly_sheet_slice = slices.Insert(monthly_sheet_slice, len(monthly_sheet_slice)-1, temp_array_to_append)
					// fmt.Println("Data: ", d, " SHEET NAME: ", sheet_name)
					if d[0] == "" {
						d = d[1:]
					}
					for i := range len(d) {
						window.cash_allocation_file.SetCellValue(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[i], cell_with_formula+temp_count), d[i])
						window.cash_allocation_file.SetCellStyle(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[i], cell_with_formula+temp_count), fmt.Sprintf("%s%d", excel_columns_alphabet[i], (cell_with_formula+temp_count)), style)

					}
					temp_count += 1
				}
				window.update_formulas(sheet_name, cell_with_formula, excel_columns_alphabet, cell_with_formula_string, temp_count, second_sheet_start_row, reconcilling_cell)

				// log_file.WriteString(fmt.Sprintf("[%s]: %s\n%s", time.Now().Local().Format("02/01/2006 15:04"), slice_of_boxi_data_slice, monthly_sheet_slice))
			}
		}
		// fmt.Println("SHEET WITH NO EMPTY ROW: ", sheet_name)

	} else {
		// If sheet name length greater than 9 characters long
		if strings.Contains(sheet_name, "-") {
			if job_code_column_column_index != -1 {
				slice_of_boxi_data_slice := filter_multiple_arrays(&boxi_row_data_slice, func(s []string) bool {
					return fmt.Sprintf("%s-%s", strings.TrimSpace(s[account_column_column_index]), strings.TrimSpace(s[job_code_column_column_index])) == strings.TrimSpace(sheet_name)
				})
				if len(slice_of_boxi_data_slice) > 0 {

					window.cash_allocation_file.InsertRows(sheet_name, cell_with_formula, len(slice_of_boxi_data_slice)+1)
					temp_count := 0
					for _, d := range slice_of_boxi_data_slice {
						// monthly_sheet_slice = slices.Insert(monthly_sheet_slice, len(monthly_sheet_slice)-1, temp_array_to_append)
						// fmt.Println("Data: ", d, " SHEET NAME: ", sheet_name)
						if d[0] == "" {
							d = d[1:]
						}
						for i := range len(d) {
							window.cash_allocation_file.SetCellValue(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[i], cell_with_formula+temp_count), d[i])
							window.cash_allocation_file.SetCellStyle(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[i], cell_with_formula+temp_count), fmt.Sprintf("%s%d", excel_columns_alphabet[i], (cell_with_formula+temp_count)), style)
						}
						temp_count += 1
					}
					window.update_formulas(sheet_name, cell_with_formula, excel_columns_alphabet, cell_with_formula_string, temp_count, second_sheet_start_row, reconcilling_cell)
				}

			} else {
				slice_of_boxi_data_slice := filter_multiple_arrays(&boxi_row_data_slice, func(s []string) bool {
					return fmt.Sprintf("%s-%s", strings.TrimSpace(s[account_column_column_index]), strings.TrimSpace(s[3])) == strings.TrimSpace(sheet_name)
				})
				if len(slice_of_boxi_data_slice) > 0 {

					window.cash_allocation_file.InsertRows(sheet_name, cell_with_formula, len(slice_of_boxi_data_slice)+1)
					temp_count := 0
					for _, d := range slice_of_boxi_data_slice {
						// monthly_sheet_slice = slices.Insert(monthly_sheet_slice, len(monthly_sheet_slice)-1, temp_array_to_append)
						// fmt.Println("Data: ", d, " SHEET NAME: ", sheet_name)
						if d[0] == "" {
							d = d[1:]
						}
						for i := range len(d) {
							window.cash_allocation_file.SetCellValue(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[i], cell_with_formula+temp_count), d[i])
							window.cash_allocation_file.SetCellStyle(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[i], cell_with_formula+temp_count), fmt.Sprintf("%s%d", excel_columns_alphabet[i], (cell_with_formula+temp_count)), style)
						}
						temp_count += 1
					}
					window.update_formulas(sheet_name, cell_with_formula, excel_columns_alphabet, cell_with_formula_string, temp_count, second_sheet_start_row, reconcilling_cell)
				}

			}
		} else {
			if job_code_column_column_index != -1 {
				slice_of_boxi_data_slice := filter_multiple_arrays(&boxi_row_data_slice, func(s []string) bool {
					return fmt.Sprintf("%s%s", strings.TrimSpace(s[account_column_column_index]), strings.TrimSpace(s[job_code_column_column_index])) == strings.TrimSpace(sheet_name)
				})
				if len(slice_of_boxi_data_slice) > 0 {

					window.cash_allocation_file.InsertRows(sheet_name, cell_with_formula, len(slice_of_boxi_data_slice)+1)
					temp_count := 0
					for _, d := range slice_of_boxi_data_slice {
						// monthly_sheet_slice = slices.Insert(monthly_sheet_slice, len(monthly_sheet_slice)-1, temp_array_to_append)
						// fmt.Println("Data: ", d, " SHEET NAME: ", sheet_name)
						if d[0] == "" {
							d = d[1:]
						}
						for i := range len(d) {
							window.cash_allocation_file.SetCellValue(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[i], cell_with_formula+temp_count), d[i])
							window.cash_allocation_file.SetCellStyle(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[i], cell_with_formula+temp_count), fmt.Sprintf("%s%d", excel_columns_alphabet[i], (cell_with_formula+temp_count)), style)
						}
						temp_count += 1
					}
					window.update_formulas(sheet_name, cell_with_formula, excel_columns_alphabet, cell_with_formula_string, temp_count, second_sheet_start_row, reconcilling_cell)
				}

			} else {
				slice_of_boxi_data_slice := filter_multiple_arrays(&boxi_row_data_slice, func(s []string) bool {
					return fmt.Sprintf("%s%s", strings.TrimSpace(s[account_column_column_index]), strings.TrimSpace(s[3])) == strings.TrimSpace(sheet_name)
				})
				if len(slice_of_boxi_data_slice) > 0 {

					window.cash_allocation_file.InsertRows(sheet_name, cell_with_formula, len(slice_of_boxi_data_slice)+1)
					temp_count := 0
					for _, d := range slice_of_boxi_data_slice {
						// monthly_sheet_slice = slices.Insert(monthly_sheet_slice, len(monthly_sheet_slice)-1, temp_array_to_append)
						// fmt.Println("Data: ", d, " SHEET NAME: ", sheet_name)
						if d[0] == "" {
							d = d[1:]
						}
						for i := range len(d) {
							window.cash_allocation_file.SetCellValue(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[i], cell_with_formula+temp_count), d[i])
							window.cash_allocation_file.SetCellStyle(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[i], cell_with_formula+temp_count), fmt.Sprintf("%s%d", excel_columns_alphabet[i], (cell_with_formula+temp_count)), style)
						}
						temp_count += 1
					}
					// cell_with_formula_string_before_semi_colon := strings.Split(cell_with_formula_string, ":")
					// fmt.Println("cell_with_formula_string_before_semi_colon")
					// fmt.Println(cell_with_formula_string_before_semi_colon)
					window.update_formulas(sheet_name, cell_with_formula, excel_columns_alphabet, cell_with_formula_string, temp_count, second_sheet_start_row, reconcilling_cell)
				}

			}
		}
	}
}

func read_cell_data_control_accounts(sheet_name string, workbook *excelize.File, boxi bool) ([][]string, string, error) {
	// check for 1 sheet
	list_sheets := workbook.GetSheetList()
	find_trust_in_sheet := filter_single_array(list_sheets, func(s string) bool {
		return s == sheet_name
	})

	find_trust_in_sheet_without_actr := filter_single_array(list_sheets, func(s string) bool {
		return !strings.Contains(strings.ToLower(s), "act")
	})
	// if len(find_trust_in_sheet) == 0 {
	// 	if strings.Contains(strings.ToLower(workbook.GetSheetName(0)), "act") {
	// 		// fmt.Println("WORKBOOK 1: ", workbook.GetSheetName(1))
	// 		data, err := workbook.GetRows(workbook.GetSheetName(1))
	// 		return data, workbook.GetSheetName(1), err
	// 	} else {
	// 		// fmt.Println("WORKBOOK 1: ", workbook.GetSheetName(0))
	// 		data, err := workbook.GetRows(workbook.GetSheetName(0))
	// 		return data, workbook.GetSheetName(0), err
	// 	}
	// } else {
	// 	data, err := workbook.GetRows(sheet_name)
	// 	return data, sheet_name, err
	// }
	if boxi {
		if len(find_trust_in_sheet_without_actr) != 0 {
			index_of_sheet_name := 0
			for i, name := range list_sheets {
				if name == find_trust_in_sheet_without_actr[0] {
					fmt.Println("WORKBOOK 1: ", workbook.GetSheetName(i))
					index_of_sheet_name = i
				}
			}

			data, err := workbook.GetRows(workbook.GetSheetName(index_of_sheet_name))
			return data, workbook.GetSheetName(index_of_sheet_name), err
		} else {
			data, err := workbook.GetRows(workbook.GetSheetName(0))
			return data, workbook.GetSheetName(0), err
		}
	} else {
		if len(find_trust_in_sheet) == 0 {
			if strings.Contains(strings.ToLower(workbook.GetSheetName(0)), "act") {
				// fmt.Println("WORKBOOK 1: ", workbook.GetSheetName(1))
				data, err := workbook.GetRows(workbook.GetSheetName(1))
				return data, workbook.GetSheetName(1), err
			} else {
				// fmt.Println("WORKBOOK 1: ", workbook.GetSheetName(0))
				data, err := workbook.GetRows(workbook.GetSheetName(0))
				return data, workbook.GetSheetName(0), err
			}
		} else {
			data, err := workbook.GetRows(sheet_name)
			return data, sheet_name, err
		}
	}
}

func (window *main_data_struct) update_formulas(sheet_name string, cell_formula_location int, excel_columns_alphabet []string, cell_formula_string string, temp_count int, second_sheet_start_row int, reconcilling_cell string) {
	cell_with_formula_string_before_semi_colon := strings.Split(cell_formula_string, ":")

	if len(cell_with_formula_string_before_semi_colon) == 1 {
		regex_get_cell := regexp.MustCompile("[A-z][0-9]+")
		cell_with_formula_string_before_semi_colon[0] = strings.Replace(cell_with_formula_string_before_semi_colon[0], ")", "", 1)
		cell_with_formula_string_before_semi_colon = append(cell_with_formula_string_before_semi_colon, fmt.Sprintf("%s)", regex_get_cell.FindAllString(cell_formula_string, -1)[0]))
	}
	// fmt.Println("cell_with_formula_string_before_semi_colon")
	// fmt.Println(cell_with_formula_string_before_semi_colon)
	regex_get_column := regexp.MustCompile("[A-z]+")

	regex_get_row_number := regexp.MustCompile("[0-9]+")
	num_convert_from_formula, err := strconv.Atoi(regex_get_row_number.FindAllString(cell_with_formula_string_before_semi_colon[1], -1)[0])
	if err != nil {
		panic("Can't convert to int")
	}
	// fmt.Println(regex_get_column.FindAllString(cell_with_formula_string_before_semi_colon[1], -1))
	temp_count_row := (num_convert_from_formula + temp_count)
	fmt.Println("cell_with_formula_string_before_semi_colon AFTER")
	fmt.Println(cell_with_formula_string_before_semi_colon)
	formula_cell_string := fmt.Sprintf("%s%d)", regex_get_column.FindAllString(cell_with_formula_string_before_semi_colon[1], -1)[0], temp_count_row)
	fmt.Println("LOCATION: ", fmt.Sprintf("%s:%s", cell_with_formula_string_before_semi_colon[0], formula_cell_string), " SHEET: ", sheet_name, " TEMP COUNT: ", temp_count_row+1, " Location to update: ", fmt.Sprintf("%s%d", regex_get_column.FindAllString(cell_with_formula_string_before_semi_colon[1], -1)[0], temp_count_row+1))
	fmt.Println("924: ", num_convert_from_formula, " TEMP COUNT ROW: ", temp_count_row, " CELL LOCATION + TEMP COUNT: ", (cell_formula_location+temp_count)+1, " CELL FORMULA LOCATION: ", cell_formula_location, " CELL FORMULA STRING: ", cell_formula_string, " CELL: ", fmt.Sprintf("%s%d", regex_get_column.FindAllString(cell_with_formula_string_before_semi_colon[1], -1)[0], (cell_formula_location+temp_count)+1), " FORMULA: ", fmt.Sprintf("%s:%s", cell_with_formula_string_before_semi_colon[0], formula_cell_string))
	window.cash_allocation_file.SetCellFormula(sheet_name, fmt.Sprintf("%s%d", regex_get_column.FindAllString(cell_with_formula_string_before_semi_colon[1], -1)[0], (cell_formula_location+temp_count)+1), fmt.Sprintf("%s:%s", cell_with_formula_string_before_semi_colon[0], formula_cell_string))
	window.cash_allocation_file.SetCellFormula(sheet_name, reconcilling_cell, fmt.Sprintf("%s:%s", cell_with_formula_string_before_semi_colon[0], formula_cell_string))
	// window.cash_allocation_file.SetCellFormula(sheet_name, fmt.Sprintf("%s%d", excel_columns_alphabet[amount_column_index-1], (cell_with_formula+temp_count)+1), )

}

// Generic functions

func filter_single_array[T any](s_array []T, compare_func func(T) bool) (return_array []T) {
	for _, s := range s_array {
		if compare_func(s) {
			return_array = append(return_array, s)
		}
	}
	return
}
func filter_multiple_arrays[T any](s_array *[][]T, compare_func func([]T) bool) (return_array [][]T) {
	for _, s := range *s_array {
		if compare_func(s) {
			return_array = append(return_array, s)
		}
	}
	return
}
func filter_multiple_arrays_return_index[T any](s_array *[][]T, compare_func func([]T) bool) int {
	for i, s := range *s_array {
		if compare_func(s) {
			return i
		}
	}
	return 0
}
func filter_multiple_arrays_return_array_of_indexes[T any](s_array *[][]T, compare_func func([]T) bool) []int {
	var list_of_indexes []int
	for i, s := range *s_array {
		if compare_func(s) {
			list_of_indexes = append(list_of_indexes, i)
		}
	}
	return list_of_indexes
}
func filter_single_arrays_return_index[T any](s_array *[]T, compare_func func(T) bool) int {
	for i, s := range *s_array {
		if compare_func(s) {
			return i
		}
	}
	return 0
}
func read_single_sheet_data(sheet_name string, workbook *excelize.File) ([][]string, error) {
	// check for 1 sheet
	list_sheets := workbook.GetSheetList()
	index_of_sheet_name := 0
	for i, name := range list_sheets {
		if name == sheet_name {
			index_of_sheet_name = i
			break
		}
	}

	data, err := workbook.GetRows(workbook.GetSheetName(index_of_sheet_name))
	return data, err
}
func convert_to_object_data(headings []string, data_single_array [][]string) mapped_data_object {
	var new_data_object = make(mapped_data_object, len(data_single_array))
	// var headers_map = make(map[string]interface{}, len(data_single_array[0]))
	// fmt.Println(data_single_array[0])
	// for j, data := range data_single_array[i] {
	// 	heading := headings[j]
	// 	headers_map[heading] = data
	// 	fmt.Println("HEADING: ", heading, " DATA: ", data)
	// }
	for i := range data_single_array {
		var headers_map = make(map[string]interface{}, len(data_single_array[i]))
		// fmt.Println(data_single_array[i])
		if len(data_single_array[i]) > len(headings) {
			for i := range len(data_single_array[i]) {
				headings = append(headings, fmt.Sprintf("null_%d", i))
			}
		}
		for j, data := range data_single_array[i] {
			// heading := strings.ToLower(strings.ReplaceAll(strings.Replace(strings.Replace(strings.TrimSpace(headings[j]), "/", "", 1), "#", "", 1), " ", "_"))
			if headings[j] != "" {
				var current_data_item string = data
				if len(current_data_item) == 0 {
					current_data_item = "NULL"
					// fmt.Println("DATA: ", data, " CURRENT ITEM: ", current_data_item)
				}
				match_amount, _ := regexp.MatchString("(^[0-9]{1,}\\.[0-9]{2}$)", strings.TrimSpace(strings.ReplaceAll(current_data_item, ",", "")))
				// fmt.Println(match)
				if match_amount {
					current_data_item = strings.TrimSpace(strings.ReplaceAll(current_data_item, ",", ""))
				}
				headers_map[strings.ToLower(strings.ReplaceAll(strings.Replace(strings.Replace(strings.TrimSpace(headings[j]), "/", "", 1), "#", "", 1), " ", "_"))] = strings.TrimSpace(current_data_item)
			}
		}
		// fmt.Println(headers_map)
		new_data_object[i] = headers_map
	}
	fmt.Println("new_data_object")
	fmt.Println(new_data_object)
	return new_data_object
}
func set_background_rect_colour(graphical_context layout.Context, size image.Point, colour color.NRGBA) layout.Dimensions {
	defer clip.Rect{Max: size}.Push(graphical_context.Ops).Pop()
	paint.ColorOp{Color: colour}.Add(graphical_context.Ops)
	paint.PaintOp{}.Add(graphical_context.Ops)
	return layout.Dimensions{Size: size}
}

func print_type[T any](value T) {
	fmt.Printf("The type of the value is %T\n", value)
}

// Layouts

func left_side_bar(theme *material.Theme, graphical_context layout.Context, buttons []widget.Clickable, button_text []string, colours_list []color.NRGBA) layout.Dimensions {
	buttons_flex_children := []layout.FlexChild{}
	for i := range buttons {
		buttons_flex_children = append(buttons_flex_children, layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
			return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
				// Here we set the minimum constraints to zero. This allows our button to be smaller
				// than the entire right-side pane in the UI. Without this change, the button is forced
				// to occupy all of the space.
				// graphical_context.Constraints.Min = image.Point{}
				// Here we inset the button a little bit on all sides.
				button := material.Button(theme, &buttons[i], button_text[i])
				button.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
				button.Background = colours_list[light_grey]
				button.TextSize = unit.Sp(12)
				return layout.UniformInset(5).Layout(graphical_context,
					button.Layout,
				)
			})

		}))
	}

	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceEnd,
	}.Layout(graphical_context,
		// layout.Flexed(2, func(graphical_context layout.Context) layout.Dimensions {
		// 	// This returns an empty left-hand pane.
		// 	return layout.Dimensions{Size: graphical_context.Constraints.Max}
		// }),
		buttons_flex_children...,
	)
}
func (window *main_data_struct) control_accounts_layout(graphical_context layout.Context, design_style *material.Theme, buttons []widget.Clickable, show_options *bool, dropdown *dropdown_struct, dropdown_menu_enum *widget.Enum) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(graphical_context,
		layout.Rigid(
			func(graphical_context layout.Context) layout.Dimensions {
				if len(window.trusts) > 0 {
					margins := layout.Inset{
						Top:    unit.Dp(25),
						Bottom: unit.Dp(0),
						Right:  unit.Dp(25),
						Left:   unit.Dp(25),
					}
					return margins.Layout(
						graphical_context,
						func(graphical_context layout.Context) layout.Dimensions {
							button := material.Button(design_style, &buttons[0], window.trust)

							button.Background = color.NRGBA{R: 232, G: 52, B: 5, A: 255}
							return button.Layout(graphical_context)
						},
					)
				} else {
					return layout.Dimensions{}
				}
			},
		),
		layout.Rigid(
			func(graphical_context layout.Context) layout.Dimensions {

				return layout.Flex{Axis: layout.Vertical}.Layout(
					graphical_context,
					layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
						// input := material.List(design_style, &widget_list)
						if *show_options {
							margins := layout.Inset{
								Top:    unit.Dp(0),
								Bottom: unit.Dp(0),
								Right:  unit.Dp(25),
								Left:   unit.Dp(25),
							}
							return margins.Layout(
								graphical_context,
								func(graphical_context layout.Context) layout.Dimensions {
									return material.List(design_style, &dropdown.list).Layout(graphical_context, len(dropdown.items), func(graphical_context layout.Context, i int) layout.Dimensions {
										// item := &dropdown.items[i]
										// button := material.Button(design_style, new(widget.Clickable), *item)
										list_radio_boxes := material.RadioButton(design_style, dropdown_menu_enum, dropdown.items[i].Name, dropdown.items[i].Trust)
										list_radio_boxes.Size = unit.Dp(12)
										// window.trust = dropdown_menu_enum.Value

										if dropdown_menu_enum.Update(graphical_context) {
											// window.information = fmt.Sprintf("Trust: %s", dropdown_menu_enum.Value)
											trust_selected := filter_single_arrays_return_index(&window.trusts, func(item trust_data_struct) bool {
												return item.Name == dropdown_menu_enum.Value
											})
											dropdown.trust_selected = window.trusts[trust_selected]
											window.trust = dropdown.trust_selected.Trust
											// window.information = fmt.Sprintf("Trust: %s", dropdown.trust_selected.Trust)
											*show_options = false
										}
										return list_radio_boxes.Layout(graphical_context)
									})
								},
							)
						}
						return layout.Dimensions{}
					}),
				)
			},
		),
		layout.Flexed(1,
			func(graphical_context layout.Context) layout.Dimensions {
				margins := layout.Inset{
					Top:    unit.Dp(12),
					Bottom: unit.Dp(0),
					Right:  unit.Dp(25),
					Left:   unit.Dp(25),
				}
				return margins.Layout(
					graphical_context,
					func(graphical_context layout.Context) layout.Dimensions {
						button := material.Button(design_style, &buttons[1], "Load BOXI File")
						// button.Color = color.NRGBA{R: 76, G: 87, B: 96, A: 255}
						button.Background = color.NRGBA{R: 190, G: 75, B: 4, A: 255}
						return button.Layout(graphical_context)
					},
				)
			},
		),
		layout.Flexed(1,
			func(graphical_context layout.Context) layout.Dimensions {
				margins := layout.Inset{
					Top:    unit.Dp(10),
					Bottom: unit.Dp(0),
					Right:  unit.Dp(25),
					Left:   unit.Dp(25),
				}
				return margins.Layout(
					graphical_context,
					func(graphical_context layout.Context) layout.Dimensions {
						button := material.Button(design_style, &buttons[2], "Load Month File")
						button.Background = color.NRGBA{R: 190, G: 128, B: 4, A: 255}
						return button.Layout(graphical_context)
					},
				)
			},
		),
	)
}
func (window *main_data_struct) bank_rec_layout(graphical_context layout.Context, design_style *material.Theme, buttons []widget.Clickable, show_options *bool, dropdown *dropdown_struct, dropdown_menu_enum *widget.Enum) layout.Dimensions {
	// editor *widget.Editor
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(graphical_context,
		layout.Rigid(
			func(graphical_context layout.Context) layout.Dimensions {
				if len(window.trusts) > 0 {
					margins := layout.Inset{
						Top:    unit.Dp(25),
						Bottom: unit.Dp(0),
						Right:  unit.Dp(25),
						Left:   unit.Dp(25),
					}
					return margins.Layout(
						graphical_context,
						func(graphical_context layout.Context) layout.Dimensions {
							button := material.Button(design_style, &buttons[0], window.trust)

							button.Background = color.NRGBA{R: 232, G: 52, B: 5, A: 255}
							return button.Layout(graphical_context)
						},
					)
				} else {
					return layout.Dimensions{}
				}
			},
		),
		layout.Rigid(
			func(graphical_context layout.Context) layout.Dimensions {

				return layout.Flex{Axis: layout.Vertical}.Layout(
					graphical_context,
					layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
						// input := material.List(design_style, &widget_list)
						if *show_options {
							margins := layout.Inset{
								Top:    unit.Dp(0),
								Bottom: unit.Dp(0),
								Right:  unit.Dp(25),
								Left:   unit.Dp(25),
							}
							return margins.Layout(
								graphical_context,
								func(graphical_context layout.Context) layout.Dimensions {
									return material.List(design_style, &dropdown.list).Layout(graphical_context, len(dropdown.items), func(graphical_context layout.Context, i int) layout.Dimensions {
										// item := &dropdown.items[i]
										// button := material.Button(design_style, new(widget.Clickable), *item)
										list_radio_boxes := material.RadioButton(design_style, dropdown_menu_enum, dropdown.items[i].Name, dropdown.items[i].Trust)
										list_radio_boxes.Size = unit.Dp(12)
										// window.trust = dropdown_menu_enum.Value

										if dropdown_menu_enum.Update(graphical_context) {
											// window.information = fmt.Sprintf("Trust: %s", dropdown_menu_enum.Value)
											trust_selected := filter_single_arrays_return_index(&window.trusts, func(item trust_data_struct) bool {
												return item.Name == dropdown_menu_enum.Value
											})
											dropdown.trust_selected = window.trusts[trust_selected]
											window.trust = dropdown.trust_selected.Trust
											// window.information = fmt.Sprintf("Trust: %s", dropdown.trust_selected.Trust)
											*show_options = false
										}
										return list_radio_boxes.Layout(graphical_context)
									})
								},
							)
						}
						return layout.Dimensions{}
					}),
				)
			},
		),
		layout.Rigid(
			func(graphical_context layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(graphical_context,
					layout.Flexed(0.5, func(graphical_context layout.Context) layout.Dimensions {
						margins := layout.Inset{
							Top:    unit.Dp(12),
							Bottom: unit.Dp(0),
							Right:  unit.Dp(5),
							Left:   unit.Dp(25),
						}
						return margins.Layout(
							graphical_context,
							func(graphical_context layout.Context) layout.Dimensions {
								button := material.Button(design_style, &buttons[1], "Load BOXI File")
								// button.Color = color.NRGBA{R: 76, G: 87, B: 96, A: 255}
								button.Background = color.NRGBA{R: 190, G: 75, B: 4, A: 255}
								button.Inset = layout.Inset{
									Top:    unit.Dp(50),
									Bottom: unit.Dp(50),
									Left:   unit.Dp(25),
									Right:  unit.Dp(25),
								}
								button.TextSize = unit.Sp(15)
								return button.Layout(graphical_context)
							},
						)
					}),
					layout.Flexed(0.5, func(graphical_context layout.Context) layout.Dimensions {
						margins := layout.Inset{
							Top:    unit.Dp(12),
							Bottom: unit.Dp(0),
							Right:  unit.Dp(25),
							Left:   unit.Dp(5),
						}
						return margins.Layout(
							graphical_context,
							func(graphical_context layout.Context) layout.Dimensions {
								button := material.Button(design_style, &buttons[2], "Load Cash Allocation File")
								button.Background = color.NRGBA{R: 190, G: 128, B: 4, A: 255}
								button.Inset = layout.Inset{
									Top:    unit.Dp(50),
									Bottom: unit.Dp(50),
									Left:   unit.Dp(25),
									Right:  unit.Dp(25),
								}
								button.TextSize = unit.Sp(15)
								return button.Layout(graphical_context)
							},
						)
					}),
				)

			},
		),
		layout.Rigid(
			func(graphical_context layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(graphical_context,
					layout.Flexed(0.5, func(graphical_context layout.Context) layout.Dimensions {
						margins := layout.Inset{
							Top:    unit.Dp(10),
							Bottom: unit.Dp(0),
							Right:  unit.Dp(5),
							Left:   unit.Dp(25),
						}
						return margins.Layout(
							graphical_context,
							func(graphical_context layout.Context) layout.Dimensions {
								button := material.Button(design_style, &buttons[3], "Load Mapping Table")
								button.Background = color.NRGBA{R: 220, G: 128, B: 150, A: 255}
								button.Inset = layout.Inset{
									Top:    unit.Dp(50),
									Bottom: unit.Dp(50),
									Left:   unit.Dp(25),
									Right:  unit.Dp(25),
								}
								button.TextSize = unit.Sp(15)
								return button.Layout(graphical_context)
							},
						)
					}),
					layout.Flexed(0.5, func(graphical_context layout.Context) layout.Dimensions {
						margins := layout.Inset{
							Top:    unit.Dp(10),
							Bottom: unit.Dp(0),
							Right:  unit.Dp(25),
							Left:   unit.Dp(5),
						}
						return margins.Layout(
							graphical_context,
							func(graphical_context layout.Context) layout.Dimensions {
								button := material.Button(design_style, &buttons[4], "Load Bank Rec File")
								button.Background = color.NRGBA{R: 120, G: 150, B: 180, A: 255}
								button.Inset = layout.Inset{
									Top:    unit.Dp(50),
									Bottom: unit.Dp(50),
									Left:   unit.Dp(25),
									Right:  unit.Dp(25),
								}
								button.TextSize = unit.Sp(15)
								return button.Layout(graphical_context)
							},
						)
					}),
				)

			},
		),
	)
}
func (window *main_data_struct) sms_layout(graphical_context layout.Context, design_style *material.Theme, text_inputs []widget.Editor) layout.Dimensions {
	buttons_flex_children := []layout.FlexChild{}
	button_text := []string{"Add Phone Number", "Add Sender", "Add Message"}
	for i := range text_inputs {
		buttons_flex_children = append(buttons_flex_children,
			layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
				text_input_sms := material.Editor(design_style, &text_inputs[i], button_text[i])
				text_input_sms.SelectionColor = colours_list[light_grey]
				text_image_point_size := layout.Inset{
					Top:    unit.Dp(25),
					Left:   unit.Dp(25),
					Right:  unit.Dp(10),
					Bottom: unit.Dp(25),
				}.Layout(graphical_context, text_input_sms.Layout).Size
				set_background_rect_colour(graphical_context, text_image_point_size, colours_list[lighter_grey])

				paint.FillShape(graphical_context.Ops, colours_list[light_grey], clip.Stroke{
					Path:  clip.Rect{Max: text_image_point_size}.Path(),
					Width: 2.5,
				}.Op())
				return layout.Inset{
					Top:    unit.Dp(25),
					Left:   unit.Dp(25),
					Right:  unit.Dp(10),
					Bottom: unit.Dp(25),
				}.Layout(graphical_context, text_input_sms.Layout)
				// button := material.Button(design_style, &buttons[i], fmt.Sprintf("Button %d", i+1))
				// button.Background = colours_list[i+1]
				// return layout.UniformInset(5).Layout(graphical_context,
				// 	button.Layout,
				// ),
			}),
		)
	}
	// fmt.Println(buttons_flex_children)
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(graphical_context,
		// layout.Flexed(2, func(graphical_context layout.Context) layout.Dimensions {
		// 	// This returns an empty left-hand pane.
		// 	return layout.Dimensions{Size: graphical_context.Constraints.Max}
		// }),
		buttons_flex_children...,
	)

}
func (window *main_data_struct) text_input_layouts(graphical_context layout.Context, design_style *material.Theme, text_inputs []widget.Editor, button_text []string) layout.Dimensions {
	buttons_flex_children := []layout.FlexChild{}
	// button_text := []string{"Add Email Address", "Add Message"}
	for i := range text_inputs {
		buttons_flex_children = append(buttons_flex_children,
			layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
				text_input_sms := material.Editor(design_style, &text_inputs[i], button_text[i])
				text_input_sms.SelectionColor = colours_list[light_grey]
				text_image_point_size := layout.Inset{
					Top:    unit.Dp(25),
					Left:   unit.Dp(25),
					Right:  unit.Dp(10),
					Bottom: unit.Dp(25),
				}.Layout(graphical_context, text_input_sms.Layout).Size
				set_background_rect_colour(graphical_context, text_image_point_size, colours_list[lighter_grey])

				paint.FillShape(graphical_context.Ops, colours_list[light_grey], clip.Stroke{
					Path:  clip.Rect{Max: text_image_point_size}.Path(),
					Width: 2.5,
				}.Op())
				return layout.Inset{
					Top:    unit.Dp(25),
					Left:   unit.Dp(25),
					Right:  unit.Dp(10),
					Bottom: unit.Dp(25),
				}.Layout(graphical_context, text_input_sms.Layout)
				// button := material.Button(design_style, &buttons[i], fmt.Sprintf("Button %d", i+1))
				// button.Background = colours_list[i+1]
				// return layout.UniformInset(5).Layout(graphical_context,
				// 	button.Layout,
				// ),
			}),
		)
	}
	// fmt.Println(buttons_flex_children)
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(graphical_context,
		// layout.Flexed(2, func(graphical_context layout.Context) layout.Dimensions {
		// 	// This returns an empty left-hand pane.
		// 	return layout.Dimensions{Size: graphical_context.Constraints.Max}
		// }),
		buttons_flex_children...,
	)

}
func right_side_layout(theme *material.Theme, graphical_context layout.Context, buttons []widget.Clickable, colours_list []color.NRGBA) layout.Dimensions {
	buttons_flex_children := []layout.FlexChild{}
	for i := range buttons {
		buttons_flex_children = append(buttons_flex_children, layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {

			return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
				// Here we set the minimum constraints to zero. This allows our button to be smaller
				// than the entire right-side pane in the UI. Without this change, the button is forced
				// to occupy all of the space.
				// graphical_context.Constraints.Min = image.Point{}
				// Here we inset the button a little bit on all sides.
				button := material.Button(theme, &buttons[i], fmt.Sprintf("Button %d", i+1))
				button.Background = colours_list[i+1]
				return layout.UniformInset(5).Layout(graphical_context,
					button.Layout,
				)
			})

		}))
	}
	// fmt.Println(buttons_flex_children)
	return layout.Flex{}.Layout(graphical_context,
		// layout.Flexed(2, func(graphical_context layout.Context) layout.Dimensions {
		// 	// This returns an empty left-hand pane.
		// 	return layout.Dimensions{Size: graphical_context.Constraints.Max}
		// }),
		buttons_flex_children...,
	)
}
func right_side_layout_one(theme *material.Theme, graphical_context layout.Context, buttons []widget.Clickable, colours_list []color.NRGBA) layout.Dimensions {
	buttons_flex_children := []layout.FlexChild{}
	for i := range buttons {
		buttons_flex_children = append(buttons_flex_children, layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {

			return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
				// Here we set the minimum constraints to zero. This allows our button to be smaller
				// than the entire right-side pane in the UI. Without this change, the button is forced
				// to occupy all of the space.
				// graphical_context.Constraints.Min = image.Point{}
				// Here we inset the button a little bit on all sides.
				button := material.Button(theme, &buttons[i], fmt.Sprintf("Button %d", i+1))
				button.Background = colours_list[i+1]
				return layout.UniformInset(5).Layout(graphical_context,
					button.Layout,
				)
			})

		}))
	}
	// fmt.Println(buttons_flex_children)
	return layout.Flex{}.Layout(graphical_context,
		// layout.Flexed(2, func(graphical_context layout.Context) layout.Dimensions {
		// 	// This returns an empty left-hand pane.
		// 	return layout.Dimensions{Size: graphical_context.Constraints.Max}
		// }),
		buttons_flex_children...,
	)
}

func (window *main_data_struct) boxi_layout(graphical_context layout.Context, design_style *material.Theme, buttons []widget.Clickable, show_options bool, dropdown *dropdown_struct, dropdown_menu_enum *widget.Enum) layout.Dimensions {
	return layout.Flex{
		Axis: layout.Vertical,
	}.Layout(graphical_context,
		layout.Rigid(
			func(graphical_context layout.Context) layout.Dimensions {
				if len(window.trusts) > 0 {
					margins := layout.Inset{
						Top:    unit.Dp(25),
						Bottom: unit.Dp(0),
						Right:  unit.Dp(25),
						Left:   unit.Dp(25),
					}
					return margins.Layout(
						graphical_context,
						func(graphical_context layout.Context) layout.Dimensions {
							button := material.Button(design_style, &buttons[0], window.trust)

							button.Background = color.NRGBA{R: 232, G: 52, B: 5, A: 255}
							return button.Layout(graphical_context)
						},
					)
				} else {
					return layout.Dimensions{}
				}
			},
		),
		layout.Rigid(
			func(graphical_context layout.Context) layout.Dimensions {

				return layout.Flex{Axis: layout.Vertical}.Layout(
					graphical_context,
					layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
						// input := material.List(design_style, &widget_list)
						if show_options {
							margins := layout.Inset{
								Top:    unit.Dp(0),
								Bottom: unit.Dp(0),
								Right:  unit.Dp(25),
								Left:   unit.Dp(25),
							}
							return margins.Layout(
								graphical_context,
								func(graphical_context layout.Context) layout.Dimensions {
									return material.List(design_style, &dropdown.list).Layout(graphical_context, len(dropdown.items), func(graphical_context layout.Context, i int) layout.Dimensions {
										// item := &dropdown.items[i]
										// button := material.Button(design_style, new(widget.Clickable), *item)
										list_radio_boxes := material.RadioButton(design_style, dropdown_menu_enum, dropdown.items[i].Name, dropdown.items[i].Trust)
										list_radio_boxes.Size = unit.Dp(12)
										// window.trust = dropdown_menu_enum.Value

										if dropdown_menu_enum.Update(graphical_context) {
											// window.information = fmt.Sprintf("Trust: %s", dropdown_menu_enum.Value)
											trust_selected := filter_single_arrays_return_index(&window.trusts, func(item trust_data_struct) bool {
												return item.Name == dropdown_menu_enum.Value
											})
											dropdown.trust_selected = window.trusts[trust_selected]
											window.trust = dropdown.trust_selected.Trust
											// window.information = fmt.Sprintf("Trust: %s", dropdown.trust_selected.Trust)
											show_options = false
										}
										return list_radio_boxes.Layout(graphical_context)
									})
								},
							)
						}
						return layout.Dimensions{}
					}),
				)
			},
		),
		layout.Flexed(1,
			func(graphical_context layout.Context) layout.Dimensions {
				margins := layout.Inset{
					Top:    unit.Dp(12),
					Bottom: unit.Dp(0),
					Right:  unit.Dp(25),
					Left:   unit.Dp(25),
				}
				return margins.Layout(
					graphical_context,
					func(graphical_context layout.Context) layout.Dimensions {
						button := material.Button(design_style, &buttons[1], "Load BOXI File")
						// button.Color = color.NRGBA{R: 76, G: 87, B: 96, A: 255}
						button.Background = color.NRGBA{R: 190, G: 75, B: 4, A: 255}
						return button.Layout(graphical_context)
					},
				)
			},
		),
		layout.Flexed(1,
			func(graphical_context layout.Context) layout.Dimensions {
				margins := layout.Inset{
					Top:    unit.Dp(10),
					Bottom: unit.Dp(0),
					Right:  unit.Dp(25),
					Left:   unit.Dp(25),
				}
				return margins.Layout(
					graphical_context,
					func(graphical_context layout.Context) layout.Dimensions {
						button := material.Button(design_style, &buttons[2], "Load Month File")
						button.Background = color.NRGBA{R: 190, G: 128, B: 4, A: 255}
						return button.Layout(graphical_context)
					},
				)
			},
		),
	)
}

// API CALLS
type data_rest_api_send_struct struct {
	To      int    `json:"to"`
	From    string `json:"from"`
	Message string `json:"msg"`
}
type voodoo_response_struct struct {
	Count      int8                             `json:"count"`
	Messages   []voodoo_response_message_struct `json:"messages"`
	Originator string                           `json:"originator"`
	Body       string                           `json:"body"`
	Credits    int                              `json:"credits"`
	Balance    int                              `json:"balance"`
}
type voodoo_response_message_struct struct {
	Id        string `json:"id"`
	Recipient int    `json:"recipient"`
	Status    string `json:"status"`
}

func (state *main_data_struct) send_sms_api_call() (status string, response voodoo_response_struct) {
	if strings.EqualFold(state.text_input_array[0].text_inputted[0:1], "0") {
		state.text_input_array[0].text_inputted = fmt.Sprintf("%s%s", "44", state.text_input_array[0].text_inputted[1:])
	}
	convert_first_input_to_int, _ := strconv.ParseInt(state.text_input_array[0].text_inputted, 0, 64)
	voodoo_sms_data := data_rest_api_send_struct{
		To:      int(convert_first_input_to_int),
		From:    state.text_input_array[1].text_inputted,
		Message: state.text_input_array[2].text_inputted,
	}
	url := "https://api.voodoosms.com/sendsms"

	// {
	// 	"to": 447000123890,
	// 	"from": "VoodooSMS",
	// 	"msg": "Testing The API",
	// 	"schedule": "3 weeks",
	// 	"external_reference": "Testing VoodooSMS",
	// 	"sandbox": true
	// }

	// json.NewEncoder(os.Stdout).Encode(`{"to":447792596594, "from: "Adnan", "msg": "Test"}`)
	data_marshal, _ := json.Marshal(voodoo_sms_data)
	// var jsonStr = []byte(`{"to":447792596594, "from: "Adnan", "msg": "Test"}`)
	// var jsonStr []byte
	// updated_slice := fmt.Appendf(jsonStr, `{"to":447792596594, "from: "%s", "msg": "%s", "sandbox": true,"external_reference": "Testing VoodooSMS"}`, state.text_input_array[0], state.text_input_array[1], state.text_input_array[2])
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(data_marshal))
	if err != nil {
		panic(err)
	}
	// request.Header.Set("X-Custom-Header", "myvalue")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", voodoo_api_key)
	client := &http.Client{}
	response_voodoo, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response_voodoo.Body.Close()

	fmt.Println("response Status:", response_voodoo.Status)
	fmt.Println("response Headers:", response_voodoo.Header)
	body, _ := io.ReadAll(response_voodoo.Body)
	fmt.Println("response Body:", string(body))
	// "count":1,
	// "originator":"Adnan",
	// "body":"Test message to send",
	// "scheduledDateTime":null,
	// "credits":1,
	// "balance":21,
	// "messages":[{"id":"126258140222409536526908414937125121122510220","recipient":447792596594,"reference":null,"status":"PENDING_SENT"}]
	var voodoo_data_struct voodoo_response_struct
	json.Unmarshal(body, &voodoo_data_struct)
	return response_voodoo.Status, voodoo_data_struct
}

type api_token_struct struct {
	AccessToken string `json:"access_token"`
}

type email_api_struct struct {
	Message message_email_api_struct `json:"message"`
}
type message_email_api_struct struct {
	Subject    string                        `json:"subject"`
	Body       body_email_api_struct         `json:"body"`
	Recipients []recipients_email_api_struct `json:"toRecipients"`
}
type body_email_api_struct struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}
type recipients_email_api_struct struct {
	Email email_addresses_api_struct `json:"emailAddress"`
}
type email_addresses_api_struct struct {
	Address string `json:"address"`
}

func (state *main_data_struct) get_access_token_graph_api_call() (access_token string) {
	graph_token_url := "https://login.microsoft.com/{TENANTID}/oauth2/v2.0/token"

	graph_token_data := url.Values{}
	graph_token_data.Add("client_id", client_id)
	graph_token_data.Add("client_secret", client_secret)
	graph_token_data.Add("scope", "https://graph.microsoft.com/.default")
	graph_token_data.Add("grant_type", "client_credentials")
	request, err := http.NewRequest("POST", graph_token_url, strings.NewReader(graph_token_data.Encode()))
	// request.Header.Set("X-Custom-Header", "myvalue")
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	client := &http.Client{}
	response_access_token_api, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response_access_token_api.Body.Close()

	fmt.Println("response Status:", response_access_token_api.Status)
	fmt.Println("response Headers:", response_access_token_api.Header)
	body, _ := io.ReadAll(response_access_token_api.Body)
	var graph_api_token api_token_struct
	json.Unmarshal(body, &graph_api_token)
	// fmt.Println("response Body:", string(body))
	fmt.Println("Token:", graph_api_token.AccessToken)
	return graph_api_token.AccessToken
}
func (state *main_data_struct) send_email_office_api_call(access_token string) (status string) {

	url := fmt.Sprintf("%susers/NoReply@ELFSInvoicePortaloutlook.onmicrosoft.com/sendMail", graph_base_uri)

	var email_data_structure email_api_struct
	email_data_structure.Message.Subject = state.text_input_array[1].text_inputted
	email_data_structure.Message.Body.ContentType = "HTML"
	email_data_structure.Message.Body.Content = fmt.Sprintln("<style>body {font-family: Century Gothic;padding: 0;margin: 0;box-sizing: border-box;}.div {background-color: rgb(", colours_list[dark_blue].R, ",", colours_list[dark_blue].G, ",", colours_list[dark_blue].B, ");text-align: center;color: #fff;border-radius: 5px 5px 0 0;margin-bottom: 0;display: flex;align-items:center;}img {width:115px;background-color: rgb(215,212,212) !important;border-radius: 5px;}.container {padding: 10px 20px;color: #333;border-radius: 0 0 5px 5px;}.inner {margin-top: 0;background-color: #D3D3D3;padding: 20px;}a {text-decoration: none;color: #fc7825;font-weight: 800;transition: 250ms ease-in-out;}a:hover {transition: 250ms ease-in-out;color: #e3063D}</style><div class='container'><div class='div'><img style='' src='", base_64_image, "'  /><h1 style='width:75%'>", state.text_input_array[1].text_inputted, "</h1></div><div class='inner'>", strings.ReplaceAll(state.text_input_array[2].text_inputted, "\n", "<br/>"), "</div>")
	var emails_array []string
	if strings.Contains(state.text_input_array[0].text_inputted, " ") {
		emails_array = strings.Split(state.text_input_array[0].text_inputted, " ")
	} else if strings.Contains(state.text_input_array[0].text_inputted, ",") {
		emails_array = strings.Split(state.text_input_array[0].text_inputted, ",")
	} else if strings.Contains(state.text_input_array[0].text_inputted, ";") {
		emails_array = strings.Split(state.text_input_array[0].text_inputted, ";")
	} else {
		emails_array = append(emails_array, state.text_input_array[0].text_inputted)
	}
	// var emails_array []string = strings.Split(state.text_input_array[0].text_inputted, ";")
	for i := range emails_array {
		email_data_structure.Message.Recipients = append(email_data_structure.Message.Recipients, recipients_email_api_struct{
			Email: email_addresses_api_struct{
				Address: strings.ReplaceAll(strings.ReplaceAll(emails_array[i], ";", ""), ",", ""),
			},
		})
	}

	// <h3>Test HEADING</h3>
	// <p>This is a paragraph</p>
	// <a href="https://google.co.uk">Google</a>
	// <h4 style='color:darkred;text-align:center'>Test HEADING 4 with Styling</h4>
	// <p>This is a paragraph Part 2</p>
	// <a href="https://google.co.uk" style='background-color:green;padding:20px;margin-top:10px;border-radius:5px;color:white'>Google</a>

	// {
	// 	"to": 447000123890,
	// 	"from": "VoodooSMS",
	// 	"msg": "Testing The API",
	// 	"schedule": "3 weeks",
	// 	"external_reference": "Testing VoodooSMS",
	// 	"sandbox": true
	// }

	// json.NewEncoder(os.Stdout).Encode(`{"to":447792596594, "from: "Adnan", "msg": "Test"}`)
	// var jsonStr = []byte(`{"to":447792596594, "from: "Adnan", "msg": "Test"}`)
	// var jsonStr []byte
	// updated_slice := fmt.Appendf(jsonStr, `{"to":447792596594, "from: "%s", "msg": "%s", "sandbox": true,"external_reference": "Testing VoodooSMS"}`, state.text_input_array[0], state.text_input_array[1], state.text_input_array[2])
	data_marshal, _ := json.Marshal(email_data_structure)
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(data_marshal))
	if err != nil {
		panic(err)
	}
	// request.Header.Set("X-Custom-Header", "myvalue")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", access_token))
	client := &http.Client{}
	response_email_api, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response_email_api.Body.Close()

	fmt.Println("response Status:", response_email_api.Status)
	fmt.Println("response Headers:", response_email_api.Header)
	body, _ := io.ReadAll(response_email_api.Body)
	fmt.Println("response Body:", string(body))
	// "count":1,
	// "originator":"Adnan",
	// "body":"Test message to send",
	// "scheduledDateTime":null,
	// "credits":1,
	// "balance":21,
	// "messages":[{"id":"126258140222409536526908414937125121122510220","recipient":447792596594,"reference":null,"status":"PENDING_SENT"}]

	return response_email_api.Status
}

const base_64_image string = "data:image/png;base64, base 64 image"
