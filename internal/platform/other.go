//go:build !darwin

package platform

func ConfigureWindow(cornerRadius float64) {}

func UpdateCornerRadius(cornerRadius float64) {}
