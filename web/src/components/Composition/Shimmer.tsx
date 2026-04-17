interface ShimmerProps {
  width?: string | number;
  height?: string | number;
  borderRadius?: string | number;
  style?: React.CSSProperties;
}

export default function Shimmer({ width = "100%", height = 20, borderRadius = 4, style = {} }: ShimmerProps) {
  const wrapperStyle: React.CSSProperties = {
    width,
    height,
    borderRadius,
    ...style,
  };

  return (
    <div className="shimmer-wrapper" style={wrapperStyle}>
      <div className="shimmer" />
    </div>
  );
}
