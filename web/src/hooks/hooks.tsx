import { RefObject, useEffect, useState } from "react";

export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const handler = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(handler);
    };
  }, [value]);

  return debouncedValue;
}

export function scrollElementHorizontally(reactRef: RefObject<HTMLElement>) {
  return () => {
    if (!reactRef.current) {
      return;
    }

    // When user scroll wheel it converts to horizontal scroll
    const handleWheel = (event: WheelEvent) => {
      event.preventDefault();
      reactRef.current.scrollLeft += event.deltaY;
    };

    reactRef.current.addEventListener("wheel", handleWheel, { passive: false });

    return () => {
      if (!reactRef.current) {
        return;
      }

      reactRef.current.removeEventListener("wheel", handleWheel);
    };
  };
}
