export default function Divider({ noMargin }: { noMargin?: boolean } = {}) {
  return <hr className="divider" data-no-margin={noMargin} />;
}
