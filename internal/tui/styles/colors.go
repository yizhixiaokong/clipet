package styles

import "image/color"

// DimColor returns the dim/secondary text color.
func DimColor() color.Color {
	return colorDim
}

// PrimaryColor returns the primary accent color.
func PrimaryColor() color.Color {
	return colorPrimary
}

// GoldColor returns the gold/highlight color.
func GoldColor() color.Color {
	return colorGold
}

// TextColor returns the default text color.
func TextColor() color.Color {
	return colorText
}
