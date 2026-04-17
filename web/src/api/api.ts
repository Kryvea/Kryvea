import axios, { AxiosError, HttpStatusCode } from "axios";
import { NavigateFunction } from "react-router";
import { toast } from "react-toastify";
import { ObjectWithId } from "../types/common.types";

type OnThenCallback<T> = (data: T) => any;
type OnThenCallbackBlob = (data: Blob, contentDisposition: string) => any;
type OnCatchCallback<T> = (error: T) => any;
type OnFinallyCallback = () => any;
type HttpErrorData = {
  error: string;
};

let navigateFn;

// Need to be called asap to populate export the navigate function for the APIs closures
export const setNavigate = (fn: typeof navigateFn) => {
  navigateFn = fn;
};

export const navigate: NavigateFunction = (...rest) => {
  if (navigateFn) {
    navigateFn(...rest);
  }
};

const UNKNOWN_HTTP_ERROR_CAUSE = "Unknown error";

//prettier-ignore
const defaultHandleCatch: OnCatchCallback<AxiosError<HttpErrorData>> = err => toast.error(err.response?.data.error || UNKNOWN_HTTP_ERROR_CAUSE);

const onCatchClosure = (onCatch: OnCatchCallback<AxiosError<HttpErrorData>>) => (err: AxiosError<HttpErrorData>) => {
  switch (err.response?.status) {
    case HttpStatusCode.Unauthorized:
      navigate("/login");
      break;
    case HttpStatusCode.BadGateway:
      toast.error("Server unreachable");
      break;
  }

  return onCatch(err);
};

const onBlobCatchClosure = (onCatch: OnCatchCallback<AxiosError<HttpErrorData>>) => async (err: AxiosError) => {
  try {
    const text = await (err.response.data as Blob).text();
    const json = JSON.parse(text);
    // replace data with parsed JSON so your handler gets it
    (err.response as any).data = json;
  } catch (parseErr) {
    console.error("Failed to parse JSON error response", parseErr);
  }

  switch (err.response?.status) {
    case HttpStatusCode.Unauthorized:
      toast.error("Session expired, please log in again");
      navigate("/login", { replace: false, state: { from: window.location.pathname } });
      break;
  }

  return onCatch(err as AxiosError<HttpErrorData>);
};

export function getData<TResponseData>(
  endpoint: string,
  onThen: OnThenCallback<TResponseData> = undefined,
  onCatch: OnCatchCallback<AxiosError<HttpErrorData>> = defaultHandleCatch,
  onFinally: OnFinallyCallback = undefined
) {
  axios
    .get<TResponseData>(endpoint)
    .then(({ data }) => onThen(data))
    .catch(onCatchClosure(onCatch))
    .finally(onFinally);
}

export function postData<TResponseData>(
  endpoint: string,
  data: any = undefined,
  onThen: OnThenCallback<TResponseData> = undefined,
  onCatch: OnCatchCallback<AxiosError<HttpErrorData>> = defaultHandleCatch,
  onFinally: OnFinallyCallback = undefined
) {
  axios
    .post<TResponseData>(endpoint, data)
    .then(({ data }) => onThen(data))
    .catch(onCatchClosure(onCatch))
    .finally(onFinally);
}

export function patchData<TResponseData>(
  endpoint: string,
  data: any = undefined,
  onThen: OnThenCallback<TResponseData> = undefined,
  onCatch: OnCatchCallback<AxiosError<HttpErrorData>> = defaultHandleCatch,
  onFinally: OnFinallyCallback = undefined
) {
  axios
    .patch<TResponseData>(endpoint, data)
    .then(({ data }) => onThen(data))
    .catch(onCatchClosure(onCatch))
    .finally(onFinally);
}

export function putData<TResponseData>(
  endpoint: string,
  data: any = undefined,
  onThen: OnThenCallback<TResponseData> = undefined,
  onCatch: OnCatchCallback<AxiosError<HttpErrorData>> = defaultHandleCatch,
  onFinally: OnFinallyCallback = undefined
) {
  axios
    .put<TResponseData>(endpoint, data)
    .then(({ data }) => onThen(data))
    .catch(onCatchClosure(onCatch))
    .finally(onFinally);
}

export function deleteData<TResponseData>(
  endpoint: string,
  onThen: OnThenCallback<TResponseData> = undefined,
  onCatch: OnCatchCallback<AxiosError<HttpErrorData>> = defaultHandleCatch,
  onFinally: OnFinallyCallback = undefined
) {
  axios
    .delete<TResponseData>(endpoint)
    .then(({ data }) => onThen(data))
    .catch(onCatchClosure(onCatch))
    .finally(onFinally);
}

export function getBlob(
  endpoint: string,
  onThen: OnThenCallbackBlob = undefined,
  onCatch: OnCatchCallback<AxiosError<HttpErrorData>> = defaultHandleCatch,
  onFinally: OnFinallyCallback = undefined
) {
  axios
    .get(endpoint, { responseType: "blob" })
    .then(({ data, headers: { "Content-Disposition": contentDisposition } }) => {
      if (data instanceof Blob === false) {
        console.error("Expected a Blob response, got data = ", data.constructor.name);
        return;
      }
      onThen(data, contentDisposition);
    })
    .catch(onBlobCatchClosure(onCatch))
    .finally(onFinally);
}

export function postDownloadBlob(
  endpoint: string,
  data: any = undefined,
  onThen: OnThenCallbackBlob = undefined,
  onCatch: OnCatchCallback<AxiosError<HttpErrorData>> = defaultHandleCatch,
  onFinally: OnFinallyCallback = undefined
) {
  axios
    .post(endpoint, data, { responseType: "blob" })
    .then(({ data, headers: { "content-disposition": contentDisposition } }) => {
      if (data instanceof Blob === false) {
        console.error("Expected a Blob response, got data = ", data.constructor.name);
        return;
      }
      onThen(data, contentDisposition);
    })
    .catch(onBlobCatchClosure(onCatch))
    .finally(onFinally);
}

export function autoUpdateArrState(setState) {
  return (data: ObjectWithId) => {
    setState((prev: ObjectWithId[]) => {
      if (!Array.isArray(prev)) {
        console.error("Expected previous state to be an array");
        return prev;
      }

      prev.map(item => {
        if (item.id !== data.id) {
          return item;
        }
        return { ...data };
      });
    });
  };
}
