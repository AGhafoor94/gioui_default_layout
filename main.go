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

var base_64_image string = "data:image/png;base64, iVBORw0KGgoAAAANSUhEUgAABYoAAAWKCAYAAABW6FGBAAAsYmNhQlgAACxianVtYgAAAB5qdW1kYzJwYQARABCAAACqADibcQNjMnBhAAAALDxqdW1iAAAAR2p1bWRjMm1hABEAEIAAAKoAOJtxA3Vybjp1dWlkOjRkMzI5ZTFiLTlhMTktNDY5ZC05MmMxLTJkZGExYzIwNzNkMAAAAAGmanVtYgAAAClqdW1kYzJhcwARABCAAACqADibcQNjMnBhLmFzc2VydGlvbnMAAAAAymp1bWIAAAAmanVtZGNib3IAEQAQgAAAqgA4m3EDYzJwYS5hY3Rpb25zAAAAAJxjYm9yoWdhY3Rpb25zgaNmYWN0aW9ua2MycGEuZWRpdGVkbXNvZnR3YXJlQWdlbnRtQWRvYmUgRmlyZWZseXFkaWdpdGFsU291cmNlVHlwZXhGaHR0cDovL2N2LmlwdGMub3JnL25ld3Njb2Rlcy9kaWdpdGFsc291cmNldHlwZS90cmFpbmVkQWxnb3JpdGhtaWNNZWRpYQAAAKtqdW1iAAAAKGp1bWRjYm9yABEAEIAAAKoAOJtxA2MycGEuaGFzaC5kYXRhAAAAAHtjYm9ypWpleGNsdXNpb25zgaJlc3RhcnQYIWZsZW5ndGgZLG5kbmFtZW5qdW1iZiBtYW5pZmVzdGNhbGdmc2hhMjU2ZGhhc2hYIBNxNziCdFjOuJslOgMKn+9085WYu+5tMk2nPdopSey7Y3BhZEgAAAAAAAAAAAAAAgdqdW1iAAAAJGp1bWRjMmNsABEAEIAAAKoAOJtxA2MycGEuY2xhaW0AAAAB22Nib3KoaGRjOnRpdGxlb0dlbmVyYXRlZCBJbWFnZWlkYzpmb3JtYXRpaW1hZ2UvcG5namluc3RhbmNlSUR4LHhtcDppaWQ6OTA4NzJiNmUtZGQ0Ny00MzFjLWJhYWUtNTkxMTBhODE0NjMzb2NsYWltX2dlbmVyYXRvcng2QWRvYmVfSWxsdXN0cmF0b3IvMjguMCBhZG9iZV9jMnBhLzAuNy42IGMycGEtcnMvMC4yNS4ydGNsYWltX2dlbmVyYXRvcl9pbmZvgb9kbmFtZXFBZG9iZSBJbGx1c3RyYXRvcmd2ZXJzaW9uZDI4LjD/aXNpZ25hdHVyZXgZc2VsZiNqdW1iZj1jMnBhLnNpZ25h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SvPIhgBjVtggH0w0vm/lmCQYiC/Y12GeCxfgYlL33btjcGFkWQu8AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA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"
