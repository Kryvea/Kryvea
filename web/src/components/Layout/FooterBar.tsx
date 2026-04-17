import { useRef, useState } from "react";
import { Link } from "react-router";

type Props = {
  className?: string;
};

export default function FooterBar({ className }: Props) {
  const heartRef = useRef<SVGSVGElement>(null);
  const [clickCount, setClickCount] = useState(0);
  const [exploded, setExploded] = useState(false);

  const handleMouseOver = () => {
    if (heartRef.current) heartRef.current.classList.add("heart-up");
  };

  const handleMouseOut = () => {
    if (heartRef.current) heartRef.current.classList.remove("heart-up");
  };

  const handleClick = () => {
    if (!exploded) {
      const counter = clickCount + 1;
      setClickCount(counter);

      // Add heartbeat class to trigger animation
      if (heartRef.current) {
        heartRef.current.classList.add("beating");

        // Remove the class after the animation ends so it can retrigger
        heartRef.current.addEventListener(
          "animationend",
          () => {
            if (heartRef.current) {
              heartRef.current.classList.remove("beating");
            }
          },
          { once: true }
        );
      }

      if (counter === 5) {
        setExploded(true);
        if (heartRef.current) {
          heartRef.current.classList.add("explode");
        }
      }
    }
  };

  return (
    <footer className={`${className} my-2 select-none font-light italic`}>
      <div className="flex justify-between">
        <div onMouseOver={handleMouseOver} onMouseOut={handleMouseOut}>
          <b>
            <Link to="https://github.com/Kryvea/Kryvea" rel="noreferrer" target="_blank">
              Kryvea
            </Link>
            &nbsp;made with&nbsp;
            <svg ref={heartRef} onClick={handleClick} viewBox="0 0 24 24" className="heart" fill="currentColor">
              <path
                d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 
                  5.42 4.42 3 7.5 3c1.74 0 3.41 0.81 4.5 2.09C13.09 
                  3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 
                  3.78-3.4 6.86-8.55 11.54L12 21.35z"
              />
            </svg>
            &nbsp;by&nbsp;
            <Link to="https://github.com/Alexius22" rel="noreferrer" target="_blank">
              Alexius
            </Link>
            {", "}
            <Link to="https://github.com/CharminDoge" rel="noreferrer" target="_blank">
              CharminDoge
            </Link>
            {" and "}
            <Link to="https://github.com/JJJJJJack" rel="noreferrer" target="_blank">
              Jack
            </Link>
            &nbsp;
          </b>
        </div>
        <div>
          <b>Version</b>: 1.0.0
        </div>
      </div>
    </footer>
  );
}
