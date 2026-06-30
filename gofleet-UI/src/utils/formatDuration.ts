/**
 * Format duration in minutes or hours
 * @param minutes - Duration in minutes
 * @returns Formatted string like "15 min" or "2.5 hr"
 */
export const formatDuration = (minutes: number): string => {
  if (isNaN(minutes) || minutes < 0) return '0 min';
  
  if (minutes >= 60) {
    const hours = minutes / 60;
    // Show 1 decimal place for hours
    return `${hours.toFixed(1)} hr`;
  } else {
    // Show minutes as whole number
    return `${Math.round(minutes)} min`;
  }
};

/**
 * Format duration with more precision
 * @param minutes - Duration in minutes
 * @param options - Formatting options
 * @returns Formatted duration string
 */
export const formatDurationPrecise = (
  minutes: number,
  options: { 
    showSeconds?: boolean; 
    decimalPlaces?: number; 
  } = {}
): string => {
  const { showSeconds = false, decimalPlaces = 1 } = options;
  
  if (isNaN(minutes) || minutes < 0) {
    return showSeconds ? '0:00' : '0 min';
  }
  
  if (showSeconds) {
    const totalSeconds = Math.round(minutes * 60);
    const hours = Math.floor(totalSeconds / 3600);
    const minutesRemainder = Math.floor((totalSeconds % 3600) / 60);
    const seconds = totalSeconds % 60;
    
    const hoursStr = hours > 0 ? `${hours}:` : '';
    const minutesStr = minutesRemainder.toString().padStart(2, '0');
    const secondsStr = seconds.toString().padStart(2, '0');
    
    return `${hoursStr}${minutesStr}:${secondsStr}`;
  } else {
    if (minutes >= 60) {
      const hours = minutes / 60;
      const value = hours.toFixed(decimalPlaces);
      return `${value} hr`;
    } else {
      const value = Math.round(minutes).toString();
      return `${value} min`;
    }
  }
};

/**
 * Format duration in hours and minutes (e.g., "2h 30m")
 * @param minutes - Duration in minutes
 * @returns Formatted string like "2h 30m"
 */
export const formatDurationHM = (minutes: number): string => {
  if (isNaN(minutes) || minutes < 0) return '0m';
  
  const hours = Math.floor(minutes / 60);
  const remainingMinutes = Math.round(minutes % 60);
  
  if (hours === 0) {
    return `${remainingMinutes}m`;
  } else if (remainingMinutes === 0) {
    return `${hours}h`;
  } else {
    return `${hours}h ${remainingMinutes}m`;
  }
};

export default formatDuration;
