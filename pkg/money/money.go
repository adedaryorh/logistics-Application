package money

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Currency represents a currency code
type Currency string

// Predefined currencies
const (
	USD Currency = "USD"
	NGN Currency = "NGN"
	EUR Currency = "EUR"
	GBP Currency = "GBP"
)

// Money represents a monetary value with currency
type Money struct {
	Amount   int64
	Currency string
}

// NewMoney creates a new Money instance
func NewMoney(amount int64, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: strings.ToUpper(currency),
	}
}

// Add adds two Money amounts, returning an error if currencies don't match
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return NewMoney(m.Amount+other.Amount, m.Currency), nil
}

func Add(a, b Money) (Money, error) {
	return a.Add(b)
}

// Subtract subtracts other from m, returning an error if currencies don't match
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return NewMoney(m.Amount-other.Amount, m.Currency), nil
}

// Multiply multiplies the amount by a multiplier
func (m Money) Multiply(multiplier int64) Money {
	return NewMoney(m.Amount*multiplier, m.Currency)
}

// Divide divides the amount by a divisor
func (m Money) Divide(divisor int64) (Money, error) {
	if divisor == 0 {
		return Money{}, ErrDivisionByZero
	}
	return NewMoney(m.Amount/divisor, m.Currency), nil
}

// Equals compares two Money amounts for equality
func (m Money) Equals(other Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

// GreaterThan returns true if m > other
func (m Money) GreaterThan(other Money) bool {
	if m.Currency != other.Currency {
		return false // Cannot compare different currencies
	}
	return m.Amount > other.Amount
}

// LessThan returns true if m < other
func (m Money) LessThan(other Money) bool {
	if m.Currency != other.Currency {
		return false // Cannot compare different currencies
	}
	return m.Amount < other.Amount
}

// Format formats the money as a string with currency symbol
func (m Money) Format() string {
	switch m.Currency {
	case string(USD):
		return formatMinorUnits("$", m.Amount)
	case string(NGN):
		return formatMinorUnits("₦", m.Amount)
	case string(EUR):
		return formatMinorUnits("€", m.Amount)
	case string(GBP):
		return formatMinorUnits("£", m.Amount)
	default:
		return fmt.Sprintf("%s %s", m.Currency, formatMinorUnits("", m.Amount))
	}
}

// String implements the Stringer interface
func (m Money) String() string {
	return m.Format()
}

// ParseCurrency converts a string to Currency
func ParseCurrency(s string) (Currency, error) {
	switch strings.ToUpper(s) {
	case "USD":
		return USD, nil
	case "NGN":
		return NGN, nil
	case "EUR":
		return EUR, nil
	case "GBP":
		return GBP, nil
	default:
		return "", ErrUnknownCurrency
	}
}

func formatMinorUnits(symbol string, amount int64) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}

	major := amount / 100
	minor := amount % 100
	majorStr := strconv.FormatInt(major, 10)

	var grouped []byte
	for i, ch := range majorStr {
		if i > 0 && (len(majorStr)-i)%3 == 0 {
			grouped = append(grouped, ',')
		}
		grouped = append(grouped, byte(ch))
	}

	return fmt.Sprintf("%s%s%s.%02d", sign, symbol, string(grouped), minor)
}

// Errors
var (
	ErrCurrencyMismatch = errors.New("currency mismatch")
	ErrDivisionByZero   = errors.New("division by zero")
	ErrUnknownCurrency  = errors.New("unknown currency")
	ErrInvalidAmount    = errors.New("invalid amount")
)
