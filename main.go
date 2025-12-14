package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"syscall"
	"unsafe"

	"gioui.org/app"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type main_window struct {
	window      *app.Window
	information string
	completed   bool
}

func main() {
	// BUILD: go build -ldflags -H=windowsgui
	go func() {
		window := new(app.Window)
		window.Option(app.Title("Application 1"))

		main_window := main_window{window: window}

		// main_layout(main_window.window)
		// kpi_ui(main_window.window)
		// side_layout_version_two(main_window.window)
		side_layout_version_three(main_window.window)

	}()
	app.Main()
}
func main_layout(main_window *app.Window) {
	main_window.Option(app.Size(unit.Dp(400), unit.Dp(500)))
	var ops op.Ops
	theme := material.NewTheme()

	var (
		button_one widget.Clickable
		button_two widget.Clickable
	)
	for {
		switch event := main_window.Event().(type) {
		case app.FrameEvent:
			graphical_context := app.NewContext(&ops, event)

			// if label_for_button_one.Clicked(graphical_context) {
			// 	fmt.Println("the label was clicked")
			// 	window.information = ""
			// }
			// if button_one.Clicked(graphical_context) {
			// 	window.information = "Bank Reconciliation"
			// }
			// layout.Stack{}.Layout(graphical_context,
			// 	layout.Stacked(func(graphical_context layout.Context) layout.Dimensions {
			// 		margins := layout.Inset{
			// 			Top:    unit.Dp(25),
			// 			Bottom: unit.Dp(25),
			// 			Right:  unit.Dp(0),
			// 			Left:   unit.Dp(0),
			// 		}
			// 		label := material.Label(theme, unit.Sp(30), "window.information")
			// 		label.Alignment = text.Middle
			// 		// Here we use the material.Clickable wrapper func to animate button clicks.
			// 		// return material.Clickable(graphical_context, &label_for_button_one, label.Layout)
			// 		return margins.Layout(graphical_context, label.Layout)
			// 	}),
			// )

			// This flex splits the window vertically.
			layout.Flex{}.Layout(graphical_context,
				layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
					margins := layout.Inset{
						Top:    unit.Dp(25),
						Bottom: unit.Dp(25),
						Right:  unit.Dp(0),
						Left:   unit.Dp(0),
					}
					label := material.Label(theme, unit.Sp(30), "window.information")
					label.Alignment = text.Middle
					// Here we use the material.Clickable wrapper func to animate button clicks.
					// return material.Clickable(graphical_context, &label_for_button_one, label.Layout)
					return margins.Layout(graphical_context, label.Layout)
				}),
			)
			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceBetween,
			}.Layout(graphical_context,
				layout.Rigid(
					func(graphical_context layout.Context) layout.Dimensions {
						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(0),
							Right:  unit.Dp(25),
							Left:   unit.Dp(25),
						}
						return margins.Layout(
							graphical_context,
							func(graphical_context layout.Context) layout.Dimensions {
								button := material.Button(theme, &button_one, "Control Accounts")
								// button.Color = color.NRGBA{R: 76, G: 87, B: 96, A: 255}
								button.TextSize = unit.Sp(20)
								button.Background = color.NRGBA{R: 209, G: 84, B: 6, A: 255}
								return button.Layout(graphical_context)
							},
						)
					},
				),
				layout.Rigid(
					func(graphical_context layout.Context) layout.Dimensions {
						margins := layout.Inset{
							Top:    unit.Dp(25),
							Bottom: unit.Dp(25),
							Right:  unit.Dp(25),
							Left:   unit.Dp(25),
						}
						return margins.Layout(
							graphical_context,
							func(graphical_context layout.Context) layout.Dimensions {
								button := material.Button(theme, &button_two, "Bank Reconciliation")
								button.TextSize = unit.Sp(20)
								// button.Color = color.NRGBA{R: 76, G: 87, B: 96, A: 255}
								button.Background = color.NRGBA{R: 209, G: 27, B: 6, A: 255}
								return button.Layout(graphical_context)
							},
						)
					},
				),
				layout.Rigid(
					func(graphical_context layout.Context) layout.Dimensions {
						margins := layout.Inset{
							Top:    unit.Dp(10),
							Bottom: unit.Dp(10),
							Right:  unit.Dp(25),
							Left:   unit.Dp(25),
						}
						return margins.Layout(
							graphical_context,
							func(graphical_context layout.Context) layout.Dimensions {
								button := material.Button(theme, &button_one, "Close")
								button.Background = color.NRGBA{R: 163, G: 22, B: 33, A: 255}
								return button.Layout(graphical_context)
							},
						)
					},
				),
			)
			event.Frame(graphical_context.Ops)
		case app.DestroyEvent:
			os.Exit(0)
		}
	}
}
func kpi_ui(window *app.Window) error {
	window.Option(app.Size(unit.Dp(1500), unit.Dp(900)))

	var operations op.Ops
	// var label_for_button_one widget.Clickable
	var button_one widget.Clickable
	var button_two widget.Clickable

	design := material.NewTheme()
	// window.information = "KPI APP"

	for {
		switch event := window.Event().(type) {
		case app.FrameEvent:
			graphical_context := app.NewContext(&operations, event)

			// if label_for_button_one.Clicked(graphical_context) {
			// 	fmt.Println("the label was clicked")
			// 	window.information = ""
			// }
			if button_one.Clicked(graphical_context) {
				// window.information = "Bank Reconciliation"
			}
			layout.Flex{}.Layout(graphical_context,
				layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
					margins := layout.Inset{
						Top:    unit.Dp(25),
						Bottom: unit.Dp(25),
						Right:  unit.Dp(0),
						Left:   unit.Dp(0),
					}
					label := material.Label(design, unit.Sp(30), "window.information")
					label.Alignment = text.Middle
					// Here we use the material.Clickable wrapper func to animate button clicks.
					// return material.Clickable(graphical_context, &label_for_button_one, label.Layout)
					return margins.Layout(graphical_context, label.Layout)
				}),
			)

			// This flex splits the window vertically.
			layout.Flex{
				Axis: layout.Horizontal,
			}.Layout(graphical_context,
				layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
					// This flex splits the bottom pane horizontally.
					return layout.Flex{}.Layout(graphical_context,
						// layout.Flexed(2, func(graphical_context layout.Context) layout.Dimensions {
						// 	// This returns an empty left-hand pane.
						// 	return layout.Dimensions{Size: graphical_context.Constraints.Max}
						// }),
						layout.Flexed(0.75, func(graphical_context layout.Context) layout.Dimensions {
							// Here we position the button at the "north" or top-center of the available space.
							return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
								// Here we set the minimum constraints to zero. This allows our button to be smaller
								// than the entire right-side pane in the UI. Without this change, the button is forced
								// to occupy all of the space.
								// graphical_context.Constraints.Min = image.Point{}
								// Here we inset the button a little bit on all sides.
								return layout.UniformInset(5).Layout(graphical_context,
									material.Button(design, &button_one, "Button1").Layout,
								)
							})

						}),
						layout.Flexed(0.5, func(graphical_context layout.Context) layout.Dimensions {
							// Here we position the button at the "north" or top-center of the available space.
							return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
								// Here we set the minimum constraints to zero. This allows our button to be smaller
								// than the entire right-side pane in the UI. Without this change, the button is forced
								// to occupy all of the space.
								// graphical_context.Constraints.Min = image.Point{}
								// Here we inset the button a little bit on all sides.
								return layout.UniformInset(5).Layout(graphical_context,
									material.Button(design, &button_two, "Button2").Layout,
								)
							})

						}),
						layout.Flexed(0.5, func(graphical_context layout.Context) layout.Dimensions {
							// Here we position the button at the "north" or top-center of the available space.
							return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
								// Here we set the minimum constraints to zero. This allows our button to be smaller
								// than the entire right-side pane in the UI. Without this change, the button is forced
								// to occupy all of the space.
								// graphical_context.Constraints.Min = image.Point{}
								// Here we inset the button a little bit on all sides.
								return layout.UniformInset(5).Layout(graphical_context,
									material.Button(design, &button_two, "Button3").Layout,
								)
							})

						}),
					)
				}),
				layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
					// This flex splits the bottom pane horizontally.
					return layout.Flex{}.Layout(graphical_context,
						// layout.Flexed(2, func(graphical_context layout.Context) layout.Dimensions {
						// 	// This returns an empty left-hand pane.
						// 	return layout.Dimensions{Size: graphical_context.Constraints.Max}
						// }),
						layout.Flexed(0.75, func(graphical_context layout.Context) layout.Dimensions {
							// Here we position the button at the "north" or top-center of the available space.
							return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
								// Here we set the minimum constraints to zero. This allows our button to be smaller
								// than the entire right-side pane in the UI. Without this change, the button is forced
								// to occupy all of the space.
								// graphical_context.Constraints.Min = image.Point{}
								// Here we inset the button a little bit on all sides.
								return layout.UniformInset(5).Layout(graphical_context,
									material.Button(design, &button_one, "Button4").Layout,
								)
							})

						}),
						layout.Flexed(0.5, func(graphical_context layout.Context) layout.Dimensions {
							// Here we position the button at the "north" or top-center of the available space.
							return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
								// Here we set the minimum constraints to zero. This allows our button to be smaller
								// than the entire right-side pane in the UI. Without this change, the button is forced
								// to occupy all of the space.
								// graphical_context.Constraints.Min = image.Point{}
								// Here we inset the button a little bit on all sides.
								return layout.UniformInset(5).Layout(graphical_context,
									material.Button(design, &button_two, "Button5").Layout,
								)
							})

						}),
						layout.Flexed(0.5, func(graphical_context layout.Context) layout.Dimensions {
							// Here we position the button at the "north" or top-center of the available space.
							return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
								// Here we set the minimum constraints to zero. This allows our button to be smaller
								// than the entire right-side pane in the UI. Without this change, the button is forced
								// to occupy all of the space.
								// graphical_context.Constraints.Min = image.Point{}
								// Here we inset the button a little bit on all sides.
								return layout.UniformInset(5).Layout(graphical_context,
									material.Button(design, &button_two, "Button6").Layout,
								)
							})

						}),
					)
				}),
			)
			event.Frame(graphical_context.Ops)
		case app.DestroyEvent:
			return event.Err
		}
	}
}
func side_layout(window *app.Window) error {
	var ops op.Ops
	var button_one widget.Clickable

	theme := material.NewTheme()
	for {
		switch event := window.Event().(type) {
		case app.FrameEvent:
			graphical_context := app.NewContext(&ops, event)

			if button_one.Clicked(graphical_context) {
				fmt.Println("button1 was clicked")
			}

			// This flex splits the window vertically.
			layout.Flex{
				Axis: layout.Vertical,
			}.Layout(graphical_context,
				layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
					label := material.Label(theme, unit.Sp(50), "hello world")
					label.Alignment = text.Middle
					// Here we use the material.Clickable wrapper func to animate button clicks.
					return label.Layout(graphical_context)
				}),
				layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
					// This flex splits the bottom pane horizontally.
					return layout.Flex{}.Layout(graphical_context,
						layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
							// This returns an empty left-hand pane.
							return layout.Dimensions{Size: graphical_context.Constraints.Min}
						}),
						layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
							// Here we position the button at the "north" or top-center of the available space.
							return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
								// Here we set the minimum constraints to zero. This allows our button to be smaller
								// than the entire right-side pane in the UI. Without this change, the button is forced
								// to occupy all of the space.
								graphical_context.Constraints.Min = image.Point{}
								// Here we inset the button a little bit on all sides.
								return layout.UniformInset(8).Layout(graphical_context,
									material.Button(theme, &button_one, "Button1").Layout,
								)
							})

						}),
					)
				}),
			)
			event.Frame(graphical_context.Ops)
		case app.DestroyEvent:
			return event.Err
		}
	}
}
func side_layout_version_two(window *app.Window) error {
	var ops op.Ops
	var (
		button_one   widget.Clickable
		button_two   widget.Clickable
		button_three widget.Clickable
	)

	theme := material.NewTheme()
	for {
		switch event := window.Event().(type) {
		case app.FrameEvent:
			graphical_context := app.NewContext(&ops, event)

			if button_one.Clicked(graphical_context) {
				fmt.Println("button1 was clicked")
			}

			// This flex splits the window vertically.
			layout.Flex{
				Axis: layout.Vertical,
			}.Layout(graphical_context,
				layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
					label := material.Label(theme, unit.Sp(35), "BOXI APP")
					label.Alignment = text.Middle
					label.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
					// Here we use the material.Clickable wrapper func to animate button clicks.
					return label.Layout(graphical_context)
				}),
				layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
					// This flex splits the bottom pane horizontally.
					return layout.Flex{
						Axis: layout.Horizontal,
					}.Layout(graphical_context,
						// layout.Flexed(0.25, func(graphical_context layout.Context) layout.Dimensions {
						// 	// Here we position the button at the "north" or top-center of the available space.
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

						layout.Flexed(0.25, func(graphical_context layout.Context) layout.Dimensions {
							// Here we position the button at the "north" or top-center of the available space.
							return layout.Flex{
								Axis:    layout.Vertical,
								Spacing: layout.SpaceEnd,
							}.Layout(graphical_context,
								// layout.Flexed(2, func(graphical_context layout.Context) layout.Dimensions {
								// 	// This returns an empty left-hand pane.
								// 	return layout.Dimensions{Size: graphical_context.Constraints.Max}
								// }),
								layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
									// Here we position the button at the "north" or top-center of the available space.
									return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
										// Here we set the minimum constraints to zero. This allows our button to be smaller
										// than the entire right-side pane in the UI. Without this change, the button is forced
										// to occupy all of the space.
										// graphical_context.Constraints.Min = image.Point{}
										// Here we inset the button a little bit on all sides.
										return layout.UniformInset(5).Layout(graphical_context,
											material.Button(theme, &button_one, "Button1").Layout,
										)
									})

								}),
								layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
									// Here we position the button at the "north" or top-center of the available space.
									return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
										// Here we set the minimum constraints to zero. This allows our button to be smaller
										// than the entire right-side pane in the UI. Without this change, the button is forced
										// to occupy all of the space.
										// graphical_context.Constraints.Min = image.Point{}
										// Here we inset the button a little bit on all sides.
										return layout.UniformInset(5).Layout(graphical_context,
											material.Button(theme, &button_two, "Button2").Layout,
										)
									})

								}),
								layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
									// Here we position the button at the "north" or top-center of the available space.
									return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
										// Here we set the minimum constraints to zero. This allows our button to be smaller
										// than the entire right-side pane in the UI. Without this change, the button is forced
										// to occupy all of the space.
										// graphical_context.Constraints.Min = image.Point{}
										// Here we inset the button a little bit on all sides.
										return layout.UniformInset(5).Layout(graphical_context,
											material.Button(theme, &button_three, "Button3").Layout,
										)
									})

								}),
							)

						}),
						layout.Flexed(0.75, func(graphical_context layout.Context) layout.Dimensions {
							label := material.Label(theme, unit.Sp(20), "DATA")
							label.Alignment = text.Middle
							// Here we use the material.Clickable wrapper func to animate button clicks.
							return label.Layout(graphical_context)
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
			)
			event.Frame(graphical_context.Ops)
		case app.DestroyEvent:
			return event.Err
		}
	}
}
func side_layout_version_three(window *app.Window) error {
	var ops op.Ops
	var (
		menu_button        widget.Clickable
		side_button_one    widget.Clickable
		side_button_two    widget.Clickable
		side_button_three  widget.Clickable
		right_button_one   widget.Clickable
		right_button_two   widget.Clickable
		right_button_three widget.Clickable
	)
	side_buttons := []widget.Clickable{
		side_button_one, side_button_two, side_button_three,
	}
	right_side_buttons := []widget.Clickable{
		right_button_one, right_button_two, right_button_three,
	}
	colours_list := []color.NRGBA{
		{R: 0, G: 0, B: 0, A: 255},       // White
		{R: 0, G: 94, B: 184, A: 255},    // Blue
		{R: 0, G: 48, B: 135, A: 255},    // Dark Blue
		{R: 0, G: 150, B: 57, A: 255},    // Green
		{R: 0, G: 103, B: 71, A: 255},    // Dark Green
		{R: 255, G: 184, B: 28, A: 255},  // Warm Yellow
		{R: 237, G: 139, B: 0, A: 255},   // Orange
		{R: 232, G: 232, B: 232, A: 255}, // Light Grey
		{R: 240, G: 240, B: 240, A: 255}, // Lighter Grey
		{R: 0, G: 0, B: 0, A: 255},       // Black
	}
	var side_bar_width float32 = 0.0
	var right_content_width float32 = 1.0
	show_side_bar := false
	show_selection := 0

	theme := material.NewTheme()
	for {
		switch event := window.Event().(type) {
		case app.FrameEvent:
			graphical_context := app.NewContext(&ops, event)
			if menu_button.Clicked(graphical_context) {
				if side_bar_width == 0.0 {
					side_bar_width = 0.15
					right_content_width = (1.0 - side_bar_width)
				} else {
					side_bar_width = 0.0
					right_content_width = (1.0 - side_bar_width)
				}
				show_side_bar = !show_side_bar
			}
			if side_buttons[0].Clicked(graphical_context) {
				fmt.Println("Button 1 was clicked")
				show_selection = 1
			}
			if side_buttons[1].Clicked(graphical_context) {
				fmt.Println("BUTTON 2 was clicked")
				show_selection = 2
			}
			if side_buttons[2].Clicked(graphical_context) {
				fmt.Println("BUTTON 3 was clicked")
				show_selection = 3
			}
			if right_side_buttons[1].Clicked(graphical_context) {
				fmt.Println("TEST CLICK")
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
							margins := layout.Inset{
								Top:    unit.Dp(13),
								Bottom: unit.Dp(13),
								Left:   unit.Dp(0),
								Right:  unit.Dp(0),
							}
							menu_button_layout := material.Button(theme, &menu_button, "Menu")
							set_background_rect_colour(graphical_context, material.Button(theme, &menu_button, "Menu").Layout(graphical_context).Size, colours_list[6])
							// // paint.Fill(&ops, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
							// // Here we use the material.Clickable wrapper func to animate button clicks.
							// // return label.Layout(graphical_context)
							menu_button_layout.CornerRadius = unit.Dp(0)
							menu_button_layout.Inset = margins
							menu_button_layout.Background = colours_list[6]
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
							label := material.Label(theme, unit.Sp(35), "HEADING")
							label.Alignment = text.Middle
							label.Color = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
							set_background_rect_colour(graphical_context, label.Layout(graphical_context).Size, colours_list[6])
							// paint.Fill(&ops, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
							// Here we use the material.Clickable wrapper func to animate button clicks.
							// return label.Layout(graphical_context)
							// return margins.Layout(graphical_context, label.Layout)
							return label.Layout(graphical_context)
						}),
					)
				}),

				layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
					// This flex splits the bottom pane horizontally.
					return layout.Flex{
						Axis: layout.Horizontal,
					}.Layout(graphical_context,
						// layout.Flexed(0.25, func(graphical_context layout.Context) layout.Dimensions {
						// 	// Here we position the button at the "north" or top-center of the available space.
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
							// Here we position the button at the "north" or top-center of the available space.
							set_background_rect_colour(graphical_context, left_side_bar(theme, graphical_context, side_buttons, colours_list).Size, colours_list[7])
							return left_side_bar(theme, graphical_context, side_buttons, colours_list)

						}),
						layout.Flexed(right_content_width, func(graphical_context layout.Context) layout.Dimensions {
							switch show_selection {
							case 1:
								return layout.Flex{
									Axis: layout.Vertical,
								}.Layout(
									graphical_context,
									layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
										label := material.Label(theme, unit.Sp(20), "TEST STRING")
										label.Alignment = text.Start
										return label.Layout(graphical_context)
									}),
									layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
										return right_side_layout(theme, graphical_context, right_side_buttons, colours_list)
									}),
									// layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
									// 	return right_side_layout_one(theme, graphical_context, right_side_buttons)
									// }),
								)
							case 2:
								return layout.Flex{
									Axis: layout.Vertical,
								}.Layout(
									graphical_context,
									layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
										margins := layout.Inset{
											Top:    unit.Dp(0),
											Bottom: unit.Dp(10),
											Left:   unit.Dp(10),
											Right:  unit.Dp(0),
										}
										label := material.Label(theme, unit.Sp(20), "TEST STRING 2")
										label.Alignment = text.Start
										return margins.Layout(graphical_context, label.Layout)
									}),
									// layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
									// 	return right_side_layout(theme, graphical_context, right_side_buttons)
									// }),
									layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
										return right_side_layout_one(theme, graphical_context, right_side_buttons, colours_list)
									}),
								)
							case 3:
								return layout.Flex{
									Axis: layout.Vertical,
								}.Layout(
									graphical_context,
									layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
										label := material.Label(theme, unit.Sp(20), "TEST STRING 3")
										label.Alignment = text.Start
										return label.Layout(graphical_context)
									}),
									layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
										return right_side_layout(theme, graphical_context, right_side_buttons, colours_list)
									}),
									layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
										return right_side_layout_one(theme, graphical_context, right_side_buttons, colours_list)
									}),
								)
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
					margins := layout.Inset{
						Top:    unit.Dp(15),
						Bottom: unit.Dp(15),
						Left:   unit.Dp(0),
						Right:  unit.Dp(0),
					}
					label := material.Label(theme, unit.Sp(15), "FOOTER")
					label.Alignment = text.Middle
					label.Color = color.NRGBA{R: 0, G: 0, B: 0, A: 255}
					set_background_rect_colour(graphical_context, margins.Layout(graphical_context, label.Layout).Size, colours_list[8])
					// paint.Fill(&ops, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
					// Here we use the material.Clickable wrapper func to animate button clicks.
					// return label.Layout(graphical_context)
					return margins.Layout(graphical_context, label.Layout)
				}),
			)
			event.Frame(graphical_context.Ops)
		case app.DestroyEvent:
			return event.Err
		}
	}
}
func left_side_bar(theme *material.Theme, graphical_context layout.Context, buttons []widget.Clickable, colours []color.NRGBA) layout.Dimensions {
	return layout.Flex{
		Axis:    layout.Vertical,
		Spacing: layout.SpaceEnd,
	}.Layout(graphical_context,
		// layout.Flexed(2, func(graphical_context layout.Context) layout.Dimensions {
		// 	// This returns an empty left-hand pane.
		// 	return layout.Dimensions{Size: graphical_context.Constraints.Max}
		// }),
		layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
			// Here we position the button at the "north" or top-center of the available space.
			return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
				// Here we set the minimum constraints to zero. This allows our button to be smaller
				// than the entire right-side pane in the UI. Without this change, the button is forced
				// to occupy all of the space.
				// graphical_context.Constraints.Min = image.Point{}
				// Here we inset the button a little bit on all sides.
				button := material.Button(theme, &buttons[0], "Button1")
				button.Background = colours[1]
				return layout.UniformInset(5).Layout(graphical_context,
					button.Layout,
				)
			})

		}),
		layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
			// Here we position the button at the "north" or top-center of the available space.
			return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
				// Here we set the minimum constraints to zero. This allows our button to be smaller
				// than the entire right-side pane in the UI. Without this change, the button is forced
				// to occupy all of the space.
				// graphical_context.Constraints.Min = image.Point{}
				// Here we inset the button a little bit on all sides.
				button := material.Button(theme, &buttons[1], "Button 2")
				button.Background = colours[2]
				return layout.UniformInset(5).Layout(graphical_context,
					button.Layout,
				)
			})

		}),
		layout.Rigid(func(graphical_context layout.Context) layout.Dimensions {
			// Here we position the button at the "north" or top-center of the available space.
			return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
				// Here we set the minimum constraints to zero. This allows our button to be smaller
				// than the entire right-side pane in the UI. Without this change, the button is forced
				// to occupy all of the space.
				// graphical_context.Constraints.Min = image.Point{}
				// Here we inset the button a little bit on all sides.
				button := material.Button(theme, &buttons[2], "Button 3")
				button.Background = colours[3]
				return layout.UniformInset(5).Layout(graphical_context,
					button.Layout,
				)
			})

		}),
	)
}
func right_side_layout(theme *material.Theme, graphical_context layout.Context, buttons []widget.Clickable, colours []color.NRGBA) layout.Dimensions {
	return layout.Flex{}.Layout(graphical_context,
		// layout.Flexed(2, func(graphical_context layout.Context) layout.Dimensions {
		// 	// This returns an empty left-hand pane.
		// 	return layout.Dimensions{Size: graphical_context.Constraints.Max}
		// }),

		layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
			// Here we position the button at the "north" or top-center of the available space.
			return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
				// Here we set the minimum constraints to zero. This allows our button to be smaller
				// than the entire right-side pane in the UI. Without this change, the button is forced
				// to occupy all of the space.
				// graphical_context.Constraints.Min = image.Point{}
				// Here we inset the button a little bit on all sides.
				button := material.Button(theme, &buttons[0], "Button 1")
				button.Background = colours[4]
				return layout.UniformInset(5).Layout(graphical_context,
					button.Layout,
				)
			})

		}),
		layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
			// Here we position the button at the "north" or top-center of the available space.
			return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
				// Here we set the minimum constraints to zero. This allows our button to be smaller
				// than the entire right-side pane in the UI. Without this change, the button is forced
				// to occupy all of the space.
				// graphical_context.Constraints.Min = image.Point{}
				// Here we inset the button a little bit on all sides.
				button := material.Button(theme, &buttons[1], "Button 2")
				button.Background = colours[5]
				return layout.UniformInset(5).Layout(graphical_context,
					button.Layout,
				)
			})

		}),
		layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
			// Here we position the button at the "north" or top-center of the available space.
			return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
				// Here we set the minimum constraints to zero. This allows our button to be smaller
				// than the entire right-side pane in the UI. Without this change, the button is forced
				// to occupy all of the space.
				// graphical_context.Constraints.Min = image.Point{}
				// Here we inset the button a little bit on all sides.
				button := material.Button(theme, &buttons[2], "Button 3")
				button.Background = colours[6]
				return layout.UniformInset(5).Layout(graphical_context,
					button.Layout,
				)
			})

		}),
	)
}
func right_side_layout_one(theme *material.Theme, graphical_context layout.Context, buttons []widget.Clickable, colours []color.NRGBA) layout.Dimensions {
	return layout.Flex{}.Layout(graphical_context,
		// layout.Flexed(2, func(graphical_context layout.Context) layout.Dimensions {
		// 	// This returns an empty left-hand pane.
		// 	return layout.Dimensions{Size: graphical_context.Constraints.Max}
		// }),

		layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
			// Here we position the button at the "north" or top-center of the available space.
			return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
				// Here we set the minimum constraints to zero. This allows our button to be smaller
				// than the entire right-side pane in the UI. Without this change, the button is forced
				// to occupy all of the space.
				// graphical_context.Constraints.Min = image.Point{}
				// Here we inset the button a little bit on all sides.
				button := material.Button(theme, &buttons[0], "Button 4")
				button.Background = colours[6]
				return layout.UniformInset(5).Layout(graphical_context,
					button.Layout,
				)
			})

		}),
		layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
			// Here we position the button at the "north" or top-center of the available space.
			return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
				// Here we set the minimum constraints to zero. This allows our button to be smaller
				// than the entire right-side pane in the UI. Without this change, the button is forced
				// to occupy all of the space.
				// graphical_context.Constraints.Min = image.Point{}
				// Here we inset the button a little bit on all sides.
				return layout.UniformInset(5).Layout(graphical_context,
					material.Button(theme, &buttons[1], "Button 5").Layout,
				)
			})

		}),
		layout.Flexed(1, func(graphical_context layout.Context) layout.Dimensions {
			// Here we position the button at the "north" or top-center of the available space.
			return layout.N.Layout(graphical_context, func(graphical_context layout.Context) layout.Dimensions {
				// Here we set the minimum constraints to zero. This allows our button to be smaller
				// than the entire right-side pane in the UI. Without this change, the button is forced
				// to occupy all of the space.
				// graphical_context.Constraints.Min = image.Point{}
				// Here we inset the button a little bit on all sides.
				return layout.UniformInset(5).Layout(graphical_context,
					material.Button(theme, &buttons[2], "Button6").Layout,
				)
			})

		}),
	)
}
func set_background_rect_colour(graphical_context layout.Context, size image.Point, colour color.NRGBA) layout.Dimensions {
	defer clip.Rect{Max: size}.Push(graphical_context.Ops).Pop()
	paint.ColorOp{Color: colour}.Add(graphical_context.Ops)
	paint.PaintOp{}.Add(graphical_context.Ops)
	return layout.Dimensions{Size: size}
}
func open_file_dialog_box() {
	syscall_file_open := syscall.NewLazyDLL("commdlg.h")
	win_structure := syscall_file_open.NewProc("OPENFILENAMEA")
	procedure := syscall_file_open.NewProc("GetOpenFileNameA")
	procedure_call, _, err := procedure.Call(uintptr(unsafe.Pointer(unsafe.Sizeof(win_structure))))
	if err != nil {
		fmt.Println("ERROR OPENING FILE DIALOG BOX")
	}
	fmt.Println(procedure_call)
}
