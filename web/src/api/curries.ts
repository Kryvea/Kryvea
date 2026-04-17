import { toast } from "react-toastify";

export const curryDownloadReport = toastId => (blob, contentDisposition) => {
  let filename = "report_export";
  if (contentDisposition) {
    const match = contentDisposition.match(/filename=?"(.+?)?"$/);
    if (match && match[1]) {
      filename = match[1];
    }
  }

  const url = window.URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  window.URL.revokeObjectURL(url);

  toast.update(toastId, {
    render: "Report generated successfully!",
    type: "success",
    isLoading: false,
    autoClose: 3000,
    closeButton: true,
  });
};
