type ScoreBarProps = {
  /** Must be a value between 0 - 10 */
  score: number;
};

export default function ScoreBar({ score }: ScoreBarProps) {
  if (score < 0) {
    score = 0;
  }
  if (score > 10) {
    score = 10;
  }

  const getScoreColor = (score: number) => {
    if (score < 4) return "bg-yellow-400";
    if (score < 7) return "bg-amber-500";
    if (score < 9) return "bg-red-500";
    return "bg-purple-500";
  };

  const getSeverity = (score: number) => {
    if (score == 0) return "Info";
    if (score < 4) return "Low";
    if (score < 7) return "Medium";
    if (score < 9) return "High";
    return "Critical";
  };

  return (
    <div className="mb-2 w-full">
      <div className="mb-2 flex justify-between text-sm font-medium">
        <span>
          Overall Score: <b>{score}</b>
        </span>
        <span>
          Severity: <b>{getSeverity(score)}</b>
        </span>
      </div>
      <div className="h-3 w-full rounded-full bg-neutral-500/25">
        <div
          className={`${getScoreColor(score)} h-3 rounded-full transition-all duration-500 ease-out`}
          style={{
            width: `${score * 10}%`,
          }}
        />
      </div>
    </div>
  );
}
