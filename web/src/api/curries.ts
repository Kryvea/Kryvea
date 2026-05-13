import { toast } from "react-toastify";

export const curryDownloadReport = toastId => (blob, contentDisposition) => {
  let filename: string | undefined;
  if (contentDisposition) {
    const ext = contentDisposition.match(/filename\*\s*=\s*[^']*'[^']*'([^;]+)/i);
    if (ext) {
      filename = decodeURIComponent(ext[1].trim());
    }
    if (!filename) {
      const legacy = contentDisposition.match(/filename\s*=\s*"([^"]+)"/i);
      if (legacy) {
        filename = legacy[1];
      }
    }
  }
  filename ??= "report_export";

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
