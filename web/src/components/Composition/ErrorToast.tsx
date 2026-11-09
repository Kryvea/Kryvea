import Accordion from "../Form/Accordion";
import Flex from "./Flex";

interface ErrorToastProps {
  message: string;
  details: string;
}

export default function ErrorToast({ message, details }: ErrorToastProps) {
  return (
    <Flex className="error-toast" col>
      <div>{message}</div>
      <Accordion title="Details">
        <div className="error-toast-details" tabIndex={0}>
          {details}
        </div>
      </Accordion>
    </Flex>
  );
}
