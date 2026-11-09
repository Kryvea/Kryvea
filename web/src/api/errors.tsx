import { Id, toast, ToastOptions } from "react-toastify";
import ErrorToast from "../components/Composition/ErrorToast";

export type HttpErrorData = {
  error?: string;
  details?: string;
  errors?: string[];
};

const UNKNOWN_HTTP_ERROR_CAUSE = "Unknown error";

export function showApiError(data?: HttpErrorData, existingToastId?: Id) {
  const errors: HttpErrorData[] = data?.errors?.length ? data.errors.map(error => ({ error })) : [data ?? {}];

  errors.forEach(({ error, details }, index) => {
    const message = error || UNKNOWN_HTTP_ERROR_CAUSE;
    const content = details ? <ErrorToast message={message} details={details} /> : message;
    const options: ToastOptions = {
      type: "error",
      isLoading: false,
      autoClose: 3000,
      closeOnClick: !details,
      closeButton: true,
      draggable: !details,
    };

    if (index === 0 && existingToastId != null && toast.isActive(existingToastId)) {
      toast.update(existingToastId, { ...options, render: content });
    } else {
      toast.error(content, options);
    }
  });
}
