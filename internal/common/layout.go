// Package common содержит общие функции и константы, используемые разными пакетами приложения.
package common

import "math"

// Константы для расчета ширины макета
const (
	// DefaultWidth - стандартная ширина макета в символах
	DefaultWidth = 80
	// MinRatio - минимальное соотношение ширины макета к ширине экрана
	MinRatio = 4.0 / 7.0
	// LayoutWrapMargin — запас символов, вычитаемый при форматировании длинных строк
	LayoutWrapMargin = 12
)

// CalculateLayoutWidth вычисляет оптимальную ширину макета на основе ширины экрана.
// Это максимальное значение из DefaultWidth символов или MinRatio от ширины экрана.
// Эта функция доступна для использования в других пакетах.
//
// @param screenWidth Ширина экрана в символах
// @return Оптимальная ширина макета в символах
func CalculateLayoutWidth(screenWidth int) int {
	return int(math.Max(float64(DefaultWidth), MinRatio*float64(screenWidth)))
}
