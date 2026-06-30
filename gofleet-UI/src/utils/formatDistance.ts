/**
 * Format distance in kilometers or meters
 * @param meters - Distance in meters
 * @returns Formatted string like "2.3 km" or "450 m"
 */
export const formatDistance = (meters: number): string => {
  if (isNaN(meters) || meters < 0) return '0 m';
  
  if (meters >= 1000) {
    const km = meters / 1000;
    // Show 1 decimal place for kilometers
    return `${km.toFixed(1)} km`;
  } else {
    // Show meters as whole number
    return `${Math.round(meters)} m`;
  }
};

/**
 * Format distance for display in UI components
 * @param meters - Distance in meters
 * @param options - Formatting options
 * @returns Formatted distance string
 */
export const formatDistanceForDisplay = (
  meters: number,
  options: { 
    useMetric?: boolean; 
    decimalPlaces?: number; 
    showUnit?: boolean 
  } = {}
): string => {
  const { useMetric = true, decimalPlaces = 1, showUnit = true } = options;
  
  if (isNaN(meters) || meters < 0) {
    return showUnit ? (useMetric ? '0 m' : '0 ft') : '0';
  }
  
  if (useMetric) {
    if (meters >= 1000) {
      const km = meters / 1000;
      const value = km.toFixed(decimalPlaces);
      return showUnit ? `${value} km` : value;
    } else {
      const value = Math.round(meters).toString();
      return showUnit ? `${value} m` : value;
    }
  } else {
    // Imperial units (feet)
    const feet = meters * 3.28084;
    if (feet >= 5280) { // 1 mile = 5280 feet
      const miles = feet / 5280;
      const value = miles.toFixed(decimalPlaces);
      return showUnit ? `${value} mi` : value;
    } else {
      const value = Math.round(feet).toString();
      return showUnit ? `${value} ft` : value;
    }
  }
};

export default formatDistance;
