import { getAvailablePeriods } from "../../../data-visualizer/src/pages/Constants.tsx";

/**
 * Wraps getAvailablePeriods to ensure all periods (including future quarters/months)
 * are enabled for filtering issues.
 */
export function fetchAvailablePeriods(periodType: string, year: string | number) {
  const periods = getAvailablePeriods(periodType, year.toString());
  return periods.map((p) => ({ ...p, disabled: false }));
}

/**
 * Utility to convert period selections, labels, or raw period strings to standard DHIS2 codes:
 * - Year + Quarter: YYYYQ# (e.g., 2026Q1)
 * - Year + Month: YYYYMM (e.g., 202606)
 * - Year + Week: YYYYWww (e.g., 2026W02)
 */
export function convertPeriodToCode(
  periodInput: unknown,
  fallbackYear?: number | string
): string {
  if (periodInput === null || periodInput === undefined || periodInput === "") {
    return "";
  }

  let inputStr = "";
  if (typeof periodInput === "object") {
    const obj = periodInput as { id?: string; label?: string; name?: string; value?: string };
    if (obj.id && /^\d{4}(Q[1-4]|(0[1-9]|1[0-2])|W\d{2})$/i.test(obj.id)) {
      return obj.id.toUpperCase();
    }
    inputStr = obj.id || obj.label || obj.name || obj.value || "";
  } else {
    inputStr = String(periodInput).trim();
  }

  if (!inputStr) return "";

  const clean = inputStr.trim();

  // 1. Quarterly code: 2026Q1, 2026-Q1, Q1 2026, Q1-2026, 2026Q01
  const qCodeMatch = clean.match(/^(\d{4})[-_\s]?[qQ]0?([1-4])$/);
  if (qCodeMatch) {
    return `${qCodeMatch[1]}Q${qCodeMatch[2]}`;
  }
  const qRevCodeMatch = clean.match(/^[qQ]0?([1-4])[-_\s]?(\d{4})$/);
  if (qRevCodeMatch) {
    return `${qRevCodeMatch[2]}Q${qRevCodeMatch[1]}`;
  }

  // 2. Weekly code: 2026W02, 2026W2, 2026-W02, W02 2026
  const wCodeMatch = clean.match(/^(\d{4})[-_\s]?[wW](\d{1,2})$/);
  if (wCodeMatch) {
    const year = wCodeMatch[1];
    const week = String(wCodeMatch[2]).padStart(2, "0");
    return `${year}W${week}`;
  }
  const wRevCodeMatch = clean.match(/^[wW](\d{1,2})[-_\s]?(\d{4})$/);
  if (wRevCodeMatch) {
    const year = wRevCodeMatch[2];
    const week = String(wRevCodeMatch[1]).padStart(2, "0");
    return `${year}W${week}`;
  }

  // 3. Monthly code: 202606 (YYYYMM 6 digits)
  const mCodeMatch = clean.match(/^(\d{4})(0[1-9]|1[0-2])$/);
  if (mCodeMatch) {
    return `${mCodeMatch[1]}${mCodeMatch[2]}`;
  }

  // 4. Monthly format: YYYY-MM, YYYY/MM, MM-YYYY, MM/YYYY
  const mDashMatch = clean.match(/^(\d{4})[-/](0[1-9]|1[0-2])$/);
  if (mDashMatch) {
    return `${mDashMatch[1]}${mDashMatch[2]}`;
  }
  const mRevDashMatch = clean.match(/^(0[1-9]|1[0-2])[-/](\d{4})$/);
  if (mRevDashMatch) {
    return `${mRevDashMatch[2]}${mRevDashMatch[1]}`;
  }

  // Extract year if present in label, otherwise use fallbackYear
  const yearMatch = clean.match(/\b(19\d\d|20\d\d)\b/);
  const year = yearMatch ? yearMatch[1] : fallbackYear ? String(fallbackYear) : "";

  // Quarter text matching: e.g. "January - March 2026", "Jan - Mar 2026", "Quarter 1", "Q1"
  const qWordMatch = clean.match(/quarter\s*0?([1-4])/i);
  if (qWordMatch && year) {
    return `${year}Q${qWordMatch[1]}`;
  }

  if (/jan(?:uary)?\s*-\s*mar(?:ch)?/i.test(clean) || /\bquarter\s*1\b/i.test(clean) || /\bq1\b/i.test(clean)) {
    if (year) return `${year}Q1`;
  }
  if (/apr(?:il)?\s*-\s*jun(?:e)?/i.test(clean) || /\bquarter\s*2\b/i.test(clean) || /\bq2\b/i.test(clean)) {
    if (year) return `${year}Q2`;
  }
  if (/jul(?:y)?\s*-\s*sep(?:tember)?/i.test(clean) || /\bquarter\s*3\b/i.test(clean) || /\bq3\b/i.test(clean)) {
    if (year) return `${year}Q3`;
  }
  if (/oct(?:ober)?\s*-\s*dec(?:ember)?/i.test(clean) || /\bquarter\s*4\b/i.test(clean) || /\bq4\b/i.test(clean)) {
    if (year) return `${year}Q4`;
  }

  // Week text matching: e.g. "Week 2 - 2026-01-05 - 2026-01-11", "Week 02 2026", "Week 2"
  const weekTextMatch = clean.match(/\bweek\s*(\d{1,2})\b/i);
  if (weekTextMatch) {
    const weekNum = String(weekTextMatch[1]).padStart(2, "0");
    if (year) return `${year}W${weekNum}`;
  }

  // Month text matching: e.g. "June 2026", "January 2026", "Jun 2026"
  const monthsMap: Record<string, string> = {
    january: "01", jan: "01",
    february: "02", feb: "02",
    march: "03", mar: "03",
    april: "04", apr: "04",
    may: "05",
    june: "06", jun: "06",
    july: "07", jul: "07",
    august: "08", aug: "08",
    september: "09", sep: "09", sept: "09",
    october: "10", oct: "10",
    november: "11", nov: "11",
    december: "12", dec: "12",
  };

  const lowerClean = clean.toLowerCase();
  for (const [mName, mNum] of Object.entries(monthsMap)) {
    if (new RegExp(`\\b${mName}\\b`, "i").test(lowerClean)) {
      if (year) return `${year}${mNum}`;
    }
  }

  return clean;
}

/**
 * Checks if a target filter period matches an issue's reporting period.
 * Performs full string matching (matching exact raw string or exact normalized period code).
 */
export function isPeriodMatching(
  targetPeriodInput: unknown,
  issuePeriodInput: unknown,
  fallbackYear?: number | string
): boolean {
  if (!targetPeriodInput) return true; // No filter selected

  const rawTarget = String(targetPeriodInput).trim();
  const rawIssue = String(issuePeriodInput ?? "").trim();

  if (!rawIssue) return false;

  // 1. Direct raw string match (case-insensitive)
  if (rawTarget.toLowerCase() === rawIssue.toLowerCase()) {
    return true;
  }

  // 2. Normalized code full string match (case-insensitive)
  const targetCode = convertPeriodToCode(targetPeriodInput, fallbackYear);
  const issueCode = convertPeriodToCode(issuePeriodInput, fallbackYear);

  if (targetCode && issueCode && targetCode.toLowerCase() === issueCode.toLowerCase()) {
    return true;
  }

  if (targetCode && rawIssue.toLowerCase() === targetCode.toLowerCase()) {
    return true;
  }

  if (issueCode && rawTarget.toLowerCase() === issueCode.toLowerCase()) {
    return true;
  }

  return false;
}
