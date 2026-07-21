import {
  AppConnectivity,
  ChartArea,
  ChartAreaStepper,
  ChartBarStacked,
  ChartBar,
  ChartColumn,
  ChartLine,
  ChartStacked,
  DataStructured,
  Diagram,
  EventSchedule,
  TableBuilt,
} from "@carbon/react/icons";

export const chartTypes = [
  {
    id: "tabular",
    label: "Pivot Table",
    icon: TableBuilt,
    description: ".",
  },
  {
    id: "column",
    label: "Column",
    icon: ChartColumn,
    description: "",
  },
  {
    id: "stackedColumn",
    label: "Stacked Column",
    icon: ChartStacked,
    description: "",
  },
  {
    id: "bar",
    label: "Bar",
    icon: ChartBar,
    description: "",
  },
  {
    id: "stackedBar",
    label: "Stacked bar",
    icon: ChartBarStacked,
    description: "",
  },
  {
    id: "line",
    label: "Line",
    icon: ChartLine,
    description: "",
  },
  {
    id: "area",
    label: "Area",
    icon: ChartArea,
    description: "",
  },
  {
    id: "stackedArea",
    label: "Stacked area",
    icon: ChartAreaStepper,
    description: "",
  },
  {
    id: "pie",
    label: "Pie",
    icon: Diagram,
    description: "",
  },
];

export const mainDimension = [
  {
    id: "2",
    label: "Data",
    icon: AppConnectivity,
    value: "data",
  },
  {
    id: "3",
    label: "Period",
    icon: EventSchedule,
    value: "period",
  },
  {
    id: "4",
    label: "Organisation Unit",
    icon: DataStructured,
    value: "orgunit",
  },
];

export const getYearRangeDescending = () => {
  const currentYear = new Date().getFullYear();
  const endYear = 2016;

  const years: number[] = [];
  for (let year = currentYear; year >= endYear; year--) {
    years.push(year);
  }

  return years;
};

export const periodType = [
  {
    label: "Weekly",
    value: "Weekly",
  },
  {
    label: "Monthly",
    value: "Monthly",
  },
  {
    label: "Quarterly",
    value: "Quarterly",
  },
];

type YearOption = {
  id: string;
  name: string;
};

type QuarterOption = {
  id: string;
  label: string;
  disabled: boolean;
};
type PeriodOption = {
  id: string;
  name: string;
};

export const getWeeksOfYear = (year: string = new Date().getFullYear().toString()) => {
  const weeks: QuarterOption[] = [];
  const startYear = parseInt(year);
  const now = new Date();
  now.setHours(0, 0, 0, 0);

  const formatDate = (date) => {
    const pad = (num) => String(num).padStart(2, "0");
    const y = date.getFullYear();
    const m = pad(date.getMonth() + 1);
    const d = pad(date.getDate());

    return `${y}-${m}-${d}`;
  };

  const getISOWeekNumber = (date) => {
    const d = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()));
    d.setUTCDate(d.getUTCDate() + 4 - (d.getUTCDay() || 7));
    const yearStart = new Date(Date.UTC(d.getUTCFullYear(), 0, 1));
    return Math.ceil(((d.getTime() - yearStart.getTime()) / 86400000 + 1) / 7);
  };

  const currentDate = new Date(startYear, 0, 1);
  const dayOfWeek = currentDate.getDay();
  const daysToAdjust = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
  currentDate.setDate(currentDate.getDate() - daysToAdjust);
  const yearEndMarker = new Date(startYear, 11, 31);

  while (currentDate.getFullYear() <= startYear) {
    if (currentDate.getFullYear() > startYear && currentDate.getDate() > 7) {
      break;
    }
    const weekStart = new Date(currentDate);
    const weekEnd = new Date(currentDate);
    weekEnd.setDate(weekEnd.getDate() + 6);
    const finalEndDate = weekEnd > yearEndMarker ? yearEndMarker : weekEnd;

    const weekNumber = getISOWeekNumber(weekStart);
    const formattedWeek = String(weekNumber).padStart(2, "0");
    const weekString = `Week ${weekNumber} - ${formatDate(weekStart)} - ${formatDate(finalEndDate)}`;
    let isDisabled = false;

    if (finalEndDate >= now) {
      isDisabled = true;
    }

    if (formatDate(finalEndDate) !== formatDate(yearEndMarker)) {
      weeks.push({
        id: `${year}W${formattedWeek}`,
        label: weekString,
        disabled: isDisabled,
      });
    }

    currentDate.setDate(currentDate.getDate() + 7);
  }

  return weeks;
};

export const getBiWeeksOfYear = (year: string = new Date().getFullYear().toString()) => {
  const biWeeks: PeriodOption[] = [];
  const startYear = parseInt(year);

  const formatDate = (date) => {
    const pad = (num) => String(num).padStart(2, "0");
    const y = date.getFullYear();
    const m = pad(date.getMonth() + 1);
    const d = pad(date.getDate());

    return `${y}-${m}-${d}`;
  };

  const getISOWeekNumber = (date) => {
    const d = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()));
    d.setUTCDate(d.getUTCDate() + 4 - (d.getUTCDay() || 7));
    const yearStart = new Date(Date.UTC(d.getUTCFullYear(), 0, 1));
    return Math.ceil(((d.getTime() - yearStart.getTime()) / 86400000 + 1) / 7);
  };

  const currentDate = new Date(startYear, 0, 1);
  const dayOfWeek = currentDate.getDay();
  const daysToAdjust = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
  currentDate.setDate(currentDate.getDate() - daysToAdjust);
  const december31stString = formatDate(new Date(startYear, 11, 31));

  while (currentDate.getFullYear() <= startYear) {
    if (currentDate.getFullYear() > startYear && currentDate.getDate() > 14) {
      break;
    }
    const biWeekStart = new Date(currentDate);
    const biWeekEnd = new Date(currentDate);
    biWeekEnd.setDate(biWeekEnd.getDate() + 13);
    const yearEnd = new Date(startYear, 11, 31);
    const finalEndDate = biWeekEnd > yearEnd ? yearEnd : biWeekEnd;
    const weekNumber = getISOWeekNumber(biWeekStart);
    const biWeekNumber = Math.ceil(weekNumber / 2);
    const biWeekString = `Bi-Week ${biWeekNumber} - ${formatDate(biWeekStart)} - ${formatDate(finalEndDate)}`;
    if (formatDate(finalEndDate) !== december31stString) {
      biWeeks.push({
        id: `${biWeekNumber}-${startYear}`,
        name: biWeekString,
      });
    }

    currentDate.setDate(currentDate.getDate() + 14);
  }

  return biWeeks;
};

export const getMonthsOfYear = (year: string = new Date().getFullYear().toString()) => {
  const months: QuarterOption[] = [];
  const startYear = parseInt(year);
  const now = new Date();
  const currentYear = now.getFullYear();
  const currentMonthIndex = now.getMonth();

  const getMonthName = (monthIndex) => {
    const date = new Date(startYear, monthIndex, 1);
    return date.toLocaleDateString("en-US", { month: "long" });
  };

  for (let monthIndex = 0; monthIndex < 12; monthIndex++) {
    const monthNumber = monthIndex + 1;
    const formattedMonth = String(monthNumber).padStart(2, "0");
    const monthName = getMonthName(monthIndex);
    const monthString = `${monthName} ${startYear}`;
    let isDisabled = false;

    if (startYear === currentYear) {
      if (monthIndex > currentMonthIndex) {
        isDisabled = true;
      }
    }

    months.push({
      id: `${startYear}${formattedMonth}`,
      label: monthString,
      disabled: isDisabled,
    });
  }

  return months;
};

export const getBiMonthsOfYear = (year: string = new Date().getFullYear().toString()) => {
  const biMonths: PeriodOption[] = [];
  const startYear = parseInt(year);

  const getMonthFullName = (monthIndex) => {
    const date = new Date(startYear, monthIndex, 1);
    return date.toLocaleDateString("en-US", { month: "long" });
  };

  for (let monthIndex = 0; monthIndex < 12; monthIndex += 2) {
    const biMonthlyNumber = monthIndex / 2 + 1;
    const startMonthName = getMonthFullName(monthIndex);
    const endMonthName = getMonthFullName(monthIndex + 1);
    const periodString = `${startMonthName} - ${endMonthName} ${startYear}`;

    biMonths.push({
      id: `${biMonthlyNumber}-${startYear}`,
      name: periodString,
    });
  }

  return biMonths;
};

export const getQuarterlyPeriodsOfYear = (year: string = new Date().getFullYear().toString()) => {
  const quarters: QuarterOption[] = [];
  const startYear = parseInt(year);
  const now = new Date();
  const currentYear = now.getFullYear();
  const currentMonthIndex = now.getMonth();
  const currentQuarterNumber = Math.floor(currentMonthIndex / 3) + 1;

  const getMonthFullName = (monthIndex) => {
    const date = new Date(startYear, monthIndex, 1);
    return date.toLocaleDateString("en-US", { month: "long" });
  };

  for (let monthIndex = 0; monthIndex < 12; monthIndex += 3) {
    const quarterNumber = monthIndex / 3 + 1;
    const startMonthName = getMonthFullName(monthIndex);
    const endMonthIndex = monthIndex + 2;
    const endMonthName = getMonthFullName(endMonthIndex);
    const periodString = `${startMonthName} - ${endMonthName} ${startYear}`;
    let isDisabled = false;

    if (startYear === currentYear) {
      if (quarterNumber > currentQuarterNumber) {
        isDisabled = true;
      }
    }

    quarters.push({
      id: `${startYear}Q${quarterNumber}`,
      label: periodString,
      disabled: isDisabled,
    });
  }

  return quarters;
};

export const getLastTenYears = (startYear: string = new Date().getFullYear().toString()) => {
  const years: YearOption[] = [];
  const endYear = parseInt(startYear) - 9;

  for (let currentYear = parseInt(startYear); currentYear >= endYear; currentYear--) {
    const yearString = String(currentYear);

    years.push({
      id: yearString,
      name: yearString,
    });
  }

  return years;
};

export const getPeriodType = (periodId) => {
  if (!periodId || typeof periodId !== "string") {
    return null;
  } else if (periodId.includes("W")) {
    return "Weekly";
  } else if (periodId.includes("Q")) {
    return "Quarterly";
  }

  return "Monthly";
};

export const getAvailablePeriods = (periodType, year) => {
  let availablePeriodArray: QuarterOption[] = [];
  if (periodType === "Weekly") {
    availablePeriodArray = getWeeksOfYear(year);
  } else if (periodType === "Monthly") {
    availablePeriodArray = getMonthsOfYear(year);
  } else if (periodType === "Quarterly") {
    availablePeriodArray = getQuarterlyPeriodsOfYear(year);
  }

  return availablePeriodArray;
};

export const levelOfCareOptions = [
    { id: "NRH", label: "NRH" },
    { id: "RRH", label: "RRH" },
    { id: "General Hospital", label: "General Hospital" },
    { id: "HCIV", label: "HC IV" },
    { id: "HC III", label: "HC III" },
    { id: "HCII", label: "HCII" },
    { id: "BCDP", label: "BCDP" },
    { id: "RBB", label: "RBB" },
    { id: "Clinic", label: "Clinic" },
    { id: "Drug Shop", label: "Drug Shop" },
    { id: "NBB", label: "NBB" },

];

export const ownershipOptions = [
  { id: "GOV", label: "GOV" },
  { id: "PFP", label: "PFP" },
  { id: "PNFP", label: "PNFP" },
];