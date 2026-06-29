package money

import "testing"

func TestAdd(t *testing.T) {
	t.Parallel()

	sum, err := Add(NewMoney(150000, "NGN"), NewMoney(25000, "NGN"))
	if err != nil {
		t.Fatalf("Add returned error: %v", err)
	}

	if sum.Amount != 175000 {
		t.Fatalf("expected amount 175000, got %d", sum.Amount)
	}
	if sum.Currency != "NGN" {
		t.Fatalf("expected currency NGN, got %s", sum.Currency)
	}
}

func TestAddCurrencyMismatch(t *testing.T) {
	t.Parallel()

	_, err := Add(NewMoney(100, "USD"), NewMoney(100, "NGN"))
	if err == nil {
		t.Fatal("expected currency mismatch error")
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		money Money
		want  string
	}{
		{name: "naira", money: NewMoney(150000, "NGN"), want: "₦1,500.00"},
		{name: "usd", money: NewMoney(1500, "USD"), want: "$15.00"},
		{name: "negative", money: NewMoney(-250050, "NGN"), want: "-₦2,500.50"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.money.String(); got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}
