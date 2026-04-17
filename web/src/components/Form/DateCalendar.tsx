import { useEffect, useState } from "react";
import DatePicker from "react-datepicker";
import "react-datepicker/dist/react-datepicker.css";
import { getUserDateFormatPattern } from "../../utils/dates";
import Grid from "../Composition/Grid";
import Label from "./Label";

interface DateCalendarProps {
  idDate: string;
  label: string;
  isRange?: boolean;
  showTime?: boolean;
  value: { start: string; end?: string };
  onChange: (value: string | { start: string; end: string }) => void;
  placeholder?: string | { start?: string; end?: string };
}

export default function DateCalendar({
  idDate,
  label,
  isRange = false,
  showTime = false,
  value,
  onChange,
  placeholder,
}: DateCalendarProps) {
  const parseDate = (dateString?: string | null): Date | null => {
    if (!dateString || dateString === "null") return null;
    const date = new Date(dateString);
    return isNaN(date.getTime()) ? null : date;
  };

  // Set time to 12:00 (noon) to avoid timezone confusion when time is not shown
  const normalizeDate = (date: Date): Date => {
    if (showTime) return date;
    const normalized = new Date(date);
    normalized.setHours(12, 0, 0, 0);
    return normalized;
  };

  const [range, setRange] = useState<[Date | null, Date | null]>([parseDate(value.start), parseDate(value.end)]);

  useEffect(() => {
    if (isRange) {
      setRange([parseDate(value.start), parseDate(value.end)]);
    }
  }, [value.start, value.end, isRange]);

  const handleChangeSingle = (date: Date | null) => {
    if (date && !isNaN(date.getTime())) {
      onChange(normalizeDate(date).toISOString());
    }
  };

  const handleChangeRange = (dates: [Date | null, Date | null]) => {
    setRange(dates);
    let [start, end] = dates;

    // Swap if user manually entered dates in wrong order
    if (start && end && start > end) {
      [start, end] = [end, start];
    }

    onChange({
      start: start ? normalizeDate(start).toISOString() : "",
      end: end ? normalizeDate(end).toISOString() : "",
    });
  };

  const dateFormat = showTime ? `${getUserDateFormatPattern()} HH:mm` : getUserDateFormatPattern();

  const placeholderText =
    typeof placeholder === "string" ? placeholder : isRange ? "Select range date" : "Select a date";

  return (
    <Grid>
      {label && <Label text={label} htmlFor={isRange ? undefined : idDate} />}
      <div>
        {isRange ? (
          <DatePicker
            selectsRange
            startDate={range[0]}
            endDate={range[1]}
            dateFormat={dateFormat}
            onChange={handleChangeRange}
            className="datepicker-input"
            placeholderText={placeholderText}
            autoComplete="off"
          />
        ) : (
          <DatePicker
            selected={parseDate(value.start)}
            onChange={handleChangeSingle}
            showTimeSelect={showTime}
            timeIntervals={15}
            todayButton={showTime ? "Today" : undefined}
            dateFormat={dateFormat}
            className="datepicker-input"
            placeholderText={placeholderText}
            id={idDate}
            autoComplete="off"
          />
        )}
      </div>
    </Grid>
  );
}
