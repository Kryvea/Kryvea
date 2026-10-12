import DescribedCode from "../Composition/DescribedCode";

export default function PasswordPolicies() {
  return (
    <DescribedCode
      subtitle={
        <ul>
          <li>
            At least <strong>10</strong> characters long
          </li>
          <li>At least one uppercase letter</li>
          <li>At least one lowercase letter</li>
          <li>At least one digit</li>
          <li>At least one special character (punctuation, symbol, or space)</li>
        </ul>
      }
    />
  );
}
