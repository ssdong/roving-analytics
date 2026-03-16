/**
 * formatDate
 * Formats a date object into a string representation "YYYY-MM-DD".
 *
 * @param {Date} date - The date object to format.
 * @returns {string} - The formatted date string.
 */
export function formatDate(date: Date): string {
  // Extracting the year from the provided date.
  const year = date.getFullYear();

  // Extracting the month, which is 0-indexed (0 for January, 1 for February, ...), hence adding 1.
  // Then, ensuring it's always 2 digits, e.g., '01' for January.
  const month = String(date.getMonth() + 1).padStart(2, '0');

  // Extracting the day and ensuring it's always 2 digits, e.g., '09' for 9th.
  const day = String(date.getDate()).padStart(2, '0');

  // Extracting hours, minutes, and seconds, and ensuring each is always 2 digits.
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  const seconds = String(date.getSeconds()).padStart(2, '0');

  // Constructing and returning the formatted date string in the desired format.
  // This specific format "YYYY-MM-DD HH:MM:SS" is used to match the format expected by ClickHouse
  // when dealing with DateTime('UTC') type columns.
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
}

export function localDateToUTC(date: Date): Date {
  // date.getTime() gives the number of milliseconds since the Unix epoch for the given local date.
  // date.getTimezoneOffset() gives the timezone offset in minutes.
  // Multiplying this offset by 60,000 converts it to milliseconds.
  // By adding this offset, we effectively convert the local date and time to UTC.
  return new Date(date.getTime() + date.getTimezoneOffset() * 60000);
}

export function formatDatYYYYMMDD(date: Date): string {
  // Extracting the year from the provided date.
  const year = date.getFullYear();

  // Extracting the month, which is 0-indexed (0 for January, 1 for February, ...), hence adding 1.
  // Then, ensuring it's always 2 digits, e.g., '01' for January.
  const month = String(date.getMonth() + 1).padStart(2, '0');

  // Extracting the day and ensuring it's always 2 digits, e.g., '09' for 9th.
  const day = String(date.getDate()).padStart(2, '0');

  // Returning the formatted date string in the format `YYYY-MM-DD`.
  return `${year}-${month}-${day}`;
}

export function utcDateStringToLocalDateString(utcDateString: string): string {
  // Create a UTC Date object at midnight for the provided date string
  const utcDate = new Date(`${utcDateString}T00:00:00Z`);

  // Convert the UTC date to the user's local timezone
  const localDate = new Date(
    utcDate.getTime() - utcDate.getTimezoneOffset() * 60000
  );

  // Extract and return the local date string
  return formatDatYYYYMMDD(localDate);
}

/**
 * daysBetween
 * Calculates the total number of days between two dates, inclusive.
 *
 * @param {Date} startDate - The start date.
 * @param {Date} endDate - The end date.
 * @returns {number} - The number of days between the start and end dates.
 */
export function daysBetween(startDate: Date, endDate: Date): number {
  const oneDay = 24 * 60 * 60 * 1000;
  return (
    Math.floor(Math.abs((endDate.getTime() - startDate.getTime()) / oneDay)) + 1
  );
}

/**
 * DataTemplateCallback type
 * A callback type to define the structure of generated data for a given date.
 *
 * @callback DataTemplateCallback
 * @param {string} date - The date string.
 * @returns {any} - The generated data structure for the given date.
 */
export type DataTemplateCallback = (date: string) => any;

/**
 * generateDays
 * Generates an array of data structures for a range of days, based on a template callback.
 *
 * @param {Date} start - The start date.
 * @param {Date} end - The end date.
 * @param {DataTemplateCallback} template - The template callback to generate the data structure.
 * @returns {any[]} - An array of generated data structures.
 */
export function generateDays(
  start: Date,
  end: Date,
  template: DataTemplateCallback
): any[] {
  // Reset the time for the start date to the beginning of the day
  start.setHours(0, 0, 0, 0);
  // Reset the time for the end date to the beginning of the day as well
  end.setHours(0, 0, 0, 0);

  const results: any[] = [];
  const currDate = new Date(end); // start with the end date

  while (currDate >= start) {
    results.unshift(template(formatDatYYYYMMDD(currDate))); // prepend the results to maintain order
    currDate.setDate(currDate.getDate() - 1);
  }

  return results;
}
