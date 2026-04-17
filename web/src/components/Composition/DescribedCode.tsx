import Grid from "./Grid";
import Subtitle from "./Subtitle";

type DescribedCodeProps = {
  className?: string;
  subtitle?: string;
  text?: string;
  children?: React.ReactNode;
};

export default function DescribedCode({ className, subtitle, text, children }: DescribedCodeProps) {
  return (
    <Grid className={`DescribedCode ${className}`}>
      <Subtitle className="opacity-50" text={subtitle} />
      {text && <code className="overflow-auto whitespace-pre text-justify">{text}</code>}
      {children}
    </Grid>
  );
}
