export function getApiErrorMessage(error, fallback = "Something went wrong") {
  if (!error) return fallback;
  if (typeof error === "string") return error;

  const data = error.response?.data;
  if (data) {
    if (typeof data === "string") return data;
    if (data.message) return data.message;
    if (data.error) return data.error;
    if (Array.isArray(data.errors) && data.errors.length > 0) {
      return data.errors.map((item) => item.message || item).join(", ");
    }
  }

  if (error.code === "ECONNABORTED") return "Request timed out. Please try again.";
  if (error.message === "Network Error") return "Cannot reach the backend server.";
  return error.message || fallback;
}

export function isApiSuccess(data) {
  return data?.status === "ok" || data?.success === true;
}
