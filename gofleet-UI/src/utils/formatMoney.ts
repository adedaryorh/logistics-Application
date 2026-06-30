/**
 * Format money in Nigerian Naira format
 * @param amount - Amount in kobo (smallest unit)
 * @returns Formatted string like "₦1,200.00"
 */
export const formatMoney = (amountInKobo: number | string): string => {
  const amount = Number(amountInKobo);
  if (isNaN(amount)) return '₦0.00';
  
  // Convert from kobo to naira
  const nairaAmount = amount / 100;
  
  return new Intl.NumberFormat('en-NG', {
    style: 'currency',
    currency: 'NGN',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(nairaAmount);
};

/**
 * Format money without currency symbol
 * @param amountInKobo - Amount in kobo
 * @returns Formatted string like "1,200.00"
 */
export const formatMoneyWithoutSymbol = (amountInKobo: number | string): string => {
  const amount = Number(amountInKobo);
  if (isNaN(amount)) return '0.00';
  
  const nairaAmount = amount / 100;
  
  return new Intl.NumberFormat('en-NG', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(nairaAmount);
};

/**
 * Format money with custom currency
 * @param amountInCents - Amount in smallest currency unit
 * @param currency - Currency code (default: NGN)
 * @returns Formatted string
 */
export const formatCurrency = {
  formatMoney,
  formatMoneyWithoutSymbol,
  formatCurrencyAmount: (amount: number, currency: string = 'NGN') => {
    return new Intl.NumberFormat('en-NG', {
      style: 'currency',
      currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(amount / 100); // Assuming amount is in cents/kobo
  }
};
