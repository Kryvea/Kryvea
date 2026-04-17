const utcDate = new Intl.DateTimeFormat(undefined, {
  timeZone: "UTC",
});

// Format date to the user's locale
export function formatDate(dateString: string) {
  return utcDate.format(new Date(dateString));
}

const utcDateTime = new Intl.DateTimeFormat(undefined, {
  day: "2-digit",
  month: "2-digit",
  year: "numeric",
  hour: "numeric",
  minute: "numeric",
  timeZone: "UTC",
  timeZoneName: "short",
});

const testDate = new Date(Date.UTC(2025, 0, 2));
const formatter = new Intl.DateTimeFormat(undefined, {
  year: "numeric",
  month: "2-digit",
  day: "2-digit",
  timeZone: "UTC",
});

export function getUserDateFormatPattern(): string {
  const formatted = formatter.format(testDate);

  // Match numbers and separators
  const parts = formatted.match(/(\d+|\D+)/g);

  if (!parts) return "MM/dd/yyyy"; // fallback

  return parts
    .map(part => {
      if (/^\d+$/.test(part)) {
        switch (part) {
          case "01":
            return "MM";
          case "02":
            return "dd";
          case "2025":
            return "yyyy";
          default:
            return part;
        }
      }
      return part;
    })
    .join("");
}
