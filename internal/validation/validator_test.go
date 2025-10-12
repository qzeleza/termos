package validation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Тесты для покрытия validation framework

func TestPasswordValidator(t *testing.T) {
	// Тестируем валидатор с минимальной длиной
	validator := NewPasswordValidator(8)
	assert.NotNil(t, validator)
	assert.Contains(t, validator.Description(), "Пароль", "Описание должно содержать 'Пароль'")

	// Тест валидных паролей, соответствующих всем требованиям
	validPasswords := []string{
		"Passw0rd!",
		"Example#123A",
		"Str0ng!Key",
	}
	for _, password := range validPasswords {
		err := validator.Validate(password)
		assert.NoError(t, err, "Валидный пароль '%s' не должен вызывать ошибку", password)
	}

	// Тест невалидных паролей
	invalidPasswords := []string{
		"short",       // слишком короткий
		"",            // пустой
		"password123", // нет заглавных и спецсимволов
		"PASSWORD!!!", // нет цифр и строчных
		"SimplePass",  // нет цифр и спецсимволов
	}
	for _, password := range invalidPasswords {
		err := validator.Validate(password)
		assert.Error(t, err, "Короткий пароль '%s' не должен проходить валидацию", password)
	}
}

func TestEmailValidator(t *testing.T) {
	validator := NewEmailValidator()
	assert.NotNil(t, validator)
	assert.Contains(t, validator.Description(), "Email", "Описание должно содержать 'Email'")

	// Валидные email адреса
	validEmails := []string{
		"user@example.com",
		"test.email@domain.co.uk",
		"admin@localhost.com",
		"user@example.org",
		"123@numbers.com",
	}
	for _, email := range validEmails {
		err := validator.Validate(email)
		assert.NoError(t, err, "Email '%s' должен проходить валидацию", email)
	}

	// Невалидные email адреса
	invalidEmails := []string{
		"invalid-email",
		"@domain.com",
		"user@",
		"",
	}
	for _, email := range invalidEmails {
		err := validator.Validate(email)
		assert.Error(t, err, "Невалидный email '%s' не должен проходить валидацию", email)
	}
}

func TestNumberValidator(t *testing.T) {
	validator := NewNumberValidator(1, 100)
	assert.NotNil(t, validator)
	assert.Contains(t, validator.Description(), "Число", "Описание должно содержать 'Число'")

	// Валидные числа
	validNumbers := []string{
		"1",
		"50",
		"100",
		"25",
	}
	for _, number := range validNumbers {
		err := validator.Validate(number)
		assert.NoError(t, err, "Число '%s' должно проходить валидацию", number)
	}

	// Невалидные числа
	invalidNumbers := []string{
		"0",    // меньше минимума
		"101",  // больше максимума
		"abc",  // не число
		"",     // пустое
		"50.5", // дробное
		"-10",  // отрицательное
	}
	for _, number := range invalidNumbers {
		err := validator.Validate(number)
		assert.Error(t, err, "Невалидное число '%s' не должно проходить валидацию", number)
	}
}

func TestIPValidator(t *testing.T) {
	// Тестируем общий IP валидатор (IPv4 и IPv6)
	ipValidator := NewIPValidator(true, true)
	assert.NotNil(t, ipValidator)
	assert.Contains(t, ipValidator.Description(), "IP", "Описание должно содержать 'IP'")

	// Валидные IP адреса
	validIPs := []string{
		"192.168.1.1", // IPv4
		"10.0.0.1",    // IPv4
		"::1",         // IPv6
		"2001:db8::1", // IPv6
	}
	for _, ip := range validIPs {
		err := ipValidator.Validate(ip)
		assert.NoError(t, err, "IP '%s' должен проходить валидацию", ip)
	}

	// Невалидные IP адреса
	invalidIPs := []string{
		"256.256.256.256", // IPv4 с неверными октетами
		"192.168.1",       // неполный IPv4
		"invalid-ip",      // не IP
		"",                // пустой
	}
	for _, ip := range invalidIPs {
		err := ipValidator.Validate(ip)
		assert.Error(t, err, "Невалидный IP '%s' не должен проходить валидацию", ip)
	}

	// Тестируем IPv4 валидатор
	ipv4Validator := NewIPv4Validator()
	assert.NotNil(t, ipv4Validator)

	err := ipv4Validator.Validate("192.168.1.1")
	assert.NoError(t, err, "IPv4 адрес должен проходить валидацию")

	err = ipv4Validator.Validate("::1")
	assert.Error(t, err, "IPv6 адрес не должен проходить IPv4 валидацию")

	// Тестируем IPv6 валидатор
	ipv6Validator := NewIPv6Validator()
	assert.NotNil(t, ipv6Validator)

	err = ipv6Validator.Validate("::1")
	assert.NoError(t, err, "IPv6 адрес должен проходить валидацию")

	err = ipv6Validator.Validate("192.168.1.1")
	assert.Error(t, err, "IPv4 адрес не должен проходить IPv6 валидацию")
}

func TestDomainValidator(t *testing.T) {
	validator := NewDomainValidator()
	assert.NotNil(t, validator)
	assert.Contains(t, validator.Description(), "Домен", "Описание должно содержать 'Домен'")

	// Валидные домены
	validDomains := []string{
		"example.com",
		"sub.domain.co.uk",
		"test-domain.org",
		"domain123.net",
	}
	for _, domain := range validDomains {
		err := validator.Validate(domain)
		assert.NoError(t, err, "Домен '%s' должен проходить валидацию", domain)
	}

	// Невалидные домены
	invalidDomains := []string{
		"",            // пустой
		"domain..com", // двойные точки
		".domain.com", // начинается с точки
		"domain.com.", // заканчивается точкой
	}
	for _, domain := range invalidDomains {
		err := validator.Validate(domain)
		assert.Error(t, err, "Невалидный домен '%s' не должен проходить валидацию", domain)
	}
}

func TestTextValidator(t *testing.T) {
	validator := NewTextValidator(5, 20)
	assert.NotNil(t, validator)
	assert.Contains(t, validator.Description(), "Текст", "Описание должно содержать 'Текст'")

	// Валидные тексты
	validTexts := []string{
		"Hello",
		"Hello World",
		"Test123",
	}
	for _, text := range validTexts {
		err := validator.Validate(text)
		assert.NoError(t, err, "Текст '%s' должен проходить валидацию", text)
	}

	// Невалидные тексты
	invalidTexts := []string{
		"abc",                         // слишком короткий
		"",                            // пустой
		"Very long text that exceeds", // слишком длинный
	}
	for _, text := range invalidTexts {
		err := validator.Validate(text)
		assert.Error(t, err, "Невалидный текст '%s' не должен проходить валидацию", text)
	}

	// Тестируем WithPattern
	updatedValidator := validator.WithPattern("^[0-9]+$")
	err := updatedValidator.Validate("123")
	assert.Error(t, err, "Текст не соответствует длине после изменения паттерна")

	// Создаем валидатор только с паттерном
	patternValidator := NewTextValidator(0, 0).WithPattern("^[0-9]+$")
	err = patternValidator.Validate("123")
	assert.NoError(t, err, "Число должно проходить валидацию с числовым паттерном")

	err = patternValidator.Validate("abc")
	assert.Error(t, err, "Буквы не должны проходить валидацию с числовым паттерном")
}

func TestCompositeValidator(t *testing.T) {
	// Создаем составной валидатор из нескольких простых
	emailValidator := NewEmailValidator()
	textValidator := NewTextValidator(5, 50)

	composite := NewCompositeValidator(AllMustPass, emailValidator, textValidator)
	assert.NotNil(t, composite)
	assert.Contains(t, composite.Description(), "Email", "Описание должно содержать описания всех валидаторов")

	// Валидное значение для всех валидаторов
	err := composite.Validate("user@example.com")
	assert.NoError(t, err, "Значение должно проходить все валидации")

	// Невалидное для email
	err = composite.Validate("invalid")
	assert.Error(t, err, "Значение не должно проходить валидацию email")

	// Тестируем пустой композитный валидатор
	emptyComposite := NewCompositeValidator(AllMustPass)
	err = emptyComposite.Validate("anything")
	assert.NoError(t, err, "Пустой композитный валидатор должен принимать любое значение")

	// Тестируем режим AnyCanPass
	anyComposite := NewCompositeValidator(AnyCanPass, emailValidator, textValidator)
	err = anyComposite.Validate("user@example.com") // Проходит email валидацию
	assert.NoError(t, err, "Должно проходить хотя бы одну валидацию")
}

func TestEdgeCases(t *testing.T) {
	// Тестируем крайние случаи

	// Пустые строки
	emailValidator := NewEmailValidator()
	err := emailValidator.Validate("")
	assert.Error(t, err, "Пустой email не должен проходить валидацию")

	// Очень длинные строки
	longString := string(make([]rune, 1000))
	for i := range longString {
		longString = string(append([]rune(longString)[:i], 'a'))
	}
	textValidator := NewTextValidator(0, 0).WithPattern(".*")
	err = textValidator.Validate(longString)
	assert.NoError(t, err, "Очень длинная строка должна проходить валидацию если соответствует паттерну")

	// Граничные значения для чисел
	numberValidator := NewNumberValidator(0, 100)
	err = numberValidator.Validate("0")
	assert.NoError(t, err, "Граничное значение 0 должно проходить валидацию")

	err = numberValidator.Validate("100")
	assert.NoError(t, err, "Граничное значение 100 должно проходить валидацию")

	// Специальные символы в текстовых валидаторах
	domainValidator := NewDomainValidator()
	// Просто проверяем, что не паникует
	assert.NotPanics(t, func() {
		if err := domainValidator.Validate("special-domain.com"); err != nil {
			t.Errorf("Ошибка валидации домена: %v", err)
		}
	})
}

func TestValidatorFunc(t *testing.T) {
	// Тестируем ValidatorFunc
	customValidator := ValidatorFunc(func(input string) error {
		if input == "test" {
			return nil
		}
		return errors.New("не равен 'test'")
	})

	err := customValidator.Validate("test")
	assert.NoError(t, err, "Кастомный валидатор должен принимать 'test'")

	err = customValidator.Validate("other")
	assert.Error(t, err, "Кастомный валидатор должен отклонять 'other'")

	desc := customValidator.Description()
	assert.Equal(t, "Пользовательская валидация", desc)
}

func TestPasswordValidatorDefaultMinLength(t *testing.T) {
	// Тестируем создание с нулевой минимальной длиной
	validator := NewPasswordValidator(0)
	assert.Equal(t, 8, validator.MinLength, "Должна быть установлена длина по умолчанию")

	validator = NewPasswordValidator(-1)
	assert.Equal(t, 8, validator.MinLength, "Должна быть установлена длина по умолчанию")
}

func TestIPValidatorDefaultBehavior(t *testing.T) {
	// Тестируем поведение по умолчанию
	validator := NewIPValidator(false, false) // Ни IPv4, ни IPv6 не разрешены
	assert.True(t, validator.allowIPv4, "IPv4 должен быть разрешен по умолчанию")
}
