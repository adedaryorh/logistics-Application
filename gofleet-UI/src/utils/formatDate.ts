/**
 * Format date for display
 * @param dateString - ISO date string or Date object
 * @param options - Formatting options
 * @returns Formatted date string
 */
export const formatDate = (
  dateString: string | Date,
  options: {
    timeZone?: string;
    dateStyle?: Intl.DateTimeFormatOptions['dateStyle'];
    timeStyle?: Intl.DateTimeFormatOptions['timeStyle'];
  } = {}
): string => {
  const date = typeof dateString === 'string' ? new Date(dateString) : dateString;
  
  if (isNaN(date.getTime())) return 'Invalid Date';
  
  const { timeZone, dateStyle = 'medium', timeStyle } = options;
  
  return new Intl.DateTimeFormat('en-NG', {
    timeZone,
    dateStyle,
    timeStyle,
  }).format(date);
};

/**
 * Format date as "Today", "Yesterday", or actual date
 * @param dateString - ISO date string or Date object
 * @returns Formatted date string
 */
export const formatDateRelative = (dateString: string | Date): string => {
  const date = typeof dateString === 'string' ? new Date(dateString) : dateString;
  const today = new Date();
  const yesterday = new Date(today);
  yesterday.setDate(today.getDate() - 1);
  
  if (isNaN(date.getTime())) return 'Invalid Date';
  
  // Check if today
  if (
    date.getDate() === today.getDate() &&
    date.getMonth() === today.getMonth() &&
    date.getFullYear() === today.getFullYear()
  ) {
    return 'Today';
  }
  
  // Check if yesterday
  if (
    date.getDate() === yesterday.getDate() &&
    date.getMonth() === yesterday.getMonth() &&
    date.getFullYear() === yesterday.getFullYear()
  ) {
    return 'Yesterday';
  }
  
  // Show date format
  return formatDate(date, { dateStyle: 'medium' });
};

/**
 * Format time only
 * @param dateString - ISO date string or Date object
 * @returns Formatted time string like "14:30"
 */
export const formatTime = (dateString: string | Date): string => {
  const date = typeof dateString === 'string' ? new Date(dateString) : dateString;
  
  if (isNaN(date.getTime())) return 'Invalid Time';
  
  return new Intl.DateTimeFormat('en-NG', {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
  }).format(date);
};

/**
 * Format date-time for chat/message timestamps
 * @param dateString - ISO date string or Date object
 * @returns Formatted string like "Today, 14:30" or "Jan 15, 14:30"
 */
export const formatMessageTimestamp = (dateString: string | Date): string => {
  const date = typeof dateString === 'string' ? new Date : dateString;
  const today = new Date();
  const yesterday = new Date(today);
  yesterday.setDate(today.getDate() - 1);
  
  if (isNaN(date.getTime())) return 'Invalid Time';
  
  // Check if today
  if (
    date.getDate() === today.getDate() &&
    date.getMonth() === today.getMonth() &&
    date.getFullYear() === today.getFullYear()
  ) {
    return `Today, ${formatTime(date)}`;
  }
  
  // Check if yesterday
  if (
    date.getDate() === yesterday.getDate() &&
    date.getMonth() === yesterday.getMonth() &&
    date.getFullYear() === yesterday.getFullYear()
  ) {
    return `Yesterday, ${formatTime(date)}`;
  }
  
  // Show date and time
  return formatDate(date, { dateStyle: 'short', timeStyle: 'short' });
};

export default {
  formatDate,
  formatDateRelative,
  formatTime,
  formatMessageTimestamp,
};
